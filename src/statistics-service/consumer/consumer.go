package consumer

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Shopify/sarama"

	"social-network/statistics-service/models"
	"social-network/statistics-service/service"
)

// Topics is the set of topics this service consumes. user_registrations is
// deliberately excluded - the statistics ER model has no "registrations"
// entity, only likes/comments/views tied to posts.
var Topics = []string{"post_views", "post_interactions", "post_comments"}

type Handler struct {
	stats *service.StatsService
}

func NewHandler(stats *service.StatsService) *Handler {
	return &Handler{stats: stats}
}

func (h *Handler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *Handler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *Handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.handleMessage(msg.Topic, msg.Value); err != nil {
			log.Printf("failed to process message from %s: %v", msg.Topic, err)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (h *Handler) handleMessage(topic string, value []byte) error {
	var event models.KafkaEvent
	if err := json.Unmarshal(value, &event); err != nil {
		return err
	}

	at, err := event.ParsedTimestamp()
	if err != nil {
		at = time.Now()
	}

	switch topic {
	case "post_views":
		return h.stats.RecordView(event.PostID, event.UserID, at)
	case "post_interactions":
		return h.stats.RecordLike(event.PostID, event.UserID, at)
	case "post_comments":
		return h.stats.RecordComment(event.PostID, event.UserID, at)
	}

	return nil
}
