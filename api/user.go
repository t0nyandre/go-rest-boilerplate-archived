package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/thanhpk/randstr"
	"gitlab.com/t0nyandre/go-rest-boilerplate/extras"
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
	r.HandleFunc("/user/confirm/{token}", res.Confirm).Methods("GET")
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

func (u *userResources) Confirm(w http.ResponseWriter, r *http.Request) {
	var user models.User
	params := mux.Vars(r)
	token := params["token"]
	redConn := u.S.Pool.Get()
	if "" == token {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
		return
	}
	id, err := redConn.Do("GET", token)
	if err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
		return
	}

	if err = u.Db.First(&user, "id = ?", id).Error; err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.BadTokenError), nil)
		return
	}

	user.Confirmed = true
	if err = u.Db.Save(&user).Error; err != nil {
		responses.NewResponse(w, 500, fmt.Errorf("Could not confirm user: %s", err.Error()), nil)
		return
	}

	_, err = redConn.Do("DEL", token)
	if err != nil {
		responses.NewResponse(w, 500, fmt.Errorf("%s", err.Error()), nil)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s%s%s", r.Host, "/user/", user.ID))
	responses.NewResponse(w, 201, nil, &user)
}

func (u *userResources) Profile(w http.ResponseWriter, r *http.Request) {
	var user models.User

	session, err := u.S.Get(r, "sid")
	if err != nil {
		responses.NewResponse(w, 500, err, nil)
	}

	id := session.Values["user_id"]

	if err := u.Db.First(&user, "id = ?", id).Error; err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.AccessDenied), nil)
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

type confirmMail struct {
	ConfirmURL string
}

// Create account if the email hasn't been used before.
// Sending our an email with a confirmation token so they can confirm the account.
func (u *userResources) Create(w http.ResponseWriter, r *http.Request) {
	var input createInput
	var confirmData confirmMail
	redConn := u.S.Pool.Get()

	json.NewDecoder(r.Body).Decode(&input)

	newUser := models.User{
		Email:    input.Email,
		Password: input.Password,
	}

	if err := u.Db.Create(&newUser).Error; err != nil {
		responses.NewResponse(w, 400, fmt.Errorf("Could not create user: %s", err.Error()), nil)
		return
	}

	token := randstr.Hex(22)
	_, err := redConn.Do("SET", token, newUser.ID, "EX", 60*60*24*3, "NX")
	if err != nil {
		responses.NewResponse(w, 500, fmt.Errorf("Could not create user: %s", err.Error()), nil)
		return
	}

	confirmData.ConfirmURL = fmt.Sprintf("%s%s/%s", os.Getenv("APP_URL"), os.Getenv("APP_CONFIRM_PATH"), token)

	err = extras.SendEmail("Tony Andre Haugen <no_reply@tonyandre.co>", []string{input.Email}, "Just one more step to go ...", extras.ConfirmAccount, confirmData)
	if err != nil {
		responses.NewResponse(w, 500, fmt.Errorf("Could not send confirmation mail: %s", err.Error()), nil)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s%s%s", r.Host, "/user/", newUser.ID))
	responses.NewResponse(w, 201, nil, newUser)
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login is where we log the user in. We check if their email exists, verify password and see if they have confirmed
// their user account. If every check is ok we generate a cookie for them "sid" and save their session in redis
func (u *userResources) Login(w http.ResponseWriter, r *http.Request) {
	var input loginInput
	var user models.User

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&input)

	if err := u.Db.First(&user, "email = ?", input.Email).Error; err != nil {
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

// Logout is used for logging the user out byt removing their session and make their cookie "sid" blank.
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
