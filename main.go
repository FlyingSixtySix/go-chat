package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type state struct {
	active   bool
	username string
}

func (s *state) timeout(conn *websocket.Conn) {
	timer := time.NewTimer(time.Minute * 5)
	go func() {
		<-timer.C
		delete(clients, conn)
	}()
}

var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]*state)

func broadcast(messageType int, message []byte) {
	for client := range clients {
		err := client.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("Error while broadcasting: %s\n", err.Error())
			err := client.Close()
			if err != nil {
				log.Printf("Error closing connection from failed broadcast: %s\n", err.Error())
			}
			clients[client].timeout(client)
		}
	}
}

type UserPacket struct {
	Username string
}

type MessagePacket struct {
	Content string
}

type Packet struct {
	PacketType string `json:"type"`
	Data       interface{}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade WebSocket connection", err)
		return
	}
	clients[conn] = &state{
		active:   true,
		username: "Guest",
	}

	for {
		messageType, messageBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var packet Packet
		err = json.Unmarshal(messageBytes, &packet)
		if err != nil {
			break
		}
		fmt.Println(packet.PacketType, packet.Data)
		broadcast(messageType, messageBytes)
	}
}

func main() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		fmt.Println(r.Header.Get("Origin"))
		switch r.Header.Get("Origin") {
		case "http://localhost:63342":
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
