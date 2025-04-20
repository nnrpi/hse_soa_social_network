package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"social-network/post-service/models"
)

type MockPromoRepository struct {
	mock.Mock
}

func (m *MockPromoRepository) Create(ctx context.Context, promo *models.Promo) error {
	args := m.Called(ctx, promo)
	return args.Error(0)
}

func (m *MockPromoRepository) GetByCode(ctx context.Context, code string) (*models.Promo, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Promo), args.Error(1)
}

func (m *MockPromoRepository) Update(ctx context.Context, promo *models.Promo) error {
	args := m.Called(ctx, promo)
	return args.Error(0)
}

func (m *MockPromoRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type PromoService struct {
	repo *MockPromoRepository
}

func NewPromoService(repo *MockPromoRepository) *PromoService {
	return &PromoService{repo: repo}
}

func (s *PromoService) CreatePromo(ctx context.Context, promo *models.Promo) error {
	return s.repo.Create(ctx, promo)
}

func TestCreatePromo(t *testing.T) {
	mockRepo := new(MockPromoRepository)
	service := NewPromoService(mockRepo)

	ctx := context.Background()
	promo := &models.Promo{
		Code:        "SPRING2025",
		Title:       "Spring promotion",
		Discount:    20.0,
		Description: "Spring promotion",
		CreatedAt:   time.Now(),
	}

	mockRepo.On("Create", ctx, promo).Return(nil)

	err := service.CreatePromo(ctx, promo)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
