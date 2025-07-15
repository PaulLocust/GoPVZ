package middleware

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/helpers"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(os.Getenv("JWT_SECRET"))

func JWTAuthMiddleware(log *slog.Logger, requiredRoles ...string) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				helpers.WriteJSONError(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenStr := authHeader

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrTokenSignatureInvalid
				}
				return secretKey, nil
			})

			if err != nil {
				log.Error("error message", sl.Err(err))
				helpers.WriteJSONError(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				log.Error("error message: Invalid token")
				helpers.WriteJSONError(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				helpers.WriteJSONError(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Проверка срока действия через GetExpirationTime()
			exp, err := claims.GetExpirationTime()
			if err != nil || exp == nil || exp.Before(time.Now()) {
				log.Error("error message", sl.Err(err))
				helpers.WriteJSONError(w, "Token expired", http.StatusUnauthorized)
				return
			}

			roleVal, ok := claims["role"].(string)
			if !ok || roleVal == "" {
				helpers.WriteJSONError(w, "Role claim missing or invalid", http.StatusUnauthorized)
				return
			}

			allowed := false
			for _, r := range requiredRoles {
				if roleVal == r {
					allowed = true
					break
				}
			}
			if !allowed {
				helpers.WriteJSONError(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
