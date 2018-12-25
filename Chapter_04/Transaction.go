package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "time"
)

type Transaction struct {
    TxIn        []TxIn   `json:"txIn"`
    TxOut       []TxOut  `json:"txOut"`
    Timestamp   string   `json:"timestamp"`
    HashedStr   string   `json:"hashedStr"`
}

func (this *Transaction) init(txIn []TxIn, txOut []TxOut) {
    this.TxIn = txIn
    this.TxOut = txOut
    this.Timestamp = fmt.Sprint(time.Now().UnixNano())
    this.HashedStr = this.calculateHash()
}

func (this Transaction) calculateHash() string {
    txInBytes, err := json.Marshal(this.TxIn);
    if err != nil {
        panic(err)
    }
    txOutBytes, err := json.Marshal(this.TxOut);
    if err != nil {
        panic(err)
    }

    recordStr := string(txInBytes) + string(txOutBytes) + this.Timestamp
    hash := sha256.New()
    hash.Write([]byte(recordStr))
    hashed := hash.Sum(nil)
    return hex.EncodeToString(hashed)
}