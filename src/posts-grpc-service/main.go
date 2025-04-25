package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pb "posts-api-service/proto"
)

type server struct {
	pb.UnimplementedPostsServiceServer
	db            *gorm.DB
	kafkaProducer sarama.SyncProducer
}

type Post struct {
	ID            int64      `gorm:"primaryKey"`
	Title         string     `gorm:"size:255;not null"`
	Content       string     `gorm:"type:text;not null"`
	UserID        int64      `gorm:"not null"`
	CreatedAt     time.Time  `gorm:"not null"`
	LikesCount    int32      `gorm:"default:0"`
	CommentsCount int32      `gorm:"default:0"`
	ViewsCount    int32      `gorm:"default:0"`
	DeletedAt     *time.Time `gorm:"index"`
}

type Comment struct {
	ID        int64      `gorm:"primaryKey"`
	PostID    int64      `gorm:"not null"`
	UserID    int64      `gorm:"not null"`
	Content   string     `gorm:"type:text;not null"`
	CreatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:"index"`
}

type Like struct {
	ID        int64     `gorm:"primaryKey"`
	PostID    int64     `gorm:"not null"`
	UserID    int64     `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

type View struct {
	ID        int64     `gorm:"primaryKey"`
	PostID    int64     `gorm:"not null"`
	UserID    int64     `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

func (s *server) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.Post, error) {
	var post Post
	if err := s.db.First(&post, req.Id).Error; err != nil {
		return nil, err
	}

	view := View{
		PostID:    post.ID,
		UserID:    0,
		CreatedAt: time.Now(),
	}
	s.db.Create(&view)

	s.db.Model(&post).Update("views_count", post.ViewsCount+1)

	viewEvent := map[string]interface{}{
		"event_type": "view",
		"post_id":    post.ID,
		"user_id":    0,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	s.sendKafkaEvent("post_views", viewEvent)

	return &pb.Post{
		Id:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		UserId:        post.UserID,
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
		LikesCount:    post.LikesCount,
		CommentsCount: post.CommentsCount,
		ViewsCount:    post.ViewsCount + 1,
	}, nil
}

func (s *server) LikePost(ctx context.Context, req *pb.LikePostRequest) (*pb.LikePostResponse, error) {
	var post Post
	if err := s.db.First(&post, req.PostId).Error; err != nil {
		return nil, err
	}

	var existingLike Like
	result := s.db.Where("post_id = ? AND user_id = ?", req.PostId, req.UserId).First(&existingLike)
	if result.Error == nil {
		return &pb.LikePostResponse{Success: false}, nil
	}

	like := Like{
		PostID:    req.PostId,
		UserID:    req.UserId,
		CreatedAt: time.Now(),
	}
	if err := s.db.Create(&like).Error; err != nil {
		return nil, err
	}

	s.db.Model(&post).Update("likes_count", post.LikesCount+1)

	likeEvent := map[string]interface{}{
		"event_type": "like",
		"post_id":    req.PostId,
		"user_id":    req.UserId,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	s.sendKafkaEvent("post_interactions", likeEvent)

	return &pb.LikePostResponse{Success: true}, nil
}

func (s *server) CommentPost(ctx context.Context, req *pb.CommentPostRequest) (*pb.Comment, error) {
	var post Post
	if err := s.db.First(&post, req.PostId).Error; err != nil {
		return nil, err
	}

	comment := Comment{
		PostID:    req.PostId,
		UserID:    req.UserId,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
	if err := s.db.Create(&comment).Error; err != nil {
		return nil, err
	}

	s.db.Model(&post).Update("comments_count", post.CommentsCount+1)

	commentEvent := map[string]interface{}{
		"event_type": "comment",
		"post_id":    req.PostId,
		"user_id":    req.UserId,
		"content":    req.Content,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	s.sendKafkaEvent("post_comments", commentEvent)

	return &pb.Comment{
		Id:        comment.ID,
		PostId:    comment.PostID,
		UserId:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *server) GetPostComments(ctx context.Context, req *pb.GetPostCommentsRequest) (*pb.CommentsResponse, error) {
	var comments []Comment
	var total int64

	offset := (req.Page - 1) * req.Limit
	s.db.Model(&Comment{}).Where("post_id = ?", req.PostId).Count(&total)

	s.db.Where("post_id = ?", req.PostId).
		Order("created_at DESC").
		Offset(int(offset)).
		Limit(int(req.Limit)).
		Find(&comments)

	pbComments := make([]*pb.Comment, len(comments))
	for i, comment := range comments {
		pbComments[i] = &pb.Comment{
			Id:        comment.ID,
			PostId:    comment.PostID,
			UserId:    comment.UserID,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		}
	}

	return &pb.CommentsResponse{
		Comments: pbComments,
		Total:    int32(total),
		Page:     req.Page,
		Limit:    req.Limit,
	}, nil
}

func (s *server) sendKafkaEvent(topic string, event map[string]interface{}) {
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event to JSON: %v", err)
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(jsonBytes),
	}

	_, _, err = s.kafkaProducer.SendMessage(msg)
	if err != nil {
		log.Printf("Error sending message to Kafka: %v", err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.AutoMigrate(&Post{}, &Comment{}, &Like{}, &View{})

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 10
	kafkaConfig.Producer.Return.Successes = true

	kafkaBrokers := []string{os.Getenv("KAFKA_BROKER")}
	if len(kafkaBrokers[0]) == 0 {
		kafkaBrokers = []string{"kafka:9092"}
	}

	producer, err := sarama.NewSyncProducer(kafkaBrokers, kafkaConfig)
	if err != nil {
		log.Fatalf("Failed to setup Kafka producer: %v", err)
	}
	defer producer.Close()

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPostsServiceServer(s, &server{
		db:            db,
		kafkaProducer: producer,
	})

	reflection.Register(s)

	log.Printf("Starting gRPC server on port %s\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
