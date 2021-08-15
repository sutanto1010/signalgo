package main

type IBackplane interface {
	Start()
	OnMessage(senderId string, message interface{})
	OnRegister(clientId string)
	OnUnregister(clientId string)
}
