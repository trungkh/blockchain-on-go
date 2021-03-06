package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "time"
)

type Block struct {
    PreviousHash string        `json:"previousHash"`
    Timestamp    string        `json:"timestamp"`
    Transactions []Transaction `json:"transactions"`
    HashedStr    string        `json:"hashedStr"`
}

func (this *Block) init(transactions []Transaction, previousHash string) {
    this.PreviousHash = previousHash
    this.Timestamp = fmt.Sprint(time.Now().UnixNano())
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
    hashed := hash.Sum(nil)
    return hex.EncodeToString(hashed)
}
