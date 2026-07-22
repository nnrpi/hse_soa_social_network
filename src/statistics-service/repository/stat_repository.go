package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"social-network/statistics-service/models"
)

var ErrNotFound = errors.New("stat not found")

// StatRepository is a raw-SQL repository shared by the likes/comments/views
// tables in ClickHouse - they're schema-identical, only the table name differs.
type StatRepository struct {
	db    *sql.DB
	table string
}

func NewLikesRepository(db *sql.DB) *StatRepository    { return &StatRepository{db: db, table: "likes"} }
func NewCommentsRepository(db *sql.DB) *StatRepository { return &StatRepository{db: db, table: "comments"} }
func NewViewsRepository(db *sql.DB) *StatRepository    { return &StatRepository{db: db, table: "views"} }

func (r *StatRepository) Init() error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		source_id          Int64,
		user_ids           Array(Int64),
		created_timestamp  DateTime,
		last_modified      DateTime
	) ENGINE = ReplacingMergeTree(last_modified)
	ORDER BY source_id`, r.table)

	_, err := r.db.Exec(query)
	return err
}

// GetBySourceID returns the current aggregated row for a post. Since
// ClickHouse's ReplacingMergeTree only collapses duplicate versions during
// background merges, FINAL forces that collapse at read time so callers
// always see the latest user_ids/last_modified for this source_id.
func (r *StatRepository) GetBySourceID(sourceID int64) (*models.Stat, error) {
	query := fmt.Sprintf(`SELECT source_id, user_ids, created_timestamp, last_modified
		FROM %s FINAL WHERE source_id = ?`, r.table)

	var stat models.Stat
	err := r.db.QueryRow(query, sourceID).Scan(
		&stat.SourceID,
		&stat.UserIDs,
		&stat.CreatedTimestamp,
		&stat.LastModified,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &stat, nil
}

// Insert writes a new version of the source_id's row. ClickHouse has no
// in-place UPDATE, so every call is a fresh row that ReplacingMergeTree will
// later collapse with earlier versions of the same source_id.
func (r *StatRepository) Insert(stat *models.Stat) error {
	query := fmt.Sprintf(`INSERT INTO %s (source_id, user_ids, created_timestamp, last_modified)
		VALUES (?, ?, ?, ?)`, r.table)

	_, err := r.db.Exec(query,
		stat.SourceID,
		stat.UserIDs,
		stat.CreatedTimestamp,
		stat.LastModified,
	)
	return err
}
