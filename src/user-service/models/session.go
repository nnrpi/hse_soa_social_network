package models

import (
	"time"
)

type Session struct {
	ID           string    `json:"-"`
	UserID       int64     `json:"-"`
	Username     string    `json:"-"`
	SessionToken string    `json:"-"`
	ExpiresAt    time.Time `json:"-"`
	CreatedAt    time.Time `json:"-"`
}

type AuthResponse struct {
	Username  string `json:"username"`
	Message   string `json:"message"`
	Token     string `json:"token,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

type LogoutRequest struct {
	Username string `json:"username" validate:"required"`
}
