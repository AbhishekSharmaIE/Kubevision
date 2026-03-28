package api

import (
	"net/http"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/handlers"
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/middleware"
	"github.com/AbhishekSharmaIE/Kubevision/internal/rbac"
	"github.com/gin-gonic/gin"
)

// registerPhase2Routes registers users, teams, permissions, Helm repos, and RBAC probe routes under /api/v1.
func registerPhase2Routes(v1 *gin.RouterGroup, cfg RouterConfig) {
	if cfg.DB == nil || cfg.JWT == nil || cfg.Redis == nil {
		return
	}

	uh := handlers.NewUserHandler(cfg.DB)
	th := handlers.NewTeamHandler(cfg.DB)
	ph := handlers.NewPermissionHandler(cfg.DB)
	hr := handlers.NewHelmReposHandler()

	// Bootstrap: first user without auth; later creates need admin JWT.
	v1.POST("/users", middleware.UserCreateAuth(cfg.JWT, cfg.Redis, cfg.DB), uh.Create)

	auth := v1.Group("")
	auth.Use(middleware.RequireAuth(cfg.JWT, cfg.Redis, cfg.DB))

	auth.GET("/users", middleware.RequireAdmin(), uh.List)
	auth.GET("/users/:id", uh.Get)
	auth.PUT("/users/:id", uh.Update)
	auth.DELETE("/users/:id", middleware.RequireAdmin(), uh.Delete)

	auth.GET("/teams", th.List)
	auth.POST("/teams", middleware.RequireAdmin(), th.Create)
	auth.GET("/teams/:id/members", th.ListMembers)
	auth.POST("/teams/:id/members", middleware.RequireAdmin(), th.AddMember)
	auth.DELETE("/teams/:id/members/:userId", middleware.RequireAdmin(), th.RemoveMember)
	auth.GET("/teams/:id", th.Get)
	auth.PUT("/teams/:id", middleware.RequireAdmin(), th.Update)
	auth.DELETE("/teams/:id", middleware.RequireAdmin(), th.Delete)

	auth.GET("/permissions/check", ph.Check)
	auth.GET("/permissions", ph.List)
	auth.GET("/permissions/:id", ph.Get)
	auth.POST("/permissions", middleware.RequireAdmin(), ph.Create)
	auth.PUT("/permissions/:id", middleware.RequireAdmin(), ph.Update)
	auth.DELETE("/permissions/:id", middleware.RequireAdmin(), ph.Delete)

	auth.GET("/helm/repos", hr.ListHelmRepos)
	auth.GET("/helm/repos/:repo/charts", hr.ListHelmRepoCharts)
	auth.POST("/helm/repos", middleware.RequireAdmin(), hr.AddHelmRepo)

	// Example cluster route protected by RBAC (read on namespace).
	auth.GET("/rbac/probe/:cluster_id/:namespace",
		middleware.RequireClusterPermission(cfg.DB, rbac.PermRead),
		func(c *gin.Context) {
			httputil.DataJSON(c, http.StatusOK, gin.H{
				"clusterId": c.Param("cluster_id"),
				"namespace": c.Param("namespace"),
				"message":   "RBAC check passed",
			})
		},
	)
}
