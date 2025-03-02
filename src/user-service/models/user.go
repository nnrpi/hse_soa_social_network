package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int64     `json:"-"`
	Username    string    `json:"username"`
	Password    string    `json:"-"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	Birthdate   time.Time `json:"birthdate"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

type UserPublic struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}

type SignInRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UpdateUserRequest struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Email       string `json:"email" validate:"omitempty,email"`
	Birthdate   string `json:"birthdate" validate:"omitempty,datetime=2025-03-01"`
	PhoneNumber string `json:"phone_number" validate:"omitempty"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
