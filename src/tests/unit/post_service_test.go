package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"social-network/post-service/models"
)

type PostFilter struct {
	Page     int
	PageSize int
	UserID   string
	Tags     []string
}

type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) GetByID(ctx context.Context, id string) (*models.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) Update(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) List(ctx context.Context, filter *PostFilter) ([]*models.Post, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Post), args.Error(1)
}

type PostService struct {
	repo *MockPostRepository
}

func NewPostService(repo *MockPostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) CreatePost(ctx context.Context, post *models.Post) error {
	return s.repo.Create(ctx, post)
}

func (s *PostService) GetPost(ctx context.Context, id string) (*models.Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PostService) UpdatePost(ctx context.Context, post *models.Post) error {
	_, err := s.repo.GetByID(ctx, post.ID)
	if err != nil {
		return err
	}
	return s.repo.Update(ctx, post)
}

func (s *PostService) DeletePost(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func TestCreatePost(t *testing.T) {
	mockRepo := new(MockPostRepository)
	service := NewPostService(mockRepo)

	ctx := context.Background()
	post := &models.Post{
		CreatorID:   123,
		Title:       "Test Post",
		Description: "This is a test post",
		CreatedAt:   time.Now(),
		IsPrivate:   false,
	}

	mockRepo.On("Create", ctx, post).Return(nil)

	err := service.CreatePost(ctx, post)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetPost(t *testing.T) {
	mockRepo := new(MockPostRepository)
	service := NewPostService(mockRepo)

	ctx := context.Background()
	postID := "post123"
	expectedPost := &models.Post{
		ID:          postID,
		CreatorID:   123,
		Title:       "Test Post",
		Description: "This is a test post",
		CreatedAt:   time.Now(),
		IsPrivate:   false,
	}

	mockRepo.On("GetByID", ctx, postID).Return(expectedPost, nil)

	post, err := service.GetPost(ctx, postID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPost, post)
	mockRepo.AssertExpectations(t)
}
