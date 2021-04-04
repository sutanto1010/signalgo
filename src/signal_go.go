// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type SignalGo struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	messages chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func (g *SignalGo) HandleMessage(msg []byte)  {
	fmt.Println(string(msg))
}

func (g *SignalGo) SendToUser(connectionId string, message interface{}) {

}
func (g *SignalGo) SendToGroup(group string, message interface{}) {

}

func NewSignalGo() *SignalGo {
	return &SignalGo{
		messages:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
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
			g.HandleMessage(message)
		}
	}
}
