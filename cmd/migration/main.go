package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		slog.Warn("couldn't load .env")
	}
	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		os.Getenv("DB_LOGIN"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_ADDRESS"),
		5432,
		os.Getenv("DB_NAME")))
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	_ = goose.SetDialect("postgres")

	if err := goose.Up(db, "./db/migrations"); err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}
