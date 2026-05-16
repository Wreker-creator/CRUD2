package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"rest/marketgrpc"
	"rest/protobuf"
	"rest/rest"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	// gRPC server runs in a goroutine so it doesn't block the HTTP server below.
	// both servers share the same store, so a single DB connection pool serves both.
	go func() {
		// gRPC runs over raw TCP, not HTTP — so we open a TCP listener directly
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen on port 50051: %v", err)
		}

		// create a bare gRPC server with no middleware or options for now
		grpcServer := grpc.NewServer()

		// register our MarketServer implementation against the generated service descriptor.
		// this is how gRPC knows to route incoming RPCs to our handler methods
		protobuf.RegisterMarketServiceServer(grpcServer, marketgrpc.NewMarketServer(store))

		// the reason for adding this is simple, grpcurl has no way to know what methods our server has
		// or what messsages it expects, reflection regiters the service's schema with the grpc serve at runtime
		// so tooling like grpcurl can discover and call it wihout us manually describing the schema on the command line
		reflection.Register(grpcServer)

		log.Println("gRPC server running on :50051")

		// Serve blocks, same as http.ListenAndServe
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// wire the store into the HTTP router
	server := rest.NewMarket(store)

	log.Println("HTTP server running on :8080")

	// ListenAndServe blocks until the server crashes or is shut down
	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
