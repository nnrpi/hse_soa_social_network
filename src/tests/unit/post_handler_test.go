package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"social-network/post-service/models"
)

type MockPostService struct {
	mock.Mock
}

func (m *MockPostService) CreatePost(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostService) GetPost(ctx context.Context, id string) (*models.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Post), args.Error(1)
}

type PostHandler struct {
	service *MockPostService
}

func NewPostHandler(service *MockPostService) *PostHandler {
	return &PostHandler{service: service}
}

type CreatePostRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	IsPrivate   bool     `json:"is_private"`
	Tags        []string `json:"tags"`
}

func (h *PostHandler) Create(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
		return
	}

	post := &models.Post{
		Title:       req.Title,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
		Tags:        req.Tags,
		CreatorID:   123,
		CreatedAt:   time.Now(),
	}

	err := h.service.CreatePost(c, post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func TestCreatePostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(MockPostService)

	handler := NewPostHandler(mockService)
	router.POST("/posts", handler.Create)

	postData := CreatePostRequest{
		Title:       "Test Post",
		Description: "This is a test post",
		IsPrivate:   false,
		Tags:        []string{"test", "example"},
	}
	jsonData, _ := json.Marshal(postData)
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer token123")

	mockService.On("CreatePost", mock.Anything, mock.AnythingOfType("*models.Post")).Return(nil)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	mockService.AssertExpectations(t)
}
