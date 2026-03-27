package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/rbac"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RequireClusterPermission enforces RBAC for routes that include :cluster_id and :namespace
// (or query params cluster_id and namespace as fallback).
func RequireClusterPermission(pool *pgxpool.Pool, minPermission string) gin.HandlerFunc {
	if err := rbac.ParsePermission(minPermission); err != nil {
		panic("middleware.RequireClusterPermission: " + err.Error())
	}
	return func(c *gin.Context) {
		uid := UserID(c)
		if uid == uuid.Nil {
			httputil.AbortWithErrorJSON(c, http.StatusUnauthorized, "not authenticated")
			return
		}
		clusterID := c.Param("cluster_id")
		if clusterID == "" {
			clusterID = c.Query("cluster_id")
		}
		ns := c.Param("namespace")
		if ns == "" {
			ns = c.Query("namespace")
		}
		if clusterID == "" || ns == "" {
			httputil.AbortWithErrorJSON(c, http.StatusBadRequest, "cluster_id and namespace are required")
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		ok, err := rbac.UserSatisfiesPermission(ctx, pool, uid, clusterID, ns, minPermission)
		if err != nil {
			httputil.AbortWithErrorJSON(c, http.StatusInternalServerError, "permission check failed")
			return
		}
		if !ok {
			httputil.AbortWithErrorJSON(c, http.StatusForbidden, "insufficient permission for cluster/namespace")
			return
		}
		c.Next()
	}
}
