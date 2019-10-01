package responses

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Error *Error       `json:"error,omitempty"`
	Body  *interface{} `json:"data,omitempty"`
}

type Error struct {
	Message string `json:"message,omitempty"`
}

type CustomResponse struct {
	Message string `json:"message"`
}

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
