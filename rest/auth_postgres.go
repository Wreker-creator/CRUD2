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
