package rest

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// holds the userstore and handlers register + login
type authHandler struct {
	store UserStore
}

// this is what the user sends, pretty simple so far
type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// what we send back
type tokenResponse struct {
	Token string `json:"token"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// this handles creating a new user
// post /auth/register
// Body: {email, password}

func (a *authHandler) handleRegister(w http.ResponseWriter, r *http.Request) {

	var req authRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// bcrypt hashes the password with a random salt (automatically)
	// cost of 14 means it takes ~100ms to hash - slow enough to resist brute force,
	// but fast enough for users. Never stores the plain text password.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		http.Error(w, "failed to process the password", http.StatusInternalServerError)
		return
	}

	role := "user"
	if r.URL.Query().Get("role") == "admin" {
		role = "admin"
	}

	if err := a.store.CreateUser(req.Email, string(hash), role); err != nil {
		log.Printf("Create user error : %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (a *authHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := a.store.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Access token — short lived, 15 minutes
	accessToken, err := generateAccessToken(user)
	if err != nil {
		http.Error(w, "failed to generate access token", http.StatusInternalServerError)
		return
	}

	// Refresh token — just a random string, not a JWT
	// We store it in the DB so we can validate and revoke it
	// crypto/rand gives us cryptographically secure random bytes — not predictable
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		http.Error(w, "failed to generate refresh token", http.StatusInternalServerError)
		return
	}
	// encode to hex so it's a safe printable string
	refreshTokenString := hex.EncodeToString(refreshTokenBytes)

	// Save refresh token to DB with 7 day expiry
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := a.store.SaveRefreshToken(user.ID, refreshTokenString, expiresAt); err != nil {
		http.Error(w, "failed to save refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	})
}

// handleRefresh validates a refresh token and issues a new access token
// POST /auth/refresh
// Body: {"refresh_token": "..."}
func (a *authHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Look up the token in the DB
	rt, err := a.store.GetRefreshToken(body.RefreshToken)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Check if it's expired — the DB stores expiry but we check it in Go
	if time.Now().After(rt.ExpiresAt) {
		// Clean up the expired token
		a.store.DeleteRefreshToken(body.RefreshToken)
		http.Error(w, "refresh token expired", http.StatusUnauthorized)
		return
	}

	// Fetch the user to get their current role
	// We need this to put the correct role in the new access token
	user, err := a.store.GetUserById(rt.UserId)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Issue a fresh access token
	accessToken, err := generateAccessToken(user)
	if err != nil {
		http.Error(w, "failed to generate access token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
	})
}

// handleLogout deletes the refresh token — true server side logout
// POST /auth/logout
// Body: {"refresh_token": "..."}
func (a *authHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// We don't error if the token doesn't exist — logout should always succeed
	// from the client's perspective
	a.store.DeleteRefreshToken(body.RefreshToken)

	w.WriteHeader(http.StatusOK)
}

// generateAccessToken is a helper used by both handleLogin and handleRefresh
// Keeps the JWT creation logic in one place instead of duplicated
func generateAccessToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}

	return token.SignedString([]byte(secret))
}
