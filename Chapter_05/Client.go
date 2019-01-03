package main

import (
    "log"

    "github.com/gorilla/websocket"
)

type Client struct {
    ws *websocket.Conn
    messages chan []byte
    connected bool

    hub *P2pServer
}

func (this *Client) writePump() {
    defer func() {
        this.setConnection(false)
        this.ws.Close()
    }()

    for {
        select {
        case message, ok := <- this.messages:
            if !ok {
                log.Println("connection closing...")
                err := this.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
                if err != nil {
                    log.Println("connection close: ", err)
                }
                return
            }
            err := this.ws.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                log.Println("connection write: ", err)
                return
            }
        }
    }
}

func (this *Client) readPump() {
    defer func() {
        if this.getConnection() {
            close(this.messages)
        }
        this.hub.unregister <- this
    }()

    for {
        _, message, err := this.ws.ReadMessage()
        if err != nil {
            log.Println("connection read: ", err)
            break
        }
        if parseMessage(message) {
            if this.hub.p2pClient.getConnection() {
                this.hub.p2pClient.messages <- message
            }
            this.hub.broadcast <- message
        }
    }
}

func (this *Client) setConnection(status bool) {
    this.connected = status
}

func (this Client) getConnection() bool {
    return this.connected
}