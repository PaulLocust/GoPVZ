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
package rest

import (
	"GoPVZ/internal/config"
	_ "GoPVZ/internal/transport/rest/docs"
	"GoPVZ/internal/transport/rest/handlers"
	"GoPVZ/internal/transport/rest/helpers"

	"GoPVZ/internal/transport/rest/middleware"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	moderator = "moderator"
	employee  = "employee"
)

func Run(cfg config.Config, log *slog.Logger, DBConn *sql.DB) {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, Go!")
	})

	mux.HandleFunc("/dummyLogin", handlers.DummyLoginHandler(log))
	mux.HandleFunc("/register", handlers.RegisterHandler(log, DBConn))
	mux.HandleFunc("/login", handlers.LoginHandler(log, DBConn))

	//mux.HandleFunc("/pvz", middleware.JWTAuthMiddleware(log, moderator)(handlers.PVZHandler(log, DBConn)))

	mux.HandleFunc("/receptions", middleware.JWTAuthMiddleware(log, employee)(handlers.ReceptionHandler(log, DBConn)))
	mux.HandleFunc("/products", middleware.JWTAuthMiddleware(log, employee)(handlers.ProductHandler(log, DBConn)))
	mux.HandleFunc("/pvz/{pvzId}/delete_last_product", middleware.JWTAuthMiddleware(log, employee)(handlers.DeleteLastProductHandler(log, DBConn)))
	mux.HandleFunc("/pvz/{pvzId}/close_last_reception", middleware.JWTAuthMiddleware(log, employee)(handlers.CloseLastReceptionHandler(log, DBConn)))
	//mux.HandleFunc("/pvz/list", middleware.JWTAuthMiddleware(log, employee, moderator)(handlers.GetPVZListHandler(log, DBConn)))

	mux.HandleFunc("/pvz", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			middleware.JWTAuthMiddleware(log, moderator)(handlers.PVZHandler(log, DBConn)).ServeHTTP(w, r)
		case http.MethodGet:
			middleware.JWTAuthMiddleware(log, employee, moderator)(handlers.GetPVZListHandler(log, DBConn)).ServeHTTP(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			helpers.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Info("Starting REST server", slog.Int("port", cfg.REST.Port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.REST.Port), mux)
	if err != nil {
		panic(err)
	}

}
