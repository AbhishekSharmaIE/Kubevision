package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/AbhishekSharmaIE/Kubevision/internal/k8s"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

// HelmHandler lists Helm releases using the cluster kubeconfig.
type HelmHandler struct {
	mgr *k8s.ClusterManager
}

// NewHelmHandler constructs HelmHandler.
func NewHelmHandler(mgr *k8s.ClusterManager) *HelmHandler {
	return &HelmHandler{mgr: mgr}
}

func (h *HelmHandler) clusterClient(c *gin.Context, clusterID string) (*k8s.ClusterClient, bool) {
	if _, err := uuid.Parse(clusterID); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid cluster id")
		return nil, false
	}
	cl, err := h.mgr.GetCluster(clusterID)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusNotFound, "cluster not connected; check registration and reachability")
		return nil, false
	}
	return cl, true
}

func helmActionConfig(namespace string, restCfg *rest.Config) (*action.Configuration, error) {
	initNS := namespace
	if initNS == "" {
		initNS = metav1.NamespaceDefault
	}
	flags := genericclioptions.NewConfigFlags(true)
	flags.Namespace = &initNS
	flags.WrapConfigFn = func(_ *rest.Config) *rest.Config {
		return rest.CopyConfig(restCfg)
	}
	cfg := new(action.Configuration)
	noopLog := func(string, ...interface{}) {}
	if err := cfg.Init(flags, initNS, "secret", noopLog); err != nil {
		return nil, err
	}
	return cfg, nil
}

func releaseToJSON(r *release.Release) gin.H {
	out := gin.H{
		"name":      r.Name,
		"namespace": r.Namespace,
		"revision":  r.Version,
	}
	if r.Info != nil {
		out["status"] = r.Info.Status.String()
		out["description"] = r.Info.Description
		if !r.Info.LastDeployed.IsZero() {
			out["updated"] = r.Info.LastDeployed.UTC().Format(time.RFC3339)
		}
	}
	if r.Chart != nil && r.Chart.Metadata != nil {
		out["chart"] = r.Chart.Metadata.Name
		out["chartVersion"] = r.Chart.Metadata.Version
		out["appVersion"] = r.Chart.Metadata.AppVersion
	}
	return out
}

// ListHelmReleases handles GET /clusters/:id/helm/releases.
// Query: namespace (omit for all namespaces), limit (default 256, max 500), allStates (true to include superseded, pending, etc.).
func (h *HelmHandler) ListHelmReleases(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	cl, ok := h.clusterClient(c, id)
	if !ok {
		return
	}

	ns := strings.TrimSpace(c.Query("namespace"))
	limit := 256
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	initNS := metav1.NamespaceDefault
	if ns != "" {
		initNS = ns
	}

	cfg, err := helmActionConfig(initNS, cl.Config)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "helm init failed: "+err.Error())
		return
	}

	list := action.NewList(cfg)
	list.AllNamespaces = (ns == "")
	list.Limit = limit
	if strings.EqualFold(c.Query("allStates"), "true") {
		list.StateMask = action.ListAll
	}

	releases, err := list.Run()
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "helm list failed: "+err.Error())
		return
	}

	out := make([]gin.H, 0, len(releases))
	for _, r := range releases {
		out = append(out, releaseToJSON(r))
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out)})
}
