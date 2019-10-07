package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aidarkhanov/nanoid"
	"github.com/gorilla/mux"
	"github.com/thanhpk/randstr"
	"gitlab.com/t0nyandre/go-rest-boilerplate/email"
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
	r.HandleFunc("/users/confirm/{token}", res.Confirm).Methods("GET")
	r.HandleFunc("/users/login", res.Login).Methods("POST")
	r.HandleFunc("/users/logout", res.Logout).Methods("POST")
	r.HandleFunc("/users/register", res.Register).Methods("POST")
	r.HandleFunc("/users/me", middleware.AuthRequired(res.Me)).Methods("GET")
}

func (res *authResources) Confirm(w http.ResponseWriter, req *http.Request) {
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

	res.loginUser(w, user)
	loginRes := loginResponse{&user, utils.GenerateAccessToken(user)}

	w.Header().Set("Location", fmt.Sprintf("%s%s%s", req.Host, "/users/", user.Username))
	responses.NewResponse(w, 201, nil, &loginRes)
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	*models.User
	Token string `json:"access_token,omitempty"`
}

// Login is where we log the user in. We check if their email exists, verify password and see if they have confirmed
// their user account. If every check is ok we generate a cookie for them "sid" and save their session in redis
func (res *authResources) Login(w http.ResponseWriter, req *http.Request) {
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

	res.loginUser(w, user)
	loginRes := loginResponse{&user, utils.GenerateAccessToken(user)}

	w.Header().Set("Location", fmt.Sprintf("%s%s%s", req.Host, "/users/", user.Username))
	responses.NewResponse(w, 200, nil, &loginRes)
}

func (res *authResources) loginUser(w http.ResponseWriter, user models.User) {
	var secure bool

	if !user.UserConfirmed() {
		responses.NewResponse(w, 401, fmt.Errorf("You haven't confirmed your account yet. Please check your email and click the link for confirmation"), nil)
		return
	}

	if user.UserLocked().Locked {
		responses.NewResponse(w, 401, fmt.Errorf("Account is locked: %s", user.UserLocked().Reason), nil)
		return
	}

	jid := nanoid.New()
	if _, err := utils.Store.Set(fmt.Sprintf("%s%s", extras.RefreshTokenPrefix, jid), utils.GenerateRefreshToken(user), time.Second*60*60*24*30).Result(); err != nil {
		fmt.Println(err.Error())
	}

	if os.Getenv("API_ENV") == "dev" {
		secure = false
	} else {
		secure = true
	}

	cookie := http.Cookie{
		Name:     "jid",
		Value:    jid,
		Path:     "/refresh-token",
		HttpOnly: true,
		MaxAge:   60 * 60 * 24 * 30, // 30 days
		Secure:   secure,            // Change this to true on production
	}
	http.SetCookie(w, &cookie)
}

type emailData struct {
	ConfirmURL string
}

// Logout is used for logging the user out byt removing their session and make their cookie "sid" blank.
func (res *authResources) Logout(w http.ResponseWriter, req *http.Request) {
	res.removeRefreshToken(w)

	responses.NewResponse(w, 200, nil, nil)
}

func (res *authResources) removeRefreshToken(w http.ResponseWriter) {
	var secure bool

	if os.Getenv("API_ENV") == "dev" {
		secure = false
	} else {
		secure = true
	}

	delCookie := http.Cookie{
		Name:     "jid",
		Value:    "",
		HttpOnly: true,
		Path:     "/refresh-token",
		MaxAge:   -1,
		Secure:   secure,
	}

	http.SetCookie(w, &delCookie)
}

type createInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Create account if the email hasn't been used before.
// Sending our an email with a confirmation token so they can confirm the account.
func (res *authResources) Register(w http.ResponseWriter, req *http.Request) {
	var input createInput
	var confirmData emailData

	json.NewDecoder(req.Body).Decode(&input)

	newUser := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
	}

	if os.Getenv("API_ADMIN_USER") == input.Email {
		newUser.Role = string(models.Admin)
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

	w.Header().Set("Location", fmt.Sprintf("%s%s%s", req.Host, "/users/", newUser.Username))
	responses.NewResponse(w, 201, nil, newUser)
}

func (res *authResources) Me(w http.ResponseWriter, req *http.Request) {
	var user models.User

	payload := req.Context().Value(middleware.LoggedInUserCtx).(utils.TokenPayload)

	if err := models.Db.First(&user, "id = ?", payload.UserID).Error; err != nil {
		responses.NewResponse(w, 401, fmt.Errorf("%s", extras.AccessDenied), nil)
		return
	}

	responses.NewResponse(w, 200, nil, user)
}
