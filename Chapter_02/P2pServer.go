package main

import (
    "encoding/binary"
    "log"
    "net/http"
    "os"
    "strconv"
    "unsafe"

    "github.com/gorilla/websocket"
)

type P2pServer struct {
    // Registered clients
    clients map[*Client]bool

    // Unregister requests
    unregister chan *Client

    // Inbound messages
    broadcast chan []byte

    // p2p client connection
    p2pClient *P2pClient

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
    this.clients = make(map[*Client]bool)
    this.unregister = make(chan *Client)
    this.broadcast = make(chan []byte)
    this.p2pClient = client
    this.done = make(chan struct{})

    var oops bool = false
    go this.unregisterHandler(&oops)
    go this.broadcastHandler(&oops)

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

    client := &Client{
        ws: conn,
        messages: make(chan []byte),
        connected: true,
        hub: this,
    }
    
    // Register requests
    this.clients[client] = true
    log.Println("Connnection established...")

    go client.writePump()
    // Syncing entire ledger if any node connected to this node
    go nodeSync(client.messages)
    client.readPump()
}

func (this *P2pServer) unregisterHandler(oops *bool) {
    defer func () {
        close(this.unregister)
        close(this.done)
    }()

    for {
        select {
        case client := <-this.unregister:
            _, ok := this.clients[client]
            if ok {
                delete(this.clients, client)
            }
            if this.numberOfClient() == 0 && *oops {
                return
            }
        }
    }
}

func (this *P2pServer) broadcastHandler(oops *bool) {
    for {
        select {
        case message, ok := <-this.broadcast:
            if !ok {
                log.Println("server closing...")
                for client := range this.clients {
                    close(client.messages)
                }
                *oops = true
                return
            }

            size := unsafe.Sizeof((*byte)(nil))
            ptr := this.convertToPointer(message[:size])
            c := (*Client)(unsafe.Pointer(ptr))
            for client := range this.clients {
                // Prevent loop back sending to client
                if client == c {
                    continue
                }
                client.messages <- message[size:]
            }
        }
    }
}

func (this *P2pServer) numberOfClient() int {
    return len(this.clients)
}

// Convert byte slice to pointer
func (this *P2pServer) convertToPointer(bytes []byte) uintptr {
    var ptr uintptr
    switch len(bytes) {
    case 4:
        ptr = uintptr(binary.LittleEndian.Uint32(bytes))
    case 8:
        ptr = uintptr(binary.LittleEndian.Uint64(bytes))
    }
    return ptr
}