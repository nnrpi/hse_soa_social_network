package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"social-network/post-service/models"
)

type PromoRepository struct {
	db *sql.DB
}

func NewPromoRepository(db *sql.DB) *PromoRepository {
	return &PromoRepository{db: db}
}

func (r *PromoRepository) Init() error {
	query := `CREATE TABLE IF NOT EXISTS promos (
        id VARCHAR(36) PRIMARY KEY,
        title VARCHAR(200) NOT NULL,
        description TEXT NOT NULL,
        creator_id INT NOT NULL,
        discount FLOAT NOT NULL,
        code VARCHAR(50) UNIQUE NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    )`
	_, err := r.db.Exec(query)
	return err
}

func (r *PromoRepository) CreatePromo(promo *models.Promo) error {
	promo.ID = uuid.New().String()
	promo.CreatedAt = time.Now()
	promo.UpdatedAt = time.Now()

	query := `INSERT INTO promos (id, title, description, creator_id, discount, code, created_at, updated_at)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(query,
		promo.ID,
		promo.Title,
		promo.Description,
		promo.CreatorID,
		promo.Discount,
		promo.Code,
		promo.CreatedAt,
		promo.UpdatedAt,
	)
	return err
}

func (r *PromoRepository) GetPromoByID(id string) (*models.Promo, error) {
	query := `SELECT id, title, description, creator_id, discount, code, created_at, updated_at
             FROM promos WHERE id = $1`

	var promo models.Promo
	err := r.db.QueryRow(query, id).Scan(
		&promo.ID,
		&promo.Title,
		&promo.Description,
		&promo.CreatorID,
		&promo.Discount,
		&promo.Code,
		&promo.CreatedAt,
		&promo.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("promo not found")
		}
		return nil, err
	}
	return &promo, nil
}

func (r *PromoRepository) UpdatePromo(promo *models.Promo) error {
	promo.UpdatedAt = time.Now()

	query := `UPDATE promos SET 
             title = $1, 
             description = $2, 
             discount = $3, 
             code = $4, 
             updated_at = $5
             WHERE id = $6 AND creator_id = $7`

	result, err := r.db.Exec(query,
		promo.Title,
		promo.Description,
		promo.Discount,
		promo.Code,
		promo.UpdatedAt,
		promo.ID,
		promo.CreatorID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("promo not found or you don't have permission to update it")
	}

	return nil
}

func (r *PromoRepository) DeletePromo(id string, creatorID int64) error {
	query := `DELETE FROM promos WHERE id = $1 AND creator_id = $2`
	result, err := r.db.Exec(query, id, creatorID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("promo not found or you don't have permission to delete it")
	}

	return nil
}

func (r *PromoRepository) ListPromos(page, pageSize int32, userID int64) ([]*models.Promo, int, error) {
	countQuery := `SELECT COUNT(*) FROM promos WHERE creator_id = $1`
	var total int
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `SELECT id, title, description, creator_id, discount, code, created_at, updated_at
             FROM promos 
             WHERE creator_id = $1
             ORDER BY created_at DESC
             LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var promos []*models.Promo
	for rows.Next() {
		var promo models.Promo
		if err := rows.Scan(
			&promo.ID,
			&promo.Title,
			&promo.Description,
			&promo.CreatorID,
			&promo.Discount,
			&promo.Code,
			&promo.CreatedAt,
			&promo.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		promos = append(promos, &promo)
	}

	return promos, total, nil
}
