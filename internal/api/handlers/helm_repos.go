package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/gin-gonic/gin"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
)

// HelmReposHandler reads/writes Helm chart repository config on the API server (not per-cluster).
// Set KUBEVISION_HELM_HOME (default ./data/helm) for repositories.yaml and index cache.
type HelmReposHandler struct {
	helmHome string
}

// NewHelmReposHandler constructs HelmReposHandler.
func NewHelmReposHandler() *HelmReposHandler {
	home := strings.TrimSpace(os.Getenv("KUBEVISION_HELM_HOME"))
	if home == "" {
		home = filepath.Join("data", "helm")
	}
	return &HelmReposHandler{helmHome: home}
}

func (h *HelmReposHandler) repoConfigPath() string {
	return filepath.Join(h.helmHome, "repositories.yaml")
}

func (h *HelmReposHandler) repoCachePath() string {
	return filepath.Join(h.helmHome, "cache")
}

func (h *HelmReposHandler) ensureHelmDirs() error {
	if err := os.MkdirAll(h.repoCachePath(), 0o755); err != nil {
		return err
	}
	return os.MkdirAll(filepath.Dir(h.repoConfigPath()), 0o755)
}

func (h *HelmReposHandler) loadRepoFile() (*repo.File, error) {
	if err := h.ensureHelmDirs(); err != nil {
		return nil, err
	}
	path := h.repoConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f := repo.NewFile()
		if werr := f.WriteFile(path, 0o600); werr != nil {
			return nil, werr
		}
		return f, nil
	}
	return repo.LoadFile(path)
}

func (h *HelmReposHandler) saveRepoFile(f *repo.File) error {
	return f.WriteFile(h.repoConfigPath(), 0o600)
}

// ListHelmRepos handles GET /helm/repos.
func (h *HelmReposHandler) ListHelmRepos(c *gin.Context) {
	f, err := h.loadRepoFile()
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "helm repos: "+err.Error())
		return
	}
	out := make([]gin.H, 0, len(f.Repositories))
	for _, e := range f.Repositories {
		if e == nil {
			continue
		}
		out = append(out, gin.H{
			"name": e.Name, "url": e.URL,
		})
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out), "helmHome": h.helmHome})
}

// ListHelmRepoCharts handles GET /helm/repos/:repo/charts.
// Query refresh=true forces re-download of index.yaml into the server cache.
func (h *HelmReposHandler) ListHelmRepoCharts(c *gin.Context) {
	repoName := strings.TrimSpace(c.Param("repo"))
	if repoName == "" {
		httputil.ErrorJSON(c, http.StatusBadRequest, "repo name required")
		return
	}

	f, err := h.loadRepoFile()
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "helm repos: "+err.Error())
		return
	}
	entry := f.Get(repoName)
	if entry == nil {
		httputil.ErrorJSON(c, http.StatusNotFound, "unknown helm repository: "+repoName)
		return
	}

	settings := cli.New()
	settings.RepositoryConfig = h.repoConfigPath()
	settings.RepositoryCache = h.repoCachePath()

	cr, err := repo.NewChartRepository(entry, getter.All(settings))
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "helm repo client: "+err.Error())
		return
	}
	cr.CachePath = settings.RepositoryCache

	idxPath := filepath.Join(cr.CachePath, helmpath.CacheIndexFile(entry.Name))
	refresh := strings.EqualFold(c.Query("refresh"), "true")
	if refresh {
		if _, err := cr.DownloadIndexFile(); err != nil {
			httputil.ErrorJSON(c, http.StatusBadGateway, "helm repo index download failed: "+err.Error())
			return
		}
	} else if _, err := os.Stat(idxPath); os.IsNotExist(err) {
		if _, err := cr.DownloadIndexFile(); err != nil {
			httputil.ErrorJSON(c, http.StatusBadGateway, "helm repo index download failed: "+err.Error())
			return
		}
	}

	idx, err := repo.LoadIndexFile(idxPath)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "helm repo index read failed: "+err.Error())
		return
	}

	type chartRow struct {
		name, version, appVersion, description string
	}
	var rows []chartRow
	for name, versions := range idx.Entries {
		for _, v := range versions {
			if v == nil {
				continue
			}
			rows = append(rows, chartRow{
				name:        name,
				version:     v.Version,
				appVersion:  v.AppVersion,
				description: v.Description,
			})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].name != rows[j].name {
			return rows[i].name < rows[j].name
		}
		return rows[i].version < rows[j].version
	})
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"name": r.name, "version": r.version, "appVersion": r.appVersion, "description": r.description,
		})
	}
	httputil.DataJSONWithMeta(c, http.StatusOK, out, gin.H{"total": len(out), "repository": repoName})
}

type helmAddRepoBody struct {
	Name                  string `json:"name" binding:"required"`
	URL                   string `json:"url" binding:"required"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	InsecureSkipTLSverify bool   `json:"insecureSkipTlsVerify"`
	PassCredentialsAll    bool   `json:"passCredentialsAll"`
}

// AddHelmRepo handles POST /helm/repos (admin only — mitigates SSRF to internal networks).
func (h *HelmReposHandler) AddHelmRepo(c *gin.Context) {
	var body helmAddRepoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "json body with name and url is required")
		return
	}
	u := strings.TrimSpace(body.URL)
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "oci://") {
		httputil.ErrorJSON(c, http.StatusBadRequest, "url must start with http://, https://, or oci://")
		return
	}

	f, err := h.loadRepoFile()
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "helm repos: "+err.Error())
		return
	}
	if f.Has(strings.TrimSpace(body.Name)) {
		httputil.ErrorJSON(c, http.StatusConflict, "repository name already exists")
		return
	}
	ent := &repo.Entry{
		Name:                  strings.TrimSpace(body.Name),
		URL:                   u,
		Username:              body.Username,
		Password:              body.Password,
		InsecureSkipTLSverify: body.InsecureSkipTLSverify,
		PassCredentialsAll:    body.PassCredentialsAll,
	}
	f.Add(ent)
	if err := h.saveRepoFile(f); err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "helm repos save failed: "+err.Error())
		return
	}
	httputil.DataJSON(c, http.StatusCreated, gin.H{"name": ent.Name, "url": ent.URL})
}
