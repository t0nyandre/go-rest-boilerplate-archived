package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-session/session"
	"github.com/gorilla/mux"
	"github.com/thanhpk/randstr"
	"gitlab.com/t0nyandre/go-rest-boilerplate/email"
	"gitlab.com/t0nyandre/go-rest-boilerplate/extras"
	"gitlab.com/t0nyandre/go-rest-boilerplate/middleware"
	"gitlab.com/t0nyandre/go-rest-boilerplate/models"
	"gitlab.com/t0nyandre/go-rest-boilerplate/responses"
	"gitlab.com/t0nyandre/go-rest-boilerplate/utils"
)

type userResources struct{}

// ServeUserRoutes creates all the resources for User
func ServeUserRoutes(r *mux.Router) {
	res := userResources{}
	r.HandleFunc("/user/all", middleware.AuthRequired(res.GetAll)).Methods("GET")
	r.HandleFunc("/user/me", middleware.AuthRequired(res.Me)).Methods("GET")
	r.HandleFunc("/user/{username}", res.Get).Methods("GET")
	r.HandleFunc("/user", res.Create).Methods("POST")
}

func (res *userResources) Get(w http.ResponseWriter, req *http.Request) {
	var user models.User

	params := mux.Vars(req)
	username := params["username"]

	if err := models.Db.First(&user, "username = ?", username).Error; err != nil {
		responses.NewResponse(w, 404, err, nil)
		return
	}

	responses.NewResponse(w, 200, nil, &user)
}

func (res *userResources) Me(w http.ResponseWriter, req *http.Request) {
	var user models.User

	store, err := session.Start(context.Background(), w, req)
	if err != nil {
		log.Println(err.Error())
	}

	id, _ := store.Get("user_id")

	if err := models.Db.First(&user, "id = ?", id).Error; err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.AccessDenied), nil)
		return
	}

	responses.NewResponse(w, 200, nil, &user)
}

func (res *userResources) GetAll(w http.ResponseWriter, req *http.Request) {
	var users []models.User

	if err := models.Db.Find(&users).Error; err != nil {
		responses.NewResponse(w, 404, err, nil)
		return
	}
	responses.NewResponse(w, 200, nil, &users)
}

type createInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type confirmMail struct {
	ConfirmURL string
}

// Create account if the email hasn't been used before.
// Sending our an email with a confirmation token so they can confirm the account.
func (res *userResources) Create(w http.ResponseWriter, req *http.Request) {
	var input createInput
	var confirmData confirmMail

	json.NewDecoder(req.Body).Decode(&input)

	newUser := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
	}

	if err := models.Db.Create(&newUser).Error; err != nil {
		responses.NewResponse(w, 400, fmt.Errorf("Could not create user: %s", err.Error()), nil)
		return
	}

	token := randstr.Hex(22)
	_, err := utils.Store.Set(fmt.Sprintf("%s%s", extras.ConfirmAccountPrefix, token), newUser.ID, time.Second*60*60*24*3).Result()
	if err != nil {
		responses.NewResponse(w, 500, fmt.Errorf("Could not create user: %s", err.Error()), nil)
		return
	}

	confirmData.ConfirmURL = fmt.Sprintf("%s%s/%s", os.Getenv("APP_URL"), os.Getenv("APP_CONFIRM_PATH"), token)

	if os.Getenv("API_ENV") != "dev" {
		go email.ConfirmAccountEmail("Tony Andre Haugen <no_reply@tonyandre.co>", []string{input.Email}, confirmData)
	} else {
		w.Header().Set("ConfirmURL", confirmData.ConfirmURL)
	}

	w.Header().Set("Location", fmt.Sprintf("%s%s%s", req.Host, "/user/", newUser.Username))
	responses.NewResponse(w, 201, nil, newUser)
}
