package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"rest/marketgrpc"
	"rest/protobuf"
	"rest/rest"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL env variable is not set")
	}

	// main.go now owns the connection pool.
	// Both stores share this single pool — no duplicate connections.
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	// Both stores receive the same *sql.DB pool
	foodStore := rest.NewPostgresFoodStore(db)
	userStore := rest.NewPostgresUserStore(db)

	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen on port 50051: %v", err)
		}

		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(marketgrpc.LoggingInterceptor),
		)

		protobuf.RegisterMarketServiceServer(grpcServer, marketgrpc.NewMarketServer(foodStore))
		reflection.Register(grpcServer)

		log.Println("gRPC server running on :50051")

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// NewMarket now receives both stores
	server := rest.NewMarket(foodStore, userStore)

	log.Println("HTTP server running on :8080")

	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
