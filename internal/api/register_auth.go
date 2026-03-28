package api

import (
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/handlers"
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(v1 *gin.RouterGroup, cfg RouterConfig) {
	if cfg.JWT == nil || cfg.DB == nil || cfg.Redis == nil {
		return
	}
	ah := handlers.NewAuthHandler(cfg.DB, cfg.JWT, cfg.Redis)

	v1.POST("/auth/login", ah.Login)
	v1.POST("/auth/refresh", ah.Refresh)
	v1.POST("/auth/logout", middleware.RequireAuth(cfg.JWT, cfg.Redis, cfg.DB), ah.Logout)
}
