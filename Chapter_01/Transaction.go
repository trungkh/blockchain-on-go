package main

type Transaction struct {
    FromAddress string   `json:"fromAddress"`
    ToAddress   string   `json:"toAddress"`
    Value       int      `json:"value"`
    Data        string   `json:"data"`
}