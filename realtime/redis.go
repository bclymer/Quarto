package realtime

import (
	"log"
	"menteslibres.net/gosexy/redis"
)

var (
	client *redis.Client
)

func ConnectRedis() *redis.Client {
	client := redis.New()
	err := client.Connect("bclymer.com", 6379)
	if err != nil {
		log.Fatalln("Connect to Redis:", err)
	}
	return client
}

func RedisPut(key, value string) {
	client.Set(key, value)
	client.Expire(key, 10)
}

func RedisGet(key string) (string, error) {
	return client.Get(key)
}
