package main

type IBackplane interface {
	Start()
	OnMessage(message Message)
	OnRegister(client *Client)
	OnUnRegister(client *Client)
}
