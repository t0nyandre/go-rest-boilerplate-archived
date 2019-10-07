package middleware

import (
	"net/http"

	"gitlab.com/t0nyandre/go-rest-boilerplate/utils"
)

var limiter = utils.NewIPRateLimiter(1, 9)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := limiter.GetLimiter(r.RemoteAddr)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
