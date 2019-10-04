package responses

import (
	"encoding/json"
	"net/http"
)

// Response is the standardized response the users will get from this API. Error, if any and response body
type Response struct {
	Error *Error       `json:"error,omitempty"`
	Body  *interface{} `json:"data,omitempty"`
}

// Error structures how I want the errors to be shown
type Error struct {
	Message string `json:"message,omitempty"`
}

// CustomResponse will generate a custom message when the http request is successful but no
// data is given. One example is the logout endpoint
type CustomResponse struct {
	Message string `json:"message"`
}

// NewResponse is standardizing the json response provided by this API
func NewResponse(w http.ResponseWriter, statusCode int, err error, data interface{}) {
	w.WriteHeader(statusCode)
	res := Response{}
	if err != nil {
		errors := Error{
			Message: err.Error(),
		}
		res.Error = &errors
	}
	if data != nil {
		res.Body = &data
	}
	json.NewEncoder(w).Encode(&res)
}
