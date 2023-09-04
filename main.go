package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]bool)

func broadcast(messageType int, message []byte) {
	for client := range clients {
		err := client.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("Error while broadcasting: %s\n", err.Error())
			err := client.Close()
			if err != nil {
				log.Printf("Error closing connection from failed broadcast: %s\n", err.Error())
			}
			delete(clients, client)
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade WebSocket connection", err)
		return
	}
	clients[conn] = true

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		broadcast(messageType, message)
	}
}

func main() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		fmt.Println(r.Header.Get("Origin"))
		switch r.Header.Get("Origin") {
		case "http://localhost:63342":
		case "":
			return true
		}
		return false
	}
	http.HandleFunc("/ws", wsHandler)
	port := 3000
	log.Printf("Listening on port %d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
