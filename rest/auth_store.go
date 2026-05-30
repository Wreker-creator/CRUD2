package rest

import "time"

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // don't return the password hash in the json response.
	Role         string `json:"role"`
}

// mapped to our refresh tokens table, 005 in migrations
type RefreshToken struct {
	ID        int
	Token     string
	UserId    int
	ExpiresAt time.Time
}

type UserStore interface {
	CreateUser(email, passwordHash string, role string) error
	GetUserByEmail(email string) (User, error)
	GetUserById(id int) (User, error)

	// for refresh token
	SaveRefreshToken(userId int, token string, expiresAt time.Time) error
	GetRefreshToken(token string) (RefreshToken, error)
	DeleteRefreshToken(token string) error
}
