package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type Server struct {
	Db *gorm.DB
}

func NewRouter(db *gorm.DB) *mux.Router {
	s := &Server{Db: db}
	r := mux.NewRouter()
	r.Use(commonMiddleware)
	r.HandleFunc("/", index).Methods("GET")
	ServeUserRoutes(s.Db, r)
	return r
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HELLO WORLD!!!"))
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
