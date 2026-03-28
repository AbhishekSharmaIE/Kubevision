package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/api/middleware"
	"github.com/AbhishekSharmaIE/Kubevision/internal/k8s"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterHandler manages registered Kubernetes clusters (DB + in-memory clients).
type ClusterHandler struct {
	pool *pgxpool.Pool
	mgr  *k8s.ClusterManager
}

// NewClusterHandler constructs ClusterHandler.
func NewClusterHandler(pool *pgxpool.Pool, mgr *k8s.ClusterManager) *ClusterHandler {
	return &ClusterHandler{pool: pool, mgr: mgr}
}

type createClusterJSON struct {
	Name             string `json:"name"`
	Environment      string `json:"environment"`
	KubeconfigBase64 string `json:"kubeconfigBase64"`
	Kubeconfig       string `json:"kubeconfig"`
}

// List returns clusters visible to the caller (admin: all; others: via team RBAC rows).
func (h *ClusterHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	uid := middleware.UserID(c)
	var rows pgx.Rows
	var err error
	if middleware.IsAdmin(c) {
		rows, err = h.pool.Query(ctx, `
			SELECT id::text, name, environment, server_url, version, created_at
			FROM clusters ORDER BY name`)
	} else {
		rows, err = h.pool.Query(ctx, `
			SELECT DISTINCT c.id::text, c.name, c.environment, c.server_url, c.version, c.created_at
			FROM clusters c
			INNER JOIN cluster_permissions cp ON cp.cluster_id = c.id::text
			INNER JOIN team_members tm ON tm.team_id = cp.team_id
			WHERE tm.user_id = $1
			ORDER BY c.name`, uid)
	}
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to list clusters")
		return
	}
	defer rows.Close()

	out := make([]gin.H, 0)
	for rows.Next() {
		var id, name, env, serverURL, ver string
		var created time.Time
		if err := rows.Scan(&id, &name, &env, &serverURL, &ver, &created); err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to read clusters")
			return
		}
		_, err := h.mgr.GetCluster(id)
		reachable := err == nil
		out = append(out, gin.H{
			"id": id, "name": name, "environment": env, "serverUrl": serverURL,
			"version": ver, "createdAt": created.UTC().Format(time.RFC3339),
			"reachable": reachable,
		})
	}
	if err := rows.Err(); err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to list clusters")
		return
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out)})
}

// Create registers a cluster (admin). Accepts multipart file `kubeconfig` or JSON body.
func (h *ClusterHandler) Create(c *gin.Context) {
	var kubeBytes []byte
	name := strings.TrimSpace(c.PostForm("name"))
	env := strings.TrimSpace(c.PostForm("environment"))

	if fh, err := c.FormFile("kubeconfig"); err == nil && fh != nil {
		f, err := fh.Open()
		if err != nil {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "could not read kubeconfig file")
			return
		}
		defer f.Close()
		kubeBytes, err = io.ReadAll(io.LimitReader(f, 1<<22))
		if err != nil {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "could not read kubeconfig file")
			return
		}
	} else {
		var body createClusterJSON
		if err := c.ShouldBindJSON(&body); err != nil {
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "send multipart kubeconfig file or JSON with name, environment, kubeconfigBase64")
			return
		}
		name = strings.TrimSpace(body.Name)
		env = strings.TrimSpace(body.Environment)
		switch {
		case body.KubeconfigBase64 != "":
			b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(body.KubeconfigBase64))
			if err != nil {
				httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid kubeconfigBase64")
				return
			}
			kubeBytes = b
		case strings.TrimSpace(body.Kubeconfig) != "":
			kubeBytes = []byte(body.Kubeconfig)
		default:
			httputil.ErrorJSON(c, httputil.StatusUnprocessable, "kubeconfig or kubeconfigBase64 required")
			return
		}
	}

	if !validateName(name) {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid name")
		return
	}
	if env == "" || len(env) > 50 {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid environment (max 50 chars)")
		return
	}
	if len(kubeBytes) == 0 {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "empty kubeconfig")
		return
	}

	serverURL, liveVer, err := peekClusterServer(kubeBytes)
	if err != nil {
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "invalid kubeconfig: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	addedBy := middleware.UserID(c)
	var rowID string
	err = h.pool.QueryRow(ctx, `
		INSERT INTO clusters (name, environment, kubeconfig, server_url, version, added_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text`,
		name, env, kubeBytes, serverURL, liveVer, addedBy).Scan(&rowID)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusConflict, "could not save cluster")
		return
	}

	if err := h.mgr.AddCluster(ctx, rowID, name, env, kubeBytes); err != nil {
		_, _ = h.pool.Exec(ctx, `DELETE FROM clusters WHERE id = $1::uuid`, rowID)
		httputil.ErrorJSON(c, httputil.StatusUnprocessable, "cluster not reachable: "+err.Error())
		return
	}

	httputil.DataJSON(c, http.StatusCreated, gin.H{
		"id": rowID, "name": name, "environment": env, "serverUrl": serverURL,
		"version": liveVer, "reachable": true,
	})
}

func peekClusterServer(kubeconfig []byte) (serverURL, version string, err error) {
	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return "", "", err
	}
	serverURL = cfg.Host
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return serverURL, "", err
	}
	v, err := cs.ServerVersion()
	if err != nil {
		return serverURL, "", err
	}
	return serverURL, v.GitVersion, nil
}

// Get returns cluster metadata and lightweight live stats when the API is reachable.
func (h *ClusterHandler) Get(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if _, err := uuid.Parse(id); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid cluster id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()

	uid := middleware.UserID(c)
	if !h.canAccessCluster(ctx, id, uid, middleware.IsAdmin(c)) {
		httputil.ErrorJSON(c, http.StatusForbidden, "no access to this cluster")
		return
	}

	var name, env, serverURL, ver string
	var created time.Time
	err := h.pool.QueryRow(ctx, `
		SELECT name, environment, server_url, version, created_at
		FROM clusters WHERE id = $1::uuid`, id).Scan(&name, &env, &serverURL, &ver, &created)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.ErrorJSON(c, http.StatusNotFound, "cluster not found")
		return
	}
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to load cluster")
		return
	}

	data := gin.H{
		"id": id, "name": name, "environment": env, "serverUrl": serverURL,
		"version": ver, "createdAt": created.UTC().Format(time.RFC3339),
		"reachable": false, "nodeCount": 0, "namespaceCount": 0,
	}

	cl, err := h.mgr.GetCluster(id)
	if err != nil {
		httputil.DataJSON(c, http.StatusOK, data)
		return
	}
	data["reachable"] = true
	if lv, e := cl.Clientset.ServerVersion(); e == nil {
		data["liveVersion"] = lv.GitVersion
	}
	if nl, e := cl.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); e == nil {
		data["nodeCount"] = len(nl.Items)
	}
	if ns, e := cl.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); e == nil {
		data["namespaceCount"] = len(ns.Items)
	}

	httputil.DataJSON(c, http.StatusOK, data)
}

func (h *ClusterHandler) canAccessCluster(ctx context.Context, clusterID string, userID uuid.UUID, admin bool) bool {
	if admin {
		var exists bool
		_ = h.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM clusters WHERE id = $1::uuid)`, clusterID).Scan(&exists)
		return exists
	}
	var ok bool
	_ = h.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM cluster_permissions cp
			INNER JOIN team_members tm ON tm.team_id = cp.team_id
			WHERE cp.cluster_id = $1 AND tm.user_id = $2
		)`, clusterID, userID).Scan(&ok)
	return ok
}

// Delete removes a cluster from DB and memory (admin).
func (h *ClusterHandler) Delete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if _, err := uuid.Parse(id); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid cluster id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	cmd, err := h.pool.Exec(ctx, `DELETE FROM clusters WHERE id = $1::uuid`, id)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "failed to delete cluster")
		return
	}
	if cmd.RowsAffected() == 0 {
		httputil.ErrorJSON(c, http.StatusNotFound, "cluster not found")
		return
	}
	h.mgr.RemoveCluster(id)
	c.Status(http.StatusNoContent)
}
