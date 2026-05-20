package rest

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// validates the token on every protected request.
// it wraps any http.Handler and returns the handler ONLY if the token is valid
func JWTMiddleWare(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing auth header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "invalid authorisation format", http.StatusUnauthorized)
			return
		}

		secret := os.Getenv("JWT_SECRET")

		// parse and validate the token.
		// this checks the signature and the expiry time automatically

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {

			// verify the signing method is what we expect
			// rejecting other methods prevents algorithm confusing attacks
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(secret), nil

		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// token is valid - pass the request to the actual handler
		next.ServeHTTP(w, r)

	})

}
