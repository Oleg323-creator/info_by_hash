package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	txHash := "0x783e170e1dda8c8a7b2a88026b5a68686f80ffccff6f5d029ef9f70166a4c27c"

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
	log.Printf("Input Data: %s\n", tx.Input)
}

func hexToDecimal(hexValue string) *big.Int {
	decimalValue := new(big.Int)
	decimalValue.SetString(strings.TrimPrefix(hexValue, "0x"), 16)
	return decimalValue
}
