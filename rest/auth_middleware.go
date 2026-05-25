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

// RoleMiddleware returns a middleware that only allows through requests
// where the JWT contains the required role.
// It must be used AFTER JWTMiddleware since it assumes the token is already validated.
func RoleMiddleware(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Read the token from the header — already validated by JWTMiddleware
			tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			secret := os.Getenv("JWT_SECRET")

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			// Extract role from claims
			role, ok := claims["role"].(string)
			if !ok || role != requiredRole {
				// 403 Forbidden = we know WHO you are, but you don't have PERMISSION
				// Different from 401 = we don't know who you are at all
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
