package main

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
)

type Transaction struct {
    FromAddress string   `json:"fromAddress"`
    ToAddress   string   `json:"toAddress"`
    Value       float32  `json:"value"`
    Data        string   `json:"data"`
    Timestamp   string   `json:"timestamp"`
    HashedStr   string   `json:"hashedStr"`
}

func (this *Transaction) init(fromAddress string, toAddress string, value float32, data string) {
    this.FromAddress = fromAddress
    this.ToAddress = toAddress
    this.Value = value
    this.Data = data
    this.Timestamp = fmt.Sprint(time.Now().UnixNano())
    // How to ensure true randomness?
    this.HashedStr = this.calculateHash()
}

func (this Transaction) calculateHash() string {
    recordStr := this.FromAddress + this.ToAddress + this.Timestamp + fmt.Sprint(this.Value)
    hash := sha256.New()
    hash.Write([]byte(recordStr))
    hashed := hash.Sum(nil)
    return hex.EncodeToString(hashed)
}