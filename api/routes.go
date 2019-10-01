package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	redisstore "gopkg.in/boj/redistore.v1"
)

type Server struct {
	Db *gorm.DB
	S  *redisstore.RediStore
}

func NewRouter(db *gorm.DB, sess *redisstore.RediStore) *mux.Router {
	s := &Server{Db: db, S: sess}
	r := mux.NewRouter()
	r.Use(commonMiddleware)
	r.HandleFunc("/", index).Methods("GET")
	ServeUserRoutes(s.Db, s.S, r)
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
