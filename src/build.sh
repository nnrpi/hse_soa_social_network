#!/bin/bash
podman compose down
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/user-service ./user-service/main.go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/post-service ./post-service/main.go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/proxy-service ./proxy-service/main.go
chmod +x ./bin/user-service ./bin/post-service ./bin/proxy-service
podman compose up --build