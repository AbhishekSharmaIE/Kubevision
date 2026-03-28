package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const defaultPerMinute = 100

// RateLimit applies a per-IP fixed window counter in Redis (100 req/min by default).
// Skips /healthz, /readyz, and /metrics.
func RateLimit(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/healthz" || path == "/readyz" || path == "/metrics" {
			c.Next()
			return
		}
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}
		bucket := time.Now().Unix() / 60
		key := "kv:rl:" + ip + ":" + strconv.FormatInt(bucket, 10)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		n, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if n == 1 {
			_ = rdb.Expire(ctx, key, 2*time.Minute).Err()
		}
		limit := defaultPerMinute
		if n > int64(limit) {
			c.Header("Retry-After", "60")
			httputil.AbortWithErrorJSON(c, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		c.Next()
	}
}
