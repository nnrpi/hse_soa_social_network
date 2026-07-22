#!/bin/bash
set -e

mkdir -p posts-api-service/proto
mkdir -p posts-grpc-service
mkdir -p bin

chmod +x build_posts_services.sh

echo "Setup completed successfully!" 