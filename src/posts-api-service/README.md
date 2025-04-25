# Posts API Service

This service provides a REST API for interacting with posts in the social network. It serves as a proxy to the gRPC-based posts-grpc-service, which contains the core business logic.

## Features

- View a post
- Like a post
- Comment on a post
- Get comments for a post with pagination

## API Endpoints

- `GET /posts/:id` - Get a post by ID
- `POST /posts/:id/like` - Like a post
- `POST /posts/:id/comments` - Add a comment to a post
- `GET /posts/:id/comments` - Get all comments for a post with pagination

## Kafka Integration

The service produces events to Kafka topics:
- User registrations
- Post views
- Post likes
- Post comments

## Architecture

This service follows a layered architecture pattern:
1. REST API (in this service)
2. gRPC API (in posts-grpc-service)
3. Data storage (PostgreSQL)
4. Event streaming (Kafka)

## Environment Variables

- `PORT` - Port for the REST API server (default: 8090)
- `POSTS_GRPC_SERVICE_URL` - URL of the gRPC service (default: posts-grpc-service:9090) 