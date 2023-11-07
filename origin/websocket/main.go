package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
    conns = make(map[*websocket.Conn]bool)
    upgrader = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
    }
)


// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(conn *websocket.Conn, isBroadcast bool) {
    for {
        // read in a message
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println(err)
            return
        }
        // print out that message for clarity
        fmt.Printf("Client %s sent: %s\n", conn.RemoteAddr() , string(p))
        // Check if message is broadcast
        if isBroadcast {
            for ws := range conns {
                go func(ws *websocket.Conn) {
                    if err := ws.WriteMessage(1, p); err != nil {
                        log.Println(err)
                    }
                }(ws)
            }
        } else { 
            if err := conn.WriteMessage(messageType, p); err != nil {
                log.Println(err)
                return
            }
        }

    }
}

// Reply to client with the same message
func echoEndpoint(w http.ResponseWriter, r *http.Request) {
    upgrader.CheckOrigin = func(r *http.Request) bool { return true }

    // upgrade this connection to a WebSocket connection
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
    }

    log.Printf("Client %s Connected\n", ws.RemoteAddr())
    err = ws.WriteMessage(1, []byte("Hi Client, this is echo endpoint!"))
    if err != nil {
        log.Println(err)
    }
    
    // listen for new messages coming through on WebSocket connection
    reader(ws, false)
}

// Send message to all clients
func broadcastEndpoint(w http.ResponseWriter, r *http.Request) {
    upgrader.CheckOrigin = func(r *http.Request) bool { return true }

    // upgrade this connection to a WebSocket connection
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
    }
    conns[ws] = true
    log.Printf("Client %s Connected\n", ws.RemoteAddr())
    err = ws.WriteMessage(1, []byte("Hi Client, this is broadcast endpoint!"))
    if err != nil {
        log.Println(err)
    }
    // listen for new messages coming through on WebSocket connection
    reader(ws, true)

}

//Periodically send server date time to client
func dataStreamEndpoint(w http.ResponseWriter, r *http.Request) {
    upgrader.CheckOrigin = func(r *http.Request) bool { return true }

    // upgrade this connection to a WebSocket connection
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
    }

    log.Printf("Client %s Connected\n", ws.RemoteAddr())
    err = ws.WriteMessage(1, []byte("Hi Client, this is data endpoint!"))
    if err != nil {
        log.Println(err)
    }

    for {
        msg := fmt.Sprintf("Server Date: %s\n", time.Now().Format(time.RFC850))
        if err = ws.WriteMessage(1, []byte(msg)); err != nil {
            log.Println(err)
        }
        time.Sleep(5 * time.Second)
    }

}
func main() {
    http.HandleFunc("/echo", echoEndpoint)
    http.HandleFunc("/broadcast", broadcastEndpoint)
    http.HandleFunc("/data", dataStreamEndpoint)
    log.Fatal(http.ListenAndServe(":8312", nil))
}