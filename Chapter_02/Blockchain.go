package main

import (
    "strings"
    "sync"
)

var mutex = &sync.Mutex{}

type Blockchain struct {
    chain               []Block
    pendingTransactions []Transaction
    difficulty          int
    miningReward        float32
}

func (this *Blockchain) init(genesis bool) {
    if genesis {
        mutex.Lock()
        this.chain = append(this.chain, *this.createGenesisBlock())
        mutex.Unlock()
    }
    this.pendingTransactions = nil
}

func (this Blockchain) createGenesisBlock() *Block {
    block := new(Block)
    block.init(nil, "0")
    return block
}

func (this *Blockchain) setDifficulty(difficulty int) {
    this.difficulty = difficulty
}

func (this *Blockchain) setMiningReward(reward float32) {
    this.miningReward = reward
}

func (this *Blockchain) mineBlock(minerAddress string) Block {
    // coinbase transaction
    tx := Transaction{
        ToAddress: minerAddress,
        Value: this.miningReward,
    }
    this.pendingTransactions = append([]Transaction{tx}, this.pendingTransactions...)

    // create new block based on hardcoded timestamp, all pending tx and previous blockhash
    block := new(Block)
    block.init(this.getPendingTransactions(),
                this.getBlock(this.getBlockHeight()).HashedStr)
    block.mineBlock(this.difficulty)

    this.addBlock(*block)
    // reset pending Transactions
    this.pendingTransactions = nil
    return *block
}

// this method is now not available for public use
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

func (this Blockchain) getAddressBalance(address string) float32 {

    var balance float32 = 0;

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