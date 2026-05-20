package rest

import (
	"database/sql"
	"fmt"
)

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{db: db}
}

func (s *PostgresUserStore) CreateUser(email, passwordHash string) error {
	_, err := s.db.Exec(`
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
	`, email, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (s *PostgresUserStore) GetUserByEmail(email string) (User, error) {
	var u User
	err := s.db.QueryRow(`
		SELECT id, email, password_hash
		FROM users
		WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash)

	if err == sql.ErrNoRows {
		return User{}, fmt.Errorf("user not found")
	}
	if err != nil {
		return User{}, fmt.Errorf("failed to fetch user: %w", err)
	}
	return u, nil
}
