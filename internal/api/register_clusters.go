package api

import (
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/handlers"
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/middleware"
	"github.com/AbhishekSharmaIE/Kubevision/internal/rbac"
	"github.com/gin-gonic/gin"
)

func registerClusterRoutes(v1 *gin.RouterGroup, cfg RouterConfig) {
	if cfg.DB == nil || cfg.JWT == nil || cfg.Redis == nil || cfg.ClusterManager == nil {
		return
	}
	ch := handlers.NewClusterHandler(cfg.DB, cfg.ClusterManager)
	rh := handlers.NewClusterResourcesHandler(cfg.ClusterManager)

	g := v1.Group("")
	g.Use(middleware.RequireAuth(cfg.JWT, cfg.Redis, cfg.DB))

	g.GET("/clusters", ch.List)
	g.POST("/clusters", middleware.RequireAdmin(), ch.Create)

	// More specific paths before /clusters/:id
	g.GET("/clusters/:id/namespaces/:namespace/pods",
		middleware.RequireClusterPermissionByID(cfg.DB, "id", "namespace", rbac.PermRead),
		rh.ListPods)
	g.GET("/clusters/:id/namespaces",
		middleware.RequireClusterScopePermission(cfg.DB, "id", rbac.PermRead),
		rh.ListNamespaces)
	g.GET("/clusters/:id/nodes",
		middleware.RequireClusterScopePermission(cfg.DB, "id", rbac.PermRead),
		rh.ListNodes)

	g.GET("/clusters/:id", ch.Get)
	g.DELETE("/clusters/:id", middleware.RequireAdmin(), ch.Delete)
}
