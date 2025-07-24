package usecase

import (
	"GoPVZ/internal/auth/entity"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtManager struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewJwtManager(secret string, duration time.Duration) *JwtManager {
	return &JwtManager{secretKey: secret, tokenDuration: duration}
}

func (jm *JwtManager) GenerateToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    string(user.Role),
		"exp":     time.Now().Add(jm.tokenDuration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jm.secretKey))
}

func (jm *JwtManager) VerifyToken(tokenStr string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jm.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
