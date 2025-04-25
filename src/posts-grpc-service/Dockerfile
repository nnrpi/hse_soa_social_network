FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN cd posts-grpc-service && go build -o /posts-grpc-service

FROM alpine:latest

WORKDIR /app
COPY --from=builder /posts-grpc-service .

EXPOSE 9090
CMD ["./posts-grpc-service"] 