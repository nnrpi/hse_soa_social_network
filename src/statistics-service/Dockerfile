FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN cd statistics-service && go build -o /statistics-service

FROM alpine:latest

WORKDIR /app
COPY --from=builder /statistics-service .

EXPOSE 8095
CMD ["./statistics-service"]
