package rest

import (
	"database/sql"
	"errors"
	"fmt"
)

type PostgresUserStore struct {
	db *sql.DB
}

// did not add dsn here nor am i establishing another connection to the db because
// in main im already calling NewPostgresFoodStore, that establishes a connection and
// verifies that the db is reachable, i dont need to do that again.
func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{db: db}
}

func (p *PostgresUserStore) CreateUser(email, passwordHash string) error {

	_, err := p.db.Exec(`
	INSERT INTO users (email, password_hash) VALUES ($1, $2)
	`, email, passwordHash)

	if err != nil {
		return fmt.Errorf("Failed to create user : %w", err)
	}

	return nil

}

func (p *PostgresUserStore) GetUserByEmail(email string) (User, error) {

	var u User

	err := p.db.QueryRow(`SELECT FROM users WHERE email = $1`, email).Scan(&u.ID, &u.Email, &u.PasswordHash)

	if errors.Is(sql.ErrNoRows, err) {
		return User{}, fmt.Errorf("user not found")
	}

	if err != nil {
		return User{}, fmt.Errorf("Failed to get user by email, '%w'", err)
	}

	return u, nil

}
