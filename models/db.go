package models

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
)

// Db is the global database connection used throughout the whole application
var Db *gorm.DB
var err error

// ModelID is a nanoID or UUID
type ModelID struct {
	ID string `json:"_id" gorm:"primary_key"`
}

// Timestamp for all the models
type Timestamp struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"null"`
}

// Connect to the database and store the gorm.DB type to the global variable Db
func Connect() *gorm.DB {
	var sslmode string
	switch os.Getenv("API_ENV") {
	case "dev":
		sslmode = "disable"
		break
	case "test":
		sslmode = "disable"
		break
	default:
		sslmode = "require"
	}

	Db, err = gorm.Open("postgres", fmt.Sprintf("host=%s port=%v user=%s dbname=%s password=%s sslmode=%s",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PASSWORD"),
		sslmode,
	))

	if err != nil {
		panic(err)
	}

	if os.Getenv("API_ENV") == "dev" {
		Db.LogMode(true)
	}

	return Db
}

// TestConnection will connect the database and redis store to their respective global variables.
// It will give a log feedback if successful or failure
func TestConnection() {
	Connect()
	Db.DropTableIfExists(&User{})
	Db.AutoMigrate(&User{})
	log.Println("Connection to database was a success!")
}
