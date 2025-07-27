package app

import (
	"GoPVZ/config"
	_ "GoPVZ/docs" // для swagger
	domainAuthControllerHttp "GoPVZ/internal/auth/controller/http"
	userEntity "GoPVZ/internal/auth/entity"
	domainAuthRepo "GoPVZ/internal/auth/repo"
	domainAuthUsecase "GoPVZ/internal/auth/usecase"
	domainPVZControllerHttp "GoPVZ/internal/pvz/controller/http"
	domainPvzRepo "GoPVZ/internal/pvz/repo"
	domainPvzUsecase "GoPVZ/internal/pvz/usecase"
	"GoPVZ/pkg/pkgHttpserver"
	"GoPVZ/pkg/pkgLogger"
	"GoPVZ/pkg/pkgMetrics"
	"GoPVZ/pkg/pkgPostgres"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
// @description Вставьте JWT токен с префиксом 'Bearer '. Пример: Bearer eyJhbGciOiJIUzI1NiIs...
func Run(cfg *config.Config) {
	log := pkgLogger.New("local")
	log.Info("Starting application")

	DBConn, err := pkgPostgres.New(cfg.PGURL.URL)
	if err != nil {
		log.Error("Failed to connect to database", pkgLogger.Err(err))
		os.Exit(1)
	}
	log.Info("Connected to PostgreSQL")

	// Запуск сервера метрик Prometheus 
	go func() {
    	addr := ":" + cfg.Prometheus.Port // Используем порт из конфига
    	http.Handle("/metrics", promhttp.Handler())
    	log.Info("Prometheus metrics server started", slog.String("port", cfg.Prometheus.Port))
		
    	if err := http.ListenAndServe(addr, nil); err != nil {
    	    log.Error("Metrics server error", pkgLogger.Err(err))
    	}
	}()

	// auth domain
	userRepo := domainAuthRepo.NewUserRepo(DBConn.Pool)
	jwtManager := domainAuthUsecase.NewJwtManager(cfg.JWT.Secret, 24*time.Hour)
	authUC := domainAuthUsecase.NewAuthUseCase(userRepo, jwtManager)

	// pvz domain
	pvzRepo := domainPvzRepo.NewPVZRepo(DBConn.Pool)
	pvzUC := domainPvzUsecase.NewPVZUseCase(pvzRepo)

	// Создаем middleware
	authMiddleware := domainAuthControllerHttp.JWTMiddleware(authUC.GetJwtManager())
	employeeOnly := domainAuthControllerHttp.RolesMiddleware(userEntity.RoleEmployee)
	moderatorOnly := domainAuthControllerHttp.RolesMiddleware(userEntity.RoleModerator)
	employeeOrModerator := domainAuthControllerHttp.RolesMiddleware(userEntity.RoleEmployee, userEntity.RoleModerator)

	server := pkgHttpserver.New(
		pkgHttpserver.Port(cfg.HTTP.Port),
		pkgHttpserver.ReadTimeout(10*time.Second),
	)
	router := server.GetRouter()

	// Добавляем middleware для метрик
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()

		pkgMetrics.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		pkgMetrics.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
		).Observe(time.Since(start).Seconds())
	})

	registerRoutes(router, authUC, pvzUC, authMiddleware, employeeOnly, moderatorOnly, employeeOrModerator)
	server.Start()
	log.Info("Server started on port ", slog.String("port", cfg.HTTP.Port))
	waitForShutdown(server, log)
}


func registerRoutes(
	router *gin.Engine,
	authUC *domainAuthUsecase.AuthUseCase,
	pvzUC *domainPvzUsecase.PVZUseCase,
	authMiddleware gin.HandlerFunc,
	employeeOnly gin.HandlerFunc,
	moderatorOnly gin.HandlerFunc,
	employeeOrModerator gin.HandlerFunc,
) {
	api := router.Group("/")

	// Auth routes (public)
	domainAuthControllerHttp.NewAuthRouter(api, authUC)

	// PVZ routes (protected)
	domainPVZControllerHttp.NewPVZRouter(api, pvzUC, authMiddleware, employeeOnly, moderatorOnly, employeeOrModerator)

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
