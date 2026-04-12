package main

import (
	"log"
	"net/http"
	"os"
	"rest/rest"
)

func main() {
	// read the DSN from the environment — injected by Docker Compose via .env
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL env variable is not set")
	}

	// open a connection pool to Postgres and verify it's reachable via Ping
	store, err := rest.NewPostgresFoodStore(dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	// wire the store into the HTTP router
	server := rest.NewMarket(store)

	log.Println("Food store API is running on :8080")

	// ListenAndServe blocks until the server crashes or is shut down
	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
