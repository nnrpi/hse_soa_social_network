package repository

import (
	"database/sql"
	"errors"
	"time"

	"social-network/user-service/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Init() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		name VARCHAR(100),
		surname VARCHAR(100),
		birthdate DATE,
		phone_number VARCHAR(20),
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`

	_, err := r.db.Exec(query)
	return err
}

func (r *UserRepository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (username, password, email, name, surname, birthdate, phone_number, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			  RETURNING id`

	err := r.db.QueryRow(query,
		user.Username,
		user.Password,
		user.Email,
		user.Name,
		user.Surname,
		user.Birthdate,
		user.PhoneNumber,
		time.Now(),
		time.Now(),
	).Scan(&user.ID)

	return err
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, password, email, name, surname, birthdate, phone_number, created_at, updated_at
			  FROM users WHERE username = $1`

	var user models.User
	var birthdate sql.NullTime

	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Name,
		&user.Surname,
		&birthdate,
		&user.PhoneNumber,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if birthdate.Valid {
		user.Birthdate = birthdate.Time
	}

	return &user, nil
}

func (r *UserRepository) UserExists(username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	err := r.db.QueryRow(query, username).Scan(&exists)
	return exists, err
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	err := r.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	query := `UPDATE users SET 
			  name = $1, 
			  surname = $2, 
			  email = $3, 
			  birthdate = $4, 
			  phone_number = $5, 
			  updated_at = $6
			  WHERE username = $7`

	_, err := r.db.Exec(query,
		user.Name,
		user.Surname,
		user.Email,
		user.Birthdate,
		user.PhoneNumber,
		time.Now(),
		user.Username,
	)

	return err
}
