package rest

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{db: db}
}

func (s *PostgresUserStore) CreateUser(email, passwordHash string, role string) error {
	_, err := s.db.Exec(`
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
	`, email, passwordHash, role)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (s *PostgresUserStore) GetUserByEmail(email string) (User, error) {
	var u User
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, role
		FROM users
		WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)

	if err == sql.ErrNoRows {
		return User{}, fmt.Errorf("user not found")
	}
	if err != nil {
		return User{}, fmt.Errorf("failed to fetch user: %w", err)
	}
	return u, nil
}

func (s *PostgresUserStore) SaveRefreshToken(userId int, token string, expiresAt time.Time) error {

	_, err := s.db.Exec(`
	INSERT INTO refresh_tokens (token, user_id, expires_at)
	VALUES ($1, $2, $3)`, token, userId, expiresAt)

	if err != nil {
		return fmt.Errorf("Failed to save the refresh token : %w", err)
	}

	return nil

}

func (s *PostgresUserStore) GetRefreshToken(token string) (RefreshToken, error) {

	var r RefreshToken

	err := s.db.QueryRow(`
		SELECT id, token, user_id, expires_at
		FROM refresh_tokens
		WHERE token = $1
	`, token).Scan(&r.ID, &r.Token, &r.UserId, &r.ExpiresAt)

	if errors.Is(sql.ErrNoRows, err) {
		return RefreshToken{}, fmt.Errorf("Refresh token not found")
	}

	if err != nil {
		return RefreshToken{}, fmt.Errorf("Failed to fetch the refresh token : %w", err)
	}

	return r, nil

}

func (s *PostgresUserStore) DeleteRefreshToken(token string) error {

	_, err := s.db.Exec(`
	DELETE FROM refresh_tokens WHERE token = $1
	`, token)

	if err != nil {
		return fmt.Errorf("Failed to delete the refresh token : %w", err)
	}

	return nil

}

func (s *PostgresUserStore) GetUserById(id int) (User, error) {

	var u User

	err := s.db.QueryRow(`
	SELECT id, email, password_hash, role 
	FROM users
	WHERE id = $1`, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)

	if errors.Is(sql.ErrNoRows, err) {
		return User{}, fmt.Errorf("User not found")
	}

	if err != nil {
		return User{}, fmt.Errorf("Failed to fetch user : %w", err)
	}

	return u, nil

}
