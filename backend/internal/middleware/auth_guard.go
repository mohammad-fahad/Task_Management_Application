package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/taskmanager/backend/internal/models"
	"github.com/taskmanager/backend/internal/service"
)

// Context keys for storing authenticated user information.
type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

// AuthGuard middleware validates JWT from HTTP-Only cookie or Authorization header.
type AuthGuard struct {
	authService *service.AuthService
}

// NewAuthGuard creates a new AuthGuard middleware.
func NewAuthGuard(authService *service.AuthService) *AuthGuard {
	return &AuthGuard{authService: authService}
}

// Authenticate is a middleware that validates the JWT and injects user info into the context.
func (m *AuthGuard) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := m.extractToken(r)
		if tokenString == "" {
			writeAuthError(w, "authentication required")
			return
		}

		userID, role, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			writeAuthError(w, "invalid or expired token")
			return
		}

		// Inject user info into request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserRoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns a middleware that checks for specific roles.
func (m *AuthGuard) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				writeAuthError(w, "authentication required")
				return
			}

			hasRole := false
			for _, allowed := range roles {
				if role == allowed {
					hasRole = true
					break
				}
			}

			if !hasRole {
				writeForbiddenError(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetUserRole extracts the user role from the request context.
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}

// extractToken tries to get the JWT from cookie first, then Authorization header.
func (m *AuthGuard) extractToken(r *http.Request) string {
	// Try cookie first
	cookie, err := r.Cookie(m.authService.GetCookieName())
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Fallback to Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// writeAuthError writes a 401 Unauthorized JSON response.
func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Error:   message,
	})
}

// writeForbiddenError writes a 403 Forbidden JSON response.
func writeForbiddenError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Error:   message,
	})
}