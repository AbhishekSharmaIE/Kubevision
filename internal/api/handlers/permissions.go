package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/middleware"
	"github.com/AbhishekSharmaIE/Kubevision/internal/rbac"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PermissionHandler serves /api/v1/permissions.
type PermissionHandler struct {
	pool *pgxpool.Pool
}

// NewPermissionHandler constructs PermissionHandler.
func NewPermissionHandler(pool *pgxpool.Pool) *PermissionHandler {
	return &PermissionHandler{pool: pool}
}

type createPermBody struct {
	TeamID      string `json:"teamId"`
	ClusterID   string `json:"clusterId"`
	Namespace   string `json:"namespace"`
	Permission  string `json:"permission"`
}

type updatePermBody struct {
	ClusterID  *string `json:"clusterId,omitempty"`
	Namespace  *string `json:"namespace,omitempty"`
	Permission *string `json:"permission,omitempty"`
}

// List handles GET /permissions.
func (h *PermissionHandler) List(c *gin.Context) {
	teamFilter := strings.TrimSpace(c.Query("team_id"))
	clusterFilter := strings.TrimSpace(c.Query("cluster_id"))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	var rows pgx.Rows
	var err error
	if middleware.IsAdmin(c) {
		q := `SELECT id, team_id, cluster_id, namespace, permission FROM cluster_permissions WHERE 1=1`
		args := []any{}
		n := 1
		if teamFilter != "" {
			tid, err := uuid.Parse(teamFilter)
			if err != nil {
				httputil.ErrorJSON(c, http.StatusBadRequest, "invalid team_id filter")
				return
			}
			q += ` AND team_id = $` + strconv.Itoa(n)
			args = append(args, tid)
			n++
		}
		if clusterFilter != "" {
			q += ` AND cluster_id = $` + strconv.Itoa(n)
			args = append(args, clusterFilter)
		}
		q += ` ORDER BY cluster_id, namespace`
		rows, err = h.pool.Query(ctx, q, args...)
	} else {
		uid := middleware.UserID(c)
		q := `
SELECT cp.id, cp.team_id, cp.cluster_id, cp.namespace, cp.permission
FROM cluster_permissions cp
INNER JOIN team_members tm ON tm.team_id = cp.team_id
WHERE tm.user_id = $1`
		args := []any{uid}
		n := 2
		if teamFilter != "" {
			tid, err := uuid.Parse(teamFilter)
			if err != nil {
				httputil.ErrorJSON(c, http.StatusBadRequest, "invalid team_id filter")
				return
			}
			q += ` AND cp.team_id = $` + strconv.Itoa(n)
			args = append(args, tid)
			n++
		}
		if clusterFilter != "" {
			q += ` AND cp.cluster_id = $` + strconv.Itoa(n)
			args = append(args, clusterFilter)
		}
		q += ` ORDER BY cp.cluster_id, cp.namespace`
		rows, err = h.pool.Query(ctx, q, args...)
	}
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to list permissions")
		return
	}
	defer rows.Close()

	list := make([]gin.H, 0)
	for rows.Next() {
		var id, teamID uuid.UUID
		var clusterID, ns, perm string
		if err := rows.Scan(&id, &teamID, &clusterID, &ns, &perm); err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to read permissions")
			return
		}
		list = append(list, gin.H{
			"id": id.String(), "teamId": teamID.String(), "clusterId": clusterID,
			"namespace": ns, "permission": perm,
		})
	}
	httputil.DataJSON(c, http.StatusOK, list)
}

// Check handles GET /permissions/check — evaluates RBAC for the current user.
func (h *PermissionHandler) Check(c *gin.Context) {
	clusterID := strings.TrimSpace(c.Query("cluster_id"))
	ns := strings.TrimSpace(c.Query("namespace"))
	need := strings.TrimSpace(c.Query("permission"))
	if clusterID == "" || ns == "" || need == "" {
		httputil.ErrorJSON(c, http.StatusBadRequest, "cluster_id, namespace, and permission query params are required")
		return
	}
	if err := validatePermissionString(need); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	ok, err := rbac.UserSatisfiesPermission(ctx, h.pool, middleware.UserID(c), clusterID, ns, need)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "permission evaluation failed")
		return
	}
	httputil.DataJSON(c, http.StatusOK, gin.H{"allowed": ok, "clusterId": clusterID, "namespace": ns, "permission": need})
}

// Get handles GET /permissions/:id.
func (h *PermissionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid permission id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var teamID uuid.UUID
	var clusterID, ns, perm string
	err = h.pool.QueryRow(ctx, `
		SELECT team_id, cluster_id, namespace, permission FROM cluster_permissions WHERE id = $1`, id).
		Scan(&teamID, &clusterID, &ns, &perm)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.ErrorJSON(c, http.StatusNotFound, "permission not found")
		return
	}
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to load permission")
		return
	}

	if !middleware.IsAdmin(c) {
		ok, err := h.userSeesTeam(ctx, middleware.UserID(c), teamID)
		if err != nil || !ok {
			httputil.ErrorJSON(c, http.StatusForbidden, "cannot view this permission")
			return
		}
	}

	httputil.DataJSON(c, http.StatusOK, gin.H{
		"id": id.String(), "teamId": teamID.String(), "clusterId": clusterID,
		"namespace": ns, "permission": perm,
	})
}

func (h *PermissionHandler) userSeesTeam(ctx context.Context, userID, teamID uuid.UUID) (bool, error) {
	var ok bool
	err := h.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM team_members WHERE user_id = $1 AND team_id = $2)`, userID, teamID).Scan(&ok)
	return ok, err
}

// Create handles POST /permissions (admin).
func (h *PermissionHandler) Create(c *gin.Context) {
	var body createPermBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid JSON body")
		return
	}
	teamID, err := uuid.Parse(strings.TrimSpace(body.TeamID))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid teamId")
		return
	}
	clusterID := strings.TrimSpace(body.ClusterID)
	if clusterID == "" || len(clusterID) > 255 {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid clusterId")
		return
	}
	ns := strings.TrimSpace(body.Namespace)
	if ns == "" || len(ns) > 253 {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "namespace required (use * for all namespaces)")
		return
	}
	if err := validatePermissionString(body.Permission); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var teamExists bool
	_ = h.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE id = $1)`, teamID).Scan(&teamExists)
	if !teamExists {
		httputil.ErrorJSON(c, http.StatusBadRequest, "team does not exist")
		return
	}

	var id uuid.UUID
	err = h.pool.QueryRow(ctx, `
		INSERT INTO cluster_permissions (team_id, cluster_id, namespace, permission)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (team_id, cluster_id, namespace) DO UPDATE SET permission = EXCLUDED.permission
		RETURNING id`,
		teamID, clusterID, ns, body.Permission).Scan(&id)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "could not create permission")
		return
	}

	httputil.DataJSON(c, http.StatusCreated, gin.H{
		"id": id.String(), "teamId": teamID.String(), "clusterId": clusterID,
		"namespace": ns, "permission": body.Permission,
	})
}

// Update handles PUT /permissions/:id (admin).
func (h *PermissionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid permission id")
		return
	}
	var body updatePermBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid JSON body")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if body.ClusterID != nil {
		v := strings.TrimSpace(*body.ClusterID)
		if v == "" || len(v) > 255 {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid clusterId")
			return
		}
		_, err = h.pool.Exec(ctx, `UPDATE cluster_permissions SET cluster_id = $1 WHERE id = $2`, v, id)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "could not update cluster_id")
			return
		}
	}
	if body.Namespace != nil {
		v := strings.TrimSpace(*body.Namespace)
		if v == "" {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid namespace")
			return
		}
		_, err = h.pool.Exec(ctx, `UPDATE cluster_permissions SET namespace = $1 WHERE id = $2`, v, id)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "could not update namespace")
			return
		}
	}
	if body.Permission != nil {
		if err := validatePermissionString(*body.Permission); err != nil {
			httputil.ErrorJSON(c, http.StatusBadRequest, err.Error())
			return
		}
		_, err = h.pool.Exec(ctx, `UPDATE cluster_permissions SET permission = $1 WHERE id = $2`, *body.Permission, id)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "could not update permission")
			return
		}
	}

	h.Get(c)
}

// Delete handles DELETE /permissions/:id (admin).
func (h *PermissionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid permission id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	cmd, err := h.pool.Exec(ctx, `DELETE FROM cluster_permissions WHERE id = $1`, id)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to delete permission")
		return
	}
	if cmd.RowsAffected() == 0 {
		httputil.ErrorJSON(c, http.StatusNotFound, "permission not found")
		return
	}
	c.Status(http.StatusNoContent)
}
