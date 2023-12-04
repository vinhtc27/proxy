package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

const (
	Echo = iota + 1
	Broadcast
)

type wsMessage struct {
	Type    int    `json:"messageType"`
	Content string `json:"content"`
}

var (
	conns    = make(map[*websocket.Conn]bool)
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		fmt.Printf("Client %s sent: %s\n", conn.RemoteAddr(), string(p))

		rcvMsg := wsMessage{}
		if err := json.Unmarshal(p, &rcvMsg); err != nil {
			log.Println(err)
			if err := conn.WriteMessage(1, []byte(`Message Format Error! (e.g. {"messageType": 1, "content": "example"})`)); err != nil {
				log.Println(err)
				return
			}
		} else if rcvMsg.Type == 0 || len(rcvMsg.Content) == 0 {
			if err := conn.WriteMessage(1, []byte(`Message Format Error! (e.g. {"messageType": 1, "content": "example"})`)); err != nil {
				log.Println(err)
				return
			}
		} else {
			switch rcvMsg.Type {
			case Broadcast:
				for ws := range conns {
					go func(ws *websocket.Conn) {
						if err := ws.WriteMessage(1, []byte(rcvMsg.Content)); err != nil {
							log.Println(err)
						}
					}(ws)
				}
			case Echo:
				if err := conn.WriteMessage(messageType, []byte(rcvMsg.Content)); err != nil {
					log.Println(err)
					return
				}
			default:
				if err := conn.WriteMessage(1, []byte(`Undefined Message Type! (1 - Echo, 2 - Broadcast)`)); err != nil {
					log.Println(err)
					return
				}
			}

		}
	}
}

// Websocket root endpoint
// Message format: {"messageType": 1, "content": "example"}
// messageType: 1 - Echo, 2 - Broadcast
func rootEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	conns[ws] = true
	log.Printf("Client %s Connected\n", ws.RemoteAddr())
	err = ws.WriteMessage(1, []byte(
		`Hi Client, this is root endpoint!
        Message format: {"messageType": 1, "content": "example"}
        MessageType: 1 - Echo, 2 - Broadcast
        `))
	if err != nil {
		log.Println(err)
	}

	// listen for new messages coming through on WebSocket connection
	reader(ws)
}

// Periodically send server date time to client
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
	r := chi.NewRouter()
	r.HandleFunc("/", rootEndpoint)
	r.HandleFunc("/data", dataStreamEndpoint)
	host := os.Args[1]

	origin := http.Server{
		Addr:    host,
		Handler: r,
	}
	log.Printf("Origin started at %s\n", host)
	if err := origin.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
