package models

import (
	"github.com/aidarkhanov/nanoid"
	"github.com/alexedwards/argon2id"
	"github.com/jinzhu/gorm"
)

// UserRole has all the account roles for this application
type UserRole string

const (
	// Admin has full control over the access system
	Admin UserRole = "admin"
	// Pro is members that support me in any kind of way
	Pro UserRole = "pro"
	// Contributer is members that helps the community out in any way
	Contributer UserRole = "contributer"
	// Member are allowed to post comments and get access to member only things
	Member UserRole = "member"
)

// User model which structures the database table and how it's represented as json
type User struct {
	ModelID
	Username  string `json:"username,omitempty" gorm:"type:varchar(100);not null;unique"`
	Email     string `json:"email,omitempty" gorm:"type:varchar(100);not null;unique_index"`
	Password  string `json:"-" gorm:"not null"`
	Role      string `json:"role,omitempty" gorm:"default:'member'"`
	Confirmed bool   `json:"confirmed,omitempty" gorm:"default:false"`
	Timestamp
}

// BeforeCreate will hash the password and create an ID for the User
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
