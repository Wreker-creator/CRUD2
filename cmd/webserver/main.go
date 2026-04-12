package main

import (
	"log"
	"net/http"
	"os"
	"rest/rest"
)

func main() {

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Database_url env variable is not set")
	}

	store, err := rest.NewPostgresFoodStore(dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database :%v", err)
	}

	server := rest.NewMarket(store)

	log.Println("Food store API IS running on 8080")

	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatalf("server failed to start, : %v", err)
	}

}
