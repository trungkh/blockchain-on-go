package main

import (
    //"strings"
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
    // premine 3 outputs - 30 and 20 and 10 coins to Alice
    var txOuts []TxOut
    txOuts = append(txOuts, TxOut{
        ToAddress: "4b83487732a84f3963bd20f61341a1a69fd9d5db6be47d0f9d92015baf8848b3beb0c447ed24b7e0b5adc310da9b6cc5f482c53bf04508f72dd7cd4818006906",
        Value: 30,
        Data: "pre-mint",
    })
    txOuts = append(txOuts, TxOut{
        ToAddress: "4b83487732a84f3963bd20f61341a1a69fd9d5db6be47d0f9d92015baf8848b3beb0c447ed24b7e0b5adc310da9b6cc5f482c53bf04508f72dd7cd4818006906",
        Value: 20,
        Data: "pre-mint",
    })
    txOuts = append(txOuts, TxOut{
        ToAddress: "4b83487732a84f3963bd20f61341a1a69fd9d5db6be47d0f9d92015baf8848b3beb0c447ed24b7e0b5adc310da9b6cc5f482c53bf04508f72dd7cd4818006906",
        Value: 10,
        Data: "pre-mint",
    })

    genesisTx := new(Transaction)
    genesisTx.init(nil, txOuts)

    block := new(Block)
    block.init([]Transaction{*genesisTx}, "0")
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
    txOuts := TxOut {
        ToAddress: minerAddress,
        Value: this.miningReward,
        Data: "coinbase tx",
    }

    tx := new(Transaction)
    tx.init(nil, []TxOut{txOuts})
    this.pendingTransactions = append([]Transaction{*tx}, this.pendingTransactions...)

    // create new block based on all pending tx and previous blockhash
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

// this is a very important function
// Get all outputs without inputs
func (this Blockchain) getUtxo(fromAddress string) []UTXO {
    var utxos []UTXO
    for _, block := range this.chain {
        for _, tx := range block.Transactions {
            // get all output tx
            for i := 0; i < len(tx.TxOut); i++ {
                if tx.TxOut[i].ToAddress == fromAddress {
                    utxos = append(utxos, UTXO{
                        TxOutIndex: i,
                        TxOutHash: tx.HashedStr,
                        ToAddress: tx.TxOut[i].ToAddress,
                        Value: tx.TxOut[i].Value,
                        Data: tx.TxOut[i].Data,
                    })
                }
            }
        }
    }

    // now filter away those utxos that has been used
    for _, block := range this.chain {
        for _, tx := range block.Transactions {
            for _, txIn := range tx.TxIn {
                // get all output tx
                for i := 0; i < len(utxos); i++ {
                    if txIn.TxOutHash == utxos[i].TxOutHash && txIn.TxOutIndex == utxos[i].TxOutIndex {
                        // remove the item
                        utxos = append(utxos[:i], utxos[i+1:]...)
                    }
                }
            }
        }
    }
    return utxos
}

func (this Blockchain) getAddressBalance(address string) float32 {
    var balance float32 = 0
    utxos := this.getUtxo(address)

    for _, utxo := range utxos {
        balance += utxo.Value
    }
    return balance
}