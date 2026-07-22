#!/bin/bash
set -e

echo "Generating proto code..."
mkdir -p posts-api-service/proto
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  posts-api-service/proto/posts.proto

echo "Building posts-api-service..."
cd posts-api-service
go build -o ../bin/posts-api-service .
cd ..

echo "Building posts-grpc-service..."
cd posts-grpc-service
go build -o ../bin/posts-grpc-service .
cd ..

echo "Done!" 