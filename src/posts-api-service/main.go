package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	r := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	grpcServiceURL := os.Getenv("POSTS_GRPC_SERVICE_URL")
	if grpcServiceURL == "" {
		grpcServiceURL = "posts-grpc-service:9090"
	}

	conn, err := grpc.Dial(grpcServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC service: %v", err)
	}
	defer conn.Close()

	client := NewPostsServiceClient(conn)

	r.GET("/posts/:id", func(c *gin.Context) {
		id := c.Param("id")
		postID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
			return
		}

		post, err := client.GetPost(context.Background(), &GetPostRequest{Id: postID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, post)
	})

	r.POST("/posts/:id/like", func(c *gin.Context) {
		id := c.Param("id")
		postID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
			return
		}

		var req struct {
			UserID int64 `json:"user_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = client.LikePost(context.Background(), &LikePostRequest{
			PostId: postID,
			UserId: req.UserID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	r.POST("/posts/:id/comments", func(c *gin.Context) {
		id := c.Param("id")
		postID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
			return
		}

		var req struct {
			UserID  int64  `json:"user_id" binding:"required"`
			Content string `json:"content" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		comment, err := client.CommentPost(context.Background(), &CommentPostRequest{
			PostId:  postID,
			UserId:  req.UserID,
			Content: req.Content,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, comment)
	})

	r.GET("/posts/:id/comments", func(c *gin.Context) {
		id := c.Param("id")
		postID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		comments, err := client.GetPostComments(context.Background(), &GetPostCommentsRequest{
			PostId: postID,
			Page:   int32(page),
			Limit:  int32(limit),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, comments)
	})

	log.Printf("Starting server on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
