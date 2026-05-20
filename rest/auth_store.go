package rest

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // don't return the password hash in the json response.
}

type UserStore interface {
	CreateUser(email, passwordHash string) error
	GetUserByEmail(email string) (User, error)
}
