package main

type TxIn struct {
    TxOutIndex  int      `json:"txOutIndex"`
    TxOutHash   string   `json:"txOutHash"`
    Signature   string   `json:"signature"`
}