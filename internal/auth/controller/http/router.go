package http

import (
	"github.com/gin-gonic/gin"
	"GoPVZ/internal/auth/usecase"
)

func NewAuthRouter(router *gin.RouterGroup, uc *usecase.AuthUseCase) {
	handler := NewAuthHandler(uc)

	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
	router.POST("/dummyLogin", handler.DummyLogin)

	// Пример защищенного маршрута с JWT и ролями
	//protected := router.Group("/")
	//protected.Use(JWTMiddleware(uc.GetJwtManager())) // JWT middleware
	//protected.Use(RolesMiddleware(/* разрешенные роли */))

}