package realtime

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"menteslibres.net/gosexy/redis"
)

const (
	cacheTime = 5
)

var (
	client *redis.Client
)

type RedisAuth struct {
	Password string `json:"password"`
}

func ConnectRedis() *redis.Client {
	var redisAuth RedisAuth
	content, err := ioutil.ReadFile("../realtime/redisAuth.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(content, &redisAuth)
	if err != nil {
		panic(err)
	}
	client = redis.New()
	err = client.Connect("bclymer.com", 6379)
	if err != nil {
		log.Fatalln("Connect to Redis:", err)
	}
	_, err = client.Auth(redisAuth.Password)
	if err != nil {
		panic(err)
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
