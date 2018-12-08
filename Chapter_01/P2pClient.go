package main

import (
    "log"
    "net/url"

    "github.com/gorilla/websocket"
)

type P2pClient struct {
    ws *websocket.Conn
    messages chan []byte
    connected bool

    server *P2pServer
    done chan struct{}
}

func (this *P2pClient) init(peers string, server *P2pServer) {
    peer, err := url.Parse(peers); 
    if err != nil {
        log.Fatal(err)
    }

    u := url.URL { Scheme: peer.Scheme, Host: peer.Host, Path: "/" }
    conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal("Dial error: ", err)
    }
    this.ws = conn
    this.server = server
    this.messages = make(chan []byte)
    this.done = make(chan struct{})

    log.Println("Connected p2p on host: ", peer.Host)
    this.setConnection(true)

    go this.writePump()
    this.readPump()
}

func (this *P2pClient) writePump() {
    defer func() {
        this.setConnection(false)
        this.ws.Close()
        close(this.done)
    }()

    for {
        select {
        case message, ok := <- this.messages:
            if !ok {
                log.Println("client closing...")
                err := this.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
                if err != nil {
                    log.Println("client close: ", err)
                }
                return
            }
            err := this.ws.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                log.Println("client write: ", err)
                return
            }
        }
    }
}

func (this *P2pClient) readPump() {
    defer func () {
        if this.getConnection() {
            close(this.messages)
        }
    }()

    for {
        _, message, err := this.ws.ReadMessage()
        if err != nil {
            log.Println("client read:", err)
            break
        }
        parseMessage(message, this.server.getConnection(), this.server.messages)
    }
}

func (this *P2pClient) setConnection(status bool) {
    this.connected = status
}

func (this *P2pClient) getConnection() bool {
    return this.connected
}