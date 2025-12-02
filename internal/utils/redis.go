package utils

import (
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client 

func InitRedis() {
	url := os.Getenv("REDIS_URL")

	opt, err := redis.ParseURL(url)
	if err != nil {
		panic("FAILED TO PARSE REDIS_URL: " + err.Error())
	}

	RedisClient = redis.NewClient(opt)
}
