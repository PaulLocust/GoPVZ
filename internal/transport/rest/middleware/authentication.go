package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const RoleKey contextKey = "userRole"
const UserIDKey contextKey = "userID"

// Ключ для подписи
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Функция для извлечения и проверки JWT из заголовка
func JWTAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "missing Authorization header", http.StatusUnauthorized)
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            http.Error(w, "invalid Authorization header format", http.StatusUnauthorized)
            return
        }

        tokenString := parts[1]

        // Парсим токен
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Проверяем метод подписи
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            http.Error(w, "invalid token claims", http.StatusUnauthorized)
            return
        }

        // Извлекаем роль и user_id из claims
        role, ok := claims["role"].(string)
        if !ok {
            http.Error(w, "role claim missing", http.StatusUnauthorized)
            return
        }

        userID, _ := claims["user_id"].(string) // если нужно

        // Кладём в контекст
        ctx := context.WithValue(r.Context(), RoleKey, role)
        ctx = context.WithValue(ctx, UserIDKey, userID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Middleware для проверки роли, как в предыдущем ответе
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
    allowed := map[string]bool{}
    for _, r := range allowedRoles {
        allowed[strings.ToLower(r)] = true
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            role, ok := r.Context().Value(RoleKey).(string)
            if !ok || !allowed[strings.ToLower(role)] {
                http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}