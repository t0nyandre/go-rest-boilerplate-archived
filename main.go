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
	"gitlab.com/t0nyandre/go-rest-boilerplate/models"
	"gitlab.com/t0nyandre/go-rest-boilerplate/utils"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	// Establish connection to Database
	models.TestConnection()
	// Establish connection to Redis
	utils.ConnectRedis()

	// Generate all the routes
	router := api.NewRouter()

	server := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%s", os.Getenv("API_HOST"), os.Getenv("API_PORT")),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("Server running on http://%s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
