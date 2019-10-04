package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-session/session"
	"github.com/gorilla/mux"
	"gitlab.com/t0nyandre/go-rest-boilerplate/extras"
	"gitlab.com/t0nyandre/go-rest-boilerplate/middleware"
	"gitlab.com/t0nyandre/go-rest-boilerplate/models"
	"gitlab.com/t0nyandre/go-rest-boilerplate/responses"
	"gitlab.com/t0nyandre/go-rest-boilerplate/utils"
)

type authResources struct{}

// ServeAuthRoutes creates all the resources for User
func ServeAuthRoutes(r *mux.Router) {
	res := authResources{}
	r.HandleFunc("/confirm/{token}", res.Confirm).Methods("GET")
	r.HandleFunc("/login", res.Login).Methods("POST")
	r.HandleFunc("/logout", middleware.AuthRequired(res.Logout)).Methods("POST")
}

func (r *authResources) Confirm(w http.ResponseWriter, req *http.Request) {
	var user models.User
	params := mux.Vars(req)
	token := params["token"]

	if "" == token {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
		return
	}
	id, err := utils.Store.Get(fmt.Sprintf("%s%s", extras.ConfirmAccountPrefix, token)).Result()
	if err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
		return
	}

	if err = models.Db.First(&user, "id = ?", id).Error; err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
		return
	}

	user.Confirmed = true
	if err = models.Db.Save(&user).Error; err != nil {
		responses.NewResponse(w, 500, fmt.Errorf("Could not confirm user: %s", err.Error()), nil)
		return
	}

	_, err = utils.Store.Del(fmt.Sprintf("%s%s", extras.ConfirmAccountPrefix, token)).Result()
	if err != nil {
		responses.NewResponse(w, 500, fmt.Errorf("%s", err.Error()), nil)
		return
	}

	store, err := session.Start(context.Background(), w, req)
	if err != nil {
		log.Println(err.Error())
	}

	store.Set("user_id", user.ID)
	store.Set("user_role", user.Role)

	if err = store.Save(); err != nil {
		log.Printf("Error saving session: %v", err)
	}

	w.Header().Set("Location", fmt.Sprintf("%s%s", req.Host, "/me"))
	responses.NewResponse(w, 201, nil, &user)
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login is where we log the user in. We check if their email exists, verify password and see if they have confirmed
// their user account. If every check is ok we generate a cookie for them "sid" and save their session in redis
func (r *authResources) Login(w http.ResponseWriter, req *http.Request) {
	var input loginInput
	var user models.User

	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Printf("%v", err)
	}

	if err = models.Db.First(&user, "email = ?", input.Email).Error; err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.WrongUserOrPassword), nil)
		return
	}

	valid := user.VerifyPassword(input.Password)
	if valid == false {
		responses.NewResponse(w, 401, fmt.Errorf("Username and/or password is incorrect ... wrong passworrd"), nil)
		return
	}

	if !user.UserConfirmed() {
		responses.NewResponse(w, 401, fmt.Errorf("You haven't confirmed your account yet. Please check your email and click the link for confirmation"), nil)
		return
	}

	store, err := session.Start(context.Background(), w, req)
	if err != nil {
		log.Println(err.Error())
	}

	store.Set("user_id", user.ID)
	store.Set("user_role", user.Role)

	if err = store.Save(); err != nil {
		log.Printf("Error saving session: %v", err)
	}

	responses.NewResponse(w, 200, nil, &user)
}

// Logout is used for logging the user out byt removing their session and make their cookie "sid" blank.
func (r *authResources) Logout(w http.ResponseWriter, req *http.Request) {
	err := session.Destroy(context.Background(), w, req)
	if err != nil {
		responses.NewResponse(w, 500, fmt.Errorf("Error logging you out: %v", err), nil)
		return
	}

	responses.NewResponse(w, 200, nil, &responses.CustomResponse{Message: "Successfully logged out"})
}
