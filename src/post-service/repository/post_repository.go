package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"social-network/post-service/models"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Init() error {
	postsQuery := `CREATE TABLE IF NOT EXISTS posts (
        id VARCHAR(36) PRIMARY KEY,
        title VARCHAR(200) NOT NULL,
        description TEXT NOT NULL,
        creator_id INT NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
        is_private BOOLEAN NOT NULL DEFAULT FALSE
    )`
	_, err := r.db.Exec(postsQuery)
	if err != nil {
		return err
	}

	tagsQuery := `CREATE TABLE IF NOT EXISTS post_tags (
        post_id VARCHAR(36) NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
        tag VARCHAR(100) NOT NULL,
        PRIMARY KEY (post_id, tag)
    )`
	_, err = r.db.Exec(tagsQuery)
	return err
}

func (r *PostRepository) CreatePost(post *models.Post) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	post.ID = uuid.New().String()
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	postQuery := `INSERT INTO posts (id, title, description, creator_id, created_at, updated_at, is_private)
                 VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(postQuery,
		post.ID,
		post.Title,
		post.Description,
		post.CreatorID,
		post.CreatedAt,
		post.UpdatedAt,
		post.IsPrivate,
	)
	if err != nil {
		return err
	}

	for _, tag := range post.Tags {
		tagQuery := `INSERT INTO post_tags (post_id, tag) VALUES ($1, $2)`
		_, err = tx.Exec(tagQuery, post.ID, tag)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostRepository) GetPostByID(id string) (*models.Post, error) {
	postQuery := `SELECT id, title, description, creator_id, created_at, updated_at, is_private
                 FROM posts WHERE id = $1`

	var post models.Post
	err := r.db.QueryRow(postQuery, id).Scan(
		&post.ID,
		&post.Title,
		&post.Description,
		&post.CreatorID,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.IsPrivate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	tagsQuery := `SELECT tag FROM post_tags WHERE post_id = $1`
	rows, err := r.db.Query(tagsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	post.Tags = tags

	return &post, nil
}

func (r *PostRepository) UpdatePost(post *models.Post) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	post.UpdatedAt = time.Now()

	postQuery := `UPDATE posts SET 
                 title = $1, 
                 description = $2, 
                 updated_at = $3, 
                 is_private = $4
                 WHERE id = $5 AND creator_id = $6`

	result, err := tx.Exec(postQuery,
		post.Title,
		post.Description,
		post.UpdatedAt,
		post.IsPrivate,
		post.ID,
		post.CreatorID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("post not found or you don't have permission to update it")
	}

	_, err = tx.Exec(`DELETE FROM post_tags WHERE post_id = $1`, post.ID)
	if err != nil {
		return err
	}

	for _, tag := range post.Tags {
		tagQuery := `INSERT INTO post_tags (post_id, tag) VALUES ($1, $2)`
		_, err = tx.Exec(tagQuery, post.ID, tag)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostRepository) DeletePost(id string, creatorID int64) error {
	query := `DELETE FROM posts WHERE id = $1 AND creator_id = $2`
	result, err := r.db.Exec(query, id, creatorID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("post not found or you don't have permission to delete it")
	}

	return nil
}

func (r *PostRepository) ListPosts(page, pageSize int32, userID int64) ([]*models.Post, int, error) {
	countQuery := `SELECT COUNT(*) FROM posts WHERE creator_id = $1 OR is_private = FALSE`
	var total int
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	postsQuery := `SELECT id, title, description, creator_id, created_at, updated_at, is_private
                  FROM posts 
                  WHERE creator_id = $1 OR is_private = FALSE
                  ORDER BY created_at DESC
                  LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(postsQuery, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Description,
			&post.CreatorID,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.IsPrivate,
		); err != nil {
			return nil, 0, err
		}
		posts = append(posts, &post)
	}

	for _, post := range posts {
		tagsQuery := `SELECT tag FROM post_tags WHERE post_id = $1`
		tagRows, err := r.db.Query(tagsQuery, post.ID)
		if err != nil {
			return nil, 0, err
		}
		defer tagRows.Close()

		var tags []string
		for tagRows.Next() {
			var tag string
			if err := tagRows.Scan(&tag); err != nil {
				return nil, 0, err
			}
			tags = append(tags, tag)
		}
		post.Tags = tags
	}

	return posts, total, nil
}
