package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AbhishekSharmaIE/Kubevision/internal/api/httputil"
	"github.com/gin-gonic/gin"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
)

const (
	helmOpMaxTimeout   = time.Hour
	helmOpDefaultWait  = 10 * time.Minute
	helmReposFileEmpty = "apiVersion: v1\nrepositories: []\n"
)

func ensureHelmRegistryClient(cfg *action.Configuration) error {
	if cfg.RegistryClient != nil {
		return nil
	}
	c, err := registry.NewClient(registry.ClientOptWriter(io.Discard))
	if err != nil {
		return fmt.Errorf("registry client: %w", err)
	}
	cfg.RegistryClient = c
	return nil
}

// tempHelmCLISettings returns isolated Helm cache/config dirs (caller must run cleanup).
func tempHelmCLISettings() (settings *cli.EnvSettings, cleanup func(), err error) {
	dir, err := os.MkdirTemp("", "kubevision-helm-*")
	if err != nil {
		return nil, nil, err
	}
	cleanup = func() { _ = os.RemoveAll(dir) }
	settings = cli.New()
	settings.RepositoryCache = dir
	settings.RepositoryConfig = filepath.Join(dir, "repositories.yaml")
	if err := os.WriteFile(settings.RepositoryConfig, []byte(helmReposFileEmpty), 0o600); err != nil {
		cleanup()
		return nil, nil, err
	}
	return settings, cleanup, nil
}

func validateHelmChartRef(chart string) error {
	chart = strings.TrimSpace(chart)
	if chart == "" {
		return fmt.Errorf("chart is required")
	}
	if filepath.IsAbs(chart) || strings.HasPrefix(chart, ".") {
		return fmt.Errorf("local chart paths are not allowed; use oci://, an http(s) chart URL, or chart name with repoUrl")
	}
	if strings.Contains(chart, "..") {
		return fmt.Errorf("invalid chart reference")
	}
	return nil
}

func helmOpTimeout(wait bool, timeoutSec int) time.Duration {
	if !wait {
		return 5 * time.Minute
	}
	if timeoutSec <= 0 {
		return helmOpDefaultWait
	}
	d := time.Duration(timeoutSec) * time.Second
	if d > helmOpMaxTimeout {
		return helmOpMaxTimeout
	}
	return d
}

func locateHelmChart(cfg *action.Configuration, settings *cli.EnvSettings, chartRef, version, repoURL, repoUser, repoPass string) (*chart.Chart, error) {
	if err := ensureHelmRegistryClient(cfg); err != nil {
		return nil, err
	}
	inst := action.NewInstall(cfg)
	inst.ChartPathOptions.Version = strings.TrimSpace(version)
	inst.ChartPathOptions.RepoURL = strings.TrimSpace(repoURL)
	inst.ChartPathOptions.Username = repoUser
	inst.ChartPathOptions.Password = repoPass
	path, err := inst.ChartPathOptions.LocateChart(strings.TrimSpace(chartRef), settings)
	if err != nil {
		return nil, err
	}
	return loader.Load(path)
}

type helmInstallBody struct {
	ReleaseName     string                 `json:"releaseName" binding:"required"`
	Chart           string                 `json:"chart" binding:"required"`
	Version         string                 `json:"version"`
	RepoURL         string                 `json:"repoUrl"`
	RepoUsername    string                 `json:"repoUsername"`
	RepoPassword    string                 `json:"repoPassword"`
	Values          map[string]interface{} `json:"values"`
	DryRun          bool                   `json:"dryRun"`
	CreateNamespace bool                   `json:"createNamespace"`
	Wait            bool                   `json:"wait"`
	TimeoutSeconds  int                    `json:"timeoutSeconds"`
	Description     string                 `json:"description"`
}

type helmUpgradeBody struct {
	Chart          string                 `json:"chart" binding:"required"`
	Version        string                 `json:"version"`
	RepoURL        string                 `json:"repoUrl"`
	RepoUsername   string                 `json:"repoUsername"`
	RepoPassword   string                 `json:"repoPassword"`
	Values         map[string]interface{} `json:"values"`
	DryRun         bool                   `json:"dryRun"`
	Wait           bool                   `json:"wait"`
	TimeoutSeconds int                    `json:"timeoutSeconds"`
	Description    string                 `json:"description"`
	ResetValues    bool                   `json:"resetValues"`
	ReuseValues    bool                   `json:"reuseValues"`
}

// HelmInstallRelease handles POST /clusters/:id/namespaces/:namespace/helm/releases.
func (h *HelmHandler) HelmInstallRelease(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	ns := strings.TrimSpace(c.Param("namespace"))
	if ns == "" {
		httputil.ErrorJSON(c, http.StatusBadRequest, "namespace required")
		return
	}
	cl, ok := h.clusterClient(c, id)
	if !ok {
		return
	}

	var body helmInstallBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid json body: releaseName and chart are required")
		return
	}
	if err := chartutil.ValidateReleaseName(body.ReleaseName); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid release name: "+err.Error())
		return
	}
	if err := validateHelmChartRef(body.Chart); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	driverName, useBoth, driverOK := parseHelmStorageQuery(c.Query("driver"))
	if !driverOK {
		httputil.ErrorJSON(c, http.StatusBadRequest, "driver must be secret, configmap, or both")
		return
	}
	if useBoth {
		httputil.ErrorJSON(c, http.StatusBadRequest, "driver=both is not supported for install; use secret or configmap")
		return
	}

	settings, cleanup, err := tempHelmCLISettings()
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "helm temp dir: "+err.Error())
		return
	}
	defer cleanup()

	cfg, err := helmActionConfig(ns, cl.Config, driverName)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "helm init failed: "+err.Error())
		return
	}

	ch, err := locateHelmChart(cfg, settings, body.Chart, body.Version, body.RepoURL, body.RepoUsername, body.RepoPassword)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "chart resolve/load failed: "+err.Error())
		return
	}

	vals := body.Values
	if vals == nil {
		vals = map[string]interface{}{}
	}

	install := action.NewInstall(cfg)
	install.ReleaseName = body.ReleaseName
	install.Namespace = ns
	install.DryRun = body.DryRun
	install.CreateNamespace = body.CreateNamespace
	install.Wait = body.Wait
	install.Timeout = helmOpTimeout(body.Wait, body.TimeoutSeconds)
	install.Description = strings.TrimSpace(body.Description)

	ctx, cancel := context.WithTimeout(c.Request.Context(), install.Timeout+2*time.Minute)
	defer cancel()

	rel, err := install.RunWithContext(ctx, ch, vals)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "helm install failed: "+err.Error())
		return
	}

	httputil.DataJSON(c, http.StatusCreated, gin.H{"release": releaseToJSON(rel)})
}

// HelmUpgradeRelease handles PUT /clusters/:id/namespaces/:namespace/helm/releases/:release.
func (h *HelmHandler) HelmUpgradeRelease(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	ns := strings.TrimSpace(c.Param("namespace"))
	releaseName := strings.TrimSpace(c.Param("release"))
	if ns == "" || releaseName == "" {
		httputil.ErrorJSON(c, http.StatusBadRequest, "namespace and release name required")
		return
	}
	cl, ok := h.clusterClient(c, id)
	if !ok {
		return
	}

	var body helmUpgradeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "invalid json body: chart is required")
		return
	}
	if err := validateHelmChartRef(body.Chart); err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	driverName, useBoth, driverOK := parseHelmStorageQuery(c.Query("driver"))
	if !driverOK {
		httputil.ErrorJSON(c, http.StatusBadRequest, "driver must be secret, configmap, or both")
		return
	}
	if useBoth {
		httputil.ErrorJSON(c, http.StatusBadRequest, "driver=both is not supported for upgrade; use secret or configmap")
		return
	}

	settings, cleanup, err := tempHelmCLISettings()
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "helm temp dir: "+err.Error())
		return
	}
	defer cleanup()

	cfg, err := helmActionConfig(ns, cl.Config, driverName)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusInternalServerError, "helm init failed: "+err.Error())
		return
	}

	ch, err := locateHelmChart(cfg, settings, body.Chart, body.Version, body.RepoURL, body.RepoUsername, body.RepoPassword)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadRequest, "chart resolve/load failed: "+err.Error())
		return
	}

	vals := body.Values
	if vals == nil {
		vals = map[string]interface{}{}
	}

	upgrade := action.NewUpgrade(cfg)
	upgrade.Namespace = ns
	upgrade.DryRun = body.DryRun
	upgrade.Wait = body.Wait
	upgrade.Timeout = helmOpTimeout(body.Wait, body.TimeoutSeconds)
	upgrade.Description = strings.TrimSpace(body.Description)
	upgrade.ResetValues = body.ResetValues
	upgrade.ReuseValues = body.ReuseValues
	upgrade.MaxHistory = 10

	ctx, cancel := context.WithTimeout(c.Request.Context(), upgrade.Timeout+2*time.Minute)
	defer cancel()

	rel, err := upgrade.RunWithContext(ctx, releaseName, ch, vals)
	if err != nil {
		httputil.ErrorJSON(c, http.StatusBadGateway, "helm upgrade failed: "+err.Error())
		return
	}

	httputil.DataJSON(c, http.StatusOK, gin.H{"release": releaseToJSON(rel)})
}
