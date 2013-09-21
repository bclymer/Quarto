package realtime

import (
	"github.com/garyburd/redigo/redis"
)

func ConnectRedis() redis.Conn {
	redisConn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	return redisConn
}
