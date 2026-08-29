package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

type Middleware struct {
	Service *AuthService
}

func NewMiddleware(service *AuthService) *Middleware {
	return &Middleware{Service: service}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header missing"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "Invalid authorization format, must be 'Bearer <token>'"}`, http.StatusUnauthorized)
			return
		}

		claims, err := m.Service.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(RoleKey).(string)
			if !ok || userRole == "" {
				http.Error(w, `{"error": "Forbidden: missing authentication context"}`, http.StatusForbidden)
				return
			}
			for _, role := range allowedRoles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error": "Forbidden: insufficient role permissions"}`, http.StatusForbidden)
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (int64, error) {
	id, ok := ctx.Value(UserIDKey).(int64)
	if !ok {
		return 0, errors.New("user ID not found in context")
	}
	return id, nil
}

func GetRoleFromContext(ctx context.Context) (string, error) {
	role, ok := ctx.Value(RoleKey).(string)
	if !ok {
		return "", errors.New("role not found in context")
	}
	return role, nil
}
