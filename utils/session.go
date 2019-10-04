package utils

import (
	"fmt"
	"os"

	"github.com/go-session/redis"
	"github.com/go-session/session"
)

// NewSession initialize the global session manager
func NewSession() {
	session.InitManager(
		session.SetStore(redis.NewRedisStore(&redis.Options{
			Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
			DB:   0,
		})),
		session.SetSecure(false),
		session.SetCookieName("sid"),
		session.SetExpired(60*60*24*7), // 7 days
		session.SetCookieLifeTime(60*60*24*7),
	)
}
