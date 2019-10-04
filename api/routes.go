package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/t0nyandre/go-rest-boilerplate/middleware"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Use(middleware.HeaderMiddleware)
	r.HandleFunc("/", index).Methods("GET")
	ServeUserRoutes(r)
	ServeAuthRoutes(r)
	return r
}

func index(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("HELLO WORLD!!!"))
}
