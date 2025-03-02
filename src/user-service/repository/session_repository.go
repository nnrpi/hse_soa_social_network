package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"social-network/user-service/models"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Init() error {
	query := `CREATE TABLE IF NOT EXISTS sessions (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		username VARCHAR(50) NOT NULL,
		session_token VARCHAR(100) UNIQUE NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`

	_, err := r.db.Exec(query)
	return err
}

func (r *SessionRepository) CreateSession(userID int64, username string, duration time.Duration) (*models.Session, error) {
	token := uuid.New().String()
	expiresAt := time.Now().Add(duration)

	session := &models.Session{
		UserID:       userID,
		Username:     username,
		SessionToken: token,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}

	query := `INSERT INTO sessions (user_id, username, session_token, expires_at, created_at)
			  VALUES ($1, $2, $3, $4, $5)
			  RETURNING id`

	err := r.db.QueryRow(query,
		session.UserID,
		session.Username,
		session.SessionToken,
		session.ExpiresAt,
		session.CreatedAt,
	).Scan(&session.ID)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *SessionRepository) GetSessionByToken(token string) (*models.Session, error) {
	query := `SELECT id, user_id, username, session_token, expires_at, created_at
			  FROM sessions WHERE session_token = $1`

	var session models.Session
	err := r.db.QueryRow(query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Username,
		&session.SessionToken,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE session_token = $1`
	_, err := r.db.Exec(query, token)
	return err
}

func (r *SessionRepository) DeleteAllUserSessions(username string) error {
	query := `DELETE FROM sessions WHERE username = $1`
	_, err := r.db.Exec(query, username)
	return err
}

func (r *SessionRepository) CleanExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at < $1`
	_, err := r.db.Exec(query, time.Now())
	return err
}
