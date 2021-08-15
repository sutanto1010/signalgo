package main

import (
	"crypto/tls"

	"github.com/go-redis/redis/v8"
)

type RedisBackplane struct {
	Host           string
	Password       string
	PrefixKey      string
	DB             int
	SyncMessageKey string
	Client         *redis.Client
}

func (r *RedisBackplane) Start() {

}

func (r *RedisBackplane) OnMessage(senderId string, message interface{}) {

}

func (r *RedisBackplane) OnUnregister(clientId string) {

}

func (r *RedisBackplane) OnRegister(clientId string) {

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
