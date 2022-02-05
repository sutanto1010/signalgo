package signalgo

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisBackplane struct {
	Host     string
	Password string
	DB       int
	Client   *redis.Client
	sg       *SignalGo
}

const (
	SignalGoOnMessage    = "SignalGo-OnMessage"
	SignalGoOnUnRegister = "SignalGo-OnUnRegister"
)

// Handle one message event
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

// Handle on unregister events
func (r *RedisBackplane) SubscribeOnUnRegister() {
	pubsubCtx := context.Background()
	pubsub := r.Client.Subscribe(pubsubCtx, SignalGoOnUnRegister)
	for msg := range pubsub.Channel() {
		var client Client
		json.Unmarshal([]byte(msg.Payload), &client)
		cacheKey := client.ID + "-events"
		var events []string
		err := json.Unmarshal([]byte(r.get(cacheKey)), &events)
		if err == nil {
			r.set(cacheKey, events, 5*time.Minute)
		}
	}
	fmt.Println("SubscribeOnUnRegister stopped")
}

func (r *RedisBackplane) Init(sg *SignalGo) {
	r.sg = sg
	go r.SubscribeOnMessage()
	go r.SubscribeOnUnRegister()
}

func (r *RedisBackplane) get(key string) string {
	return r.Client.Get(context.Background(), key).Val()
}

func (r *RedisBackplane) set(key string, value interface{}, duration time.Duration) {
	payload, _ := json.Marshal(value)
	r.Client.Set(context.Background(), key, string(payload), duration)
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
			r.set(cacheKey, message.Client.Events, 0)
		}
	}
	r.Client.Publish(context.Background(), SignalGoOnMessage, string(data))
}

func (r *RedisBackplane) OnUnRegister(client *Client) {
	data, _ := json.Marshal(client)
	r.Client.Publish(context.Background(), SignalGoOnUnRegister, string(data))
}

func (r *RedisBackplane) OnRegister(client *Client) {
	cacheKey := client.ID + "-events"
	var events []string
	err := json.Unmarshal([]byte(r.get(cacheKey)), &events)
	if err == nil {
		for _, event := range events {
			r.sg.eventClients[event] = append(r.sg.eventClients[event], client)
		}
	}
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
