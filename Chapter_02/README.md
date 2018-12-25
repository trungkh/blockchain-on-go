# Chapter 2: Consensus (Proof of Work) and Incentivisation

> For our timestamp network, we implement the proof-of-work by incrementing a nonce in the block until a value is found that gives the block's hash the required zero bits. - Bitcoin's Whitepaper

In a decentralised system, Consensus is achieved if majority of the nodes agree on a common truth. What if some nodes are dishonest? In networking, this is the typical [Byzantine General's Problem](https://en.wikipedia.org/wiki/Byzantine_fault_tolerance). If all the generals are attacking the city, how could they coordinate the attack so that they act at the same time? The challenge is not only ensuring the message get passed through, but also to mitigate traitors.

There are many [Consensus Algorithm](https://www.coindesk.com/short-guide-blockchain-consensus-protocols/) in the Blockchain ecosystem. In Bitcoin, [Proof-of-Work](https://en.bitcoin.it/wiki/Proof_of_work) is used as a probabilistic solution to the Byzantine's problem. 

## Proof of Work

Proof of Work is about using computing power to find a cryptographic solution (a hash) for the right to generate the next block in the Blockchain, ie you proof that the block you created is legit. **The hash should be hard to find but easy to verify**.

In Bitcoin, the right hash is one with a specific number of zeros prefixing it. The difficulty property defines how many prefixing zeros the hash must have.

Q1. Write a mineBlock function that accepts a difficulty parameter in the Block struct. this.hash must also be guaranteed unique.

[Block.go](Block.go)

## Incentivisation

> By convention, the first transaction in a block is a special transaction that starts a new coin owned by the creator of the block. This adds an incentive for nodes to support the network, and provides a way to initially distribute coins into circulation, since there is no central authority to issue them. - Bitcoin's Whitepaper

One very clever thing that Satoshi did was to reward nodes for becoming a block producer. This move attracts global participation, achieving successful decentralisation.

Anyone could become a miner if they are willing to use their computational power. They will be rewarded with some currency (block reward, transaction fees) for finding the next block. In the case of Bitcoin, its 12.5 BTC at the time of writing. 

It is also important to note that due to the limited supply and price of Bitcoin, there is no incentive to cheat. Attacking the network successfully would require more than 51% of malicious nodes which would be very costly and achieves no monetary incentives.

Q2. In Blockchain.js, add a mineBlock function that accepts a minerAddress argument. Add a new transaction before all other pending transactions. This is the coinbase transaction to reward the miner only.

[Blockchain.go](Transaction.go)

Next, let us set the mining difficulty, reward and change the api endpoint in main.js from createBlock to mineBlock,

[main.go](main.go)

In mineBlock api, we make sure we provide the miner Address. Once a block is produced, we broadcast the data. All nodes receiving the data must verify the data before accepting it into the chain. Once this protocol is established, we will have a consistent decentralised ledger.

## 51% Attack

If a malicious miner controls a large proportion of hashing power, he will always have the higher possibility of finding the next block, meaning he can chose not to broadcast his new block immediately but instead at about the same time a new block is found in the main chain (he needs to be very precise about his announcement and lucky to be always ahead of the main chain). Next, he spends his funds in the main chain but don’t record it in his chain. After a number of confirmations, there are now 2 competing chains, one from the malicious miner and one from the main chain. Once the malicious miner is happy with all his spendings in the main chain and after X confirmations, he confidently mines the next block and broadcast it immediately with a different history (without his spending transactions that were recorded in the main chain before). Because the malicious miner has the longest chain, the Bitcoin protocol ensures that all the other miners will use it as the best chain and propagate it across the network. The malicious miner has essentially reversed the transaction, changed the history in the latest main chain and achieves double spending.

https://people.xiph.org/~greg/attack_success.html gives us an idea of the possibility of double spending by a malicious miner based on the hash power they have. For example,

* A miner with hash power of 25% of the global network has 52% of double spending possibility on 1 confirmation. They will have 20% of double spending on 3 confirmations and 5% of double spending possibility on 6 confirmations.
* At 50% of the hashrate, the miner has 100% of double spending possibility forever (many people call this the 51% attack).

If a miner holds more than 50% of the total hashing power, they don’t really have to care about the other miners as they will win out in the long run and take over the main chain. They could double spend any time they want. However once that happens, other miners can choose not to connect to the malicious miner, causing 2 soft forks in the chain.

![51% attack and double spending](../resources/51percent.png)

## Spy Mining

Look at [this block](https://etherscan.io/block/6693829). Why are there no transactions in this block? Can you find other empty blocks mined by the same pool?

Empty blocks are often included in the blockchain by large mining pools. Antpool, F2Pool, and BWPool always do that. By mining empty blocks, the miner do not get any transaction fees. Why would they do that?

When a pool finds a new block, it will propagate the block header to its miners before broadcasting the other nodes in the network. This means that other nodes are slightly slower in mining the next block, ie the pool which mines the block gets a headstart to mine the next block. Doing so allows the pool to gain a headstart over other miners at hashing the next block. 

It is possible for a pool to have a spy miner in its competitor's pool. When the spy miner receives a block header from the pool, it will pass on this information to its own pool. As the pool doesn't know the full transactions in the block, it will mine empty blocks, claiming the new block with just 1 coinbase transaction.

Spy miners can have [detrimental effects](https://cryptocrimson.com/sharkpool-threatens-to-attack-all-bch-forks-except-for-craig-wrights-satoshis-vision/) to the network - Terrorism in the Blockchain?

## Testing

In Terminal 1, start the node

```
$ make
$ ./blockchain
```

In Terminal 2, start the node

```
$ PEER_NO=1 PEER=ws://localhost:6000 ./blockchain
```

In Terminal 3, start the node

```
$ PEER_NO=2 PEER=ws://localhost:6001 ./blockchain
```

In Terminal 4, 

```
# Lets add a tx to node 3. Is the mining successful? Why?
$ curl -H "Content-type:application/json" --data '{"fromAddress" : "alice", "toAddress" : "bob", "value" : 40}' http://localhost:3002/createTransaction

$ curl -H "Content-type:application/json" --data '{"minerAddress":"miner"}' http://localhost:3002/mineBlock

# check the chain in all the nodes. They should be the same
$ curl http://localhost:3002/getBlockchain
$ curl http://localhost:3001/getBlockchain
$ curl http://localhost:3000/getBlockchain

# get the balance of alice, bob and miner from node 1.
$ curl http://localhost:3000/getBalance/alice
$ curl http://localhost:3000/getBalance/bob
$ curl http://localhost:3000/getBalance/miner
```

## Short Quiz

? In Proof-of-Work, which of the following is false?

a) Any node can create new blocks
B) Any new node must be approved by the network
c) It is possible to find a block with the same hash

? A spy Miner has better chance of finding a new block with lesser hashing power.

a) True
B) False

? The biggest damage that a spy miner can do is to delay the transaction speed.

A) True
b) False

? There should be almost no spy miners for Bitcoin around the year 2140.

A) True
b) False