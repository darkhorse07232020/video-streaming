package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clientSenders = make(map[*websocket.Conn]bool) // connected clients
var clientReceivers = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte) // broadcast channel

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func hello(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "hello\n")
}

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func handleRecievers(w http.ResponseWriter, r *http.Request) {
	//Upgrade initial GET request to a websocket
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {

		log.Fatal("in receeiver", err)
	}

	// Register our new client
	clientReceivers[ws] = true
	log.Printf("receiver conneceted")

	// // Make sure we close the connection when the function returns
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Register our new client
	clientSenders[ws] = true

	for {
		// Read in a new message as JSON and map it to a Message object
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			ws.Close()
			delete(clientSenders, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- p
	}

	// // Make sure we close the connection when the function returns
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		p := <-broadcast

		// Send it out to every client that is currently connected
		for client := range clientReceivers {
			if err := client.WriteMessage(1, p); err != nil {
				log.Printf("error sending message: %v", err)
				client.Close()
				delete(clientReceivers, client)
			}
		}
	}
}

func main() {

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/headers", headers)
	// Configure websocket route
	http.HandleFunc("/wsout", handleConnections)
	http.HandleFunc("/wsin", handleRecievers)

	// Start listening for incoming  messages
	go handleMessages()

	log.Println("http server started on :8090")

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
