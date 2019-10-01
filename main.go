package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"gitlab.com/t0nyandre/go-rest-boilerplate/api"
	"gitlab.com/t0nyandre/go-rest-boilerplate/db"
	"gitlab.com/t0nyandre/go-rest-boilerplate/models"
)

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		panic(err)
	}

	conn := db.NewDBConnection()
	conn.DropTableIfExists(models.User{})
	conn.AutoMigrate(models.User{})

	sess := db.NewStore()

	router := api.NewRouter(conn, sess)

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%s", os.Getenv("API_HOST"), os.Getenv("API_PORT")),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("Server running on http://%s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
