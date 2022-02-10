package signalgo

import (
	"encoding/json"
	"log"
)

type SignalGo struct {
	SignalGoInstanceID string
	// Registered clients.
	clients map[string]*Client
	// Inbound messages from the clients.
	messages chan Message
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister   chan *Client
	groupClients map[string][]*Client
	eventClients map[string][]*Client
	backplane    IBackplane
}

// Close websocket connection
func (sg *SignalGo) CloseClient(c *Client) {
	delete(sg.clients, c.ID)
	for _, event := range c.Events {
		var temp []*Client
		for _, client := range sg.eventClients[event] {
			if client.ID != c.ID {
				temp = append(temp, client)
			}
		}
		sg.eventClients[event] = temp
	}
	close(c.send)
	if sg.backplane != nil {
		sg.backplane.OnUnRegister(c)
	}
}

// Handle incoming message
func (sg *SignalGo) HandleIncomingMessage(msg Message) {
	var payload Payload
	err := json.Unmarshal(msg.Body, &payload)
	if err != nil {
		log.Println(err)
	}
	switch payload.MessageType {
	case EventRegistration:
		sg.eventClients[payload.Event] = append(sg.eventClients[payload.Event], msg.Client)
		msg.Client.Events = append(msg.Client.Events, payload.Event)
	case UserMessage:
		for _, client := range sg.eventClients[payload.Event] {
			client.Write(payload.MessageType, payload.Event, payload.Message)
		}
	}
}

// Send message to use (by using user id), the message can by anything (interface{})
func (sg *SignalGo) SendToUser(connectionId string, message interface{}) {
	panic("Implement me!")
}

// Send message to group, the message can by anything (interface{})
func (sg *SignalGo) SendToGroup(group string, message interface{}) {
	panic("Implement me!")
}
func (sg *SignalGo) SendToEvent(eventName string, message interface{}) {
	msg, _ := json.Marshal(message)
	payload := Payload{
		MessageType: UserMessage,
		Event:       eventName,
		Message:     string(msg),
	}
	for _, client := range sg.eventClients[eventName] {
		client.Write(payload.MessageType, payload.Event, payload.Message)
	}
}

// Create new SignalGo instance
func NewSignalGo() *SignalGo {
	return &SignalGo{
		SignalGoInstanceID: NewID(),
		messages:           make(chan Message),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		clients:            make(map[string]*Client),
		eventClients:       make(map[string][]*Client),
		groupClients:       make(map[string][]*Client),
	}
}

// Register backplane
func (g *SignalGo) UseBackplane(backplane IBackplane) {
	g.backplane = backplane
}

// Start SignalGo instance
func (g *SignalGo) Run() {
	for {
		select {
		case client := <-g.register:
			g.clients[client.ID] = client
			log.Printf("Register: %v", client.ID)
			if g.backplane != nil {
				g.backplane.OnRegister(client)
			}
		case client := <-g.unregister:
			if _, ok := g.clients[client.ID]; ok {
				log.Printf("Unregister: %v", client.ID)
				g.CloseClient(client)
			}

		case message := <-g.messages:
			g.HandleIncomingMessage(message)
			total := len(g.messages)
			for i := 0; i < total; i++ {
				g.HandleIncomingMessage(<-g.messages)
			}
			if g.backplane != nil {
				message.SignalGoInstanceID = g.SignalGoInstanceID
				g.backplane.OnMessage(message)
			}
		}
	}
}

type Message struct {
	Client             *Client
	Body               []byte
	SignalGoInstanceID string
}
