package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthClaims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func CreateNewAuthToken(id string, username string) (string, error) {
	claims := &AuthClaims{
		Id:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			Issuer:    "CMD",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey, exist := os.LookupEnv("SECRET_KEY")
	if !exist {
		panic("SECRET_KEY not found")
	}

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", errors.New("failed to sign token")
	}

	return signedToken, nil
}
