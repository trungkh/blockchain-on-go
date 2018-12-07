package main

import (
    "log"
    "net/http"
    "os"
    "strconv"

    "github.com/gorilla/websocket"
)

type P2pServer struct {
    ws *websocket.Conn
    messages chan []byte
    connected bool

    client *P2pClient
    done chan struct{}
}

func (this *P2pServer) init(client *P2pClient) {
    http.HandleFunc("/", this.handleMessage)

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
    this.client = client

    log.Println("Listening websocket p2p on port: ", wsPort)
    if err := http.ListenAndServe(/*addr*/":" + wsPort, nil); err != nil {
        log.Fatal(err)
    }
}

func (this *P2pServer) handleMessage(w http.ResponseWriter, r *http.Request) {
    conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
    if err != nil {
        log.Println("upgrade: ", err)
        return
    }
    this.ws = conn
    this.messages = make(chan []byte)
    this.done = make(chan struct{})

    log.Println("Connnection established...")
    this.setConnection(true)

    go this.writePump()
    go nodeSync(this.messages)
    this.readPump()
}

func (this *P2pServer) writePump() {
    defer func() {
        this.setConnection(false)
        this.ws.Close()
        close(this.done)
    }()

    for {
        select {
        case message, ok := <- this.messages:
            if !ok {
                log.Println("server closing...")
                err := this.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
                if err != nil {
                    log.Println("server close: ", err)
                }
                return
            }
            err := this.ws.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                log.Println("server write: ", err)
                return
            }
        }
    }
}

func (this *P2pServer) readPump() {
    defer func() {
        if this.getConnection() {
            close(this.messages)
        }
    }()

    for {
        _, message, err := this.ws.ReadMessage()
        if err != nil {
            log.Println("server read: ", err)
            break
        }
        log.Println("Syncing Block: ", string(message))
        
        parseMessage(message, this.client.getConnection(), this.client.messages)
    }
}

func (this *P2pServer) setConnection(status bool) {
    this.connected = status
}

func (this *P2pServer) getConnection() bool {
    return this.connected
}