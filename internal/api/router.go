package api

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/middleware"
	"github.com/AbhishekSharmaIE/Kubevision/internal/auth"
	"github.com/AbhishekSharmaIE/Kubevision/internal/k8s"
	"github.com/AbhishekSharmaIE/Kubevision/internal/metrics"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// RouterConfig wires core dependencies into the HTTP API.
type RouterConfig struct {
	ClusterManager   *k8s.ClusterManager
	MetricsCollector *metrics.Collector
	DB               *pgxpool.Pool
	Redis            *redis.Client
	JWT              *auth.JWT
}

// NewRouter returns the Gin engine with health, metrics, and API stubs.
func NewRouter(cfg RouterConfig) *gin.Engine {
	if os.Getenv("LOG_LEVEL") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.RequestID())
	if cfg.Redis != nil {
		r.Use(middleware.RateLimit(cfg.Redis))
	}
	r.Use(middleware.CORS())
	if cfg.DB != nil {
		r.Use(middleware.Audit(cfg.DB))
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":    "KubeVision",
			"version": os.Getenv("APP_VERSION"),
			"docs":    "https://github.com/AbhishekSharmaIE/Kubevision",
		})
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if cfg.DB != nil {
			if err := cfg.DB.Ping(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "postgres": err.Error()})
				return
			}
		}
		if cfg.Redis != nil {
			if err := cfg.Redis.Ping(ctx).Err(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "redis": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/version", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": gin.H{"version": os.Getenv("APP_VERSION")}})
		})
	}

	registerAuthRoutes(v1, cfg)
	registerClusterRoutes(v1, cfg)
	registerPhase2Routes(v1, cfg)

	return r
}

// ListenAddr returns host:port from PORT (default 8080).
func ListenAddr() string {
	p := os.Getenv("PORT")
	if p == "" {
		p = "8080"
	}
	if _, err := strconv.Atoi(p); err != nil {
		return ":8080"
	}
	return ":" + p
}
