package unit

import (
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"social-network/user-service/models"
	"social-network/user-service/repository"
)

func TestUserRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err, fmt.Sprintf("Failed to create mock database: %v", err))
	defer db.Close()
	userRepo := repository.NewUserRepository(db)
	t.Run("CreateUser", func(t *testing.T) {
		user := &models.User{
			Username:    "testuser",
			Password:    "hashedpassword",
			Email:       "test@example.com",
			Name:        "Test",
			Surname:     "User",
			PhoneNumber: "1234567890",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mock.ExpectQuery("INSERT INTO users").
			WithArgs(
				user.Username,
				user.Password,
				user.Email,
				user.Name,
				user.Surname,
				user.Birthdate,
				user.PhoneNumber,
				sqlmock.AnyArg(), // CreatedAt
				sqlmock.AnyArg(), // UpdatedAt
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err := userRepo.CreateUser(user)
		assert.Nil(t, err, fmt.Sprintf("Expected no error, got %v", err))
		assert.Equal(t, int(user.ID), int(1), fmt.Sprintf("Expected user ID to be 1, got %d", user.ID))
		err = mock.ExpectationsWereMet()
		assert.Nil(t, err, fmt.Sprintf("Unfulfilled expectations: %v", err))
	})
	t.Run("GetUserByUsername", func(t *testing.T) {
		birthdate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
		createdAt := time.Now().Add(-24 * time.Hour)
		updatedAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "username", "password", "email", "name", "surname", "birthdate",
			"phone_number", "created_at", "updated_at",
		}).AddRow(
			1, "testuser", "hashedpassword", "test@example.com", "Test", "User",
			birthdate, "1234567890", createdAt, updatedAt,
		)

		mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
			WithArgs("testuser").
			WillReturnRows(rows)

		user, err := userRepo.GetUserByUsername("testuser")

		assert.Nil(t, err, fmt.Sprintf("Expected no error, got %v", err))
		assert.NotNil(t, user, "Expected user to be returned, got nil")
		assert.True(t, user.ID == 1 && user.Username == "testuser" && user.Email == "test@example.com", "User data doesn't match expected values")
		err = mock.ExpectationsWereMet()
		assert.Nil(t, err, fmt.Sprintf("Unfulfilled expectations: %v", err))
	})

	t.Run("UserExists", func(t *testing.T) {
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("testuser").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := userRepo.UserExists("testuser")
		assert.Nil(t, err, fmt.Sprintf("Expected no error, got %v", err))
		assert.True(t, exists, "Expected user to exist, got false")
		err = mock.ExpectationsWereMet()
		assert.Nil(t, err, fmt.Sprintf("Unfulfilled expectations: %v", err))
	})
}

func TestSessionRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err, fmt.Sprintf("Failed to create mock database: %v", err))
	defer db.Close()

	sessionRepo := repository.NewSessionRepository(db)

	t.Run("CreateSession", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO sessions").
			WithArgs(
				sqlmock.AnyArg(), // UserID
				"testuser",
				sqlmock.AnyArg(), // SessionToken
				sqlmock.AnyArg(), // ExpiresAt
				sqlmock.AnyArg(), // CreatedAt
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		session, err := sessionRepo.CreateSession(1, "testuser", 24*time.Hour)

		assert.Nil(t, err, fmt.Sprintf("Expected no error, got %v", err))
		assert.NotNil(t, session, "Expected session to be created, got nil")
		assert.True(t, session.Username == "testuser" && session.UserID == 1, "Session data doesn't match expected values")
		assert.NotEqual(t, session.SessionToken, "", "Expected session token to be generated")
		err = mock.ExpectationsWereMet()
		assert.Nil(t, err, fmt.Sprintf("Unfulfilled expectations: %v", err))
	})

	t.Run("GetSessionByToken", func(t *testing.T) {
		token := "test-token-123"
		expiresAt := time.Now().Add(24 * time.Hour)
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "user_id", "username", "session_token", "expires_at", "created_at",
		}).AddRow(
			1, 1, "testuser", token, expiresAt, createdAt,
		)

		mock.ExpectQuery("SELECT (.+) FROM sessions WHERE session_token = ?").
			WithArgs(token).
			WillReturnRows(rows)

		session, err := sessionRepo.GetSessionByToken(token)

		assert.Nil(t, err, fmt.Sprintf("Expected no error, got %v", err))
		assert.NotNil(t, session, "Expected session to be returned, got nil")
		assert.True(t, session.SessionToken == token && session.Username == "testuser", "Session data doesn't match expected values")
		err = mock.ExpectationsWereMet()
		assert.Nil(t, err, fmt.Sprintf("Unfulfilled expectations: %v", err))
	})
}
