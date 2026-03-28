package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Context keys for auth (JWT phase).
const (
	ContextUserID    = "authUserID"
	ContextUserEmail = "authUserEmail"
	ContextUserName  = "authUserName"
	ContextIsAdmin   = "authIsAdmin"
	ContextClaims    = "authClaims"
)

func setAuthContext(c *gin.Context, claims *auth.Claims, emailFromDB, nameFromDB string) {
	uid, _ := uuid.Parse(claims.UserID)
	c.Set(ContextClaims, claims)
	c.Set(ContextUserID, uid)
	if emailFromDB != "" {
		c.Set(ContextUserEmail, emailFromDB)
	} else {
		c.Set(ContextUserEmail, claims.Email)
	}
	c.Set(ContextUserName, nameFromDB)
	c.Set(ContextIsAdmin, claims.IsAdmin)
}

// RequireAuth validates JWT + Redis session, then loads fresh user row into context.
func RequireAuth(jwtSvc *auth.JWT, rdb *redis.Client, pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c)
		if !ok {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "missing Authorization bearer token")
			return
		}
		claims, err := jwtSvc.ValidateToken(token)
		if err != nil {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		uid, err := uuid.Parse(claims.UserID)
		if err != nil || claims.RegisteredClaims.ID == "" {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "invalid token claims")
			return
		}
		if !auth.ValidateSession(c.Request.Context(), rdb, uid, claims.RegisteredClaims.ID) {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "session revoked or expired")
			return
		}
		var email, name string
		var isAdmin bool
		err = pool.QueryRow(c.Request.Context(), `
			SELECT email, name, COALESCE(is_admin, false) FROM users WHERE id = $1`, uid).
			Scan(&email, &name, &isAdmin)
		if err != nil {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "user not found")
			return
		}
		claims.IsAdmin = isAdmin
		claims.Email = email
		setAuthContext(c, claims, email, name)
		c.Next()
	}
}

// UserCreateAuth skips auth when there are zero users (bootstrap); otherwise enforces RequireAuth.
func UserCreateAuth(jwtSvc *auth.JWT, rdb *redis.Client, pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		var n int64
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
			httputil.AbortWithErrorJSON(c, http.StatusInternalServerError, "database error")
			return
		}
		if n == 0 {
			c.Next()
			return
		}
		RequireAuth(jwtSvc, rdb, pool)(c)
	}
}

func bearerToken(c *gin.Context) (string, bool) {
	raw := c.GetHeader("Authorization")
	const p = "Bearer "
	if !strings.HasPrefix(raw, p) {
		return "", false
	}
	t := strings.TrimSpace(strings.TrimPrefix(raw, p))
	return t, t != ""
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

// GetClaims returns JWT claims from context (nil if unauthenticated).
func GetClaims(c *gin.Context) *auth.Claims {
	v, ok := c.Get(ContextClaims)
	if !ok {
		return nil
	}
	cl, _ := v.(*auth.Claims)
	return cl
}
