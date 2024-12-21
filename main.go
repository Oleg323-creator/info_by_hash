package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"

	_ "github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
)

type TransactionData struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Nonce            string `json:"nonce"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Value            string `json:"value"`
	Input            string `json:"input"`
}

type TransactionReceipt struct {
	Result struct {
		GasUsed string `json:"gasUsed"`
	} `json:"result"`
}

var erc20ABI = `[{"constant":false,"inputs":[{"name":"recipient","type":"address"},
	{"name":"amount","type":"uint256"}],"name":"transfer","outputs":
	[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	txHash := "0x783e170e1dda8c8a7b2a88026b5a68686f80ffccff6f5d029ef9f70166a4c27c"

	if err != nil {
		log.Fatalf("Error connecting to the Ethereum client: %v", err)
	}

	url := fmt.Sprintf("https://api-sepolia.etherscan.io/api?module=proxy&action="+
		"eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, os.Getenv("API_KEY"))

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching transaction info: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	var response struct {
		Result TransactionData `json:"result"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error unmarshalling response: %v", err)
	}

	tx := response.Result
	log.Printf("Transaction Details:\n")
	log.Printf("Hash: %s\n", tx.Hash)
	log.Printf("From: %s\n", tx.From)
	log.Printf("To: %s\n", tx.To)
	log.Printf("Value (in Wei): %s\n", hexToDecimal(tx.Value))
	log.Printf("Block Number: %s\n", hexToDecimal(tx.BlockNumber))
	log.Printf("Gas: %s\n", hexToDecimal(tx.Gas))

	gasPriceWei := hexToDecimal(tx.GasPrice)
	gasPriceGwei := new(big.Float).Quo(new(big.Float).SetInt(gasPriceWei), big.NewFloat(1e9)) //вот это порнуха конечно
	log.Printf("Gas Price (in Gwei): %s\n", gasPriceGwei.Text('f', 9))

	log.Printf("Nonce: %s\n", hexToDecimal(tx.Nonce))
	log.Printf("Transaction Index: %s\n", hexToDecimal(tx.TransactionIndex))

	// GETTING TX  FEES
	receiptUrl := fmt.Sprintf("https://api-sepolia.etherscan.io/api?module=proxy&action="+
		"eth_getTransactionReceipt&txhash=%s&apikey=%s", txHash, os.Getenv("API_KEY"))

	receiptResp, err := http.Get(receiptUrl)
	if err != nil {
		log.Fatalf("Error fetching transaction receipt: %v", err)
	}
	defer receiptResp.Body.Close()

	receiptBody, err := ioutil.ReadAll(receiptResp.Body)
	if err != nil {
		log.Fatalf("Error reading receipt response body: %v", err)
	}

	var receiptResponse TransactionReceipt
	err = json.Unmarshal(receiptBody, &receiptResponse)
	if err != nil {
		log.Fatalf("Error unmarshalling receipt: %v", err)
	}

	gasUsed := hexToDecimal(receiptResponse.Result.GasUsed)

	feesWei := new(big.Int).Mul(gasUsed, gasPriceWei)

	// CONVERT TO ETH
	commissionETH := new(big.Float).Quo(new(big.Float).SetInt(feesWei), big.NewFloat(1e18))

	log.Printf("Transaction Fee (in ETH): %s\n", commissionETH.Text('f', 18))

	log.Printf("Input Data (Raw): %s\n", tx.Input)

	inputData := tx.Input
	if len(inputData) > 10 { // 10 BECAUSE FIRST 10 SYMBOLS IS FUNC SIGNATURE
		data := inputData[10:]

		addressBytes := common.Hex2Bytes(data[:64]) // RECIPTER ADRESS
		amountBytes := common.Hex2Bytes(data[64:])  // TOKEN AMOUNT

		var recipientAddress common.Address
		copy(recipientAddress[:], addressBytes[12:32])

		amount := new(big.Int)
		amount.SetBytes(amountBytes)

		log.Printf("Decoded Input Data:\n")
		log.Printf("Recipient Address: %s\n", recipientAddress.Hex())
		log.Printf("Amount: %s\n", amount.String())
	}
}

func hexToDecimal(hexStr string) *big.Int {
	value := new(big.Int)
	value.SetString(hexStr[2:], 16)
	return value
}
