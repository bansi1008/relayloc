package auth

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Middleware struct {
	secret []byte
}

func NewMiddleware(secret string) *Middleware {
	return &Middleware{
		secret: []byte(secret),
	}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var tokenString string

		cookie, err := r.Cookie("access_token")
		if err == nil {
			tokenString = cookie.Value
		}

		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")

			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}

			return m.secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		sub, err := token.Claims.GetSubject()
		if err != nil {
			http.Error(w, "invalid token subject", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(sub)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusUnauthorized)
			return
		}

		ctx := ContextWithUserID(r.Context(), userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
