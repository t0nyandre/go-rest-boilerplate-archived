package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"gitlab.com/t0nyandre/go-rest-boilerplate/extras"
	"gitlab.com/t0nyandre/go-rest-boilerplate/responses"
	"gitlab.com/t0nyandre/go-rest-boilerplate/utils"
)

type ContextType string

const LoggedInUserCtx ContextType = "AuthUser"

// AuthRequired is the middleware used to protect an endpoint
func AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authorization := r.Header.Get("Authorization")

		token := strings.Split(fmt.Sprintf("%s", authorization), " ")
		if len(token) < 2 {
			responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
			return
		}

		payload, err := utils.ValidateAccessToken(token[1])
		if err != nil {
			responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
			return
		}

		ctx = context.WithValue(ctx, LoggedInUserCtx, payload)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
