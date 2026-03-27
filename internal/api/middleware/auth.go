package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Context keys for auth.
const (
	ContextUserID    = "authUserID"
	ContextUserEmail = "authUserEmail"
	ContextUserName  = "authUserName"
	ContextIsAdmin   = "authIsAdmin"
)

// BearerUser loads the caller from Authorization: Bearer <user-uuid>.
// Interim identity mechanism until JWT login ships (Phase 2 slice).
func BearerUser(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "missing Authorization header")
			return
		}
		const p = "Bearer "
		if !strings.HasPrefix(raw, p) {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "Authorization must be Bearer <user-id>")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(raw, p))
		uid, err := uuid.Parse(token)
		if err != nil {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "invalid bearer user id")
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		var email, name string
		var isAdmin bool
		err = pool.QueryRow(ctx, `
			SELECT email, name, COALESCE(is_admin, false) FROM users WHERE id = $1`, uid).
			Scan(&email, &name, &isAdmin)
		if err != nil {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "unknown or inactive user")
			return
		}
		c.Set(ContextUserID, uid)
		c.Set(ContextUserEmail, email)
		c.Set(ContextUserName, name)
		c.Set(ContextIsAdmin, isAdmin)
		c.Next()
	}
}

// RequireAdmin aborts unless the authenticated user is an admin.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(ContextIsAdmin)
		if !ok {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "not authenticated")
			return
		}
		admin, _ := v.(bool)
		if !admin {
			httputil.AbortWithErrorJSON(c, http.StatusForbidden, "admin required")
			return
		}
		c.Next()
	}
}

// UserID returns the authenticated user id or uuid.Nil if missing.
func UserID(c *gin.Context) uuid.UUID {
	v, ok := c.Get(ContextUserID)
	if !ok {
		return uuid.Nil
	}
	u, _ := v.(uuid.UUID)
	return u
}

// IsAdmin returns whether the current user is admin.
func IsAdmin(c *gin.Context) bool {
	v, ok := c.Get(ContextIsAdmin)
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

// OptionalBearerUser loads the user when Authorization is present; does not abort if the header is missing.
func OptionalBearerUser(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			c.Next()
			return
		}
		const p = "Bearer "
		if !strings.HasPrefix(raw, p) {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "Authorization must be Bearer <user-id>")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(raw, p))
		uid, err := uuid.Parse(token)
		if err != nil {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "invalid bearer user id")
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		var email, name string
		var isAdmin bool
		err = pool.QueryRow(ctx, `
			SELECT email, name, COALESCE(is_admin, false) FROM users WHERE id = $1`, uid).
			Scan(&email, &name, &isAdmin)
		if err != nil {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "unknown or inactive user")
			return
		}
		c.Set(ContextUserID, uid)
		c.Set(ContextUserEmail, email)
		c.Set(ContextUserName, name)
		c.Set(ContextIsAdmin, isAdmin)
		c.Next()
	}
}
