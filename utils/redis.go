package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
)

// Store holds the global redis client
var Store *redis.Client
var err error

// ConnectRedis will generate a Redis Client which will be held in the global Store variable
func ConnectRedis() {
	Store = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := Store.Ping().Result()
	if err != nil {
		panic(err)
	}
	log.Println("Connection to redis was a success!")
}
