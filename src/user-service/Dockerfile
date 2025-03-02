
FROM --platform=linux/amd64 golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o user-service ./user-service/main.go
FROM --platform=linux/amd64 alpine:latest
RUN apk --no-cache add ca-certificates file
WORKDIR /app
COPY --from=builder /app/user-service .
RUN /bin/sh -c "echo 'Binary architecture:' && file /app/user-service"
RUN chmod +x /app/user-service
CMD ["/app/user-service"]
