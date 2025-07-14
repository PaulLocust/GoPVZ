package middleware

import (
    "context"
    "net/http"
    "strings"
    "errors"
    "os"

    "github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(os.Getenv("JWT_SECRET"))

type key int

const (
    userCtxKey key = iota
)

type User struct {
    UserID int
    Role   string
}

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "authorization header required", http.StatusUnauthorized)
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "authorization header format must be Bearer {token}", http.StatusUnauthorized)
            return
        }
        tokenStr := parts[1]

        user, err := parseToken(tokenStr)
        if err != nil {
            http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), userCtxKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func parseToken(tokenStr string) (*User, error) {
    token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        // Проверяем алгоритм подписи
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return secretKey, nil
    })
    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, errors.New("token is invalid")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("cannot parse claims")
    }

    // Извлекаем user_id и role из claims
    userID, ok := claims["user_id"].(int) // jwt/v5 парсит числа в float64
    if !ok {
        return nil, errors.New("user_id claim missing or invalid")
    }
    role, ok := claims["role"].(string)
    if !ok {
        return nil, errors.New("role claim missing or invalid")
    }

    return &User{
        UserID: userID,
        Role:   role,
    }, nil
}

// Функция для получения пользователя из контекста
func GetUserFromContext(ctx context.Context) (*User, bool) {
    user, ok := ctx.Value(userCtxKey).(*User)
    return user, ok
}