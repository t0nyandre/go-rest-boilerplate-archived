package utils

import (
	"fmt"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/t0nyandre/go-rest-boilerplate/extras"
	"github.com/t0nyandre/go-rest-boilerplate/models"
)

// AccessTokenClaims enough to validate the user and give the user his access
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"user_role"`
	jwt.StandardClaims
}

type TokenPayload struct {
	UserID string
	Role   string
}

// GenerateAccessToken with an ID and a Role. The Access token will only be valid for 10 minutes
func GenerateAccessToken(user models.User) string {
	// Create the Claims
	claims := AccessTokenClaims{
		user.ID,
		user.Role,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + (60 * 10), // 10 minutes
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	ss, _ := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))
	return ss
}

func ValidateAccessToken(tokenString string) (interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signingmethod")
		}

		return []byte(os.Getenv("ACCESS_TOKEN_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		payload := TokenPayload{
			UserID: claims["user_id"].(string),
			Role:   claims["user_role"].(string),
		}
		return payload, nil
	}
	return nil, fmt.Errorf(string(extras.BadTokenError))
}

type RefreshTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func GenerateRefreshToken(user models.User) string {
	// Create the Claims
	claims := RefreshTokenClaims{
		user.ID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + (60 * 60 * 24 * 30), // 30 days
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	ss, _ := token.SignedString([]byte(os.Getenv("REFRESH_TOKEN_SECRET")))
	return ss
}

func ValidateRefreshToken(tokenString string) (interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signingmethod")
		}

		return []byte(os.Getenv("REFRESH_TOKEN_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(string)
		return userID, nil
	}

	return nil, fmt.Errorf("Invalid refresh token")
}
