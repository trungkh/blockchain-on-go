package main

type Transaction struct {
    FromAddress string   `json:"fromAddress"`
    ToAddress   string   `json:"toAddress"`
    Value       float32  `json:"value"`
    Data        string   `json:"data"`
}