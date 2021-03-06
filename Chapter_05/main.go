package main

import (
    "encoding/hex"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "time"

    //"github.com/davecgh/go-spew/spew"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
)

var p2pServer = P2pServer{}
var p2pClient = P2pClient{}
var blockchain = Blockchain{}

func main() {
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    if err := godotenv.Load(); err != nil {
        log.Fatal(err)
    }

    genKeys := flag.Bool("k", false, "generate new key pair")
    flag.Parse()
    if *genKeys {
        wallet := new(Wallet)
        wallet.init()
        log.Printf("Private Key: %x\n", wallet.getPrivateKey())
        log.Printf(" Public Key: %x\n", wallet.getPublicKey())
        //wallet.testSigning()
        return
    }

    peers := os.Getenv("PEER")
    if peers != "" {
        blockchain.init(false)
        go p2pClient.init(peers, &p2pServer)
    } else {
        blockchain.init(true)
    }
    blockchain.setDifficulty(3)
    blockchain.setMiningReward(12.5)

    go p2pServer.init(&p2pClient)
    go runWebService()

    for {
        select {
        case <-interrupt:
            log.Println("Interrupted!")
            if !p2pClient.getConnection() && p2pServer.numberOfClient() == 0 {
                return
            }
            if p2pClient.getConnection() {
                close(p2pClient.messages)
            }
            if p2pServer.numberOfClient() > 0 {
                close(p2pServer.broadcast)
            }
        case <-p2pClient.done:
            if p2pServer.numberOfClient() == 0 {
                return
            }
        case <-p2pServer.done:
            if !p2pClient.getConnection() {
                return
            }
        }
    }
}

// web service
func runWebService() {
    router := mux.NewRouter()
    router.HandleFunc("/ping", handlePing).Methods("GET")
    router.HandleFunc("/createTransaction", handleCreateTransaction).Methods("POST")
    router.HandleFunc("/mineBlock", handleMineBlock).Methods("POST")
    router.HandleFunc("/getUtxo/{address}", handleGetUtxo).Methods("GET")
    router.HandleFunc("/getBlockchain", handleGetBlockchain).Methods("GET")
    router.HandleFunc("/getBalance/{address}", handleGetBalance).Methods("GET")

    httpPort := os.Getenv("HTTP_BASE_PORT")
    peerNo := os.Getenv("PEER_NO")
    if peerNo != "" {
        h, err := strconv.Atoi(httpPort)
        if err != nil {
            log.Fatal(err)
        }
        p, err := strconv.Atoi(peerNo)
        if err != nil {
            log.Fatal(err)
        }
        httpPort = strconv.Itoa(h + p)
    }
    s := &http.Server{
        Addr:           ":" + httpPort,
        Handler:        router,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    log.Println("Listening http on port: ", httpPort)
    if err := s.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}

func handlePing(w http.ResponseWriter, r *http.Request) {
    bytes := []byte("I'm alive!\n")
    log.Print(string(bytes))
    w.Write(bytes)
}

func handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
    var result map[string]interface{}

    if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // check balance
    if blockchain.getAddressBalance(result["fromAddress"].(string)) - float32(result["value"].(float64)) < 0 {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("Not enough funds!\n"))
        return
    }

    // Get the UTXO and decide how many utxo to sign. this is the most tricky part. What problems can you see?
    utxos := blockchain.getUtxo(result["fromAddress"].(string))
    var accumulatedUTXOVal float32 = 0
    var requiredUTXOs []UTXO
    for _, utxo := range utxos {
        requiredUTXOs = append(requiredUTXOs, utxo)
        accumulatedUTXOVal += utxo.Value
        if accumulatedUTXOVal - float32(result["value"].(float64)) > 0 {
            break
        }
    }

    wallet := new(Wallet)
    var txIns []TxIn
    var requiredUTXOValue float32 = 0
    privByte, _ := hex.DecodeString(result["privKey"].(string))
    pubByte, _ := hex.DecodeString(result["fromAddress"].(string))
    // lets spend the the UTXO by signing them. What is wrong with this?
    for _, requireUTXO := range requiredUTXOs {
        sign := wallet.sign([]byte(requireUTXO.TxOutHash), privByte)
        if wallet.verifySignature([]byte(requireUTXO.TxOutHash), sign, pubByte) {
            txIn := TxIn {
                TxOutIndex: requireUTXO.TxOutIndex,
                TxOutHash: requireUTXO.TxOutHash,
                Signature: hex.EncodeToString(sign),
            }
            requiredUTXOValue += requireUTXO.Value
            txIns = append(txIns, txIn)
        } else {
            w.WriteHeader(http.StatusBadRequest)
            w.Write([]byte("You are not the owner of the funds!\n"))
            return
        }
    }

    // Check if change is required. In the real world, the change is normally send to a new address.
    txOuts := append([]TxOut{}, TxOut {
        ToAddress: result["toAddress"].(string),
        Value: float32(result["value"].(float64)),
        Data: result["data"].(string),
    })

    changeValue := requiredUTXOValue - float32(result["value"].(float64))
    if changeValue > 0 {
        txOuts = append(txOuts, TxOut {
            ToAddress: result["fromAddress"].(string),
            Value: changeValue,
            Data: "change",
        })
    }

    tx := new(Transaction)
    tx.init(txIns, txOuts)
    blockchain.createTransaction(*tx)
    bytes, err := json.MarshalIndent(blockchain.getPendingTransactions(), "", "  ")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Current pending transactions: " + string(bytes) + "\n"))
}

func handleMineBlock(w http.ResponseWriter, r *http.Request) {
    if len(blockchain.getPendingTransactions()) > 0 {
        var result map[string]interface{}

        if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        if result["minerAddress"] == nil {
            w.WriteHeader(http.StatusBadRequest)
            w.Write([]byte("Miner's address is NULL\n"))
            return
        }
        
        block := blockchain.mineBlock(result["minerAddress"].(string))

        bytes, err := json.MarshalIndent(block, "", "  ")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        log.Println("Created Block: ", string(bytes))

        if p2pServer.numberOfClient() > 0 {
            p2pServer.broadcast <- bytes
        }
        if p2pClient.getConnection() {
            p2pClient.messages <- bytes
        }

        w.WriteHeader(http.StatusCreated)
        w.Write([]byte("New block created: " + string(bytes) + "\n"))
    } else {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("Cannot create empty block\n"))
    }
}

func handleGetUtxo(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    balance := fmt.Sprint(blockchain.getUtxo(params["address"]))
    w.Write([]byte(params["address"] + " balance is: " + balance + "\n"))
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
    //json.NewEncoder(w).Encode(blockchain.getBlockchain())
    bytes, err := json.MarshalIndent(blockchain.getBlockchain(), "", "  ")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Write(bytes)
}

func handleGetBalance(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    balance := fmt.Sprint(blockchain.getAddressBalance(params["address"]))
    w.Write([]byte(params["address"] + " balance is: " + balance + "\n"))
}

func nodeSync(messages chan []byte) {
    for _, block := range blockchain.getBlockchain() {
        message, err := json.MarshalIndent(block, "", "  ")
        if err != nil {
            log.Println("parsing: ", err)
            return
        }
        messages <- message
    }
}

func parseMessage(message []byte) bool {
    var result map[string]interface{}
    json.Unmarshal(message, &result)
    
    if result["previousHash"] == nil {
        return false
    }

    if currentHeight := blockchain.getBlockHeight(); currentHeight >= 0 {
        if blockchain.getBlock(currentHeight).HashedStr != result["previousHash"] {
            return false
        }
    }

    // Consensus action to verify result from the source node 
    var block Block
    json.Unmarshal(message, &block)
    if block.HashedStr != block.calculateHash() {
        return false
    }

    log.Println("Syncing Block: ", string(message))

    blockchain.addBlock(block)
    return true
}