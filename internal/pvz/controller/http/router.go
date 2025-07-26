package http

import (
	"GoPVZ/internal/pvz/usecase"

	"github.com/gin-gonic/gin"
)

func NewPVZRouter(
    router *gin.RouterGroup,
    uc *usecase.PVZUseCase,
    authMiddleware gin.HandlerFunc,
    employeeOnly gin.HandlerFunc,
    moderatorOnly gin.HandlerFunc,
    employeeOrModerator gin.HandlerFunc,
) {
    handler := NewPVZHandler(uc)
    
    // Protected routes group
    protected := router.Group("/")
    protected.Use(authMiddleware)
    
    // Routes for employees only
    employeeRoutes := protected.Group("/")
    employeeRoutes.Use(employeeOnly)
	employeeRoutes.POST("/receptions", handler.CreateReception)
	employeeRoutes.POST("/products", handler.CreateProduct)
	employeeRoutes.POST("/pvz/:pvzId/delete_last_product", handler.DeleteLastProduct)
	employeeRoutes.POST("/pvz/:pvzId/close_last_reception", handler.CloseReception)
    
    // Routes for moderators only
    moderatorRoutes := protected.Group("/")
    moderatorRoutes.Use(moderatorOnly)
	moderatorRoutes.POST("/pvz", handler.CreatePVZ)
    
    // Routes for both employees and moderators
    commonRoutes := protected.Group("/")
    commonRoutes.Use(employeeOrModerator)
    commonRoutes.GET("/pvz", handler.GetPVZsWithReceptions)
}