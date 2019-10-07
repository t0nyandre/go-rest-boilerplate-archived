package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/t0nyandre/go-rest-boilerplate/extras"
	"github.com/t0nyandre/go-rest-boilerplate/middleware"
	"github.com/t0nyandre/go-rest-boilerplate/models"
	"github.com/t0nyandre/go-rest-boilerplate/responses"
	"github.com/t0nyandre/go-rest-boilerplate/utils"
)

type tokenRes struct {
	Token string `json:"access_token,omitempty"`
}

func NewRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Use(middleware.HeaderMiddleware)
	r.HandleFunc("/refresh-token", RefreshToken).Methods("POST")
	ServeAuthRoutes(r)
	return r
}

func RefreshToken(w http.ResponseWriter, req *http.Request) {
	var user models.User

	cookie, err := req.Cookie("jid")
	if err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
		return
	}

	secret := strings.Split(fmt.Sprintf("%v", cookie), "=")[1]

	token, err := utils.Store.Get(fmt.Sprintf("%s%s", extras.RefreshTokenPrefix, secret)).Result()
	if err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
		return
	}

	userID, err := utils.ValidateRefreshToken(token)
	if err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
		return
	}

	if err := models.Db.First(&user, "id = ?", userID).Error; err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
		return
	}

	if user.UserDisabled() {
		responses.NewResponse(w, 401, fmt.Errorf("Account is disabled"), nil)
		return
	}

	accessToken := &tokenRes{utils.GenerateAccessToken(user)}

	w.Header().Set("Location", fmt.Sprintf("%s%s%s", req.Host, "/users/", user.Username))
	responses.NewResponse(w, 200, nil, &accessToken)
}
