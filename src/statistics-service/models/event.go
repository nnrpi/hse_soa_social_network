package models

import "time"

// KafkaEvent is the shape produced by posts-grpc-service into
// post_views, post_interactions and post_comments.
type KafkaEvent struct {
	EventType string `json:"event_type"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content,omitempty"`
	Timestamp string `json:"timestamp"`
}

func (e KafkaEvent) ParsedTimestamp() (time.Time, error) {
	return time.Parse(time.RFC3339, e.Timestamp)
}
