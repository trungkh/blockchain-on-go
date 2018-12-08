package main

import (
    "encoding/json"
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

var blockchain = Blockchain{}
var p2pServer = P2pServer{}
var p2pClient = P2pClient{}

func main() {
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    if err := godotenv.Load(); err != nil {
        log.Fatal(err)
    }

    peers := os.Getenv("PEER")
    if peers != "" {
        blockchain.init(false)
        go p2pClient.init(peers, &p2pServer)
    } else {
        blockchain.init(true)
    }
    go p2pServer.init(&p2pClient)
    go runWebService()

    for {
        select {
        case <-interrupt:
            log.Println("Interrupted!")
            if !p2pClient.getConnection() && !p2pServer.getConnection() {
                return
            }
            if p2pClient.getConnection() {
                close(p2pClient.messages)
            }
            if p2pServer.getConnection() {
                close(p2pServer.messages)
            }
        case <-p2pClient.done:
            if !p2pServer.getConnection() {
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
    router.HandleFunc("/createBlock", handleCreateBlock).Methods("GET")
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
    log.Println("I'm alive!")
    w.Write([]byte("I'm alive!\n"))
}

func handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
    var tx Transaction
    if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if (Transaction{}) == tx {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("Transaction is NULL\n"))
        return
    }

    blockchain.createTransaction(tx)
    bytes, err := json.Marshal(blockchain.getPendingTransactions())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Current pending transactions: " + string(bytes) + "\n"))
}

func handleCreateBlock(w http.ResponseWriter, r *http.Request) {
    if len(blockchain.getPendingTransactions()) > 0 {
        block := new(Block)
        block.init(blockchain.getPendingTransactions(),
                    blockchain.getBlock(blockchain.getBlockHeight()).HashedStr)
        blockchain.addBlock(*block)

        bytes, err := json.MarshalIndent(block, "", "  ")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        log.Println("Created Block: ", string(bytes))

        if p2pServer.getConnection() {
            p2pServer.messages <- bytes
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
    balance := strconv.Itoa(blockchain.getAddressBalance(params["address"]))
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

func parseMessage(message []byte, connected bool, messages chan []byte) {
    var result map[string]interface{}
    json.Unmarshal(message, &result)
    
    if result["previousHash"] == nil {
        return
    }

    if currentHeight := blockchain.getBlockHeight(); currentHeight >= 0 &&
        blockchain.getBlock(currentHeight).HashedStr != result["previousHash"] {
        return
    }

    log.Println("Syncing Block: ", string(message))

    var parsedBlock Block
    json.Unmarshal(message, &parsedBlock)
    blockchain.addBlock(parsedBlock)
    if connected {
        messages <- message
    }
}