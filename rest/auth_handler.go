package rest

import (
	"encoding/json"
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

	if err := a.store.CreateUser(req.Email, string(hash)); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (a *authHandler) handleLogin(w http.ResponseWriter, r *http.Request) {

	var req authRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := a.store.GetUserByEmail(req.Email)
	if err != nil {
		// we dont reveal whether the email exists or not, "invalid credentials" covers everything
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// returns nil if the password matches else an error is returned
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,

		// exp is the time of expiry of the token
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	// sign the token with our secret key using HMAC-SHA256
	// Anyone can READ a JWT
	// but only someone with the secret key can sign it
	// that's how we know a token was issued by us and not forged

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		http.Error(w, "Failed to sign the token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResponse{Token: tokenString})

}
