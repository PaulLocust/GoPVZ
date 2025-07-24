package app

import (
	"GoPVZ/config"
	_ "GoPVZ/docs" // для swagger
	"GoPVZ/internal/auth/controller/http"
	"GoPVZ/internal/auth/repo"
	"GoPVZ/internal/auth/usecase"
	"GoPVZ/pkg/pkgLogger"
	"GoPVZ/pkg/pkgPostgres"
	"GoPVZ/pkg/pkgHttpserver"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"

	ginSwagger "github.com/swaggo/gin-swagger"
)

// Package rest GoPVZ REST API
//
// @title Backend service GoPVZ
// @version 1.0.0
// @description Сервис для управления ПВЗ и приемкой товаров
// @host localhost:8080
// @BasePath /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT авторизация с Bearer схемой
func Run(cfg *config.Config) {

	log := pkgLogger.New("local")
	log.Info("Starting application", slog.Any("config", cfg))

	DBConn, err := pkgPostgres.New(cfg.PGURL.URL)
	if err != nil {
		log.Error("Failed to connect to database", pkgLogger.Err(err))
		os.Exit(1)
	}
	log.Info("Connected to PostgreSQL")

	// auth
	jwtManager := usecase.NewJwtManager(cfg.JWT.Secret, 24*time.Hour)
	userRepo := repo.NewUserRepo(DBConn.Pool)
	authUC := usecase.NewAuthUseCase(userRepo, jwtManager)

	server := pkgHttpserver.New(
		pkgHttpserver.Port(cfg.HTTP.Port),
		pkgHttpserver.ReadTimeout(10*time.Second),
	)
	router := server.GetRouter()

	registerRoutes(router, authUC)
	server.Start()
	log.Info("Server started on port " + cfg.HTTP.Port)
	waitForShutdown(server, log)
}

func registerRoutes(router *gin.Engine, authUC *usecase.AuthUseCase) {
	api := router.Group("/")

	// Auth
	http.NewAuthRouter(api, authUC)

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})
}

func waitForShutdown(server *pkgHttpserver.Server, log *pkgLogger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	if err := server.Shutdown(); err != nil {
		log.Error("Error during shutdown", pkgLogger.Err(err))
	} else {
		log.Info("Server exited gracefully")
	}
}
