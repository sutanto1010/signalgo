package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type RedisBackplane struct {
	Host      string
	Password  string
	PrefixKey string
	DB        int
	Client    *redis.Client
	sg        *SignalGo
}

const (
	SignalGoOnMessage    = "SignalGo-OnMessage"
	SignalGoOnUnRegister = "SignalGo-OnUnRegister"
)

func (r *RedisBackplane) SubscribeOnMessage() {
	pubsubCtx := context.Background()
	pubsub := r.Client.Subscribe(pubsubCtx, SignalGoOnMessage)
	for msg := range pubsub.Channel() {
		var message Message
		err := json.Unmarshal([]byte(msg.Payload), &message)
		if err == nil {
			if message.SignalGoInstanceID != r.sg.SignalGoInstanceID {
				r.sg.HandleIncomingMessage(message)
			}
		}
	}
	fmt.Println("SubscribeOnMessage stopped")
}
func (r *RedisBackplane) SubscribeOnUnRegister() {
	pubsubCtx := context.Background()
	pubsub := r.Client.Subscribe(pubsubCtx, SignalGoOnUnRegister)
	for msg := range pubsub.Channel() {
		var client Client
		json.Unmarshal([]byte(msg.Payload), &client)
		cacheKey := client.ID
		r.del(cacheKey + "*")
	}
	fmt.Println("SubscribeOnUnRegister stopped")
}

func (r *RedisBackplane) Init(sg *SignalGo) {
	r.sg = sg
	go r.SubscribeOnMessage()
	go r.SubscribeOnUnRegister()
}

func (r *RedisBackplane) set(key string, value interface{}) {
	payload, _ := json.Marshal(value)
	r.Client.Set(context.Background(), key, string(payload), 0)
}
func (r *RedisBackplane) del(key string) {
	cmd := r.Client.Keys(context.Background(), key)
	keys, _ := cmd.Result()
	for _, k := range keys {
		r.Client.Del(context.Background(), k)
	}
}

func (r *RedisBackplane) OnMessage(message Message) {
	data, _ := json.Marshal(message)
	var payload Payload
	err := json.Unmarshal(message.Body, &payload)
	if err == nil {
		if payload.MessageType == EventRegistration {
			cacheKey := message.Client.ID + "-events"
			r.set(cacheKey, message.Client.Events)
		}
	}
	r.Client.Publish(context.Background(), SignalGoOnMessage, string(data))
}

func (r *RedisBackplane) OnUnRegister(client *Client) {
	data, _ := json.Marshal(client)
	r.Client.Publish(context.Background(), SignalGoOnUnRegister, string(data))
}

func (r *RedisBackplane) OnRegister(client *Client) {

}

func NewRedisBackplane(
	redisHost string,
	redisPassword string,
	redisDB int,
	useTLS bool,
) RedisBackplane {
	options := &redis.Options{
		Addr:     redisHost,
		DB:       redisDB,
		Password: redisPassword,
	}

	if useTLS {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	obj := RedisBackplane{
		Host:     redisHost,
		Password: redisPassword,
		DB:       redisDB,
		Client:   redis.NewClient(options),
	}

	return obj
}
