package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "strings"
    "time"
)

type Block struct {
    PreviousHash string        `json:"previousHash"`
    Timestamp    string        `json:"timestamp"`
    Transactions []Transaction `json:"transactions"`
    HashedStr    string        `json:"hashedStr"`
    Nonce        int64         `json:"nonce"`
}

func (this *Block) init(transactions []Transaction, previousHash string) {
    this.PreviousHash = previousHash
    this.Timestamp = fmt.Sprint(time.Now().UnixNano())
    this.Transactions = transactions
    this.HashedStr = this.calculateHash()
    this.Nonce = 0
}

func (this Block) calculateHash() string {
    bytes, err := json.Marshal(this.Transactions);
    if err != nil {
        panic(err)
    }
    recordStr := this.PreviousHash + this.Timestamp + string(bytes) + fmt.Sprint(this.Nonce)
    hash := sha256.New()
    hash.Write([]byte(recordStr))
    hashed := hash.Sum(nil)
    return hex.EncodeToString(hashed)
}

// these few lines are the revolution
// the difficulty is the number of 0s that the hash must start with.
// The nonce creates the randomness for new hashes
func (this *Block) mineBlock(difficulty int) {    
    for !strings.HasPrefix(this.HashedStr, strings.Repeat("0", difficulty)) {
        this.Nonce++
        this.HashedStr = this.calculateHash()
    }
}
