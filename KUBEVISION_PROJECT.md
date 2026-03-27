# 🚀 KubeVision — Kubernetes Cluster Management Dashboard

> **A production-grade, real-time Kubernetes observability platform with a custom Helm chart repository, live cluster metrics, multi-cluster support, and a sleek dark-mode dashboard UI powered by React + Grafana + Prometheus.**

---

## 📋 Table of Contents

1. [Project Overview](#1-project-overview)
2. [Tech Stack & Architecture](#2-tech-stack--architecture)
3. [Directory Structure](#3-directory-structure)
4. [Phase 1 — Project Scaffolding](#4-phase-1--project-scaffolding)
5. [Phase 2 — Helm Chart Repository](#5-phase-2--helm-chart-repository)
6. [Phase 3 — Kubernetes Backend (Go Operator)](#6-phase-3--kubernetes-backend-go-operator)
7. [Phase 4 — Metrics Stack (Prometheus + Grafana)](#7-phase-4--metrics-stack-prometheus--grafana)
8. [Phase 5 — React Dashboard Frontend](#8-phase-5--react-dashboard-frontend)
9. [Phase 6 — WebSocket Real-Time Layer](#9-phase-6--websocket-real-time-layer)
10. [Phase 7 — Docker & Containerization](#10-phase-7--docker--containerization)
11. [Phase 8 — Kubernetes Manifests & Deployment](#11-phase-8--kubernetes-manifests--deployment)
12. [Phase 9 — CI/CD Pipeline (GitHub Actions)](#12-phase-9--cicd-pipeline-github-actions)
13. [Phase 10 — Alerting & Notification System](#13-phase-10--alerting--notification-system)
14. [Phase 11 — RBAC & Multi-Tenancy](#14-phase-11--rbac--multi-tenancy)
15. [Phase 12 — Local Dev Setup (Kind/Minikube)](#15-phase-12--local-dev-setup-kindminikube)
16. [Environment Variables Reference](#16-environment-variables-reference)
17. [API Reference](#17-api-reference)
18. [Testing Strategy](#18-testing-strategy)
19. [Cursor AI Instructions](#19-cursor-ai-instructions)

---

## 1. Project Overview

**KubeVision** is a self-hosted, open-source Kubernetes cluster management dashboard that gives DevOps/Platform engineers a single pane of glass to:

- 📊 **Monitor** real-time CPU, memory, network, and pod metrics across multiple clusters
- 🎛️ **Manage** deployments, services, configmaps, secrets, namespaces via a clean UI
- 📦 **Browse & install** Helm charts from a custom-built, self-hosted Helm chart repository
- 🔔 **Receive alerts** via Slack/PagerDuty/email when pods crash, nodes go NotReady, or resource limits are breached
- 🔐 **Enforce RBAC** — team-scoped access with read/write roles per namespace
- 🧠 **AI-assisted** troubleshooting — paste a failing pod log and get a root-cause suggestion (Claude API integration)

### What makes this different from Lens, K9s, Rancher?

| Feature | KubeVision | Lens | K9s | Rancher |
|---|---|---|---|---|
| Self-hosted | ✅ | ✅ | ✅ | ✅ |
| Custom Helm repo | ✅ Built-in | ❌ | ❌ | Partial |
| Real-time WebSocket metrics | ✅ | ❌ | Terminal only | ❌ |
| AI log analysis | ✅ | ❌ | ❌ | ❌ |
| Multi-cluster | ✅ | ✅ | ❌ | ✅ |
| Embeddable Grafana panels | ✅ Native | ❌ | ❌ | Partial |
| Single binary deploy | ✅ | ❌ | ✅ | ❌ |

---

## 2. Tech Stack & Architecture

### Frontend
- **React 18** + **TypeScript** — UI framework
- **Vite** — build tool
- **TailwindCSS** — styling (dark theme, custom design system)
- **Recharts + D3.js** — custom metric charts
- **Framer Motion** — animations
- **Zustand** — global state
- **React Query (TanStack)** — data fetching + caching
- **xterm.js** — embedded terminal for kubectl exec
- **Monaco Editor** — YAML editing with Kubernetes schema validation

### Backend
- **Go 1.22** — main API server
- **Gin** — HTTP router
- **gorilla/websocket** — real-time metrics streaming
- **client-go** — official Kubernetes Go client
- **controller-runtime** — custom operator/controller
- **Helm SDK (helm.sh/helm/v3)** — programmatic Helm operations

### Observability Stack
- **Prometheus** — metrics collection
- **Grafana** — embedded dashboards
- **AlertManager** — alert routing
- **kube-state-metrics** — Kubernetes object metrics
- **node-exporter** — host-level metrics
- **Loki** — log aggregation
- **Promtail** — log shipping

### Infrastructure
- **Docker** + **Docker Compose** — local development
- **Kind** or **Minikube** — local Kubernetes cluster
- **Helm** — packaging and deployment
- **GitHub Actions** — CI/CD
- **GitHub Pages / Chartmuseum** — Helm chart repository hosting
- **Nginx** — ingress + static asset serving

### Database
- **PostgreSQL** — user accounts, RBAC, audit logs, alert rules
- **Redis** — WebSocket session state, metric caching, rate limiting

---

## 3. Directory Structure

```
kubevision/
├── .github/
│   └── workflows/
│       ├── ci.yml                    # Run tests on PR
│       ├── release.yml               # Tag → build → push Docker image
│       └── helm-release.yml          # Package and publish Helm charts
│
├── charts/                           # Custom Helm chart repository
│   ├── index.yaml                    # Helm repo index (auto-generated)
│   ├── kubevision/                   # Main KubeVision Helm chart
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   ├── values.schema.json        # JSON Schema validation for values
│   │   └── templates/
│   │       ├── _helpers.tpl
│   │       ├── deployment.yaml
│   │       ├── service.yaml
│   │       ├── ingress.yaml
│   │       ├── configmap.yaml
│   │       ├── secret.yaml
│   │       ├── serviceaccount.yaml
│   │       ├── clusterrole.yaml
│   │       ├── clusterrolebinding.yaml
│   │       ├── hpa.yaml
│   │       └── NOTES.txt
│   └── monitoring-stack/             # Bundled Prometheus+Grafana chart
│       ├── Chart.yaml
│       └── ...
│
├── cmd/
│   └── server/
│       └── main.go                   # Entry point
│
├── internal/
│   ├── api/
│   │   ├── router.go                 # Gin router setup
│   │   ├── middleware/
│   │   │   ├── auth.go               # JWT validation
│   │   │   ├── rbac.go               # Namespace-level RBAC
│   │   │   ├── ratelimit.go          # Redis-backed rate limiting
│   │   │   └── audit.go              # Audit log middleware
│   │   └── handlers/
│   │       ├── clusters.go           # Multi-cluster management
│   │       ├── pods.go               # Pod CRUD + logs + exec
│   │       ├── deployments.go        # Deployment management
│   │       ├── services.go           # Service management
│   │       ├── namespaces.go         # Namespace management
│   │       ├── nodes.go              # Node info + drain/cordon
│   │       ├── helm.go               # Helm release management
│   │       ├── metrics.go            # Prometheus metrics proxy
│   │       ├── alerts.go             # Alert rule CRUD
│   │       ├── auth.go               # Login/logout/token refresh
│   │       ├── users.go              # User management
│   │       └── ai.go                 # AI log analysis endpoint
│   │
│   ├── k8s/
│   │   ├── client.go                 # Multi-cluster client factory
│   │   ├── watcher.go                # Informer/watch event streaming
│   │   ├── exec.go                   # WebSocket pod exec
│   │   └── operator/
│   │       ├── controller.go         # Custom resource controller
│   │       └── types.go              # CRD type definitions
│   │
│   ├── helm/
│   │   ├── client.go                 # Helm action client wrapper
│   │   ├── repo.go                   # Chart repo management
│   │   └── release.go                # Install/upgrade/rollback/uninstall
│   │
│   ├── metrics/
│   │   ├── collector.go              # Prometheus scrape + aggregation
│   │   ├── websocket.go              # Real-time metric streaming
│   │   └── cache.go                  # Redis metric caching
│   │
│   ├── alerts/
│   │   ├── engine.go                 # Alert evaluation loop
│   │   ├── notifier.go               # Slack/PagerDuty/email dispatcher
│   │   └── rules.go                  # Alert rule parsing
│   │
│   ├── auth/
│   │   ├── jwt.go                    # JWT issue/validate/refresh
│   │   ├── oidc.go                   # OIDC provider integration
│   │   └── session.go                # Redis session management
│   │
│   ├── db/
│   │   ├── postgres.go               # DB connection pool
│   │   ├── redis.go                  # Redis client
│   │   └── migrations/               # SQL migration files
│   │       ├── 001_init.sql
│   │       ├── 002_rbac.sql
│   │       └── 003_audit_logs.sql
│   │
│   └── ai/
│       └── analyzer.go               # Claude API log analysis
│
├── frontend/
│   ├── index.html
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── router.tsx                # React Router v6 routes
│       │
│       ├── design-system/            # Custom component library
│       │   ├── tokens.ts             # Color, spacing, type tokens
│       │   ├── Button/
│       │   ├── Card/
│       │   ├── Badge/
│       │   ├── Table/
│       │   ├── Modal/
│       │   ├── Tooltip/
│       │   └── Terminal/             # xterm.js wrapper
│       │
│       ├── pages/
│       │   ├── Overview/             # Cluster overview / home
│       │   ├── Nodes/                # Node list + detail
│       │   ├── Workloads/            # Deployments, StatefulSets, DaemonSets
│       │   ├── Pods/                 # Pod list + logs + exec
│       │   ├── Services/             # Services + Ingresses
│       │   ├── HelmReleases/         # Helm release management
│       │   ├── HelmRepo/             # Chart browser
│       │   ├── Metrics/              # Custom charts + Grafana embeds
│       │   ├── Alerts/               # Alert rules + firing alerts
│       │   ├── Logs/                 # Loki log explorer
│       │   ├── Settings/             # Cluster connections, RBAC, users
│       │   └── AIAnalyzer/           # AI log troubleshooter
│       │
│       ├── stores/
│       │   ├── cluster.store.ts      # Active cluster state
│       │   ├── auth.store.ts         # Auth state
│       │   └── alerts.store.ts       # Live alert state
│       │
│       ├── hooks/
│       │   ├── useWebSocket.ts       # Generic WS hook
│       │   ├── useMetrics.ts         # Real-time metric streams
│       │   ├── useKubectl.ts         # REST API wrapper hooks
│       │   └── useAlerts.ts          # Alert subscription hook
│       │
│       ├── api/
│       │   ├── client.ts             # Axios instance + interceptors
│       │   ├── clusters.api.ts
│       │   ├── pods.api.ts
│       │   ├── metrics.api.ts
│       │   ├── helm.api.ts
│       │   └── ai.api.ts
│       │
│       └── utils/
│           ├── formatBytes.ts
│           ├── formatDuration.ts
│           ├── k8sColors.ts          # Status → color mapping
│           └── yamlValidation.ts
│
├── monitoring/
│   ├── prometheus/
│   │   ├── prometheus.yml            # Scrape config
│   │   └── rules/
│   │       ├── kubernetes.yml        # K8s alerting rules
│   │       └── kubevision.yml        # App-specific rules
│   ├── grafana/
│   │   ├── datasources/
│   │   │   └── prometheus.yml
│   │   └── dashboards/
│   │       ├── cluster-overview.json
│   │       ├── node-metrics.json
│   │       ├── pod-metrics.json
│   │       └── helm-releases.json
│   ├── alertmanager/
│   │   └── alertmanager.yml
│   └── loki/
│       └── loki-config.yml
│
├── deploy/
│   ├── docker-compose.yml            # Full local stack
│   ├── docker-compose.dev.yml        # Dev overrides (hot reload)
│   └── kind-cluster.yaml             # Kind cluster config
│
├── scripts/
│   ├── setup-local.sh                # One-command local setup
│   ├── seed-demo-data.sh             # Seed demo cluster data
│   ├── package-charts.sh             # Build + index Helm charts
│   └── generate-certs.sh            # Self-signed TLS for local dev
│
├── Dockerfile                        # Multi-stage: Go API + React build
├── Dockerfile.dev                    # Dev image with hot reload
├── Makefile                          # Common dev commands
├── go.mod
├── go.sum
└── README.md
```

---

## 4. Phase 1 — Project Scaffolding

### Step 1.1 — Initialize the repo

```bash
mkdir kubevision && cd kubevision
git init
go mod init github.com/yourorg/kubevision
```

### Step 1.2 — Create the Makefile

Create `Makefile` with the following targets:

```makefile
.PHONY: dev build test lint docker-build helm-package setup-local

# Start full local stack
dev:
	docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.dev.yml up --build

# Build Go binary
build:
	go build -ldflags="-s -w -X main.Version=$(shell git describe --tags --always)" -o bin/kubevision ./cmd/server

# Run all tests
test:
	go test ./... -v -race -coverprofile=coverage.out

# Lint
lint:
	golangci-lint run ./...

# Build Docker image
docker-build:
	docker build -t kubevision:$(shell git describe --tags --always) .

# Package Helm charts
helm-package:
	bash scripts/package-charts.sh

# One-command local setup
setup-local:
	bash scripts/setup-local.sh
```

### Step 1.3 — Initialize the React frontend

```bash
cd frontend
npm create vite@latest . -- --template react-ts
npm install
npm install -D tailwindcss @tailwindcss/typography postcss autoprefixer
npm install framer-motion zustand @tanstack/react-query axios
npm install recharts d3 @types/d3
npm install xterm xterm-addon-fit xterm-addon-web-links
npm install @monaco-editor/react
npm install react-router-dom@6
npm install js-yaml @types/js-yaml
npm install react-hot-toast
npx tailwindcss init -p
```

### Step 1.4 — Install Go dependencies

```bash
go get github.com/gin-gonic/gin
go get github.com/gorilla/websocket
go get k8s.io/client-go@v0.29.0
go get k8s.io/api@v0.29.0
go get sigs.k8s.io/controller-runtime@v0.17.0
go get helm.sh/helm/v3
go get github.com/prometheus/client_golang
go get github.com/golang-jwt/jwt/v5
go get github.com/jackc/pgx/v5
go get github.com/redis/go-redis/v9
go get github.com/slack-go/slack
go get github.com/robfig/cron/v3
```

---

## 5. Phase 2 — Helm Chart Repository

### Step 2.1 — Create the main KubeVision Helm Chart

Create `charts/kubevision/Chart.yaml`:

```yaml
apiVersion: v2
name: kubevision
description: KubeVision — Kubernetes Cluster Management Dashboard
type: application
version: 0.1.0
appVersion: "1.0.0"
keywords:
  - kubernetes
  - monitoring
  - dashboard
  - helm
  - observability
home: https://github.com/yourorg/kubevision
sources:
  - https://github.com/yourorg/kubevision
maintainers:
  - name: YourName
    email: you@yourorg.com
dependencies:
  - name: postgresql
    version: "13.x.x"
    repository: https://charts.bitnami.com/bitnami
    condition: postgresql.enabled
  - name: redis
    version: "18.x.x"
    repository: https://charts.bitnami.com/bitnami
    condition: redis.enabled
  - name: prometheus
    version: "25.x.x"
    repository: https://prometheus-community.github.io/helm-charts
    condition: prometheus.enabled
  - name: grafana
    version: "7.x.x"
    repository: https://grafana.github.io/helm-charts
    condition: grafana.enabled
```

### Step 2.2 — Create `charts/kubevision/values.yaml`

```yaml
# ──────────────────────────────────────────────
# KubeVision Helm Values
# ──────────────────────────────────────────────

replicaCount: 2

image:
  repository: ghcr.io/yourorg/kubevision
  pullPolicy: IfNotPresent
  tag: ""  # Defaults to .Chart.AppVersion

serviceAccount:
  create: true
  name: "kubevision"
  annotations: {}

service:
  type: ClusterIP
  port: 8080
  metricsPort: 9090

ingress:
  enabled: true
  className: "nginx"
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: kubevision.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: kubevision-tls
      hosts:
        - kubevision.yourdomain.com

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

# ── App Configuration ──────────────────────────
config:
  logLevel: info
  jwtSecret: ""          # OVERRIDE in production — use a secret
  sessionTTL: 24h
  multiCluster:
    enabled: true
  ai:
    enabled: false        # Set to true and provide apiKey for AI features
    apiKey: ""

# ── Database ──────────────────────────────────
postgresql:
  enabled: true
  auth:
    database: kubevision
    username: kubevision
    password: changeme   # OVERRIDE in production

# ── Cache ─────────────────────────────────────
redis:
  enabled: true
  auth:
    enabled: false

# ── Monitoring ────────────────────────────────
prometheus:
  enabled: true
  alertmanager:
    enabled: true

grafana:
  enabled: true
  adminPassword: changeme
  persistence:
    enabled: true
    size: 10Gi
  dashboardProviders:
    dashboardproviders.yaml:
      apiVersion: 1
      providers:
        - name: kubevision
          type: file
          options:
            path: /var/lib/grafana/dashboards/kubevision
  dashboardsConfigMaps:
    kubevision: kubevision-grafana-dashboards
```

### Step 2.3 — Create key Helm templates

Create `charts/kubevision/templates/clusterrole.yaml`:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{ include "kubevision.fullname" . }}
  labels:
    {{- include "kubevision.labels" . | nindent 4 }}
rules:
  # Read all core resources
  - apiGroups: [""]
    resources:
      - pods
      - pods/log
      - pods/exec
      - nodes
      - namespaces
      - services
      - endpoints
      - configmaps
      - persistentvolumes
      - persistentvolumeclaims
      - events
    verbs: ["get", "list", "watch"]
  # Read apps resources
  - apiGroups: ["apps"]
    resources:
      - deployments
      - statefulsets
      - daemonsets
      - replicasets
    verbs: ["get", "list", "watch", "update", "patch"]
  # Helm secrets
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]
  # Metrics
  - apiGroups: ["metrics.k8s.io"]
    resources: ["pods", "nodes"]
    verbs: ["get", "list"]
```

### Step 2.4 — Package and index charts script

Create `scripts/package-charts.sh`:

```bash
#!/bin/bash
set -e

CHARTS_DIR="./charts"
OUTPUT_DIR="./charts"

echo "📦 Packaging Helm charts..."

# Update dependencies for each chart
for chart in "$CHARTS_DIR"/*/; do
  if [ -f "$chart/Chart.yaml" ]; then
    echo "  → Updating dependencies for $chart"
    helm dependency update "$chart"
    helm package "$chart" --destination "$OUTPUT_DIR"
  fi
done

echo "📋 Generating repo index..."
helm repo index "$OUTPUT_DIR" --url https://yourorg.github.io/kubevision/charts

echo "✅ Done! Charts packaged and index generated."
```

---

## 6. Phase 3 — Kubernetes Backend (Go Operator)

### Step 6.1 — Main entry point

Create `cmd/server/main.go`:

```go
package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/yourorg/kubevision/internal/api"
    "github.com/yourorg/kubevision/internal/db"
    "github.com/yourorg/kubevision/internal/k8s"
    "github.com/yourorg/kubevision/internal/metrics"
)

var Version = "dev"

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    slog.Info("Starting KubeVision", "version", Version)

    // Initialize DB connections
    pgPool := db.MustConnectPostgres()
    defer pgPool.Close()

    redisClient := db.MustConnectRedis()
    defer redisClient.Close()

    // Run DB migrations
    if err := db.RunMigrations(pgPool); err != nil {
        slog.Error("Failed to run migrations", "error", err)
        os.Exit(1)
    }

    // Initialize Kubernetes multi-cluster manager
    clusterManager := k8s.NewClusterManager(pgPool)
    if err := clusterManager.LoadClusters(context.Background()); err != nil {
        slog.Warn("Could not load clusters from DB", "error", err)
    }

    // Initialize metrics collector
    metricsCollector := metrics.NewCollector(clusterManager, redisClient)
    go metricsCollector.Start(context.Background())

    // Initialize and start HTTP server
    router := api.NewRouter(api.RouterConfig{
        ClusterManager:   clusterManager,
        MetricsCollector: metricsCollector,
        DB:               pgPool,
        Redis:            redisClient,
    })

    srv := &http.Server{
        Addr:         ":8080",
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        slog.Info("HTTP server listening", "addr", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("Server error", "error", err)
            os.Exit(1)
        }
    }()

    <-quit
    slog.Info("Shutting down gracefully...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        slog.Error("Server forced to shutdown", "error", err)
    }

    slog.Info("Server stopped")
}
```

### Step 6.2 — Multi-cluster client factory

Create `internal/k8s/client.go`:

```go
package k8s

import (
    "context"
    "fmt"
    "sync"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ClusterClient holds clients for a single cluster
type ClusterClient struct {
    ID          string
    Name        string
    Environment string // production, staging, dev
    Clientset   *kubernetes.Clientset
    MetricsClient *metricsv1beta1.Clientset
    Config      *rest.Config
}

// ClusterManager manages connections to multiple Kubernetes clusters
type ClusterManager struct {
    mu       sync.RWMutex
    clusters map[string]*ClusterClient
    db       DB
}

func NewClusterManager(db DB) *ClusterManager {
    return &ClusterManager{
        clusters: make(map[string]*ClusterClient),
        db:       db,
    }
}

// AddCluster registers a new cluster from a kubeconfig bytes
func (m *ClusterManager) AddCluster(ctx context.Context, id, name, env string, kubeconfigBytes []byte) error {
    config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
    if err != nil {
        return fmt.Errorf("invalid kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return fmt.Errorf("creating clientset: %w", err)
    }

    metricsClient, err := metricsv1beta1.NewForConfig(config)
    if err != nil {
        return fmt.Errorf("creating metrics client: %w", err)
    }

    // Verify connectivity
    if _, err := clientset.ServerVersion(); err != nil {
        return fmt.Errorf("cluster unreachable: %w", err)
    }

    m.mu.Lock()
    defer m.mu.Unlock()

    m.clusters[id] = &ClusterClient{
        ID:            id,
        Name:          name,
        Environment:   env,
        Clientset:     clientset,
        MetricsClient: metricsClient,
        Config:        config,
    }

    return nil
}

// GetCluster retrieves a cluster client by ID
func (m *ClusterManager) GetCluster(id string) (*ClusterClient, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    c, ok := m.clusters[id]
    if !ok {
        return nil, fmt.Errorf("cluster %q not found", id)
    }
    return c, nil
}

// ListClusters returns all registered clusters
func (m *ClusterManager) ListClusters() []*ClusterClient {
    m.mu.RLock()
    defer m.mu.RUnlock()

    result := make([]*ClusterClient, 0, len(m.clusters))
    for _, c := range m.clusters {
        result = append(result, c)
    }
    return result
}
```

### Step 6.3 — Real-time pod exec via WebSocket

Create `internal/k8s/exec.go`:

```go
package k8s

import (
    "fmt"
    "net/http"

    "github.com/gorilla/websocket"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/client-go/tools/remotecommand"
    "k8s.io/kubectl/pkg/scheme"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

// ExecPod opens a WebSocket connection and streams a shell into the given pod/container
func (c *ClusterClient) ExecPod(w http.ResponseWriter, r *http.Request, namespace, pod, container string) error {
    wsConn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return fmt.Errorf("WebSocket upgrade: %w", err)
    }
    defer wsConn.Close()

    req := c.Clientset.CoreV1().RESTClient().Post().
        Resource("pods").
        Namespace(namespace).
        Name(pod).
        SubResource("exec").
        VersionedParams(&corev1.PodExecOptions{
            Container: container,
            Command:   []string{"/bin/sh"},
            Stdin:     true,
            Stdout:    true,
            Stderr:    true,
            TTY:       true,
        }, scheme.ParameterCodec)

    exec, err := remotecommand.NewSPDYExecutor(c.Config, "POST", req.URL())
    if err != nil {
        return err
    }

    stream := &wsStream{conn: wsConn}

    return exec.StreamWithContext(r.Context(), remotecommand.StreamOptions{
        Stdin:             stream,
        Stdout:            stream,
        Stderr:            stream,
        Tty:               true,
        TerminalSizeQueue: stream,
    })
}

// wsStream adapts a WebSocket connection to remotecommand's stream interfaces
type wsStream struct {
    conn *websocket.Conn
}

func (s *wsStream) Read(p []byte) (int, error) {
    _, msg, err := s.conn.ReadMessage()
    if err != nil {
        return 0, err
    }
    return copy(p, msg), nil
}

func (s *wsStream) Write(p []byte) (int, error) {
    err := s.conn.WriteMessage(websocket.BinaryMessage, p)
    return len(p), err
}

func (s *wsStream) Next() *remotecommand.TerminalSize {
    // Terminal resize handling omitted for brevity — implement via JSON messages
    return nil
}
```

---

## 7. Phase 4 — Metrics Stack (Prometheus + Grafana)

### Step 7.1 — Prometheus scrape config

Create `monitoring/prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: local

rule_files:
  - /etc/prometheus/rules/*.yml

alerting:
  alertmanagers:
    - static_configs:
        - targets: ["alertmanager:9093"]

scrape_configs:
  # Kubernetes API server
  - job_name: kubernetes-apiservers
    kubernetes_sd_configs:
      - role: endpoints
    scheme: https
    tls_config:
      ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
    relabel_configs:
      - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: default;kubernetes;https

  # Kubernetes nodes
  - job_name: kubernetes-nodes
    scheme: https
    tls_config:
      ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
    kubernetes_sd_configs:
      - role: node
    relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)

  # kube-state-metrics
  - job_name: kube-state-metrics
    static_configs:
      - targets: ["kube-state-metrics:8080"]

  # node-exporter
  - job_name: node-exporter
    kubernetes_sd_configs:
      - role: endpoints
    relabel_configs:
      - source_labels: [__meta_kubernetes_endpoints_name]
        action: keep
        regex: node-exporter

  # KubeVision app metrics
  - job_name: kubevision
    static_configs:
      - targets: ["kubevision:9090"]
```

### Step 7.2 — Alerting rules

Create `monitoring/prometheus/rules/kubernetes.yml`:

```yaml
groups:
  - name: kubernetes.pods
    rules:
      - alert: PodCrashLooping
        expr: |
          rate(kube_pod_container_status_restarts_total[15m]) * 60 * 15 > 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Pod {{ $labels.namespace }}/{{ $labels.pod }} is crash looping"
          description: "Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} has restarted {{ $value | printf \"%.0f\" }} times in the last 15 minutes."

      - alert: PodNotReady
        expr: |
          kube_pod_status_ready{condition="true"} == 0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Pod {{ $labels.namespace }}/{{ $labels.pod }} not ready"

      - alert: PodOOMKilled
        expr: |
          kube_pod_container_status_last_terminated_reason{reason="OOMKilled"} == 1
        for: 0m
        labels:
          severity: critical
        annotations:
          summary: "Pod {{ $labels.pod }} was OOMKilled"

  - name: kubernetes.nodes
    rules:
      - alert: NodeNotReady
        expr: kube_node_status_condition{condition="Ready",status="true"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Node {{ $labels.node }} is not ready"

      - alert: NodeHighCPU
        expr: |
          (1 - avg by(instance)(rate(node_cpu_seconds_total{mode="idle"}[5m]))) * 100 > 85
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High CPU on node {{ $labels.instance }}: {{ $value | printf \"%.1f\" }}%"

      - alert: NodeHighMemory
        expr: |
          (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 90
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High memory on node {{ $labels.instance }}: {{ $value | printf \"%.1f\" }}%"

      - alert: NodeDiskPressure
        expr: |
          (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"}) * 100 < 10
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Low disk space on {{ $labels.instance }}: {{ $value | printf \"%.1f\" }}% remaining"
```

### Step 7.3 — AlertManager config

Create `monitoring/alertmanager/alertmanager.yml`:

```yaml
global:
  resolve_timeout: 5m
  slack_api_url: "${SLACK_WEBHOOK_URL}"

route:
  receiver: default
  group_by: [alertname, cluster, namespace]
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  routes:
    - match:
        severity: critical
      receiver: critical-alerts
      continue: true
    - match:
        severity: warning
      receiver: warning-alerts

receivers:
  - name: default
    slack_configs:
      - channel: "#kubevision-alerts"
        title: "{{ .CommonLabels.alertname }}"
        text: "{{ range .Alerts }}{{ .Annotations.description }}{{ end }}"

  - name: critical-alerts
    slack_configs:
      - channel: "#kubevision-critical"
        color: "danger"
        title: "🔴 CRITICAL: {{ .CommonLabels.alertname }}"
        text: "{{ range .Alerts }}{{ .Annotations.description }}{{ end }}"
    pagerduty_configs:
      - service_key: "${PAGERDUTY_KEY}"

  - name: warning-alerts
    slack_configs:
      - channel: "#kubevision-alerts"
        color: "warning"
        title: "⚠️ WARNING: {{ .CommonLabels.alertname }}"
        text: "{{ range .Alerts }}{{ .Annotations.description }}{{ end }}"

inhibit_rules:
  - source_match:
      severity: critical
    target_match:
      severity: warning
    equal: [alertname, cluster, namespace]
```

---

## 8. Phase 5 — React Dashboard Frontend

### Step 8.1 — Design System Tokens

Create `frontend/src/design-system/tokens.ts`:

```typescript
export const tokens = {
  colors: {
    // Background layers
    bg: {
      base:    '#0a0c10',  // deepest background
      surface: '#0f1218',  // card surfaces
      raised:  '#161b22',  // elevated cards
      overlay: '#1c2128',  // modals, dropdowns
      border:  '#30363d',  // borders
    },
    // Accent — electric teal
    accent: {
      default: '#00d4aa',
      muted:   '#00d4aa33',
      dim:     '#00d4aa15',
    },
    // Status colors
    status: {
      running:   '#3fb950',
      pending:   '#e3b341',
      failed:    '#f85149',
      unknown:   '#8b949e',
      succeeded: '#58a6ff',
      evicted:   '#d29922',
    },
    // Severity
    severity: {
      critical: '#f85149',
      high:     '#ff7b72',
      medium:   '#e3b341',
      low:      '#3fb950',
      info:     '#58a6ff',
    },
    // Text
    text: {
      primary:   '#e6edf3',
      secondary: '#8b949e',
      tertiary:  '#6e7681',
      inverse:   '#0a0c10',
    },
  },
  
  fonts: {
    mono: '"JetBrains Mono", "Fira Code", monospace',
    sans: '"IBM Plex Sans", system-ui, sans-serif',
    display: '"Space Grotesk", sans-serif',
  },

  spacing: {
    '1': '4px',   '2': '8px',   '3': '12px',  '4': '16px',
    '5': '20px',  '6': '24px',  '8': '32px',  '10': '40px',
    '12': '48px', '16': '64px',
  },

  radius: {
    sm: '4px', md: '8px', lg: '12px', xl: '16px', full: '9999px',
  },

  shadows: {
    sm:  '0 1px 2px rgba(0,0,0,0.5)',
    md:  '0 4px 12px rgba(0,0,0,0.4)',
    lg:  '0 8px 24px rgba(0,0,0,0.6)',
    glow: '0 0 20px rgba(0, 212, 170, 0.15)',
  },
} as const;
```

### Step 8.2 — Cluster Overview Page

Create `frontend/src/pages/Overview/index.tsx`:

```tsx
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { ClusterHealthCard } from './ClusterHealthCard';
import { NodeGrid } from './NodeGrid';
import { TopPodsByCPU } from './TopPodsByCPU';
import { RecentEvents } from './RecentEvents';
import { MetricSparklines } from './MetricSparklines';
import { clustersApi } from '@/api/clusters.api';
import { useClusterStore } from '@/stores/cluster.store';

const containerVariants = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.08 }
  }
};

const itemVariants = {
  hidden: { opacity: 0, y: 16 },
  show: { opacity: 1, y: 0, transition: { duration: 0.4, ease: 'easeOut' } }
};

export function OverviewPage() {
  const { activeClusterId } = useClusterStore();

  const { data: overview, isLoading } = useQuery({
    queryKey: ['cluster-overview', activeClusterId],
    queryFn: () => clustersApi.getOverview(activeClusterId),
    refetchInterval: 15_000,
    enabled: !!activeClusterId,
  });

  if (isLoading) return <OverviewSkeleton />;

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="p-6 space-y-6"
    >
      {/* Header */}
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-text-primary font-display">
            Cluster Overview
          </h1>
          <p className="text-text-secondary text-sm mt-1">
            {overview?.name} · {overview?.version} · {overview?.region}
          </p>
        </div>
        <ClusterSelector />
      </motion.div>

      {/* Health cards row */}
      <motion.div variants={itemVariants} className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <ClusterHealthCard
          label="Nodes"
          value={overview?.nodes.ready}
          total={overview?.nodes.total}
          status={overview?.nodes.ready === overview?.nodes.total ? 'healthy' : 'degraded'}
          icon="server"
        />
        <ClusterHealthCard
          label="Pods Running"
          value={overview?.pods.running}
          total={overview?.pods.total}
          status="healthy"
          icon="box"
        />
        <ClusterHealthCard
          label="Pods Failed"
          value={overview?.pods.failed}
          status={overview?.pods.failed > 0 ? 'critical' : 'healthy'}
          icon="alert-circle"
          highlight={overview?.pods.failed > 0}
        />
        <ClusterHealthCard
          label="Namespaces"
          value={overview?.namespaces}
          status="healthy"
          icon="layers"
        />
      </motion.div>

      {/* Metric sparklines */}
      <motion.div variants={itemVariants}>
        <MetricSparklines clusterId={activeClusterId} />
      </motion.div>

      {/* Main grid */}
      <motion.div variants={itemVariants} className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <div className="xl:col-span-2 space-y-6">
          <NodeGrid nodes={overview?.nodes.items ?? []} />
          <TopPodsByCPU clusterId={activeClusterId} />
        </div>
        <div>
          <RecentEvents clusterId={activeClusterId} />
        </div>
      </motion.div>
    </motion.div>
  );
}
```

### Step 8.3 — Real-time Metric Chart Component

Create `frontend/src/pages/Metrics/LiveMetricChart.tsx`:

```tsx
import { useEffect, useRef, useState } from 'react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { useWebSocket } from '@/hooks/useWebSocket';
import { tokens } from '@/design-system/tokens';

interface MetricPoint {
  timestamp: number;
  value: number;
}

interface LiveMetricChartProps {
  clusterId: string;
  metric: 'cpu' | 'memory' | 'network_in' | 'network_out';
  label: string;
  unit: string;
  color?: string;
}

export function LiveMetricChart({ clusterId, metric, label, unit, color = tokens.colors.accent.default }: LiveMetricChartProps) {
  const [data, setData] = useState<MetricPoint[]>([]);
  const MAX_POINTS = 60; // 5 minutes at 5s intervals

  const { lastMessage } = useWebSocket(
    `/ws/metrics?cluster=${clusterId}&metric=${metric}`
  );

  useEffect(() => {
    if (!lastMessage) return;
    const point = JSON.parse(lastMessage.data) as MetricPoint;
    setData(prev => [...prev.slice(-(MAX_POINTS - 1)), point]);
  }, [lastMessage]);

  const latest = data[data.length - 1]?.value ?? 0;

  const formatXAxis = (timestamp: number) =>
    new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });

  const formatTooltip = (value: number) => [`${value.toFixed(2)} ${unit}`, label];

  return (
    <div className="bg-surface rounded-lg border border-border p-4">
      <div className="flex items-center justify-between mb-4">
        <span className="text-text-secondary text-sm font-medium">{label}</span>
        <span className="text-text-primary font-mono text-lg font-semibold" style={{ color }}>
          {latest.toFixed(2)} {unit}
        </span>
      </div>
      <ResponsiveContainer width="100%" height={120}>
        <AreaChart data={data} margin={{ top: 0, right: 0, left: -20, bottom: 0 }}>
          <defs>
            <linearGradient id={`gradient-${metric}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={color} stopOpacity={0.3} />
              <stop offset="95%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <XAxis
            dataKey="timestamp"
            tickFormatter={formatXAxis}
            tick={{ fill: tokens.colors.text.tertiary, fontSize: 10 }}
            tickLine={false}
            axisLine={false}
            interval="preserveStartEnd"
          />
          <YAxis
            tick={{ fill: tokens.colors.text.tertiary, fontSize: 10 }}
            tickLine={false}
            axisLine={false}
          />
          <Tooltip
            formatter={formatTooltip}
            contentStyle={{
              background: tokens.colors.bg.overlay,
              border: `1px solid ${tokens.colors.bg.border}`,
              borderRadius: tokens.radius.md,
              color: tokens.colors.text.primary,
            }}
          />
          <Area
            type="monotone"
            dataKey="value"
            stroke={color}
            strokeWidth={2}
            fill={`url(#gradient-${metric})`}
            dot={false}
            animationDuration={300}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
```

---

## 9. Phase 6 — WebSocket Real-Time Layer

### Step 9.1 — Backend WebSocket metric streamer

Create `internal/metrics/websocket.go`:

```go
package metrics

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // TODO: validate origin in production
    },
}

type MetricPoint struct {
    Timestamp int64   `json:"timestamp"`
    Value     float64 `json:"value"`
    Metric    string  `json:"metric"`
    ClusterID string  `json:"clusterId"`
}

// StreamMetrics streams real-time metric data via WebSocket
func (c *Collector) StreamMetrics(w http.ResponseWriter, r *http.Request) {
    clusterID := r.URL.Query().Get("cluster")
    metric := r.URL.Query().Get("metric")

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        slog.Error("WebSocket upgrade failed", "error", err)
        return
    }
    defer conn.Close()

    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    // Send initial history (last 5 minutes)
    history := c.GetHistory(clusterID, metric, 60)
    for _, point := range history {
        if err := conn.WriteJSON(point); err != nil {
            return
        }
    }

    for {
        select {
        case <-r.Context().Done():
            return
        case <-ticker.C:
            val, err := c.FetchMetric(r.Context(), clusterID, metric)
            if err != nil {
                slog.Warn("Failed to fetch metric", "error", err)
                continue
            }

            point := MetricPoint{
                Timestamp: time.Now().UnixMilli(),
                Value:     val,
                Metric:    metric,
                ClusterID: clusterID,
            }

            // Cache in Redis
            c.cachePoint(r.Context(), point)

            if err := conn.WriteJSON(point); err != nil {
                return
            }
        }
    }
}
```

### Step 9.2 — Frontend WebSocket hook

Create `frontend/src/hooks/useWebSocket.ts`:

```typescript
import { useEffect, useRef, useState, useCallback } from 'react';

type ReadyState = 'connecting' | 'open' | 'closing' | 'closed';

interface UseWebSocketReturn {
  lastMessage: MessageEvent | null;
  readyState: ReadyState;
  sendMessage: (data: string) => void;
}

const WS_BASE = import.meta.env.VITE_WS_URL ?? 'ws://localhost:8080';
const RECONNECT_INTERVAL = 3000;
const MAX_RECONNECT_ATTEMPTS = 10;

export function useWebSocket(path: string): UseWebSocketReturn {
  const [lastMessage, setLastMessage] = useState<MessageEvent | null>(null);
  const [readyState, setReadyState] = useState<ReadyState>('connecting');
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectCountRef = useRef(0);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout>>();

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const url = `${WS_BASE}${path}`;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setReadyState('open');
      reconnectCountRef.current = 0;
    };

    ws.onmessage = (event) => {
      setLastMessage(event);
    };

    ws.onclose = () => {
      setReadyState('closed');
      if (reconnectCountRef.current < MAX_RECONNECT_ATTEMPTS) {
        reconnectCountRef.current++;
        reconnectTimerRef.current = setTimeout(connect, RECONNECT_INTERVAL);
      }
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [path]);

  useEffect(() => {
    connect();
    return () => {
      clearTimeout(reconnectTimerRef.current);
      wsRef.current?.close();
    };
  }, [connect]);

  const sendMessage = useCallback((data: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(data);
    }
  }, []);

  return { lastMessage, readyState, sendMessage };
}
```

---

## 10. Phase 7 — Docker & Containerization

### Step 10.1 — Multi-stage Dockerfile

Create `Dockerfile`:

```dockerfile
# ─── Stage 1: Build React frontend ───────────────────────────────────────────
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci --frozen-lockfile

COPY frontend/ ./
RUN npm run build

# ─── Stage 2: Build Go binary ─────────────────────────────────────────────────
FROM golang:1.22-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy built frontend into Go embed path
COPY --from=frontend-builder /app/frontend/dist ./internal/api/static

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o /kubevision \
    ./cmd/server

# ─── Stage 3: Minimal runtime image ──────────────────────────────────────────
FROM scratch

COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=go-builder /kubevision /kubevision

EXPOSE 8080 9090

ENTRYPOINT ["/kubevision"]
```

### Step 10.2 — Docker Compose for local dev

Create `deploy/docker-compose.yml`:

```yaml
version: "3.9"

services:
  # ── KubeVision API ──────────────────────────────────────────────────────────
  kubevision:
    build:
      context: ..
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      DATABASE_URL: postgres://kubevision:kubevision@postgres:5432/kubevision
      REDIS_URL: redis://redis:6379
      JWT_SECRET: dev-secret-change-in-prod
      LOG_LEVEL: debug
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ~/.kube:/root/.kube:ro  # Mount local kubeconfig for development
    networks:
      - kubevision

  # ── PostgreSQL ──────────────────────────────────────────────────────────────
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: kubevision
      POSTGRES_USER: kubevision
      POSTGRES_PASSWORD: kubevision
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U kubevision"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - kubevision

  # ── Redis ───────────────────────────────────────────────────────────────────
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    networks:
      - kubevision

  # ── Prometheus ──────────────────────────────────────────────────────────────
  prometheus:
    image: prom/prometheus:v2.50.0
    command:
      - --config.file=/etc/prometheus/prometheus.yml
      - --storage.tsdb.path=/prometheus
      - --web.enable-lifecycle
      - --storage.tsdb.retention.time=30d
    ports:
      - "9099:9090"
    volumes:
      - ../monitoring/prometheus:/etc/prometheus:ro
      - prometheus-data:/prometheus
    networks:
      - kubevision

  # ── Grafana ─────────────────────────────────────────────────────────────────
  grafana:
    image: grafana/grafana:10.3.0
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_USERS_ALLOW_SIGN_UP: "false"
      GF_FEATURE_TOGGLES_ENABLE: "ngalert"
    volumes:
      - ../monitoring/grafana/datasources:/etc/grafana/provisioning/datasources:ro
      - ../monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - grafana-data:/var/lib/grafana
    depends_on:
      - prometheus
    networks:
      - kubevision

  # ── AlertManager ────────────────────────────────────────────────────────────
  alertmanager:
    image: prom/alertmanager:v0.27.0
    ports:
      - "9093:9093"
    volumes:
      - ../monitoring/alertmanager:/etc/alertmanager:ro
    networks:
      - kubevision

  # ── Loki ────────────────────────────────────────────────────────────────────
  loki:
    image: grafana/loki:2.9.0
    ports:
      - "3100:3100"
    volumes:
      - ../monitoring/loki:/etc/loki:ro
      - loki-data:/loki
    networks:
      - kubevision

volumes:
  postgres-data:
  redis-data:
  prometheus-data:
  grafana-data:
  loki-data:

networks:
  kubevision:
    driver: bridge
```

---

## 11. Phase 8 — Kubernetes Manifests & Deployment

### Step 11.1 — Install KubeVision via Helm

```bash
# Add the KubeVision Helm repo
helm repo add kubevision https://yourorg.github.io/kubevision/charts
helm repo update

# Install with custom values
helm install kubevision kubevision/kubevision \
  --namespace kubevision \
  --create-namespace \
  --set config.jwtSecret=$(openssl rand -hex 32) \
  --set ingress.hosts[0].host=kubevision.yourdomain.com \
  --set grafana.adminPassword=$(openssl rand -hex 16) \
  --values my-values.yaml

# Verify installation
kubectl get all -n kubevision
helm status kubevision -n kubevision
```

### Step 11.2 — Upgrade with new chart version

```bash
helm upgrade kubevision kubevision/kubevision \
  --namespace kubevision \
  --reuse-values \
  --set image.tag=v1.2.0 \
  --atomic \
  --timeout 5m
```

### Step 11.3 — Rollback on failure

```bash
helm history kubevision -n kubevision
helm rollback kubevision 2 -n kubevision  # Roll back to revision 2
```

---

## 12. Phase 9 — CI/CD Pipeline (GitHub Actions)

### Step 12.1 — CI workflow

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test-go:
    name: Test Go
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: kubevision_test
          POSTGRES_USER: kubevision
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true
      - name: Run tests
        run: go test ./... -v -race -coverprofile=coverage.out
        env:
          DATABASE_URL: postgres://kubevision:test@localhost:5432/kubevision_test
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out

  test-frontend:
    name: Test Frontend
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: npm
          cache-dependency-path: frontend/package-lock.json
      - run: npm ci
      - run: npm run type-check
      - run: npm run lint
      - run: npm run test

  lint-helm:
    name: Lint Helm Charts
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: azure/setup-helm@v4
      - run: helm lint charts/kubevision
      - run: helm template kubevision charts/kubevision | kubectl create --dry-run=client -f -

  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: fs
          scan-ref: '.'
          severity: 'CRITICAL,HIGH'
```

### Step 12.2 — Release workflow

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      packages: write

    steps:
      - uses: actions/checkout@v4

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          build-args: |
            VERSION=${{ github.ref_name }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Package and publish Helm charts
        run: |
          helm dependency update charts/kubevision
          helm package charts/kubevision --version ${{ github.ref_name }}
          helm repo index . --url https://yourorg.github.io/kubevision/charts

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v1
        with:
          files: "*.tgz"
          generate_release_notes: true
```

---

## 13. Phase 10 — Alerting & Notification System

### Step 13.1 — Alert engine

Create `internal/alerts/engine.go`:

```go
package alerts

import (
    "context"
    "log/slog"
    "time"

    "github.com/robfig/cron/v3"
)

type Rule struct {
    ID          string
    Name        string
    Description string
    Query       string        // PromQL expression
    Threshold   float64
    Comparator  string        // gt, lt, gte, lte, eq
    Duration    time.Duration // Alert must fire for this long before notifying
    Labels      map[string]string
    Severity    string // critical, high, medium, low
    Channels    []string // slack, pagerduty, email
    ClusterID   string
}

type FiringAlert struct {
    Rule      *Rule
    Value     float64
    FiredAt   time.Time
    Message   string
}

type Engine struct {
    rules     []*Rule
    notifier  *Notifier
    prometheus PrometheusClient
    cron      *cron.Cron
    firing    map[string]*FiringAlert // ruleID → firing alert
}

func NewEngine(prometheus PrometheusClient, notifier *Notifier) *Engine {
    return &Engine{
        prometheus: prometheus,
        notifier:   notifier,
        cron:       cron.New(cron.WithSeconds()),
        firing:     make(map[string]*FiringAlert),
    }
}

func (e *Engine) Start(ctx context.Context) {
    e.cron.AddFunc("*/30 * * * * *", func() { // Every 30 seconds
        e.evaluate(ctx)
    })
    e.cron.Start()
    <-ctx.Done()
    e.cron.Stop()
}

func (e *Engine) evaluate(ctx context.Context) {
    for _, rule := range e.rules {
        val, err := e.prometheus.Query(ctx, rule.Query)
        if err != nil {
            slog.Warn("Alert rule query failed", "rule", rule.Name, "error", err)
            continue
        }

        firing := e.compare(val, rule.Threshold, rule.Comparator)

        if firing {
            if _, alreadyFiring := e.firing[rule.ID]; !alreadyFiring {
                alert := &FiringAlert{
                    Rule:    rule,
                    Value:   val,
                    FiredAt: time.Now(),
                    Message: formatAlertMessage(rule, val),
                }
                e.firing[rule.ID] = alert
                go e.notifier.Send(ctx, alert)
            }
        } else {
            if _, wasFiring := e.firing[rule.ID]; wasFiring {
                delete(e.firing, rule.ID)
                go e.notifier.SendResolved(ctx, rule)
            }
        }
    }
}

func (e *Engine) compare(value, threshold float64, comparator string) bool {
    switch comparator {
    case "gt":  return value > threshold
    case "gte": return value >= threshold
    case "lt":  return value < threshold
    case "lte": return value <= threshold
    case "eq":  return value == threshold
    default:    return false
    }
}
```

---

## 14. Phase 11 — RBAC & Multi-Tenancy

### Step 14.1 — Database schema for RBAC

Create `internal/db/migrations/002_rbac.sql`:

```sql
-- Teams
CREATE TABLE teams (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Users
CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email        VARCHAR(255) NOT NULL UNIQUE,
    name         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    oidc_subject VARCHAR(255),
    is_admin     BOOLEAN DEFAULT FALSE,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login   TIMESTAMP WITH TIME ZONE
);

-- Team memberships
CREATE TABLE team_members (
    team_id    UUID REFERENCES teams(id) ON DELETE CASCADE,
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(50) NOT NULL DEFAULT 'member', -- admin, member, viewer
    PRIMARY KEY (team_id, user_id)
);

-- Cluster permissions (namespace-level RBAC)
CREATE TABLE cluster_permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id     UUID REFERENCES teams(id) ON DELETE CASCADE,
    cluster_id  VARCHAR(255) NOT NULL,
    namespace   VARCHAR(255) NOT NULL, -- '*' for all namespaces
    permission  VARCHAR(50)  NOT NULL, -- read, write, admin
    UNIQUE (team_id, cluster_id, namespace)
);

-- Registered clusters
CREATE TABLE clusters (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(255) NOT NULL,
    environment   VARCHAR(50) NOT NULL, -- production, staging, dev
    kubeconfig    BYTEA NOT NULL,       -- Encrypted kubeconfig
    server_url    VARCHAR(255) NOT NULL,
    version       VARCHAR(50),
    added_by      UUID REFERENCES users(id),
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit log
CREATE TABLE audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id),
    cluster_id  VARCHAR(255),
    namespace   VARCHAR(255),
    resource    VARCHAR(255) NOT NULL,
    action      VARCHAR(50)  NOT NULL, -- get, list, create, update, delete, exec
    details     JSONB,
    ip_address  INET,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_cluster ON audit_logs(cluster_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

---

## 15. Phase 12 — Local Dev Setup (Kind/Minikube)

### Step 15.1 — Kind cluster config

Create `deploy/kind-cluster.yaml`:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: kubevision-dev

nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
      - containerPort: 30080
        hostPort: 30080
        protocol: TCP

  - role: worker
    labels:
      node-role: worker
      workload: general

  - role: worker
    labels:
      node-role: worker
      workload: monitoring
```

### Step 15.2 — One-command setup script

Create `scripts/setup-local.sh`:

```bash
#!/bin/bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error()   { echo -e "${RED}[✗]${NC} $1"; }

echo ""
echo "  ██╗  ██╗██╗   ██╗██████╗ ███████╗██╗   ██╗██╗███████╗██╗ ██████╗ ███╗   ██╗"
echo "  ██║ ██╔╝██║   ██║██╔══██╗██╔════╝██║   ██║██║██╔════╝██║██╔═══██╗████╗  ██║"
echo "  █████╔╝ ██║   ██║██████╔╝█████╗  ██║   ██║██║███████╗██║██║   ██║██╔██╗ ██║"
echo "  ██╔═██╗ ██║   ██║██╔══██╗██╔══╝  ╚██╗ ██╔╝██║╚════██║██║██║   ██║██║╚██╗██║"
echo "  ██║  ██╗╚██████╔╝██████╔╝███████╗  ╚████╔╝ ██║███████║██║╚██████╔╝██║ ╚████║"
echo "  ╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝   ╚═══╝  ╚═╝╚══════╝╚═╝ ╚═════╝ ╚═╝  ╚═══╝"
echo ""

# ── Check prerequisites ──────────────────────────────────────────────────────
log_info "Checking prerequisites..."

for cmd in docker kubectl kind helm go node npm; do
    if ! command -v "$cmd" &>/dev/null; then
        log_error "$cmd is required but not installed"
        exit 1
    fi
    log_success "$cmd found: $(command -v $cmd)"
done

# ── Create Kind cluster ──────────────────────────────────────────────────────
log_info "Creating Kind cluster..."
if kind get clusters | grep -q "kubevision-dev"; then
    log_warn "Kind cluster 'kubevision-dev' already exists, skipping..."
else
    kind create cluster --config deploy/kind-cluster.yaml
    log_success "Kind cluster created"
fi

# ── Install ingress-nginx ────────────────────────────────────────────────────
log_info "Installing ingress-nginx..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=90s
log_success "ingress-nginx ready"

# ── Install metrics-server ───────────────────────────────────────────────────
log_info "Installing metrics-server..."
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm upgrade --install metrics-server metrics-server/metrics-server \
    --namespace kube-system \
    --set args[0]="--kubelet-insecure-tls"
log_success "metrics-server installed"

# ── Start Docker Compose stack ───────────────────────────────────────────────
log_info "Starting Docker Compose stack..."
docker compose -f deploy/docker-compose.yml up -d
log_success "Docker stack running"

# ── Seed demo data ───────────────────────────────────────────────────────────
log_info "Seeding demo data..."
sleep 5  # Wait for DB to be ready
bash scripts/seed-demo-data.sh

# ── Done ─────────────────────────────────────────────────────────────────────
echo ""
log_success "🎉 KubeVision is running!"
echo ""
echo "  Dashboard:    http://localhost:8080"
echo "  Grafana:      http://localhost:3000  (admin/admin)"
echo "  Prometheus:   http://localhost:9099"
echo "  AlertManager: http://localhost:9093"
echo ""
echo "  Default login: admin@kubevision.local / admin"
echo ""
```

---

## 16. Environment Variables Reference

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | PostgreSQL connection string |
| `REDIS_URL` | ✅ | — | Redis connection string |
| `JWT_SECRET` | ✅ | — | Secret for signing JWTs (min 32 chars) |
| `SESSION_TTL` | ❌ | `24h` | JWT token lifetime |
| `LOG_LEVEL` | ❌ | `info` | `debug`, `info`, `warn`, `error` |
| `PORT` | ❌ | `8080` | HTTP server port |
| `METRICS_PORT` | ❌ | `9090` | Prometheus metrics port |
| `PROMETHEUS_URL` | ❌ | `http://prometheus:9090` | Prometheus server URL |
| `GRAFANA_URL` | ❌ | `http://grafana:3000` | Grafana URL for embedding |
| `GRAFANA_API_KEY` | ❌ | — | Grafana API key for dashboard provisioning |
| `LOKI_URL` | ❌ | `http://loki:3100` | Loki URL for log queries |
| `SLACK_WEBHOOK_URL` | ❌ | — | Slack webhook for alerts |
| `PAGERDUTY_KEY` | ❌ | — | PagerDuty service key |
| `AI_API_KEY` | ❌ | — | Anthropic API key for log analysis |
| `AI_ENABLED` | ❌ | `false` | Enable AI features |
| `OIDC_ISSUER_URL` | ❌ | — | OIDC provider URL (e.g. Dex, Okta) |
| `OIDC_CLIENT_ID` | ❌ | — | OIDC client ID |
| `OIDC_CLIENT_SECRET` | ❌ | — | OIDC client secret |
| `ENCRYPTION_KEY` | ❌ | — | AES-256 key for encrypting kubeconfigs in DB |

---

## 17. API Reference

### Authentication

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/auth/login` | Email/password login → JWT |
| `POST` | `/api/v1/auth/refresh` | Refresh access token |
| `POST` | `/api/v1/auth/logout` | Invalidate session |
| `GET` | `/api/v1/auth/oidc/callback` | OIDC callback |

### Clusters

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/clusters` | List all registered clusters |
| `POST` | `/api/v1/clusters` | Register a new cluster (upload kubeconfig) |
| `GET` | `/api/v1/clusters/:id` | Get cluster overview + health |
| `DELETE` | `/api/v1/clusters/:id` | Remove cluster |
| `GET` | `/api/v1/clusters/:id/nodes` | List nodes |
| `GET` | `/api/v1/clusters/:id/events` | Cluster events |

### Workloads

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/clusters/:id/namespaces/:ns/pods` | List pods |
| `GET` | `/api/v1/clusters/:id/namespaces/:ns/pods/:pod/logs` | Stream pod logs |
| `GET` | `/api/v1/clusters/:id/namespaces/:ns/pods/:pod/exec` | WebSocket exec |
| `GET` | `/api/v1/clusters/:id/namespaces/:ns/deployments` | List deployments |
| `PATCH` | `/api/v1/clusters/:id/namespaces/:ns/deployments/:name/scale` | Scale deployment |
| `POST` | `/api/v1/clusters/:id/namespaces/:ns/deployments/:name/restart` | Rolling restart |

### Helm

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/clusters/:id/helm/releases` | List Helm releases |
| `POST` | `/api/v1/clusters/:id/helm/releases` | Install Helm chart |
| `PUT` | `/api/v1/clusters/:id/helm/releases/:name` | Upgrade release |
| `DELETE` | `/api/v1/clusters/:id/helm/releases/:name` | Uninstall release |
| `GET` | `/api/v1/clusters/:id/helm/releases/:name/history` | Release history |
| `POST` | `/api/v1/clusters/:id/helm/releases/:name/rollback` | Rollback release |
| `GET` | `/api/v1/helm/repos` | List chart repositories |
| `GET` | `/api/v1/helm/repos/:repo/charts` | List charts in repo |

### Metrics (WebSocket)

| Path | Description |
|---|---|
| `WS /ws/metrics?cluster=:id&metric=cpu` | Real-time CPU usage |
| `WS /ws/metrics?cluster=:id&metric=memory` | Real-time memory usage |
| `WS /ws/metrics?cluster=:id&metric=network_in` | Network ingress bytes |
| `WS /ws/events?cluster=:id` | Real-time Kubernetes events |
| `WS /ws/pods?cluster=:id&namespace=:ns` | Real-time pod status changes |

### AI

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/ai/analyze-logs` | Submit pod logs for AI analysis |

---

## 18. Testing Strategy

### Unit Tests (Go)
```bash
# Run all unit tests with race detection
go test ./... -v -race

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests (Go)
```bash
# Run integration tests against a real Kind cluster
KIND_CLUSTER=kubevision-dev go test ./... -tags=integration -v
```

### Frontend Tests
```bash
cd frontend
# Unit tests (Vitest)
npm run test

# E2E tests (Playwright)
npm run test:e2e

# Type checking
npm run type-check
```

### Helm Chart Testing
```bash
# Lint
helm lint charts/kubevision

# Template render check
helm template kubevision charts/kubevision --debug

# Dry-run install against Kind cluster
helm install kubevision charts/kubevision \
    --namespace kubevision \
    --create-namespace \
    --dry-run \
    --debug
```

### Load Testing
```bash
# Use k6 for API load testing
k6 run tests/load/cluster-overview.js
```

---

## 19. Cursor AI Instructions

> **Read this section carefully before opening Cursor AI.**

### How to use this document with Cursor

1. **Open Cursor** in the `kubevision/` directory after running `make setup-local`
2. **Use Cursor Chat (Cmd+L)** with the following prompts, in order:

---

#### Prompt 1 — Scaffold the project

```
Using the directory structure in KUBEVISION_PROJECT.md, create all the files and directories.
Start with:
1. go.mod with all dependencies listed
2. The full Makefile
3. cmd/server/main.go
4. internal/api/router.go with all routes registered
5. internal/db/postgres.go and redis.go
Do not skip any file. Create stubs for any handler not yet implemented.
```

#### Prompt 2 — Implement the Kubernetes client layer

```
Implement the full internal/k8s/ package:
- client.go: multi-cluster manager as described
- watcher.go: use SharedInformerFactory to watch pods, nodes, deployments, events
  and broadcast events to registered WebSocket listeners
- exec.go: full pod exec WebSocket streaming implementation
- operator/: CRD types and reconciler loop skeleton

Use client-go v0.29.0 patterns. Add proper context cancellation and graceful shutdown.
```

#### Prompt 3 — Implement API handlers

```
Implement all handlers in internal/api/handlers/:
- Each handler should read clusterID from the URL parameter
- Look up the cluster client from ClusterManager
- Call the Kubernetes API
- Return JSON responses with proper status codes and error handling
- Add the audit log middleware call for all write operations
- Implement pagination (limit/offset) for list endpoints

Use the API reference table in KUBEVISION_PROJECT.md for all routes.
```

#### Prompt 4 — Build the React frontend

```
Build the complete React frontend in frontend/src/:
1. Set up the design system with all tokens from tokens.ts
2. Create a Sidebar navigation component with:
   - Cluster selector dropdown at the top
   - Nav items: Overview, Nodes, Workloads, Pods, Services, Helm Releases, Chart Browser, Metrics, Logs, Alerts, AI Analyzer, Settings
   - Connection status indicator (WebSocket status)
3. Build the OverviewPage with all sub-components listed in the file
4. Build the PodListPage with:
   - Filterable/sortable table
   - Status badges with correct colors from tokens
   - Click-to-expand row with pod details, containers, and quick actions (logs, exec, delete)
5. Build the LiveMetricChart component with WebSocket subscription
6. Apply the dark theme using Tailwind classes matching the design tokens

Use Framer Motion for page transitions and card appearances.
Use TanStack Query for all API calls with proper loading/error states.
```

#### Prompt 5 — Implement real-time streaming

```
Implement the full WebSocket layer:

Backend (internal/metrics/websocket.go):
- Upgrade HTTP to WebSocket
- Subscribe to Prometheus for the requested metric
- Poll every 5 seconds and send MetricPoint JSON
- Cache last 60 points in Redis per cluster+metric key
- Send historical data on connect

Backend (internal/k8s/watcher.go):
- Use SharedInformerFactory for pods, nodes, deployments
- Maintain a hub of WebSocket connections per cluster
- Broadcast add/update/delete events as JSON to all connected clients

Frontend (src/hooks/useWebSocket.ts):
- Implement auto-reconnect with exponential backoff
- Implement the useMetrics hook that subscribes to cluster metrics
- Implement the useK8sEvents hook for real-time pod/node events
```

#### Prompt 6 — Helm integration

```
Implement the full Helm integration:

Backend (internal/helm/):
- client.go: create Helm action.Configuration per cluster
- release.go: Install, Upgrade, Rollback, Uninstall with proper options
- repo.go: AddRepo, RemoveRepo, UpdateRepos, SearchCharts

Handlers (internal/api/handlers/helm.go):
- All endpoints from the API reference
- Upgrade should diff current vs new values and return the diff

Frontend (src/pages/HelmReleases/):
- Release list with status, chart name, version, namespace, last deployed
- Release detail modal with: values editor (Monaco), upgrade form, history table, rollback button
- Chart browser page with search, filters, install dialog with values form
```

#### Prompt 7 — AI log analyzer

```
Build the AI log analysis feature:

Backend (internal/ai/analyzer.go):
- POST /api/v1/ai/analyze-logs
- Accept: { podName, namespace, clusterId, logs: string }
- Call the Claude API (claude-sonnet-4-20250514) with a system prompt:
  "You are a Kubernetes expert. Analyze the following pod logs and provide:
   1. Root cause of any errors (be specific)
   2. Severity: critical/high/medium/low
   3. Step-by-step remediation commands
   4. Whether this is a known Kubernetes issue
   Format response as JSON: { rootCause, severity, remediation: string[], knownIssue: boolean, knownIssueName? }"
- Stream the response back to the frontend

Frontend (src/pages/AIAnalyzer/):
- Textarea for pasting logs (or auto-populated from a pod)
- "Analyze" button
- Animated streaming response display
- Rendered remediation steps with copy-to-clipboard kubectl commands
- Severity badge with appropriate color
```

#### Prompt 8 — Docker, Helm chart, and CI/CD

```
Create all the Docker and deployment files:
1. Multi-stage Dockerfile using Go embed for the frontend
2. docker-compose.yml with all services (kubevision, postgres, redis, prometheus, grafana, alertmanager, loki)
3. All Helm chart files in charts/kubevision/templates/ matching the structure
4. values.schema.json with JSON schema validation for all values
5. All 3 GitHub Actions workflows (ci.yml, release.yml, helm-release.yml)
6. scripts/setup-local.sh and scripts/package-charts.sh

Make sure the Helm chart passes helm lint and helm template.
```

#### Prompt 9 — Polish and production-readiness

```
Add production-readiness features:
1. Rate limiting middleware using Redis sliding window (100 req/min per user)
2. Request ID middleware (X-Request-ID header)
3. Structured logging (log/slog) on all handlers with request ID correlation
4. Prometheus metrics endpoint (/metrics) exposing:
   - HTTP request duration histogram
   - Active WebSocket connections gauge
   - Kubernetes API call duration histogram per cluster
5. Health check endpoints: /healthz (liveness) and /readyz (readiness)
6. Graceful shutdown: drain WebSocket connections, close DB pools
7. Frontend: error boundary component, global toast notifications for WS disconnect/reconnect
8. PodDisruptionBudget in Helm chart
9. NetworkPolicy in Helm chart
10. Resource limits/requests in all Helm templates
```

---

### General Cursor Tips for this project

- **Always have `internal/api/router.go` open** when working on handlers — Cursor uses it as context
- **Use `@codebase` mentions** when asking about architecture decisions
- **Run `make test` after each phase** to catch regressions early
- **Use Cursor's diff view** for Helm template changes — YAML indentation matters
- For WebSocket debugging: use `wscat -c ws://localhost:8080/ws/metrics?cluster=local&metric=cpu`

---

*Built with ❤️ by the KubeVision team. Star us on GitHub if this saves your on-call nights.*
