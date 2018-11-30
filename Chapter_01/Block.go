package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type Block struct {
	PreviousHash string        `json:"previousHash"`
	Timestamp    string        `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	HashedStr    string        `json:"hashedStr"`
}

func (this *Block) Init(timestamp string, transactions []Transaction, previousHash string) {
	this.PreviousHash = previousHash
	this.Timestamp = timestamp
	this.Transactions = transactions
	this.HashedStr = this.calculateHash()
}

func (this Block) calculateHash() string {
	bytes, err := json.Marshal(this.Transactions);
	if err != nil {
		panic(err)
	}
	recordStr := this.PreviousHash + this.Timestamp + string(bytes)
	hash := sha256.New()
	hash.Write([]byte(recordStr))
	hashedStr := hash.Sum(nil)
	return hex.EncodeToString(hashedStr)
}
