// Package rest GoPVZ REST API
//
// @title backend service
// @version 1.0.0
// @description Сервис для управления ПВЗ и приемкой товаров
// @host localhost:8080
// @BasePath /
//
// @securityDefinitions.apikey bearerAuth
// @in header
// @name Authorization
// @description JWT авторизация с Bearer схемой
package rest

import (
	"GoPVZ/internal/config"
	"GoPVZ/internal/transport/rest/handlers"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "GoPVZ/internal/transport/rest/docs"
)

func Run(cfg config.Config, log *slog.Logger, DBConn *sql.DB) {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, Go!")
	})

	mux.HandleFunc("/dummyLogin", handlers.DummyLoginHandler(log))
	mux.HandleFunc("/register", handlers.RegisterHandler(log, DBConn))
	mux.HandleFunc("/login", handlers.LoginHandler(log, DBConn))

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.REST.Port), mux)
	if err != nil {
		panic(err)
	}

}
