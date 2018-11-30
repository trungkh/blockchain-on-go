package main

import (
    "encoding/json"
    "log"
    "net/http"
    "net/url"
    "os"
    "os/signal"
    "strconv"
    "time"

    //"github.com/davecgh/go-spew/spew"
    "github.com/gorilla/websocket"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
)

var blockchain = Blockchain{}
// Create channel to receive messages from all connections
var upMessages = make(chan []byte)
var downMessages = make(chan []byte)
var serverConnected bool = false
var clientConnected bool = false

func serverConnectionSwitch(status bool) {
    serverConnected = status
}

func clientConnectionSwitch(status bool) {
    clientConnected = status
}

func main() {
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)
    done := make(chan struct{})

    if err := godotenv.Load(); err != nil {
        log.Fatal(err)
    }
    defer close(upMessages)
    defer close(downMessages)

    blockchain.Init()

    go runWebService()
    go initP2PServer()

    peers := os.Getenv("PEERS")
    if peers != "" {
        go initP2PClients(peers, interrupt, done)
        for {
            select {
            case <-done:
                return
            }
        }
    } else {
        for {
            select {
            case <-interrupt:
                log.Println("Interrupted!")
                close(done)
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
        block.Init(time.Now().Format("20060102150405"), blockchain.getPendingTransactions(), blockchain.getBlock(blockchain.getBlockHeight()).HashedStr)
        blockchain.addBlock(*block)

        bytes, err := json.MarshalIndent(block, "", "  ")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        log.Println("Created Block: ", string(bytes))

        if serverConnected {
            upMessages <- bytes
        }
        if clientConnected {
            downMessages <- bytes
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

// web socket server
func initP2PServer() {
    http.HandleFunc("/", handleMessage)

    wsPort := os.Getenv("P2P_BASE_PORT")
    peerNo := os.Getenv("PEER_NO")
    if peerNo != "" {
        w, err := strconv.Atoi(wsPort)
        if err != nil {
            log.Fatal(err)
        }
        p, err := strconv.Atoi(peerNo)
        if err != nil {
            log.Fatal(err)
        }
        wsPort = strconv.Itoa(w + p)
    }
    //s := &http.Server{
    //	Addr:           ":" + wsPort
    //}

    log.Println("Listening websocket p2p on port: ", wsPort)
    if err := http.ListenAndServe(/*addr*/":" + wsPort, nil); err != nil {
        log.Fatal(err)
    }
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
    serverConn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
    if err != nil {
        log.Println("upgrade: ", err)
        return
    }
    defer serverConn.Close()
    defer clientConnectionSwitch(false)

    log.Println("Connnection established...")
    clientConnectionSwitch(true)

    go func() {
        for _, block := range blockchain.getBlockchain()[1:] {
            message, err := json.MarshalIndent(block, "", "  ")
            if err != nil {
                log.Println("parsing: ", err)
                return
            }
            err = serverConn.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                log.Println("server write: ", err)
                return
            }
        }

        for {
            select {
            case message := <- downMessages:
                err = serverConn.WriteMessage(websocket.TextMessage, message)
                if err != nil {
                    log.Println("server write: ", err)
                    break
                }
            }
        }
    }()

    for {
        _, message, err := serverConn.ReadMessage()
        if err != nil {
            log.Println("server read: ", err)
            break
        }
        log.Println("Syncing Block: ", string(message))
        
        var result map[string]interface{}
        json.Unmarshal(message, &result)
        
        if result["previousHash"] != nil {
            currentBlock := blockchain.getBlock(blockchain.getBlockHeight())
            if currentBlock.HashedStr == result["previousHash"] {
                var parsedBlock Block
                json.Unmarshal(message, &parsedBlock)
                blockchain.addBlock(parsedBlock)
                if serverConnected {
                    upMessages <- message
                }
            }
        }
    }
}

// web socket client
func initP2PClients(peers string, interrupt chan os.Signal, done chan struct{}) {
    peer, err := url.Parse(peers); 
    if err != nil {
        log.Fatal(err)
    }

    u := url.URL { Scheme: peer.Scheme, Host: peer.Host, Path: "/" }
    clientConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal("Dial error: ", err)
    }
    defer clientConn.Close()
    defer close(done)

    log.Println("Connected p2p on host: ", peer.Host)
    serverConnectionSwitch(true)

    go func() {
        for {
            _, message, err := clientConn.ReadMessage()
            if err != nil {
                log.Println("client read:", err)
                return
            }
            log.Println("Syncing Block: ", string(message))
        
            var result map[string]interface{}
            json.Unmarshal(message, &result)
            
            if result["previousHash"] != nil {
                currentBlock := blockchain.getBlock(blockchain.getBlockHeight())
                if currentBlock.HashedStr == result["previousHash"] {
                    var parsedBlock Block
                    json.Unmarshal(message, &parsedBlock)
                    blockchain.addBlock(parsedBlock)
                    if clientConnected {
                        downMessages <- message
                    }
                }
            }
        }
    }()

    for {
        select {
        case message := <- upMessages:
            err := clientConn.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                log.Println("client write: ", err)
                return
            }
        case <-interrupt:
            log.Println("Interrupted!")
            // Cleanly close the connection by sending a close message and then
            // waiting (with timeout) for the server to close the connection.
            err := clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
            if err != nil {
                log.Println("client write close: ", err)
            }
            return
        }
    }
}