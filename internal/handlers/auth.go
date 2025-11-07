package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// contextKey is a personal type for context UUID keys
type contextUUIDKey string

// userUUIDKey is a unique key for user_id in context
const userUUIDKey contextUUIDKey = "user_id"

// authMiddleware used as middleware for authentication
func (h Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_id")
		var (
			tokenString string
			userUUID    uuid.UUID
		)

		if err != nil || cookie == nil {
			// If no cookie - creating new userUUID and JWT token
			userUUID = h.auth.GenerateUserUUID()
			tokenString, err = h.auth.BuildJWTString(userUUID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				h.zlog.Error().Msgf("Error building JWT string: %v", err)
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "user_id",
				Value:    tokenString,
				Path:     "/",
				HttpOnly: true,
				Secure:   false, // true when HTTPS in prod
				SameSite: http.SameSiteLaxMode,
			})
		} else {
			// Checking existing cookie
			userUUID, err = h.auth.ValidateUserUUID(cookie.Value)
			if err != nil {
				h.zlog.Error().Msgf("Error validating user UUID: %v", err)
				http.Error(w, "Invalid cookie", http.StatusUnauthorized)
				return
			}
		}
		ctxUserUUID := userUUID.String()
		ctx := context.WithValue(r.Context(), userUUIDKey, ctxUserUUID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
