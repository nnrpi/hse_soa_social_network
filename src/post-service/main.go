package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "social-network/post-service/proto/post-service/pb"
	"social-network/post-service/repository"
	"social-network/post-service/service"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "social_network")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	postRepo := repository.NewPostRepository(db)
	err = postRepo.Init()
	if err != nil {
		log.Fatalf("Failed to initialize post repository: %v", err)
	}

	promoRepo := repository.NewPromoRepository(db)
	err = promoRepo.Init()
	if err != nil {
		log.Fatalf("Failed to initialize promo repository: %v", err)
	}

	port := getEnv("PORT", "9000")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterPostServiceServer(server, service.NewPostServiceServer(postRepo, promoRepo))
	reflection.Register(server)

	log.Printf("Post service starting on :%s...", port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
