package middleware

import (
	"fmt"
	"log"
	"os"

	redis "github.com/go-redis/redis"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

func RateLimit() (*stdlib.Middleware, error) {
	rate, err := limiter.NewRateFromFormatted("8-S")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       1,
	})

	// Create a store with the redis client.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "limiter:",
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Create a new middleware with the limiter instance.
	return stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true))), nil
}
