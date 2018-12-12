package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "time"
    "unsafe"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
)

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
        go p2pClient.init(peers, &p2pServer)
    }
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
    message := []byte("I'm alive!\n")
    log.Print(string(message))

    if p2pServer.numberOfClient() > 0 {
        bytes := make([]byte, unsafe.Sizeof((*byte)(nil)))
        p2pServer.broadcast <- append(bytes, message...)
    }
    if p2pClient.getConnection() {
        p2pClient.messages <- message
    }

    w.Write(message)
}

func parseMessage(message []byte) bool {
    log.Print(string(message))
    return true
}