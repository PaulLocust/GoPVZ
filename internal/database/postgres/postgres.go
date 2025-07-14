package postgres

import (
	"GoPVZ/internal/config"
	"GoPVZ/internal/lib/sl"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

var DBConn *sql.DB

func ConnectToPostgresDB(cfg config.Config, log *slog.Logger) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)

	var err error
	DBConn, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error("error message", sl.Err(err))
		panic(err)
	}
	log.Info("Connected to database", slog.String("dbname", cfg.DB.Name))

	err = DBConn.Ping()
	if err != nil {
		log.Error("error message", sl.Err(err))
		panic(err)
	}

}
