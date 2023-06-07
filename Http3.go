package main

import (
	"github.com/gorilla/handlers"
	_ "github.com/jackc/pgx/stdlib"
	"log"
	"main/db"
	"main/routing"
	"net/http"
	"os"
)

func main() {
	conn, errDB := db.NewDB()
	if errDB != nil {
		panic(errDB)
	}

	router := routing.NewRouter(routing.Route{DB: conn})
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Fatal(http.ListenAndServe(":5000", loggedRouter))
}
