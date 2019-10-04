package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-session/session"
	"gitlab.com/t0nyandre/go-rest-boilerplate/extras"
	"gitlab.com/t0nyandre/go-rest-boilerplate/models"
	"gitlab.com/t0nyandre/go-rest-boilerplate/responses"
)

type authUser struct {
	UserID string
	Role   string
}

// AuthRequired is the middleware used to protect an endpoint
func AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		var authUser authUser

		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			responses.NewResponse(w, 401, fmt.Errorf("%s", extras.AccessDenied), nil)
			return
		}

		value, ok := store.Get("user_id")
		if !ok {
			session.Destroy(context.Background(), w, r)
			responses.NewResponse(w, 401, fmt.Errorf("%s", extras.AccessDenied), nil)
			return
		}

		authUser.UserID = fmt.Sprintf("%s", value)

		if err := models.Db.First(&user, "id = ?", authUser.UserID).Error; err != nil {
			session.Destroy(context.Background(), w, r)
			responses.NewResponse(w, 401, fmt.Errorf("%s", extras.AccessDenied), nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
