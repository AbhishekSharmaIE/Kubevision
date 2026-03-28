package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Audit persists write operations to audit_logs after the handler runs.
func Audit(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		method := c.Request.Method
		switch method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			return
		}

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v1/auth/login") {
			return
		}

		uid := UserID(c)
		var userUUID any
		if uid != uuid.Nil {
			userUUID = uid
		}

		resource := auditResource(path)
		action := auditAction(method)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		ip := net.ParseIP(c.ClientIP())
		_, _ = pool.Exec(ctx, `
			INSERT INTO audit_logs (user_id, cluster_id, namespace, resource, action, ip_address)
			VALUES ($1, NULL, NULL, $2, $3, $4)`,
			userUUID, resource, action, ip)
	}
}

func auditResource(path string) string {
	path = strings.TrimPrefix(path, "/api/v1/")
	if path == "" {
		return "/"
	}
	parts := strings.SplitN(path, "/", 2)
	return parts[0]
}

func auditAction(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return strings.ToLower(method)
	}
}
