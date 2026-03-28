package api

import (
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/handlers"
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

func registerClusterRoutes(v1 *gin.RouterGroup, cfg RouterConfig) {
	if cfg.DB == nil || cfg.JWT == nil || cfg.Redis == nil || cfg.ClusterManager == nil {
		return
	}
	ch := handlers.NewClusterHandler(cfg.DB, cfg.ClusterManager)

	g := v1.Group("")
	g.Use(middleware.RequireAuth(cfg.JWT, cfg.Redis, cfg.DB))

	g.GET("/clusters", ch.List)
	g.POST("/clusters", middleware.RequireAdmin(), ch.Create)
	g.GET("/clusters/:id", ch.Get)
	g.DELETE("/clusters/:id", middleware.RequireAdmin(), ch.Delete)
}
