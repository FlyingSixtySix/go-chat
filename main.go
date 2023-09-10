package main

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

type AppState struct {
	Messages []Message `json:"messages"`
}

type UserState struct {
	Username string `json:"username"`
}

type UserPacket struct {
	Username string `json:"username"`
}

type MessagePacket struct {
	Content string `json:"content"`
}

type OnlinePacket struct {
	Count int `json:"count"`
}

type Packet struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]*UserState)
var appState = AppState{}

func broadcast(messageType int, packet *Packet) {
	data, err := json.Marshal(packet)
	if err != nil {
		log.Printf("failed to marshal outgoing message: %s\n", err.Error())
		return
	}
	for conn := range clients {
		err := conn.WriteMessage(messageType, data)
		if err != nil {
			if !strings.Contains(err.Error(), "close sent") {
				log.Printf("Error while broadcasting: %s\n", err.Error())
			}
			err := conn.Close()
			if err != nil {
				log.Printf("Error closing connection from failed broadcast: %s\n", err.Error())
			}
			delete(clients, conn)
		}
	}
}

func broadcastOnline() {
	broadcast(websocket.TextMessage, &Packet{
		Type: "online",
		Data: &OnlinePacket{
			Count: len(clients),
		},
	})
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection %s\n", err)
		return
	}
	clients[conn] = &UserState{
		Username: "Guest",
	}

	broadcastOnline()

	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "close 1001") {
				delete(clients, conn)
				break
			}
			log.Printf("failed to read message: %s\n", err.Error())
			break
		}
		if messageType == websocket.CloseMessage {
			log.Printf("messageType is CloseMessage %d %d\n", messageType, websocket.CloseMessage)
			delete(clients, conn)
			break
		}
		var packet Packet
		err = json.Unmarshal(data, &packet)
		if err != nil {
			log.Printf("failed to unmarshal: %s\n", err.Error())
			break
		}

		log.Printf("packet: %s\n", packet)

		switch packet.Type {
		case "user":
			var userPacket UserPacket
			err = mapstructure.Decode(packet.Data, &userPacket)
			if err != nil {
				log.Printf("failed to decode user packet: %s\n", err.Error())
				break
			}
			clients[conn].Username = userPacket.Username
			break
		case "message":
			var messagePacket MessagePacket
			err = mapstructure.Decode(packet.Data, &messagePacket)
			if err != nil {
				log.Printf("failed to decode message packet: %s\n", err.Error())
				break
			}
			message := Message{
				Username: clients[conn].Username,
				Content:  messagePacket.Content,
			}
			appState.Messages = append(appState.Messages, message)
			broadcast(websocket.TextMessage, &Packet{
				Type: "message",
				Data: message,
			})
			break
		}
	}
}

func main() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		switch r.Header.Get("Origin") {
		case "http://localhost:63342":
			return true
		}
		return false
	}
	http.HandleFunc("/ws", wsHandler)
	onlineBroadcastTimer := time.NewTicker(time.Second * 5)
	go func() {
		for {
			select {
			case <-onlineBroadcastTimer.C:
				broadcastOnline()
			}
		}
	}()
	port := 3000
	log.Printf("Listening on port %d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
