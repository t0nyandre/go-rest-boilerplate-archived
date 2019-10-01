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
)

type userResources struct {
	Db *gorm.DB
}

func ServeUserRoutes(db *gorm.DB, r *mux.Router) {
	res := &userResources{Db: db}
	r.HandleFunc("/user/all", res.GetAll).Methods("GET")
	r.HandleFunc("/user/{id}", res.Get).Methods("GET")
	r.HandleFunc("/user", res.Create).Methods("POST")
	r.HandleFunc("/login", res.Login).Methods("POST")
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

func (u *userResources) Login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("loginUser"))
}
