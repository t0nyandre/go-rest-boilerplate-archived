package db

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
)

type ModelID struct {
	ID string `json:"_id" gorm:"primary_key"`
}

type Timestamp struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"null"`
}

func NewDBConnection() *gorm.DB {
	var sslmode string
	if os.Getenv("API_ENV") == "dev" {
		sslmode = "disable"
	} else {
		sslmode = "require"
	}

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%v user=%s dbname=%s password=%s sslmode=%s",
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

	return db
}
