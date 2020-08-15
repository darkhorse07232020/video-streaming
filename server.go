package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Define our message object
type Message struct {
	Data string `json:"data"`
}

var clientSenders = make(map[*websocket.Conn]bool) // connected clients
var clientReceivers = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message) // broadcast channel

// Configure the upgrader
// :CheckOrigin: func(r *http.Request) bool {
// 	return true
// }

var upgrader = websocket.Upgrader{}

// upgrader.CheckOrigin = func(r *http.Request) bool { return true }

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
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clientReceivers[ws] = true
	log.Printf("receiver conneceted")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clientSenders[ws] = true

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clientSenders, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clientReceivers {
			log.Println("sending message: ")
			err := client.WriteJSON(msg)
			if err != nil {
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
