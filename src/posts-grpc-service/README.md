# Posts gRPC Service

This service provides the core functionality for working with posts in the social network through a gRPC API. It handles data storage in PostgreSQL and produces events to Kafka.

## Features

- Core post functionality (view, like, comment)
- Database interactions with PostgreSQL
- Event production to Kafka

## gRPC Methods

- `GetPost` - Get a post by ID
- `LikePost` - Like a post
- `CommentPost` - Add a comment to a post
- `GetPostComments` - Get comments for a post with pagination

## Data Models

- Post
- Comment
- Like
- View

## Kafka Events

The service produces the following events to Kafka topics:
- Post views (topic: post_views)
- Post likes (topic: post_interactions)
- Post comments (topic: post_comments)

## Environment Variables

- `PORT` - Port for the gRPC server (default: 9090)
- `DB_HOST` - PostgreSQL host
- `DB_PORT` - PostgreSQL port
- `DB_USER` - PostgreSQL username
- `DB_PASSWORD` - PostgreSQL password
- `DB_NAME` - PostgreSQL database name
- `KAFKA_BROKER` - Kafka broker address (default: kafka:9092) 