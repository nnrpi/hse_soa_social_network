package unit

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"social-network/statistics-service/models"
	"social-network/statistics-service/repository"
)

// arrayAwareConverter extends the default sqlmock converter with support for
// []int64, which backs ClickHouse's Array(Int64) columns - the default
// driver.ValueConverter rejects any slice that isn't []byte.
type arrayAwareConverter struct{}

func (arrayAwareConverter) ConvertValue(v interface{}) (driver.Value, error) {
	if ids, ok := v.([]int64); ok {
		return ids, nil
	}
	return driver.DefaultParameterConverter.ConvertValue(v)
}

func TestStatRepository(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.ValueConverterOption(arrayAwareConverter{}))
	assert.Nil(t, err, fmt.Sprintf("Failed to create mock database: %v", err))
	defer db.Close()

	likesRepo := repository.NewLikesRepository(db)

	t.Run("Init", func(t *testing.T) {
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS likes").WillReturnResult(sqlmock.NewResult(0, 0))

		err := likesRepo.Init()
		assert.Nil(t, err, fmt.Sprintf("Expected no error, got %v", err))
		err = mock.ExpectationsWereMet()
		assert.Nil(t, err, fmt.Sprintf("Unfulfilled expectations: %v", err))
	})

	t.Run("Insert", func(t *testing.T) {
		stat := &models.Stat{
			SourceID:         1,
			UserIDs:          []int64{10},
			CreatedTimestamp: time.Now(),
			LastModified:     time.Now(),
		}

		mock.ExpectExec("INSERT INTO likes").
			WithArgs(stat.SourceID, stat.UserIDs, stat.CreatedTimestamp, stat.LastModified).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := likesRepo.Insert(stat)
		assert.Nil(t, err, fmt.Sprintf("Expected no error, got %v", err))
		err = mock.ExpectationsWereMet()
		assert.Nil(t, err, fmt.Sprintf("Unfulfilled expectations: %v", err))
	})

	t.Run("GetBySourceID found", func(t *testing.T) {
		created := time.Now().Add(-time.Hour)
		lastModified := time.Now()

		rows := mock.NewRows([]string{"source_id", "user_ids", "created_timestamp", "last_modified"}).
			AddRow(int64(1), []int64{10, 20}, created, lastModified)

		mock.ExpectQuery("SELECT (.+) FROM likes FINAL WHERE source_id = ?").
			WithArgs(int64(1)).
			WillReturnRows(rows)

		stat, err := likesRepo.GetBySourceID(1)
		assert.Nil(t, err, fmt.Sprintf("Expected no error, got %v", err))
		assert.NotNil(t, stat, "Expected stat to be returned, got nil")
		assert.Equal(t, int64(1), stat.SourceID)
		assert.Equal(t, []int64{10, 20}, stat.UserIDs)
		err = mock.ExpectationsWereMet()
		assert.Nil(t, err, fmt.Sprintf("Unfulfilled expectations: %v", err))
	})

	t.Run("GetBySourceID not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM likes FINAL WHERE source_id = ?").
			WithArgs(int64(2)).
			WillReturnError(sql.ErrNoRows)

		stat, err := likesRepo.GetBySourceID(2)
		assert.Nil(t, stat, "Expected no stat to be returned")
		assert.True(t, errors.Is(err, repository.ErrNotFound), fmt.Sprintf("Expected ErrNotFound, got %v", err))
		err = mock.ExpectationsWereMet()
		assert.Nil(t, err, fmt.Sprintf("Unfulfilled expectations: %v", err))
	})
}
