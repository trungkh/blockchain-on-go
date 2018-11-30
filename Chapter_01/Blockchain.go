package main

import (
	"strings"
	"sync"
)

var mutex = &sync.Mutex{}

type Blockchain struct {
    chain               []Block
    pendingTransactions []Transaction
}

func (this *Blockchain) Init() {
	mutex.Lock()
	this.chain = append(this.chain, *this.createGenesisBlock())
	mutex.Unlock()
	this.pendingTransactions = nil
}

func (this Blockchain) createGenesisBlock() *Block {
	block := new(Block)
	block.Init("this_is_genesis_address", nil, "0")
	return block
}

func (this *Blockchain) addBlock(block Block) {
	mutex.Lock()
	this.chain = append(this.chain, block)
	mutex.Unlock()
	// reset pending Transactions
	this.pendingTransactions = nil
}

func (this Blockchain) getBlock(id int) Block {
	return this.chain[id]
}

func (this Blockchain) getBlockHeight() int {
	return len(this.chain) - 1
}

func (this Blockchain) getBlockchain() []Block {
	return this.chain
}

func (this *Blockchain) createTransaction(transaction Transaction) {
	this.pendingTransactions = append(this.pendingTransactions, transaction)
}

func (this Blockchain) getPendingTransactions() []Transaction {
	return this.pendingTransactions
}

func (this Blockchain) getAddressBalance(address string) int {

	balance := 0;

	for _, block := range this.chain {
		for _, transaction := range  block.Transactions {
			if strings.Compare(transaction.FromAddress, address) == 0 {
				balance -= transaction.Value;
			}
			if strings.Compare(transaction.ToAddress, address) == 0 {
				balance += transaction.Value;
			}
		}
	}
	return balance;
}