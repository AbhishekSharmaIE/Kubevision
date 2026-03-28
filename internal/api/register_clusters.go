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
	hh := handlers.NewHelmHandler(cfg.ClusterManager)

	g := v1.Group("")
	g.Use(middleware.RequireAuth(cfg.JWT, cfg.Redis, cfg.DB))

	g.GET("/clusters", ch.List)
	g.POST("/clusters", middleware.RequireAdmin(), ch.Create)

	// More specific paths before /clusters/:id
	g.GET("/clusters/:id/namespaces/:namespace/pods/:pod/logs",
		middleware.RequireClusterPermissionByID(cfg.DB, "id", "namespace", rbac.PermRead),
		rh.GetPodLogs)
	g.GET("/clusters/:id/namespaces/:namespace/pods/:pod",
		middleware.RequireClusterPermissionByID(cfg.DB, "id", "namespace", rbac.PermRead),
		rh.GetPod)
	g.GET("/clusters/:id/namespaces/:namespace/deployments/:deployment",
		middleware.RequireClusterPermissionByID(cfg.DB, "id", "namespace", rbac.PermRead),
		rh.GetDeployment)
	g.PATCH("/clusters/:id/namespaces/:namespace/deployments/:deployment/scale",
		middleware.RequireClusterPermissionByID(cfg.DB, "id", "namespace", rbac.PermWrite),
		rh.ScaleDeployment)
	g.POST("/clusters/:id/namespaces/:namespace/deployments/:deployment/restart",
		middleware.RequireClusterPermissionByID(cfg.DB, "id", "namespace", rbac.PermWrite),
		rh.RestartDeployment)
	g.GET("/clusters/:id/namespaces/:namespace/helm/releases/:release/history",
		middleware.RequireClusterPermissionByID(cfg.DB, "id", "namespace", rbac.PermRead),
		hh.HelmReleaseHistory)
	g.GET("/clusters/:id/namespaces/:namespace/deployments",
		middleware.RequireClusterPermissionByID(cfg.DB, "id", "namespace", rbac.PermRead),
		rh.ListDeployments)
	g.GET("/clusters/:id/namespaces/:namespace/pods",
		middleware.RequireClusterPermissionByID(cfg.DB, "id", "namespace", rbac.PermRead),
		rh.ListPods)
	g.GET("/clusters/:id/events",
		middleware.RequireClusterScopePermission(cfg.DB, "id", rbac.PermRead),
		rh.ListEvents)
	g.GET("/clusters/:id/namespaces",
		middleware.RequireClusterScopePermission(cfg.DB, "id", rbac.PermRead),
		rh.ListNamespaces)
	g.GET("/clusters/:id/nodes",
		middleware.RequireClusterScopePermission(cfg.DB, "id", rbac.PermRead),
		rh.ListNodes)
	g.GET("/clusters/:id/helm/releases",
		middleware.RequireClusterScopePermission(cfg.DB, "id", rbac.PermRead),
		hh.ListHelmReleases)

	g.GET("/clusters/:id", ch.Get)
	g.DELETE("/clusters/:id", middleware.RequireAdmin(), ch.Delete)
}
