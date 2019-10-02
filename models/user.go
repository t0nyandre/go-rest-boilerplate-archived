package models

import (
	"github.com/aidarkhanov/nanoid"
	"github.com/alexedwards/argon2id"
	"github.com/jinzhu/gorm"
	"gitlab.com/t0nyandre/go-rest-boilerplate/db"
)

// User model
type User struct {
	db.ModelID
	Email     string `json:"email" gorm:"type:varchar(100);not null;unique_index"`
	Password  string `json:"-" gorm:"not null"`
	Role      string `json:"role" gorm:"default:'Member'"`
	Confirmed bool   `json:"confirmed" gorm:"default:false"`
	db.Timestamp
}

// BeforeSave will hash the password with Argon2ID algorithm
func (user *User) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("ID", nanoid.New())
	scope.SetColumn("Password", user.HashPassword())
	return nil
}

// HashPassword will take a plain text password and hash it with argon2id algorithm
func (user *User) HashPassword() string {
	customParams := argon2id.Params{
		Iterations:  3,
		Memory:      4096,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := argon2id.CreateHash(user.Password, &customParams)
	if err != nil {
		panic(err.Error())
	}

	return hash
}

// VerifyPassword is comparing the plain text password to the hased password in the database
func (user *User) VerifyPassword(password string) bool {
	match, err := argon2id.ComparePasswordAndHash(password, user.Password)
	if err != nil {
		return false
	}

	return match
}

// UserConfirmed checks if user has confrimed his account
func (user *User) UserConfirmed() bool {
	return user.Confirmed
}
