FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN cd posts-api-service && go build -o /posts-api-service

FROM alpine:latest

WORKDIR /app
COPY --from=builder /posts-api-service .

EXPOSE 8090
CMD ["./posts-api-service"] 