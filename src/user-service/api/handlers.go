package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"social-network/user-service/models"
	"social-network/user-service/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/signin", h.SignIn).Methods("POST")
	router.HandleFunc("/auth/login", h.LoginWithSession).Methods("POST")
	router.HandleFunc("/auth/logout", h.Logout).Methods("POST")
	router.HandleFunc("/users/profile", h.SessionAuthMiddleware(h.GetUserProfileBySession)).Methods("GET")
	router.HandleFunc("/users/update", h.SessionAuthMiddleware(h.UpdateUserProfileBySession)).Methods("PUT")
	router.HandleFunc("/users/{username}", h.GetUserPublicProfile).Methods("GET")
}

func (h *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req models.SignInRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := Validate(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.userService.SignIn(&req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Error creating user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := Validate(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.userService.Login(&req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Login failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetUserPublicProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	profile, err := h.userService.GetUserPublicProfile(username)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error retrieving user profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func extractBasicAuth(r *http.Request) (username, password string, ok bool) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", "", false
	}

	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return "", "", false
	}

	encoded := auth[len(prefix):]
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", false
	}

	credentials := string(decoded)
	values := strings.SplitN(credentials, ":", 2)
	if len(values) != 2 {
		return "", "", false
	}

	return values[0], values[1], true
}

func (h *UserHandler) LoginWithSession(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := Validate(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.userService.LoginWithSession(&req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Login failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    response.Token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "Already logged out"})
			return
		}
		http.Error(w, "Error reading cookie", http.StatusBadRequest)
		return
	}

	err = h.userService.Logout(cookie.Value)
	if err != nil {
		http.Error(w, "Logout failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (h *UserHandler) SessionAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthorized: No session token", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Error reading cookie", http.StatusBadRequest)
			return
		}

		session, err := h.userService.ValidateSession(cookie.Value)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "username", session.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *UserHandler) GetUserProfileBySession(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	user, err := h.userService.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "Error retrieving user profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUserProfileBySession(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := Validate(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedUser, err := h.userService.UpdateUserProfileBySession(username, &req)
	if err != nil {
		http.Error(w, "Error updating profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}
