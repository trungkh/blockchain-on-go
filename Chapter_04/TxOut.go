package main

type TxOut struct {
    ToAddress   string   `json:"toAddress"`
    Value       float32  `json:"value"`
    Data        string   `json:"data"`
}