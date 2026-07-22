package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"social-network/statistics-service/models"
	"social-network/statistics-service/repository"
	"social-network/statistics-service/service"
)

type MockStatRepo struct {
	mock.Mock
}

func (m *MockStatRepo) GetBySourceID(sourceID int64) (*models.Stat, error) {
	args := m.Called(sourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stat), args.Error(1)
}

func (m *MockStatRepo) Insert(stat *models.Stat) error {
	args := m.Called(stat)
	return args.Error(0)
}

func TestRecordView(t *testing.T) {
	now := time.Now()

	t.Run("new post", func(t *testing.T) {
		views := new(MockStatRepo)
		views.On("GetBySourceID", int64(1)).Return(nil, repository.ErrNotFound)
		views.On("Insert", mock.MatchedBy(func(s *models.Stat) bool {
			return s.SourceID == 1 && assert.ObjectsAreEqual([]int64{42}, s.UserIDs) && s.CreatedTimestamp.Equal(now) && s.LastModified.Equal(now)
		})).Return(nil)

		svc := service.NewStatsService(new(MockStatRepo), new(MockStatRepo), views)
		err := svc.RecordView(1, 42, now)

		assert.NoError(t, err)
		views.AssertExpectations(t)
	})

	t.Run("existing post appends user", func(t *testing.T) {
		created := now.Add(-time.Hour)
		existing := &models.Stat{SourceID: 1, UserIDs: []int64{7}, CreatedTimestamp: created, LastModified: created}

		views := new(MockStatRepo)
		views.On("GetBySourceID", int64(1)).Return(existing, nil)
		views.On("Insert", mock.MatchedBy(func(s *models.Stat) bool {
			return assert.ObjectsAreEqual([]int64{7, 42}, s.UserIDs) && s.CreatedTimestamp.Equal(created) && s.LastModified.Equal(now)
		})).Return(nil)

		svc := service.NewStatsService(new(MockStatRepo), new(MockStatRepo), views)
		err := svc.RecordView(1, 42, now)

		assert.NoError(t, err)
		views.AssertExpectations(t)
	})
}

func TestRecordLikeAndComment(t *testing.T) {
	now := time.Now()

	likes := new(MockStatRepo)
	likes.On("GetBySourceID", int64(5)).Return(nil, repository.ErrNotFound)
	likes.On("Insert", mock.AnythingOfType("*models.Stat")).Return(nil)

	comments := new(MockStatRepo)
	comments.On("GetBySourceID", int64(5)).Return(nil, repository.ErrNotFound)
	comments.On("Insert", mock.AnythingOfType("*models.Stat")).Return(nil)

	svc := service.NewStatsService(likes, comments, new(MockStatRepo))

	assert.NoError(t, svc.RecordLike(5, 1, now))
	assert.NoError(t, svc.RecordComment(5, 1, now))
	likes.AssertExpectations(t)
	comments.AssertExpectations(t)
}

func TestGetPostStats(t *testing.T) {
	t.Run("no data yet", func(t *testing.T) {
		likes := new(MockStatRepo)
		likes.On("GetBySourceID", int64(9)).Return(nil, repository.ErrNotFound)
		comments := new(MockStatRepo)
		comments.On("GetBySourceID", int64(9)).Return(nil, repository.ErrNotFound)
		views := new(MockStatRepo)
		views.On("GetBySourceID", int64(9)).Return(nil, repository.ErrNotFound)

		svc := service.NewStatsService(likes, comments, views)
		summary, err := svc.GetPostStats(9)

		assert.NoError(t, err)
		assert.Equal(t, &models.PostStatsSummary{PostID: 9}, summary)
	})

	t.Run("aggregates counts and unique counts", func(t *testing.T) {
		likes := new(MockStatRepo)
		likes.On("GetBySourceID", int64(3)).Return(&models.Stat{SourceID: 3, UserIDs: []int64{1, 2}}, nil)
		comments := new(MockStatRepo)
		comments.On("GetBySourceID", int64(3)).Return(&models.Stat{SourceID: 3, UserIDs: []int64{1, 1, 2}}, nil)
		views := new(MockStatRepo)
		views.On("GetBySourceID", int64(3)).Return(&models.Stat{SourceID: 3, UserIDs: []int64{1, 1, 1}}, nil)

		svc := service.NewStatsService(likes, comments, views)
		summary, err := svc.GetPostStats(3)

		assert.NoError(t, err)
		assert.Equal(t, &models.PostStatsSummary{
			PostID:              3,
			ViewsCount:          3,
			UniqueViewsCount:    1,
			LikesCount:          2,
			UniqueLikesCount:    2,
			CommentsCount:       3,
			UniqueCommentsCount: 2,
		}, summary)
	})
}
