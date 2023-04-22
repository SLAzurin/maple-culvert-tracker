package db

import (
	"database/sql"
	"net/url"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func init() {
	var err error
	var connStr = "postgresql://" + os.Getenv("POSTGRES_USER") + ":" + url.QueryEscape(os.Getenv("POSTGRES_PASSWORD")) + "@" + os.Getenv("CLIENT_POSTGRES_HOST") + ":" + os.Getenv("CLIENT_POSTGRES_PORT") + "/" + os.Getenv("POSTGRES_DB") + "?sslmode=disable"
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
}
