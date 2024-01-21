package main

import (
	"authentication/cmd/api/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
)

const webPort = "80"

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting authentication service")

	// TODO connect to DB

	// set up config
	app := Config{}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Panic(err)

	}

}

// Creating a connection to DB

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Now we are going to add Posgres to our Docker composed.yml file.
// We need to make sure that it's available before we return the database connection,
// because this service might start out before the database does.

func connectTDB() *sql.DB {

	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")

		}
	}
}
