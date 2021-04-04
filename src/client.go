// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rogpeppe/fastuuid"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	ID     string
	Groups []string
	hub    *SignalGo
	conn   *websocket.Conn
	send   chan []byte
}

func (c *Client) JoinGroup(group string) {
	c.Groups = append(c.Groups, group)
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		msgType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		log.Println(msgType)
		c.hub.messages <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) SendID() {
	payload:= CreatePayload(0,"",c.ID)
	err := c.conn.WriteMessage(websocket.BinaryMessage, payload)
	if err != nil {
		return
	}
}

// serveWs handles websocket requests from the peer.

func serveWs(hub *SignalGo, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	id := r.URL.Query().Get("id")
	isNewConnection :=id==""
	if id == "" {
		idGenerator := fastuuid.MustNewGenerator()
		id = idGenerator.Hex128()
	}
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		ID:   id,
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	if isNewConnection {
		client.SendID()
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func CreatePayload(
	messageType int,
	event string,
	body interface{}) []byte {
	msgBody, _ := json.Marshal(body)
	payload := Payload{
		MessageType: messageType,
		Event:       event,
		Message:     string(msgBody),
	}
	data, _ := json.Marshal(payload)
	return data
}

type Payload struct {
	//0=Register (System)
	//1=Message
	//2=Group
	//3=Event Registration
	MessageType int    `json:"t"`
	Event       string `json:"e"`
	//Can be message OR group name
	Message     string `json:"m"`
}
