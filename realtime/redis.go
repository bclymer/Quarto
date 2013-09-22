package realtime

import (
	"log"
	"menteslibres.net/gosexy/redis"
)

const (
	cacheTime = 5
)

var (
	client *redis.Client
)

func ConnectRedis() *redis.Client {
	client = redis.New()
	err := client.Connect("bclymer.com", 6379)
	if err != nil {
		log.Fatalln("Connect to Redis:", err)
	}
	return client
}

func RedisPut(key, value string) {
	client.Set(key, value)
	client.Expire(key, cacheTime)
}

func RedisGet(key string) (string, error) {
	return client.Get(key)
}
