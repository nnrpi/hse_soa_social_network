package service

import (
	"errors"
	"fmt"
	"time"

	"social-network/user-service/models"
	"social-network/user-service/repository"
)

type UserService struct {
	repo        *repository.UserRepository
	sessionRepo *repository.SessionRepository
}

func NewUserService(repo *repository.UserRepository, sessionRepo *repository.SessionRepository) *UserService {
	return &UserService{
		repo:        repo,
		sessionRepo: sessionRepo,
	}
}

func (s *UserService) SignIn(req *models.SignInRequest) (*models.AuthResponse, error) {
	fmt.Println(1)
	if s == nil {
		panic("s is nil")
	}
	if s.repo == nil {
		panic("s.repo is nil")
	}
	if req == nil {
		panic("req is nil")
	}

	exists, err := s.repo.UserExists(req.Username)
	fmt.Println(2)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	exists, err = s.repo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := models.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:  req.Username,
		Password:  hashedPassword,
		Email:     req.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Username: user.Username,
		Message:  "User created successfully",
	}, nil
}

func (s *UserService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, err
	}

	if !models.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	return &models.AuthResponse{
		Username: user.Username,
		Message:  "Login successful",
	}, nil
}

func (s *UserService) GetUserFullProfile(username, password string) (*models.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	if !models.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *UserService) GetUserPublicProfile(username string) (*models.UserPublic, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return &models.UserPublic{
		Username: user.Username,
		Name:     user.Name,
		Surname:  user.Surname,
	}, nil
}

func (s *UserService) LoginWithSession(req *models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, err
	}

	if !models.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	session, err := s.sessionRepo.CreateSession(user.ID, user.Username, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	expiresFormatted := session.ExpiresAt.Format(time.RFC3339)

	return &models.AuthResponse{
		Username:  user.Username,
		Message:   "Login successful",
		Token:     session.SessionToken,
		ExpiresAt: expiresFormatted,
	}, nil
}

func (s *UserService) Logout(token string) error {
	return s.sessionRepo.DeleteSession(token)
}

func (s *UserService) ValidateSession(token string) (*models.Session, error) {
	session, err := s.sessionRepo.GetSessionByToken(token)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		s.sessionRepo.DeleteSession(token)
		return nil, errors.New("session expired")
	}

	return session, nil
}

func (s *UserService) GetUserBySession(token string) (*models.User, error) {
	session, err := s.ValidateSession(token)
	if err != nil {
		return nil, err
	}

	return s.repo.GetUserByUsername(session.Username)
}

func (s *UserService) UpdateUserProfileBySession(username string, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Surname != "" {
		user.Surname = req.Surname
	}
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.repo.EmailExists(req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("email already used by another account")
		}
		user.Email = req.Email
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.Birthdate != "" {
		birthdate, err := time.Parse("2025-03-01", req.Birthdate)
		if err != nil {
			return nil, errors.New("invalid birthdate format")
		}
		user.Birthdate = birthdate
	}

	err = s.repo.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.repo.GetUserByUsername(username)
}

func (s *UserService) ValidateCredentials(username, password string) (bool, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return false, err
	}

	return models.CheckPasswordHash(password, user.Password), nil
}

func (s *UserService) UserExists(username string) (bool, error) {
	return s.repo.UserExists(username)
}

func (s *UserService) EmailExists(email string) (bool, error) {
	return s.repo.EmailExists(email)
}
