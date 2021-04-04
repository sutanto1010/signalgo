// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type SignalGo struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	messages chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
	groupClients map[string][]*Client
	eventClients map[string][]*Client
}

func (g *SignalGo) HandleIncomingMessage(msg Message)  {
	var payload Payload
	err := json.Unmarshal(msg.Body, &payload)
	if err!=nil{
		log.Println(err)
	}
	switch payload.MessageType {
		case 3:
			g.eventClients[payload.Event]=append(g.eventClients[payload.Event],msg.Client)
			break
		case 1:
			data:=CreatePayload(payload.MessageType,payload.Event,payload.Message)
			for _, client := range g.eventClients[payload.Event] {
				client.conn.WriteMessage(websocket.BinaryMessage, data)
			}
			break

	}
}

func (g *SignalGo) SendToUser(connectionId string, message interface{}) {
	panic("Implement me!")
}
func (g *SignalGo) SendToGroup(group string, message interface{}) {
	panic("Implement me!")
}

func NewSignalGo() *SignalGo {
	return &SignalGo{
		messages:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		eventClients: make(map[string][]*Client),
		groupClients: make(map[string][]*Client),
	}
}

func (g *SignalGo) Run() {
	for {
		select {
		case client := <-g.register:
			g.clients[client.ID] = client
			log.Printf("Register: %v", client.ID)
		case client := <-g.unregister:
			if _, ok := g.clients[client.ID]; ok {
				log.Printf("Unregister: %v", client.ID)
				delete(g.clients, client.ID)
				close(client.send)
			}
		case message := <-g.messages:
			g.HandleIncomingMessage(message)
		}
	}
}

type Message struct {
	Client *Client
	Body []byte
}
