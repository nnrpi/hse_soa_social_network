package unit

import (
	"testing"

	"social-network/user-service/api"
	"social-network/user-service/models"

	"github.com/stretchr/testify/assert"
)

func TestPasswordHashing(t *testing.T) {
	password := "testPassword123"

	hashedPassword, err := models.HashPassword(password)
	assert.Nil(t, err, "Failed to hash password")
	assert.NotEqual(t, hashedPassword, password, "Hashed password should be different from original")
	assert.True(t, models.CheckPasswordHash(password, hashedPassword), "Password verification failed for correct password")
	assert.False(t, models.CheckPasswordHash("wrongPassword", hashedPassword), "Password verification passed for incorrect password")
}

func TestUserModelValidation(t *testing.T) {
	validate := func(s interface{}) error {
		return api.Validate(s)
	}
	tests := []struct {
		name      string
		req       models.SignInRequest
		shouldErr bool
	}{
		{
			name: "Valid SignInRequest",
			req: models.SignInRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			shouldErr: false,
		},
		{
			name: "Short Username",
			req: models.SignInRequest{
				Username: "te", // too short username
				Email:    "test@example.com",
				Password: "password123",
			},
			shouldErr: true,
		},
		{
			name: "Invalid Email",
			req: models.SignInRequest{
				Username: "testuser",
				Email:    "not-an-email",
				Password: "password123",
			},
			shouldErr: true,
		},
		{
			name: "Short Password",
			req: models.SignInRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "pass", // Too short password
			},
			shouldErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(tc.req)
			hasError := (err != nil)
			if hasError != tc.shouldErr {
				t.Errorf("Expected validation error: %v, got: %v", tc.shouldErr, hasError)
				if err != nil {
					t.Logf("Validation error: %v", err)
				}
			}
		})
	}
}
