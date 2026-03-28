package handlers

import (
	"net/http"
	"sort"
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

func helmActionConfig(namespace string, restCfg *rest.Config, helmDriver string) (*action.Configuration, error) {
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
	if err := cfg.Init(flags, initNS, helmDriver, noopLog); err != nil {
		return nil, err
	}
	return cfg, nil
}

// parseHelmStorageQuery returns driver name for action.Configuration.Init, or useBoth to list secret + configmap and merge.
func parseHelmStorageQuery(raw string) (driver string, useBoth bool, ok bool) {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case "", "secret", "secrets":
		return "secret", false, true
	case "configmap", "configmaps":
		return "configmap", false, true
	case "both", "all":
		return "", true, true
	default:
		return "", false, false
	}
}

func runHelmList(cfg *action.Configuration, allNamespaces bool, limit int, allStates bool) ([]*release.Release, error) {
	list := action.NewList(cfg)
	list.AllNamespaces = allNamespaces
	list.Limit = limit
	if allStates {
		list.StateMask = action.ListAll
	}
	return list.Run()
}

type helmReleaseKey struct {
	ns, name string
}

// mergeHelmReleasesByLatest keeps one release per (namespace, name), preferring the higher revision.
func mergeHelmReleasesByLatest(a, b []*release.Release) []*release.Release {
	m := make(map[helmReleaseKey]*release.Release)
	for _, rel := range a {
		if rel == nil {
			continue
		}
		k := helmReleaseKey{rel.Namespace, rel.Name}
		if cur, ok := m[k]; !ok || rel.Version > cur.Version {
			m[k] = rel
		}
	}
	for _, rel := range b {
		if rel == nil {
			continue
		}
		k := helmReleaseKey{rel.Namespace, rel.Name}
		if cur, ok := m[k]; !ok || rel.Version > cur.Version {
			m[k] = rel
		}
	}
	out := make([]*release.Release, 0, len(m))
	for _, rel := range m {
		out = append(out, rel)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Namespace != out[j].Namespace {
			return out[i].Namespace < out[j].Namespace
		}
		return out[i].Name < out[j].Name
	})
	return out
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
// Query: namespace (omit for all namespaces), limit (default 256, max 500), allStates (true to include superseded, pending, etc.),
// driver: secret (default), configmap, or both (merge secret + configmap backends, dedupe by namespace/name using latest revision).
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

	driver, useBoth, driverOK := parseHelmStorageQuery(c.Query("driver"))
	if !driverOK {
		httputil.ErrorJSON(c, http.StatusBadRequest, "driver must be secret, configmap, or both")
		return
	}

	initNS := metav1.NamespaceDefault
	if ns != "" {
		initNS = ns
	}
	allNS := (ns == "")
	allStates := strings.EqualFold(c.Query("allStates"), "true")

	var releases []*release.Release

	if useBoth {
		var sec, cm []*release.Release
		var errSec, errCM error
		if cfgS, e := helmActionConfig(initNS, cl.Config, "secret"); e != nil {
			errSec = e
		} else {
			sec, errSec = runHelmList(cfgS, allNS, limit, allStates)
		}
		if cfgC, e := helmActionConfig(initNS, cl.Config, "configmap"); e != nil {
			errCM = e
		} else {
			cm, errCM = runHelmList(cfgC, allNS, limit, allStates)
		}
		if errSec != nil && errCM != nil {
			httputil.ErrorJSON(c, http.StatusBadGateway, "helm list failed (secret: "+errSec.Error()+"; configmap: "+errCM.Error()+")")
			return
		}
		releases = mergeHelmReleasesByLatest(sec, cm)
		if len(releases) > limit {
			releases = releases[:limit]
		}
	} else {
		cfg, err := helmActionConfig(initNS, cl.Config, driver)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusInternalServerError, "helm init failed: "+err.Error())
			return
		}
		releases, err = runHelmList(cfg, allNS, limit, allStates)
		if err != nil {
			httputil.ErrorJSON(c, http.StatusBadGateway, "helm list failed: "+err.Error())
			return
		}
	}

	out := make([]gin.H, 0, len(releases))
	for _, r := range releases {
		out = append(out, releaseToJSON(r))
	}
	meta := gin.H{"total": len(out)}
	if useBoth {
		meta["driver"] = "both"
	} else {
		meta["driver"] = driver
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, meta)
}
