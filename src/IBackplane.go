package main

type IBackplane interface {
	Start()
	OnMessage(senderId string, message interface{})
}
