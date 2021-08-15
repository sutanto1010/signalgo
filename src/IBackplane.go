package main

type IBackplane interface {
	Init(sg *SignalGo)
	OnMessage(message Message)
	OnRegister(client *Client)
	OnUnRegister(client *Client)
}
