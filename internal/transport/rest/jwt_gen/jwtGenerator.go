package jwt_gen

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


var secretKey = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT(role string) (string, error) {
	claims := jwt.MapClaims{
		"role": role,                                  // "employee" или "moderator"
		"exp":  time.Now().Add(24 * time.Hour).Unix(), // Срок действия
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
