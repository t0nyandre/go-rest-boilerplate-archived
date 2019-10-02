package db

import (
	"fmt"
	"os"

	"gitlab.com/t0nyandre/go-rest-boilerplate/extras"
	redisstore "gopkg.in/boj/redistore.v1"
)

func NewStore() *redisstore.RediStore {
	store, err := redisstore.NewRediStore(10, "tcp", fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")), os.Getenv("REDIS_PASSWORD"), []byte(os.Getenv("SESSION_KEY")))
	if err != nil {
		panic(err)
	}

	store.Options.HttpOnly = true
	store.SetKeyPrefix(string(extras.SessionPrefix))
	store.SetMaxAge(60 * 60 * 24 * 7) // 7days

	return store
}
