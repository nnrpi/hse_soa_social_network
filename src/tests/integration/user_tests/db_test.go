package integration

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"

	"social-network/user-service/models"
	"social-network/user-service/repository"
)

type DBTestSuite struct {
	suite.Suite
	db          *sql.DB
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	testUsers   []*models.User
}

func (s *DBTestSuite) SetupSuite() {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		s.T().Skip("Skipping database integration tests (set INTEGRATION_TEST=true to run)")
	}
	dbConnString := os.Getenv("TEST_DB_CONNECTION")
	if dbConnString == "" {
		dbConnString = "postgres://postgres:postgres@localhost:5432/socialnetwork_test?sslmode=disable"
	}

	var err error
	s.db, err = sql.Open("postgres", dbConnString)
	if err != nil {
		s.T().Fatalf("Failed to connect to database: %v", err)
	}
	err = s.db.Ping()
	if err != nil {
		s.T().Fatalf("Failed to ping database: %v", err)
	}
	s.userRepo = repository.NewUserRepository(s.db)
	s.sessionRepo = repository.NewSessionRepository(s.db)
	err = s.userRepo.Init()
	if err != nil {
		s.T().Fatalf("Failed to initialize user repository: %v", err)
	}
	err = s.sessionRepo.Init()
	if err != nil {
		s.T().Fatalf("Failed to initialize session repository: %v", err)
	}

	s.cleanTestData()
}

func (s *DBTestSuite) TearDownSuite() {
	if s.db != nil {
		s.cleanTestData()
		s.db.Close()
	}
}

func (s *DBTestSuite) SetupTest() {
	s.createTestUsers()
}

func (s *DBTestSuite) TearDownTest() {
	s.cleanTestData()
}

func (s *DBTestSuite) cleanTestData() {
	_, err := s.db.Exec("DELETE FROM sessions WHERE username LIKE 'test_%'")
	if err != nil {
		s.T().Logf("Warning: Failed to clean up test sessions: %v", err)
	}
	_, err = s.db.Exec("DELETE FROM users WHERE username LIKE 'test_%'")
	if err != nil {
		s.T().Logf("Warning: Failed to clean up test users: %v", err)
	}
	s.testUsers = nil
}

func (s *DBTestSuite) createTestUsers() {
	for i := 1; i <= 3; i++ {
		username := fmt.Sprintf("test_user_%d", i)
		email := fmt.Sprintf("test%d@example.com", i)
		password, err := models.HashPassword(fmt.Sprintf("password%d", i))
		s.Require().NoError(err, "Failed to hash password")
		user := &models.User{
			Username:    username,
			Password:    password,
			Email:       email,
			Name:        fmt.Sprintf("Test%d", i),
			Surname:     fmt.Sprintf("User%d", i),
			PhoneNumber: fmt.Sprintf("123456789%d", i),
			Birthdate:   time.Date(1990, time.January, i, 0, 0, 0, 0, time.UTC),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err = s.userRepo.CreateUser(user)
		s.Require().NoError(err, "Failed to create test user")
		s.testUsers = append(s.testUsers, user)
	}
}

func (s *DBTestSuite) TestUserRepositoryCreateUser() {
	user := &models.User{
		Username:    "test_create_user",
		Password:    "hashedpassword",
		Email:       "test_create@example.com",
		Name:        "Test",
		Surname:     "Create",
		PhoneNumber: "9876543210",
		Birthdate:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := s.userRepo.CreateUser(user)
	s.NoError(err, "Failed to create user")
	s.Greater(user.ID, int64(0), "User ID should be set after creation")
	_, err = s.db.Exec("DELETE FROM users WHERE username = 'test_create_user'")
	s.NoError(err, "Failed to clean up created user")
}

func (s *DBTestSuite) TestUserRepositoryGetUserByUsername() {
	user, err := s.userRepo.GetUserByUsername(s.testUsers[0].Username)
	s.NoError(err, "Failed to get user by username")
	s.NotNil(user, "User should not be nil")
	s.Equal(s.testUsers[0].Username, user.Username, "Username should match")
	s.Equal(s.testUsers[0].Email, user.Email, "Email should match")

	_, err = s.userRepo.GetUserByUsername("nonexistent_user")
	s.Error(err, "Should get error for non-existent user")
}

func (s *DBTestSuite) TestUserRepositoryUserExists() {
	exists, err := s.userRepo.UserExists(s.testUsers[0].Username)
	s.NoError(err, "Error checking if user exists")
	s.True(exists, "User should exist")

	exists, err = s.userRepo.UserExists("nonexistent_user")
	s.NoError(err, "Error checking if non-existent user exists")
	s.False(exists, "Non-existent user should not exist")
}

func (s *DBTestSuite) TestUserRepositoryEmailExists() {
	exists, err := s.userRepo.EmailExists(s.testUsers[0].Email)
	s.NoError(err, "Error checking if email exists")
	s.True(exists, "Email should exist")

	exists, err = s.userRepo.EmailExists("nonexistent@example.com")
	s.NoError(err, "Error checking if non-existent email exists")
	s.False(exists, "Non-existent email should not exist")
}

func (s *DBTestSuite) TestUserRepositoryUpdateUser() {
	user, err := s.userRepo.GetUserByUsername(s.testUsers[0].Username)
	s.NoError(err, "Failed to get user for update")
	user.Name = "Updated Name"
	user.Surname = "Updated Surname"
	user.Email = "updated@example.com"
	user.PhoneNumber = "5555555555"
	user.Birthdate = time.Date(1995, 5, 5, 0, 0, 0, 0, time.UTC)
	user.UpdatedAt = time.Now()
	err = s.userRepo.UpdateUser(user)
	s.NoError(err, "Failed to update user")
	updatedUser, err := s.userRepo.GetUserByUsername(s.testUsers[0].Username)
	s.NoError(err, "Failed to get updated user")
	s.Equal("Updated Name", updatedUser.Name, "Name should be updated")
	s.Equal("Updated Surname", updatedUser.Surname, "Surname should be updated")
	s.Equal("updated@example.com", updatedUser.Email, "Email should be updated")
	s.Equal("5555555555", updatedUser.PhoneNumber, "Phone number should be updated")
	expectedDate := time.Date(1995, 5, 5, 0, 0, 0, 0, time.UTC)
	s.True(expectedDate.Equal(updatedUser.Birthdate.UTC()),
		"Birthdate should be updated. Expected %v, got %v",
		expectedDate, updatedUser.Birthdate.UTC())
}

func (s *DBTestSuite) TestSessionRepositoryCreateSession() {
	userID := s.testUsers[0].ID
	username := s.testUsers[0].Username
	duration := 1 * time.Hour
	session, err := s.sessionRepo.CreateSession(userID, username, duration)
	s.NoError(err, "Failed to create session")
	s.NotNil(session, "Session should not be nil")
	s.Equal(username, session.Username, "Session username should match")
	s.NotEmpty(session.SessionToken, "Session token should not be empty")
	s.True(session.ExpiresAt.After(time.Now()), "Session expiry should be in the future")

	retrievedSession, err := s.sessionRepo.GetSessionByToken(session.SessionToken)
	s.NoError(err, "Failed to retrieve session")
	s.Equal(session.SessionToken, retrievedSession.SessionToken, "Session tokens should match")
	s.Equal(username, retrievedSession.Username, "Session username should match")
}

func (s *DBTestSuite) TestSessionRepositoryGetSessionByToken() {
	userID := s.testUsers[0].ID
	username := s.testUsers[0].Username
	session, err := s.sessionRepo.CreateSession(userID, username, 1*time.Hour)
	s.NoError(err, "Failed to create session for test")

	retrievedSession, err := s.sessionRepo.GetSessionByToken(session.SessionToken)
	s.NoError(err, "Failed to get session by token")
	s.NotNil(retrievedSession, "Session should not be nil")
	s.Equal(session.SessionToken, retrievedSession.SessionToken, "Session tokens should match")
	s.Equal(username, retrievedSession.Username, "Session username should match")

	_, err = s.sessionRepo.GetSessionByToken("nonexistent_token")
	s.Error(err, "Should get error for non-existent token")
}

func (s *DBTestSuite) TestSessionRepositoryDeleteSession() {
	userID := s.testUsers[0].ID
	username := s.testUsers[0].Username
	session, err := s.sessionRepo.CreateSession(userID, username, 1*time.Hour)
	s.NoError(err, "Failed to create session for test")
	err = s.sessionRepo.DeleteSession(session.SessionToken)
	s.NoError(err, "Failed to delete session")
	_, err = s.sessionRepo.GetSessionByToken(session.SessionToken)
	s.Error(err, "Deleted session should not be retrievable")
}

func (s *DBTestSuite) TestSessionRepositoryDeleteAllUserSessions() {
	userID := s.testUsers[0].ID
	username := s.testUsers[0].Username
	for i := 0; i < 3; i++ {
		_, err := s.sessionRepo.CreateSession(userID, username, 1*time.Hour)
		s.NoError(err, "Failed to create session %d for test", i)
	}
	err := s.sessionRepo.DeleteAllUserSessions(username)
	s.NoError(err, "Failed to delete all user sessions")

	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE username = $1", username).Scan(&count)
	s.NoError(err, "Failed to count sessions after deletion")
	s.Equal(0, count, "All sessions should be deleted")
}

func (s *DBTestSuite) TestSessionRepositoryCleanExpiredSessions() {
	userID := s.testUsers[0].ID
	username := s.testUsers[0].Username

	expiredSession, err := s.sessionRepo.CreateSession(userID, username, -1*time.Hour)
	s.NoError(err, "Failed to create expired session")

	validSession, err := s.sessionRepo.CreateSession(userID, username, 1*time.Hour)
	s.NoError(err, "Failed to create valid session")

	err = s.sessionRepo.CleanExpiredSessions()
	s.NoError(err, "Failed to clean expired sessions")

	_, err = s.sessionRepo.GetSessionByToken(expiredSession.SessionToken)
	s.Error(err, "Expired session should be deleted")

	_, err = s.sessionRepo.GetSessionByToken(validSession.SessionToken)
	s.NoError(err, "Valid session should still exist")
}

func TestDBSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}
