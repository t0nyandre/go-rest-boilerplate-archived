package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"gitlab.com/t0nyandre/go-rest-boilerplate/models"
	"gitlab.com/t0nyandre/go-rest-boilerplate/responses"
	redisstore "gopkg.in/boj/redistore.v1"
)

type userResources struct {
	Db *gorm.DB
	S  *redisstore.RediStore
}

func ServeUserRoutes(db *gorm.DB, s *redisstore.RediStore, r *mux.Router) {
	res := &userResources{Db: db, S: s}
	r.HandleFunc("/user/all", res.GetAll).Methods("GET")
	r.HandleFunc("/user/profile", res.Profile).Methods("GET")
	r.HandleFunc("/user/{id}", res.Get).Methods("GET")
	r.HandleFunc("/user", res.Create).Methods("POST")
	r.HandleFunc("/login", res.Login).Methods("POST")
	r.HandleFunc("/logout", res.Logout).Methods("POST")
}

func (u *userResources) Get(w http.ResponseWriter, r *http.Request) {
	var user models.User
	params := mux.Vars(r)
	id := params["id"]
	if "" == id {
		log.Println("ID is empty")
		return
	}
	if err := u.Db.First(&user, "id = ?", id).Error; err != nil {
		responses.NewResponse(w, 404, err, nil)
		return
	}
	responses.NewResponse(w, 200, nil, &user)
}

func (u *userResources) Profile(w http.ResponseWriter, r *http.Request) {
	var user models.User

	session, err := u.S.Get(r, "sid")
	if err != nil {
		responses.NewResponse(w, 500, err, nil)
	}

	id := session.Values["user_id"]

	if err := u.Db.First(&user, "id = ?", id).Error; err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("Access denied. Please login to get access to this data"), nil)
		return
	}
	responses.NewResponse(w, 200, nil, &user)
}

func (u *userResources) GetAll(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	if err := u.Db.Find(&users).Error; err != nil {
		responses.NewResponse(w, 404, err, nil)
		return
	}
	responses.NewResponse(w, 200, nil, &users)
}

type createInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *userResources) Create(w http.ResponseWriter, r *http.Request) {
	var input createInput

	json.NewDecoder(r.Body).Decode(&input)

	newUser := models.User{
		Email:    input.Email,
		Password: input.Password,
	}

	if err := u.Db.Create(&newUser).Error; err != nil {
		responses.NewResponse(w, 400, fmt.Errorf("Could not create user: %s", err.Error()), nil)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s%s%s", r.Host, "/user/", newUser.ID))
	responses.NewResponse(w, 201, nil, &newUser)
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *userResources) Login(w http.ResponseWriter, r *http.Request) {
	var input loginInput
	var user models.User

	json.NewDecoder(r.Body).Decode(&input)

	if err := u.Db.First(&user, "email = ?", input.Email).Error; err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("Username and/or password is incorrect"), nil)
		return
	}

	valid := user.VerifyPassword(input.Password)
	if valid == false {
		responses.NewResponse(w, 401, fmt.Errorf("Username and/or password is incorrect"), nil)
		return
	}

	if !user.UserConfirmed() {
		responses.NewResponse(w, 401, fmt.Errorf("You haven't confirmed your account yet. Please check your email and click the link for confirmation"), nil)
		return
	}

	session, err := u.S.Get(r, "sid")
	if err != nil {
		log.Println(err.Error())
	}

	session.Values["user_id"] = user.ID
	session.Values["user_role"] = user.Role

	if err = session.Save(r, w); err != nil {
		log.Printf("Error saving session: %v", err)
	}

	responses.NewResponse(w, 200, nil, &user)
}

func (u *userResources) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := u.S.Get(r, "sid")
	if err != nil {
		responses.NewResponse(w, 500, err, nil)
		return
	}

	session.Options.MaxAge = -1
	if err = session.Save(r, w); err != nil {
		responses.NewResponse(w, 500, fmt.Errorf("Error saving session: %v", err), nil)
		return
	}

	responses.NewResponse(w, 200, nil, &responses.CustomResponse{Message: "Successfully logged out"})
}
