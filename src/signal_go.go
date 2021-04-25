package main
import (
	"encoding/json"
	"log"
)
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
	backplane IBackplane
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
			for _, client := range g.eventClients[payload.Event] {
				client.Write(payload.MessageType,payload.Event,payload.Message)
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

func (g *SignalGo) UseBackplane(backplane IBackplane)  {
	g.backplane=backplane
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
			total:= len(g.messages)
			for i := 0; i < total; i++ {
				g.HandleIncomingMessage(<-g.messages)
			}
		}
	}
}

type Message struct {
	Client *Client
	Body []byte
}
