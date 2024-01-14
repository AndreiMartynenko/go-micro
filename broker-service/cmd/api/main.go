package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "80"

type Config struct{}

func main() {
	app := Config{}

	log.Println("STarting broker service on port %s", webPort)

	// define http server
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
	}

	// start server

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}
