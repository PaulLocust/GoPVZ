package rest

import (
	"GoPVZ/internal/config"
	"GoPVZ/internal/transport/rest/handlers"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
)

func Run(cfg config.Config, log *slog.Logger, DBConn *sql.DB) {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, Go!")
	})

	mux.HandleFunc("/dummyLogin", handlers.DummyLoginHandler(log))
	mux.HandleFunc("/register", handlers.RegisterHandler(log, DBConn))

	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.REST.Port), mux)
	if err != nil {
		panic(err)
	}

}
