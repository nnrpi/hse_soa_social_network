FROM --platform=linux/amd64 golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o proxy-service ./proxy-service/main.go
FROM --platform=linux/amd64 alpine:latest
RUN apk --no-cache add ca-certificates file
WORKDIR /app
COPY --from=builder /app/proxy-service .
RUN /bin/sh -c "echo 'Binary architecture:' && file /app/proxy-service"
RUN chmod +x /app/proxy-service
CMD ["/app/proxy-service"]
