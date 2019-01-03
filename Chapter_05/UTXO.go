package main

type UTXO struct {
    TxOutIndex  int      `json:"txOutIndex"`
    TxOutHash   string   `json:"txOutHash"`
    ToAddress   string   `json:"toAddress"`
    Value       float32  `json:"value"`
    Data        string   `json:"data"`
}