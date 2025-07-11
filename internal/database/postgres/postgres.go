package postgres

import (
	"GoPVZ/internal/config"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DBConn *sql.DB

func ConnectToPostgresDB(cfg config.Config) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)
	
	var err error
	DBConn, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer DBConn.Close()

	err = DBConn.Ping()
	if err != nil {
		panic(err)
	}

}
