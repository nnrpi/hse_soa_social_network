package service

import (
	"errors"
	"time"

	"social-network/statistics-service/models"
	"social-network/statistics-service/repository"
)

// StatRepo is satisfied by *repository.StatRepository; kept as an interface
// so tests can mock it without a real (or sqlmock) database.
type StatRepo interface {
	GetBySourceID(sourceID int64) (*models.Stat, error)
	Insert(stat *models.Stat) error
}

type StatsService struct {
	likesRepo    StatRepo
	commentsRepo StatRepo
	viewsRepo    StatRepo
}

func NewStatsService(likesRepo, commentsRepo, viewsRepo StatRepo) *StatsService {
	return &StatsService{
		likesRepo:    likesRepo,
		commentsRepo: commentsRepo,
		viewsRepo:    viewsRepo,
	}
}

func (s *StatsService) RecordView(postID, userID int64, at time.Time) error {
	return record(s.viewsRepo, postID, userID, at)
}

func (s *StatsService) RecordLike(postID, userID int64, at time.Time) error {
	return record(s.likesRepo, postID, userID, at)
}

func (s *StatsService) RecordComment(postID, userID int64, at time.Time) error {
	return record(s.commentsRepo, postID, userID, at)
}

func record(repo StatRepo, postID, userID int64, at time.Time) error {
	existing, err := repo.GetBySourceID(postID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	if existing == nil {
		return repo.Insert(&models.Stat{
			SourceID:         postID,
			UserIDs:          []int64{userID},
			CreatedTimestamp: at,
			LastModified:     at,
		})
	}

	return repo.Insert(&models.Stat{
		SourceID:         postID,
		UserIDs:          append(existing.UserIDs, userID),
		CreatedTimestamp: existing.CreatedTimestamp,
		LastModified:     at,
	})
}

func (s *StatsService) GetPostStats(postID int64) (*models.PostStatsSummary, error) {
	summary := &models.PostStatsSummary{PostID: postID}

	views, err := statOrZero(s.viewsRepo, postID)
	if err != nil {
		return nil, err
	}
	summary.ViewsCount = len(views.UserIDs)
	summary.UniqueViewsCount = len(models.UniqueInt64(views.UserIDs))

	likes, err := statOrZero(s.likesRepo, postID)
	if err != nil {
		return nil, err
	}
	summary.LikesCount = len(likes.UserIDs)
	summary.UniqueLikesCount = len(models.UniqueInt64(likes.UserIDs))

	comments, err := statOrZero(s.commentsRepo, postID)
	if err != nil {
		return nil, err
	}
	summary.CommentsCount = len(comments.UserIDs)
	summary.UniqueCommentsCount = len(models.UniqueInt64(comments.UserIDs))

	return summary, nil
}

func statOrZero(repo StatRepo, postID int64) (*models.Stat, error) {
	stat, err := repo.GetBySourceID(postID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return &models.Stat{SourceID: postID}, nil
		}
		return nil, err
	}
	return stat, nil
}
