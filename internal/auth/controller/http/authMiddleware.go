package http

import (
    "slices"
	"net/http"
    "strings"
    "GoPVZ/internal/auth/usecase"
    "GoPVZ/internal/auth/entity"
    "github.com/gin-gonic/gin"
)

func JWTMiddleware(jm *usecase.JwtManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        auth := c.GetHeader("Authorization")
        parts := strings.SplitN(auth, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "missing or invalid auth header"})
            return
        }
        claims, err := jm.VerifyToken(parts[1])
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
            return
        }

        c.Set("user_id", (*claims)["user_id"])
        c.Set("user_email", (*claims)["email"])
        c.Set("user_role", (*claims)["role"])
        c.Next()
    }
}

func RolesMiddleware(allowedRoles ...entity.Role) gin.HandlerFunc {
    return func(c *gin.Context) {
        roleVal, _ := c.Get("user_role")
        role := entity.Role(roleVal.(string))
        if slices.Contains(allowedRoles, role) {
                c.Next()
                return
            }
        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "forbidden"})
    }
}