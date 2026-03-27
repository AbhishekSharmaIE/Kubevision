# 🤖 Cursor AI — Git-Integrated Build Prompts for KubeVision
# Every file created → staged → committed → pushed. No exceptions.
ok
---

## ⚙️ ONE-TIME SETUP — Run this first in Cursor Terminal before anything else

```bash
# 1. Create the GitHub repo (install gh CLI first: https://cli.github.com)
gh auth login
gh repo create kubevision --public --description "KubeVision — Real-time Kubernetes Cluster Management Dashboard" --clone
cd kubevision

# 2. Set your git identity if not already set
git config user.name "Your Name"
git config user.email "you@yourorg.com"

# 3. Create a .gitignore immediately
cat > .gitignore << 'EOF'
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Go
*.test
*.out
coverage.out
vendor/

# Node
node_modules/
frontend/dist/
frontend/.env.local

# Env files
.env
.env.*
!.env.example

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Helm
charts/*.tgz
!charts/index.yaml

# Docker
.docker/

# Kubernetes secrets
*-secret.yaml
kubeconfig
*.kubeconfig
EOF

git add .gitignore
git commit -m "chore: initial repo setup with .gitignore"
git push origin main
```

---

## 🔴 MASTER RULE FOR CURSOR — PASTE THIS AT THE TOP OF EVERY PROMPT

```
ABSOLUTE RULE: After creating or modifying ANY file, you MUST immediately run these git commands via the terminal tool:

  git add <the exact files you just created or modified>
  git commit -m "<type>(<scope>): <what you did>"
  git push origin main

Never batch multiple phases into one commit. Every logical unit of work = one commit = one push.
Commit message format: https://www.conventionalcommits.org
Types: feat | fix | chore | docs | refactor | test | ci | style | build

If a file already exists and you're modifying it, show me the diff first, then commit.
If any git command fails, stop and show me the error before continuing.
```

---

## 📋 PROMPT 0 — Git Pre-Flight Check (Run before Phase 1)

Paste into Cursor Chat:

```
@KUBEVISION_PROJECT.md

Before we write a single line of application code, set up the git workflow.

Using the terminal tool, run these commands one by one and show me each output:

1. Verify we're in the right directory:
   pwd && ls

2. Confirm git is initialized and connected to remote:
   git remote -v
   git status
   git log --oneline -5

3. Create the full directory scaffold (no files yet, just dirs):
   mkdir -p cmd/server
   mkdir -p internal/{api/{middleware,handlers},k8s/{operator},helm,metrics,alerts,auth,db/migrations,ai}
   mkdir -p frontend/src/{design-system,pages/{Overview,Nodes,Workloads,Pods,Services,HelmReleases,HelmRepo,Metrics,Logs,Alerts,AIAnalyzer,Settings},stores,hooks,api,utils}
   mkdir -p charts/{kubevision/templates,monitoring-stack}
   mkdir -p monitoring/{prometheus/rules,grafana/{datasources,dashboards},alertmanager,loki}
   mkdir -p deploy scripts .github/workflows

4. Create a .env.example file:
   touch .env.example
   
   Write these contents into it:
   DATABASE_URL=postgres://kubevision:kubevision@localhost:5432/kubevision
   REDIS_URL=redis://localhost:6379
   JWT_SECRET=change-me-to-a-32-char-random-string
   LOG_LEVEL=debug
   PORT=8080
   METRICS_PORT=9090
   PROMETHEUS_URL=http://localhost:9099
   GRAFANA_URL=http://localhost:3000
   GRAFANA_API_KEY=
   LOKI_URL=http://localhost:3100
   SLACK_WEBHOOK_URL=
   PAGERDUTY_KEY=
   AI_API_KEY=
   AI_ENABLED=false
   OIDC_ISSUER_URL=
   OIDC_CLIENT_ID=
   OIDC_CLIENT_SECRET=
   ENCRYPTION_KEY=

5. Commit and push the scaffold:
   git add -A
   git commit -m "chore(scaffold): initialize full directory structure and env example"
   git push origin main
   git log --oneline -3

Show me the output of every command. Do not skip any step.
```

---

## 📋 PROMPT 1 — Go Module + Makefile + Entry Point

```
@KUBEVISION_PROJECT.md

RULE: After EVERY file you create, immediately run:
  git add <filename> && git commit -m "feat(<scope>): <description>" && git push origin main

Phase 1: Project scaffolding — Go backend foundation.

Create these files in this exact order, committing after each one:

─── FILE 1: go.mod ────────────────────────────────────────────────────────────
Create go.mod with module github.com/yourorg/kubevision, Go 1.22.
After creating: git add go.mod && git commit -m "chore(go): initialize go module" && git push origin main

─── FILE 2: Makefile ──────────────────────────────────────────────────────────
Create the full Makefile from the project doc with all targets:
dev, build, test, lint, docker-build, helm-package, setup-local, generate-mocks, migrate-up, migrate-down.

Add one more target not in the doc:
  git-status:
    @git status
    @git log --oneline -5

After creating: git add Makefile && git commit -m "chore(make): add Makefile with all dev targets" && git push origin main

─── FILE 3: cmd/server/main.go ────────────────────────────────────────────────
Create the full main.go from the project doc.
After creating: git add cmd/server/main.go && git commit -m "feat(server): add main entry point with graceful shutdown" && git push origin main

─── FILE 4: Install Go dependencies ───────────────────────────────────────────
Run in terminal:
  go get github.com/gin-gonic/gin
  go get github.com/gorilla/websocket
  go get k8s.io/client-go@v0.29.0
  go get k8s.io/api@v0.29.0
  go get k8s.io/metrics@v0.29.0
  go get sigs.k8s.io/controller-runtime@v0.17.0
  go get helm.sh/helm/v3
  go get github.com/prometheus/client_golang/prometheus
  go get github.com/prometheus/client_golang/prometheus/promhttp
  go get github.com/golang-jwt/jwt/v5
  go get github.com/jackc/pgx/v5
  go get github.com/redis/go-redis/v9
  go get github.com/slack-go/slack
  go get github.com/robfig/cron/v3
  go get go.uber.org/zap
  go mod tidy

After: git add go.mod go.sum && git commit -m "chore(deps): add all Go dependencies" && git push origin main

─── FILE 5: internal/db/postgres.go ───────────────────────────────────────────
Create a full PostgreSQL connection pool using pgx/v5 pool.
Include: MustConnectPostgres(), GetPool(), health check ping on startup.
Read DATABASE_URL from environment.

After: git add internal/db/postgres.go && git commit -m "feat(db): add PostgreSQL connection pool" && git push origin main

─── FILE 6: internal/db/redis.go ──────────────────────────────────────────────
Create Redis client using go-redis/v9.
Include: MustConnectRedis(), Ping health check, helper Set/Get/Del/Exists wrappers.
Read REDIS_URL from environment.

After: git add internal/db/redis.go && git commit -m "feat(db): add Redis client with helpers" && git push origin main

─── FILE 7: internal/db/migrations/ ───────────────────────────────────────────
Create all 3 SQL migration files:
  - 001_init.sql: create extensions (uuid-ossp, pgcrypto), users table, clusters table
  - 002_rbac.sql: full RBAC schema from project doc (teams, team_members, cluster_permissions, audit_logs)
  - 003_alert_rules.sql: alert_rules table with columns: id, name, description, query (promql), threshold, comparator, duration, severity, channels (jsonb), cluster_id, created_by, enabled, created_at, updated_at

Also create internal/db/migrate.go with RunMigrations() that applies all .sql files in order.

After: 
  git add internal/db/migrations/ internal/db/migrate.go
  git commit -m "feat(db): add SQL migrations for init, RBAC, and alert rules"
  git push origin main

After all files are done, run:
  go build ./... 2>&1
Show me the output. If there are errors, fix them and commit the fix:
  git add -A && git commit -m "fix(build): resolve build errors after initial scaffold" && git push origin main
```

---

## 📋 PROMPT 2 — Kubernetes Client Layer

```
@KUBEVISION_PROJECT.md

RULE: Every file = its own git commit + push immediately after creation.

Phase 2: Kubernetes multi-cluster client layer.

─── FILE 1: internal/k8s/client.go ────────────────────────────────────────────
Full multi-cluster manager from the project doc. Add these extras:
- RemoveCluster(id string) method
- HealthCheck(ctx, id string) error method that pings the cluster API server
- ClusterStatus struct { ID, Name, Ready bool, Version, NodeCount int, LastChecked time.Time }
- GetStatus(id string) ClusterStatus method

git add internal/k8s/client.go && git commit -m "feat(k8s): add multi-cluster client manager with health checks" && git push origin main

─── FILE 2: internal/k8s/watcher.go ───────────────────────────────────────────
Implement a real-time event broadcaster using client-go SharedInformerFactory.
Watch: Pods, Nodes, Deployments, Services, Events (core/v1).
Pattern:
  - WatcherHub struct that holds: map[clusterID]*clusterWatcher
  - clusterWatcher uses SharedInformerFactory, runs informers for each resource
  - On Add/Update/Delete: serialize the object to a WatchEvent{Type, Object, ClusterID, Timestamp}
  - Broadcast WatchEvent to all registered WebSocket subscribers via a channel
  - Register(clusterID, conn) and Unregister(clusterID, conn) methods

git add internal/k8s/watcher.go && git commit -m "feat(k8s): add SharedInformer-based real-time event broadcaster" && git push origin main

─── FILE 3: internal/k8s/exec.go ──────────────────────────────────────────────
Full pod exec WebSocket implementation from the project doc.
Add resize support: if first byte of a message is 0x01, parse it as JSON TerminalSize{Width, Height}.
Otherwise treat it as stdin data.

git add internal/k8s/exec.go && git commit -m "feat(k8s): add WebSocket pod exec with terminal resize support" && git push origin main

─── FILE 4: internal/k8s/operator/types.go ────────────────────────────────────
Define a custom CRD type: ClusterAlert
  - TypeMeta, ObjectMeta
  - Spec: { AlertName, Query, Threshold, Comparator, Severity, Channels []string }
  - Status: { State (firing|resolved|unknown), LastEvaluated, FiredAt, Message }
Register it with scheme.

git add internal/k8s/operator/types.go && git commit -m "feat(operator): define ClusterAlert CRD types" && git push origin main

─── FILE 5: internal/k8s/operator/controller.go ───────────────────────────────
Skeleton reconciler using controller-runtime.
Implement Reconcile(ctx, req) that:
1. Fetches the ClusterAlert object
2. Queries Prometheus with the spec query
3. Evaluates the threshold
4. Updates the Status subresource
5. Requeues every 30 seconds via ctrl.Result{RequeueAfter: 30 * time.Second}

git add internal/k8s/operator/controller.go && git commit -m "feat(operator): add ClusterAlert reconciler with Prometheus evaluation" && git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
go build ./internal/k8s/...
If errors: fix → git add -A && git commit -m "fix(k8s): resolve compilation errors" && git push origin main

Final status check:
git log --oneline -8
```

---

## 📋 PROMPT 3 — Auth, JWT, Middleware

```
@KUBEVISION_PROJECT.md

RULE: One file = one commit = one push. Show me each git output.

Phase 3: Authentication, JWT, and API middleware.

─── FILE 1: internal/auth/jwt.go ──────────────────────────────────────────────
Implement JWT using golang-jwt/jwt/v5:
- Claims struct: { UserID, Email, IsAdmin, TeamIDs []string, StandardClaims }
- IssueToken(user) (string, error) — HS256, configurable expiry from env
- ValidateToken(tokenString) (*Claims, error)
- RefreshToken(tokenString) (string, error) — validate old, issue new with fresh expiry

git add internal/auth/jwt.go && git commit -m "feat(auth): add JWT issue/validate/refresh" && git push origin main

─── FILE 2: internal/auth/session.go ──────────────────────────────────────────
Redis-backed session store:
- StoreSession(ctx, userID, token, ttl) error
- ValidateSession(ctx, userID, token) bool
- RevokeSession(ctx, userID, token) error
- RevokeAllSessions(ctx, userID) error

git add internal/auth/session.go && git commit -m "feat(auth): add Redis-backed session store" && git push origin main

─── FILE 3: internal/api/middleware/auth.go ────────────────────────────────────
Gin middleware:
- RequireAuth() gin.HandlerFunc — extract Bearer token, validate JWT, store Claims in context
- RequireAdmin() gin.HandlerFunc — check IsAdmin from claims
- GetClaims(c *gin.Context) *auth.Claims — helper

git add internal/api/middleware/auth.go && git commit -m "feat(middleware): add JWT auth middleware" && git push origin main

─── FILE 4: internal/api/middleware/rbac.go ────────────────────────────────────
Namespace-level RBAC middleware:
- RequirePermission(permission string) gin.HandlerFunc
- Reads clusterID and namespace from URL params
- Queries cluster_permissions table to check if user's team has the required permission
- Permissions: "read", "write", "admin"
- Admins bypass RBAC

git add internal/api/middleware/rbac.go && git commit -m "feat(middleware): add namespace-level RBAC middleware" && git push origin main

─── FILE 5: internal/api/middleware/ratelimit.go ───────────────────────────────
Redis sliding window rate limiter:
- 100 requests per minute per user (key: "ratelimit:{userID}")
- Returns 429 with Retry-After header when exceeded
- Skip rate limiting for /healthz and /readyz

git add internal/api/middleware/ratelimit.go && git commit -m "feat(middleware): add Redis sliding window rate limiter" && git push origin main

─── FILE 6: internal/api/middleware/audit.go ───────────────────────────────────
Audit log middleware:
- Log every request to audit_logs table: userID, clusterID, namespace, resource (from path), action (GET→list/get, POST→create, PUT/PATCH→update, DELETE→delete), IP address, timestamp
- Only log write operations (POST, PUT, PATCH, DELETE) and exec endpoints

git add internal/api/middleware/audit.go && git commit -m "feat(middleware): add audit log middleware" && git push origin main

─── FILE 7: internal/api/middleware/requestid.go ──────────────────────────────
Generate UUID X-Request-ID header on every request.
Attach to context and structured logger.

git add internal/api/middleware/requestid.go && git commit -m "feat(middleware): add request ID middleware" && git push origin main

─── FILE 8: internal/api/router.go ─────────────────────────────────────────────
Full Gin router with ALL routes from the API reference in the project doc.
Apply middleware in this order: RequestID → RateLimit → CORS → Auth (on protected routes) → RBAC (on cluster routes) → Audit (on write routes).

Route groups:
  /healthz          → health check (no auth)
  /readyz           → readiness check (no auth)
  /metrics          → Prometheus metrics (no auth)
  /api/v1/auth/...  → auth handlers (no auth middleware)
  /api/v1/...       → all other handlers (RequireAuth)
  /ws/...           → WebSocket handlers (RequireAuth via query param token)

git add internal/api/router.go && git commit -m "feat(api): add full Gin router with all routes and middleware chain" && git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
go build ./...
git log --oneline -10
```

---

## 📋 PROMPT 4 — All API Handlers (Go)

```
@KUBEVISION_PROJECT.md

RULE: Each handler file = its own commit. Run git add + commit + push after EACH file.

Phase 4: Implement every API handler. Use this response pattern for all:
  - Success: c.JSON(http.StatusOK, gin.H{"data": result})
  - Error:   c.JSON(http.StatusXXX, gin.H{"error": err.Error(), "requestId": requestID})
  - Pagination: accept ?limit=50&offset=0, return { data, total, limit, offset }

─── FILE 1: internal/api/handlers/auth.go ──────────────────────────────────────
POST /auth/login  → validate email+password against users table, bcrypt compare, issue JWT, store session
POST /auth/refresh → validate old token, issue new
POST /auth/logout  → revoke session in Redis

git add internal/api/handlers/auth.go && git commit -m "feat(handlers): add auth login/refresh/logout handlers" && git push origin main

─── FILE 2: internal/api/handlers/clusters.go ──────────────────────────────────
GET    /clusters              → list all clusters user has access to
POST   /clusters              → accept multipart kubeconfig upload, register cluster, encrypt kubeconfig
GET    /clusters/:id          → overview: node count, pod count, version, health status, namespace count
DELETE /clusters/:id          → remove cluster registration
GET    /clusters/:id/nodes    → list all nodes with: name, status, roles, CPU%, memory%, pod count, conditions
GET    /clusters/:id/events   → last 100 events, filterable by namespace and type (Warning/Normal)

git add internal/api/handlers/clusters.go && git commit -m "feat(handlers): add cluster CRUD and node/event handlers" && git push origin main

─── FILE 3: internal/api/handlers/pods.go ──────────────────────────────────────
GET    /clusters/:id/namespaces/:ns/pods                         → list pods with status, containers, restart count, age, node
GET    /clusters/:id/namespaces/:ns/pods/:pod                    → pod detail with full spec + status
GET    /clusters/:id/namespaces/:ns/pods/:pod/logs               → stream logs (SSE), accept ?container=&tailLines=&follow=true
GET    /clusters/:id/namespaces/:ns/pods/:pod/exec               → WebSocket exec (upgrade and call k8s.ExecPod)
DELETE /clusters/:id/namespaces/:ns/pods/:pod                    → delete pod
GET    /clusters/:id/namespaces/:ns/pods/:pod/containers         → list containers in pod

git add internal/api/handlers/pods.go && git commit -m "feat(handlers): add pod list/detail/logs/exec/delete handlers" && git push origin main

─── FILE 4: internal/api/handlers/deployments.go ───────────────────────────────
GET    /clusters/:id/namespaces/:ns/deployments          → list with: name, replicas, ready, age, image, strategy
GET    /clusters/:id/namespaces/:ns/deployments/:name    → full deployment detail + events
PATCH  /clusters/:id/namespaces/:ns/deployments/:name/scale    → body: {replicas: int}
POST   /clusters/:id/namespaces/:ns/deployments/:name/restart  → patch annotation to trigger rolling restart
PUT    /clusters/:id/namespaces/:ns/deployments/:name          → apply updated YAML (strategic merge patch)

git add internal/api/handlers/deployments.go && git commit -m "feat(handlers): add deployment management handlers" && git push origin main

─── FILE 5: internal/api/handlers/namespaces.go ────────────────────────────────
GET    /clusters/:id/namespaces      → list namespaces with resource quota and pod count
POST   /clusters/:id/namespaces      → create namespace
DELETE /clusters/:id/namespaces/:ns  → delete namespace (with confirmation param ?confirm=true)

git add internal/api/handlers/namespaces.go && git commit -m "feat(handlers): add namespace CRUD handlers" && git push origin main

─── FILE 6: internal/api/handlers/services.go ──────────────────────────────────
GET /clusters/:id/namespaces/:ns/services    → list services with type, clusterIP, externalIP, ports, age
GET /clusters/:id/namespaces/:ns/ingresses   → list ingresses with hosts, paths, TLS status

git add internal/api/handlers/services.go && git commit -m "feat(handlers): add service and ingress list handlers" && git push origin main

─── FILE 7: internal/api/handlers/nodes.go ─────────────────────────────────────
GET  /clusters/:id/nodes/:node              → node detail: labels, taints, conditions, allocatable vs capacity
POST /clusters/:id/nodes/:node/cordon       → cordon node
POST /clusters/:id/nodes/:node/uncordon     → uncordon node
POST /clusters/:id/nodes/:node/drain        → drain node (evict all non-daemonset pods)

git add internal/api/handlers/nodes.go && git commit -m "feat(handlers): add node detail, cordon, uncordon, drain handlers" && git push origin main

─── FILE 8: internal/api/handlers/helm.go ──────────────────────────────────────
All Helm handlers from the API reference:
GET    releases, POST install, PUT upgrade (diff values first), DELETE uninstall
GET    release history, POST rollback
GET    repos list, POST add repo, DELETE remove repo
GET    charts in repo with search, GET chart detail + default values

git add internal/api/handlers/helm.go && git commit -m "feat(handlers): add complete Helm release and repo management handlers" && git push origin main

─── FILE 9: internal/api/handlers/metrics.go ───────────────────────────────────
GET  /clusters/:id/metrics/summary   → current CPU%, memory%, network, pod counts (from Prometheus)
GET  /clusters/:id/metrics/history   → PromQL range query proxy, accept ?query=&start=&end=&step=
WS   /ws/metrics                     → upgrade and call metrics.StreamMetrics

git add internal/api/handlers/metrics.go && git commit -m "feat(handlers): add metrics summary, history, and WebSocket streaming" && git push origin main

─── FILE 10: internal/api/handlers/alerts.go ───────────────────────────────────
GET    /alerts             → list all alert rules for user's clusters
POST   /alerts             → create alert rule
PUT    /alerts/:id         → update rule
DELETE /alerts/:id         → delete rule
GET    /alerts/firing      → currently firing alerts across all user's clusters

git add internal/api/handlers/alerts.go && git commit -m "feat(handlers): add alert rule CRUD and firing alerts endpoint" && git push origin main

─── FILE 11: internal/api/handlers/users.go ────────────────────────────────────
GET    /users              → list users (admin only)
POST   /users              → create user (admin only)
PUT    /users/:id          → update user
DELETE /users/:id          → delete user (admin only)
GET    /users/me           → current user profile
PUT    /users/me           → update own profile

git add internal/api/handlers/users.go && git commit -m "feat(handlers): add user management handlers" && git push origin main

─── FILE 12: internal/api/handlers/ai.go ───────────────────────────────────────
POST /ai/analyze-logs:
  - Accept: { podName, namespace, clusterId, logs string, containers []string }
  - If AI_ENABLED=false, return 503
  - Call Anthropic claude-sonnet-4-20250514 with system prompt defined in project doc
  - Return structured JSON response
  - Log usage to a new ai_analysis_logs table (podName, clusterId, tokensUsed, createdAt)

git add internal/api/handlers/ai.go && git commit -m "feat(handlers): add AI log analysis handler with Claude integration" && git push origin main

─── FILE 13: internal/api/handlers/health.go ───────────────────────────────────
GET /healthz → always 200 {"status": "ok"}
GET /readyz  → check DB ping + Redis ping, return 200 if both up, 503 if either down

git add internal/api/handlers/health.go && git commit -m "feat(handlers): add liveness and readiness health check handlers" && git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
go build ./...
go vet ./...
git log --oneline -15
```

---

## 📋 PROMPT 5 — Metrics Collector + WebSocket Hub

```
@KUBEVISION_PROJECT.md

RULE: Every file = commit + push immediately.

Phase 5: Real-time metrics collection and WebSocket streaming.

─── FILE 1: internal/metrics/collector.go ──────────────────────────────────────
Prometheus query client:
- NewCollector(clusterManager, redisClient)
- FetchMetric(ctx, clusterID, metric string) (float64, error)
  Metrics map:
    "cpu"         → avg(1 - rate(node_cpu_seconds_total{mode="idle"}[5m])) by cluster * 100
    "memory"      → (1 - node_memory_MemAvailable_bytes/node_memory_MemTotal_bytes) * 100
    "network_in"  → sum(rate(node_network_receive_bytes_total[5m]))
    "network_out" → sum(rate(node_network_transmit_bytes_total[5m]))
    "pods_running" → count(kube_pod_status_phase{phase="Running"})
    "pods_failed"  → count(kube_pod_status_phase{phase="Failed"})
- GetTopPodsByCPU(ctx, clusterID, namespace, limit) → []PodMetric
- GetNodeMetrics(ctx, clusterID) → []NodeMetric
- Start(ctx) background goroutine that pre-fetches and caches all metrics every 15s

git add internal/metrics/collector.go && git commit -m "feat(metrics): add Prometheus metric collector with caching" && git push origin main

─── FILE 2: internal/metrics/cache.go ──────────────────────────────────────────
Redis metric cache:
- cachePoint(ctx, MetricPoint) — store in Redis list "metrics:{clusterID}:{metric}", trim to 60 entries
- GetHistory(clusterID, metric, limit) []MetricPoint — fetch from Redis, deserialize
- CacheNodeMetrics(ctx, clusterID, metrics) 
- GetCachedNodeMetrics(ctx, clusterID) ([]NodeMetric, error)

git add internal/metrics/cache.go && git commit -m "feat(metrics): add Redis-backed metric history cache" && git push origin main

─── FILE 3: internal/metrics/websocket.go ──────────────────────────────────────
Full WebSocket metric streamer from project doc:
- Upgrade HTTP → WebSocket
- Send historical data first (last 60 points from Redis)
- Then poll Prometheus every 5s and push new MetricPoint
- Handle client disconnect gracefully
- Support multiple concurrent clients per cluster+metric pair using a Hub pattern:
  Hub{ subscribers map[string][]chan MetricPoint }
  Subscribe(clusterID, metric) chan MetricPoint
  Unsubscribe(clusterID, metric, ch)
  The collector publishes to the hub every 5s; each WS handler reads from its channel

git add internal/metrics/websocket.go && git commit -m "feat(metrics): add WebSocket streaming hub for real-time metrics" && git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
go build ./internal/metrics/...
git log --oneline -5
```

---

## 📋 PROMPT 6 — Alerts Engine + Notifier

```
@KUBEVISION_PROJECT.md

RULE: Every file = commit + push.

Phase 6: Alert engine and notification dispatcher.

─── FILE 1: internal/alerts/engine.go ──────────────────────────────────────────
Full alert engine from project doc. Extend with:
- LoadRulesFromDB(ctx, db) — load all enabled rules from alert_rules table on startup
- SaveFiringState(ctx) — persist firing alerts to Redis so they survive restarts
- REST API to add/update/remove rules at runtime without restart
- Per-rule "duration" evaluation: alert must be firing for X minutes before notifying (pending state)
- States: inactive → pending → firing → resolved

git add internal/alerts/engine.go && git commit -m "feat(alerts): add full alert evaluation engine with pending state" && git push origin main

─── FILE 2: internal/alerts/notifier.go ────────────────────────────────────────
Multi-channel notifier:
- Send(ctx, alert *FiringAlert) — dispatch to all channels in rule.Channels
- SendResolved(ctx, rule *Rule) — send resolution notification
- Implement 3 channels:
  slack: use slack-go, format message with Block Kit (colored attachment, fields for value/threshold/cluster)
  pagerduty: HTTP POST to PagerDuty Events API v2
  email: use net/smtp with HTML template (table with alert details)
- Retry failed notifications up to 3 times with 10s backoff

git add internal/alerts/notifier.go && git commit -m "feat(alerts): add Slack, PagerDuty, and email notification channels" && git push origin main

─── FILE 3: internal/alerts/rules.go ───────────────────────────────────────────
- ParsePromQL(query string) error — validate PromQL syntax via Prometheus API
- ValidateRule(rule *Rule) []string — return list of validation errors
- FormatAlertMessage(rule, value) string — human-readable alert message template

git add internal/alerts/rules.go && git commit -m "feat(alerts): add PromQL validation and rule helpers" && git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
go build ./internal/alerts/...
git log --oneline -5
```

---

## 📋 PROMPT 7 — Helm Integration Layer

```
@KUBEVISION_PROJECT.md

RULE: Every file = commit + push.

Phase 7: Full Helm SDK integration.

─── FILE 1: internal/helm/client.go ────────────────────────────────────────────
Create Helm action.Configuration per cluster:
- NewHelmClient(clusterClient *k8s.ClusterClient, namespace string) (*HelmClient, error)
- Use the cluster's rest.Config to build the Helm k8s driver
- Cache clients per (clusterID, namespace) to avoid recreating

git add internal/helm/client.go && git commit -m "feat(helm): add per-cluster Helm action client factory" && git push origin main

─── FILE 2: internal/helm/release.go ───────────────────────────────────────────
Implement all release operations:
- ListReleases(ctx, clusterID, namespace) ([]ReleaseInfo, error)
- GetRelease(ctx, clusterID, namespace, name) (*ReleaseInfo, error)
- GetHistory(ctx, clusterID, namespace, name) ([]ReleaseRevision, error)
- Install(ctx, clusterID, namespace, chartRef, releaseName, values map) error
- Upgrade(ctx, clusterID, namespace, name, chartRef, values map) (*DiffResult, error)
  — before upgrading, compute a diff of old vs new values and return it
- Rollback(ctx, clusterID, namespace, name, revision int) error
- Uninstall(ctx, clusterID, namespace, name) error

ReleaseInfo struct: Name, Namespace, Chart, Version, AppVersion, Status, Revision, LastDeployed, Values, Notes

git add internal/helm/release.go && git commit -m "feat(helm): implement full Helm release lifecycle (install/upgrade/rollback/uninstall)" && git push origin main

─── FILE 3: internal/helm/repo.go ──────────────────────────────────────────────
Chart repository management:
- AddRepo(ctx, name, url, username, password string) error
- RemoveRepo(ctx, name string) error
- UpdateRepos(ctx) error
- SearchCharts(ctx, keyword string) ([]ChartInfo, error)
- GetChart(ctx, repoName, chartName, version string) (*ChartDetail, error)
- GetDefaultValues(ctx, repoName, chartName, version string) (string, error) — return YAML string

git add internal/helm/repo.go && git commit -m "feat(helm): add Helm chart repo management and search" && git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
go build ./internal/helm/...
git log --oneline -5
```

---

## 📋 PROMPT 8 — React Frontend Foundation

```
@KUBEVISION_PROJECT.md

RULE: Every file = commit + push. For frontend, commit logical groups (a component + its types).

Phase 8: React frontend — foundation, design system, layout.

─── STEP 1: Initialize Vite + install deps ─────────────────────────────────────
In terminal:
  cd frontend
  npm create vite@latest . -- --template react-ts --force
  npm install
  npm install -D tailwindcss @tailwindcss/typography postcss autoprefixer
  npm install framer-motion zustand @tanstack/react-query axios
  npm install recharts
  npm install xterm xterm-addon-fit xterm-addon-web-links xterm-addon-search
  npm install @monaco-editor/react
  npm install react-router-dom
  npm install js-yaml @types/js-yaml
  npm install react-hot-toast
  npm install lucide-react
  npm install @types/node -D
  npx tailwindcss init -p

cd ..
git add frontend/package.json frontend/package-lock.json frontend/vite.config.ts
git commit -m "feat(frontend): initialize Vite React TypeScript project with all dependencies"
git push origin main

─── STEP 2: Tailwind config ─────────────────────────────────────────────────────
Configure frontend/tailwind.config.ts with:
- Dark mode: 'class'
- Content: ['./index.html', './src/**/*.{ts,tsx}']
- Extend theme with all colors from tokens.ts (bg, accent, status, severity, text)
- Extend fontFamily with mono, sans, display
- Add custom boxShadow for 'glow'
- Add Google Fonts in index.html: IBM Plex Sans + JetBrains Mono

git add frontend/tailwind.config.ts frontend/index.html
git commit -m "feat(frontend): configure Tailwind with KubeVision design tokens"
git push origin main

─── STEP 3: Design system tokens ────────────────────────────────────────────────
Create frontend/src/design-system/tokens.ts (full file from project doc).
Create frontend/src/design-system/index.ts that re-exports everything.

git add frontend/src/design-system/
git commit -m "feat(design-system): add color, typography, spacing, and shadow tokens"
git push origin main

─── STEP 4: Global CSS ──────────────────────────────────────────────────────────
Create frontend/src/index.css with:
- @tailwind base/components/utilities
- CSS custom properties for all token colors (--color-bg-base, etc.)
- Global dark background: html { background: #0a0c10; }
- Scrollbar styling (thin, dark)
- Selection color: accent green
- Smooth transitions: *, *::before, *::after { transition: colors 150ms ease }
- Custom .font-mono, .font-display classes

git add frontend/src/index.css
git commit -m "feat(frontend): add global CSS with custom properties and dark theme"
git push origin main

─── STEP 5: App shell + Router ──────────────────────────────────────────────────
Create frontend/src/router.tsx with React Router v6 routes for all pages.
Create frontend/src/App.tsx with:
- QueryClientProvider (TanStack Query)
- RouterProvider
- Toaster (react-hot-toast, dark theme)
- Global error boundary

Create frontend/src/main.tsx — entry point, render App.

git add frontend/src/router.tsx frontend/src/App.tsx frontend/src/main.tsx
git commit -m "feat(frontend): add app shell with router and global providers"
git push origin main

─── STEP 6: Stores ──────────────────────────────────────────────────────────────
Create frontend/src/stores/cluster.store.ts (Zustand):
  State: activeClusterId, clusters[], setActiveCluster, setClusters
  Persist activeClusterId to localStorage

Create frontend/src/stores/auth.store.ts (Zustand):
  State: user, token, isAuthenticated, login(user,token), logout
  Persist token to localStorage (not user object)

Create frontend/src/stores/alerts.store.ts (Zustand):
  State: firingAlerts[], addAlert, removeAlert, clearAlerts

git add frontend/src/stores/
git commit -m "feat(stores): add cluster, auth, and alerts Zustand stores"
git push origin main

─── STEP 7: API client ──────────────────────────────────────────────────────────
Create frontend/src/api/client.ts:
  - Axios instance with baseURL from VITE_API_URL env
  - Request interceptor: attach Authorization Bearer token from auth store
  - Response interceptor: on 401, clear auth store and redirect to /login
  - On network error, show toast

Create frontend/src/api/clusters.api.ts — all cluster API calls
Create frontend/src/api/pods.api.ts — all pod API calls  
Create frontend/src/api/metrics.api.ts — metrics + PromQL proxy
Create frontend/src/api/helm.api.ts — Helm releases + repos
Create frontend/src/api/ai.api.ts — AI analyze endpoint

git add frontend/src/api/
git commit -m "feat(api): add Axios client and all API modules with TypeScript types"
git push origin main

─── STEP 8: Shared hooks ────────────────────────────────────────────────────────
Create frontend/src/hooks/useWebSocket.ts (full implementation from project doc with reconnect)
Create frontend/src/hooks/useMetrics.ts — wraps useWebSocket for metric streams
Create frontend/src/hooks/useK8sEvents.ts — wraps useWebSocket for k8s events
Create frontend/src/hooks/useAlerts.ts — poll /alerts/firing every 30s, update alerts store

git add frontend/src/hooks/
git commit -m "feat(hooks): add WebSocket, metrics, events, and alerts hooks"
git push origin main

─── STEP 9: Utility functions ───────────────────────────────────────────────────
Create frontend/src/utils/:
  formatBytes.ts      — bytes → "1.2 GB" / "450 MB"
  formatDuration.ts   — seconds → "2h 15m" / age → "3d ago"
  k8sColors.ts        — pod/node status → token color key
  yamlValidation.ts   — parse YAML, return errors with line numbers
  cn.ts               — className merge utility (clsx + tailwind-merge)

git add frontend/src/utils/
git commit -m "feat(utils): add formatting, k8s color mapping, and className utilities"
git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
cd frontend && npm run type-check 2>&1
cd .. && git log --oneline -12
```

---

## 📋 PROMPT 9 — Sidebar Layout + Core UI Components

```
@KUBEVISION_PROJECT.md

RULE: Each component file = commit + push.

Phase 9: Layout components and design system components.

─── FILE 1: Sidebar + Layout ────────────────────────────────────────────────────
Create frontend/src/components/Layout/Sidebar.tsx:
  - Fixed left sidebar, 240px wide, bg-bg-surface border-r border-bg-border
  - Top: KubeVision logo (SVG of a hexagon with K, accent green) + version badge
  - Below logo: ClusterSelector dropdown (shows all clusters, active cluster highlighted with green dot)
  - Nav items with lucide-react icons:
      / Overview (LayoutDashboard icon)
      /nodes (Server)
      /workloads (Boxes)
      /pods (Box)
      /services (Network)
      /helm/releases (Package)
      /helm/charts (BookOpen)
      /metrics (Activity)
      /logs (FileText)
      /alerts (Bell) — show red badge with count if firingAlerts.length > 0
      /ai (Sparkles)
      /settings (Settings)
  - Bottom: user avatar, name, email, logout button
  - Active route: accent green left border + bg-accent-dim background
  - Animate nav item hover with Framer Motion

Create frontend/src/components/Layout/TopBar.tsx:
  - 60px tall header, shows current page title + breadcrumb
  - Right: WebSocket status indicator (green=connected, yellow=reconnecting, red=disconnected)
  - Right: Notification bell (shows firing alerts count)
  - Right: Cluster health summary chips

Create frontend/src/components/Layout/AppLayout.tsx:
  - Sidebar (fixed) + TopBar (fixed top) + main content area (scrollable)

git add frontend/src/components/Layout/
git commit -m "feat(layout): add Sidebar with cluster selector, nav, TopBar with WS status" && git push origin main

─── FILE 2: Design system components ────────────────────────────────────────────
Create these components with full TypeScript props and dark theme styling:

frontend/src/design-system/Badge/index.tsx
  Props: variant (running|pending|failed|unknown|succeeded|warning|info), size (sm|md), children
  Styled as pill badges with status colors from tokens

frontend/src/design-system/Card/index.tsx
  Props: children, className, glow (boolean — adds box-shadow glow on hover), padding

frontend/src/design-system/Button/index.tsx
  Props: variant (primary|secondary|ghost|danger), size, loading, icon, disabled

frontend/src/design-system/Table/index.tsx
  Props: columns (id, header, accessor, cell renderer), data, loading skeleton, empty state, sortable

frontend/src/design-system/Modal/index.tsx
  Framer Motion scale+opacity animation, backdrop blur, close on Escape key

frontend/src/design-system/Terminal/index.tsx
  xterm.js wrapper: FitAddon, WebLinksAddon, dark theme matching tokens

frontend/src/design-system/YamlEditor/index.tsx
  Monaco Editor: kubernetes schema validation, dark theme, copy button, format button

frontend/src/design-system/Skeleton/index.tsx
  Animated shimmer placeholder for loading states, variants: text, card, table-row

git add frontend/src/design-system/
git commit -m "feat(design-system): add Badge, Card, Button, Table, Modal, Terminal, YamlEditor, Skeleton" && git push origin main

─── FILE 3: StatusBadge for K8s resources ───────────────────────────────────────
Create frontend/src/components/K8s/StatusBadge.tsx:
  Takes pod/node/deployment status string → maps to correct Badge variant using k8sColors util
  Includes blinking dot for Running status

Create frontend/src/components/K8s/ResourceAge.tsx:
  Takes creationTimestamp → displays "3d ago", "5h ago" etc with tooltip showing exact date

Create frontend/src/components/K8s/ContainerList.tsx:
  Shows list of containers in a pod with: name, image, status, restart count, ready status

git add frontend/src/components/K8s/
git commit -m "feat(components): add K8s-specific StatusBadge, ResourceAge, ContainerList" && git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
cd frontend && npm run type-check && npm run build
git log --oneline -8
```

---

## 📋 PROMPT 10 — All Pages (Overview, Pods, Nodes, Deployments)

```
@KUBEVISION_PROJECT.md

RULE: Each page folder = one commit + push.

Phase 10: Main dashboard pages.

─── PAGE 1: Login Page ──────────────────────────────────────────────────────────
frontend/src/pages/Login/index.tsx:
  - Full-screen dark login page
  - KubeVision logo centered
  - Email + Password fields (no form tag — use React state + onClick)
  - "Sign in" button with loading state
  - On success: store token in auth store, redirect to /
  - Framer Motion: card slides up from bottom

git add frontend/src/pages/Login/ && git commit -m "feat(pages): add Login page with JWT auth" && git push origin main

─── PAGE 2: Overview / Home ─────────────────────────────────────────────────────
frontend/src/pages/Overview/:
  index.tsx          — page container with staggered animation
  ClusterHealthCard.tsx — metric summary card (value, total, status, trend arrow)
  NodeGrid.tsx       — grid of node cards showing CPU bar, memory bar, pod count, status badge
  TopPodsByCPU.tsx   — ranked list of pods by CPU%, with sparkline bars
  RecentEvents.tsx   — scrollable list of K8s warning events with severity icons
  MetricSparklines.tsx — row of 4 LiveMetricChart components (CPU, Memory, Net In, Net Out)

Full implementation from project doc. All components use TanStack Query with 15s refetch.

git add frontend/src/pages/Overview/ && git commit -m "feat(pages): add Overview page with health cards, node grid, events" && git push origin main

─── PAGE 3: Nodes Page ──────────────────────────────────────────────────────────
frontend/src/pages/Nodes/:
  index.tsx — sortable table of all nodes
  NodeDetailDrawer.tsx — slide-in drawer with full node info:
    - Conditions list (MemoryPressure, DiskPressure, PIDPressure, Ready)
    - Capacity vs Allocatable table (CPU, Memory, Pods, Ephemeral Storage)  
    - Labels + Taints display
    - Cordon/Uncordon/Drain action buttons with confirmation dialog

git add frontend/src/pages/Nodes/ && git commit -m "feat(pages): add Nodes page with detail drawer, cordon/drain actions" && git push origin main

─── PAGE 4: Pods Page ───────────────────────────────────────────────────────────
frontend/src/pages/Pods/:
  index.tsx — filterable table: search by name, filter by namespace/status
  PodRow.tsx — expandable row showing containers
  PodDetailModal.tsx — full pod spec viewer + events
  PodLogsDrawer.tsx — SSE log streaming with:
    - Container selector dropdown
    - Auto-scroll toggle
    - Search/filter input (highlight matching text)
    - Download logs button
    - Line count counter
  PodExecModal.tsx — xterm.js terminal connected to WebSocket exec endpoint

git add frontend/src/pages/Pods/ && git commit -m "feat(pages): add Pods page with logs streaming and exec terminal" && git push origin main

─── PAGE 5: Workloads Page ──────────────────────────────────────────────────────
frontend/src/pages/Workloads/:
  index.tsx — tab bar: Deployments | StatefulSets | DaemonSets | CronJobs
  DeploymentList.tsx — table with scale slider inline, restart button, rollout status bar
  ScaleModal.tsx — modal with slider 0-50 replicas, shows current vs desired
  DeploymentDetailModal.tsx — tabs: Overview | YAML | Events | History

git add frontend/src/pages/Workloads/ && git commit -m "feat(pages): add Workloads page with deployment scale, restart, YAML view" && git push origin main

─── PAGE 6: Services Page ───────────────────────────────────────────────────────
frontend/src/pages/Services/:
  index.tsx — tabs: Services | Ingresses
  ServiceTable.tsx — type badge (ClusterIP/NodePort/LoadBalancer), port list
  IngressTable.tsx — host/path routing table, TLS status badge

git add frontend/src/pages/Services/ && git commit -m "feat(pages): add Services and Ingresses page" && git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
cd frontend && npm run type-check && npm run build
git log --oneline -8
```

---

## 📋 PROMPT 11 — Helm, Metrics, Alerts, AI Pages

```
@KUBEVISION_PROJECT.md

RULE: Each page = commit + push.

Phase 11: Advanced feature pages.

─── PAGE 1: Helm Releases ───────────────────────────────────────────────────────
frontend/src/pages/HelmReleases/:
  index.tsx — release table: name, chart, version, namespace, status badge, last deployed, actions
  ReleaseDetailModal.tsx — 4 tabs:
    Overview: current values in Monaco viewer, notes panel
    Upgrade: Monaco YAML editor for new values, diff preview panel, Upgrade button
    History: revision table with rollback button per row
    Resources: list of K8s resources created by this release
  InstallChartModal.tsx — search chart, select version, namespace input, Monaco values editor, Install button

git add frontend/src/pages/HelmReleases/ && git commit -m "feat(pages): add Helm Releases page with upgrade, rollback, and install" && git push origin main

─── PAGE 2: Chart Browser ───────────────────────────────────────────────────────
frontend/src/pages/HelmRepo/:
  index.tsx — search bar + repo filter tabs + chart grid
  ChartCard.tsx — card with: icon (first letter avatar), name, description, latest version, install button
  ChartDetailModal.tsx — README renderer, version selector, default values Monaco viewer, Install button

git add frontend/src/pages/HelmRepo/ && git commit -m "feat(pages): add Helm Chart Browser with install flow" && git push origin main

─── PAGE 3: Metrics Dashboard ───────────────────────────────────────────────────
frontend/src/pages/Metrics/:
  index.tsx — 2-column grid of live charts
  LiveMetricChart.tsx — full implementation from project doc with Recharts AreaChart
  PromQLExplorer.tsx — text input for raw PromQL, time range picker, submit → Recharts LineChart
  GrafanaEmbed.tsx — iframe embed of Grafana dashboard with configured panels, auth via token

git add frontend/src/pages/Metrics/ && git commit -m "feat(pages): add Metrics page with live charts and PromQL explorer" && git push origin main

─── PAGE 4: Log Explorer (Loki) ─────────────────────────────────────────────────
frontend/src/pages/Logs/:
  index.tsx — Loki log query interface
  LogSearchBar.tsx — LogQL input, namespace filter, pod filter, time range picker
  LogStream.tsx — virtualized log list (react-window) with: timestamp, log level badge, message, expandable JSON
  LogLevelFilter.tsx — toggle buttons: ERROR | WARN | INFO | DEBUG

git add frontend/src/pages/Logs/ && git commit -m "feat(pages): add Loki log explorer with LogQL and log level filtering" && git push origin main

─── PAGE 5: Alerts Page ─────────────────────────────────────────────────────────
frontend/src/pages/Alerts/:
  index.tsx — tabs: Firing Now | All Rules
  FiringAlertsList.tsx — cards per alert: severity icon, name, description, value vs threshold, fired since, silence button
  AlertRuleTable.tsx — all rules with: name, query, threshold, severity, channels, enabled toggle
  AlertRuleFormModal.tsx — create/edit form:
    Name, Description, PromQL query input (with validate button), threshold, comparator dropdown,
    duration, severity selector, channel multi-select (slack/pagerduty/email), cluster selector

git add frontend/src/pages/Alerts/ && git commit -m "feat(pages): add Alerts page with firing view and rule CRUD form" && git push origin main

─── PAGE 6: AI Log Analyzer ─────────────────────────────────────────────────────
frontend/src/pages/AIAnalyzer/:
  index.tsx — split panel layout:
    Left: pod selector (cluster → namespace → pod → container) with "Load Logs" button
          OR large textarea for manual paste
    Right: analysis result panel
  AnalysisResult.tsx — display:
    Severity badge (critical/high/medium/low) with color
    Root Cause section (bold, prominent)
    Is Known Issue badge
    Remediation Steps: numbered list with each step in a code block + copy button
    "Analyze Again" button

git add frontend/src/pages/AIAnalyzer/ && git commit -m "feat(pages): add AI Log Analyzer with pod selector and analysis result display" && git push origin main

─── PAGE 7: Settings Page ───────────────────────────────────────────────────────
frontend/src/pages/Settings/:
  index.tsx — tabs: Clusters | Users & Teams | Integrations | Profile
  ClusterSettings.tsx — add cluster (upload kubeconfig), list clusters, delete cluster, test connection
  UserManagement.tsx — user table, invite user modal, role assignment
  IntegrationsSettings.tsx — Slack webhook input, PagerDuty key, test notification button
  ProfileSettings.tsx — name, email, change password form

git add frontend/src/pages/Settings/ && git commit -m "feat(pages): add Settings page with cluster, user, integrations, and profile tabs" && git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
cd frontend && npm run type-check && npm run build
git log --oneline -10
```

---

## 📋 PROMPT 12 — Docker, Monitoring Stack, Helm Chart

```
@KUBEVISION_PROJECT.md

RULE: Every file = commit + push.

Phase 12: Containerization, monitoring config, and Helm chart.

─── STEP 1: Dockerfile ──────────────────────────────────────────────────────────
Create the multi-stage Dockerfile from the project doc.
Add to main.go: use go:embed for the frontend/dist directory so it's served from the binary.

git add Dockerfile cmd/server/main.go internal/api/router.go
git commit -m "build(docker): add multi-stage Dockerfile with embedded frontend"
git push origin main

─── STEP 2: Dockerfile.dev ──────────────────────────────────────────────────────
Create Dockerfile.dev:
  - FROM golang:1.22-alpine
  - Install air (go install github.com/cosmtrek/air@latest) for hot reload
  - COPY . .
  - CMD ["air", "-c", ".air.toml"]

Create .air.toml for Go hot reload config.

git add Dockerfile.dev .air.toml
git commit -m "build(docker): add development Dockerfile with Air hot reload"
git push origin main

─── STEP 3: Docker Compose ──────────────────────────────────────────────────────
Create deploy/docker-compose.yml (full file from project doc).
Create deploy/docker-compose.dev.yml (overrides for dev: use Dockerfile.dev, mount source as volume).

git add deploy/docker-compose.yml deploy/docker-compose.dev.yml
git commit -m "build(compose): add full Docker Compose stack for local development"
git push origin main

─── STEP 4: Prometheus config ───────────────────────────────────────────────────
Create monitoring/prometheus/prometheus.yml (full from project doc).
Create monitoring/prometheus/rules/kubernetes.yml (full from project doc).
Create monitoring/prometheus/rules/kubevision.yml:
  - KubeVisionHighErrorRate: rate of 5xx responses > 5%
  - KubeVisionWebSocketConnectionsDrop: active WS connections drops by >50% in 5 min
  - KubeVisionAPILatencyHigh: p99 latency > 2s

git add monitoring/prometheus/
git commit -m "feat(monitoring): add Prometheus config and alerting rules"
git push origin main

─── STEP 5: AlertManager config ─────────────────────────────────────────────────
Create monitoring/alertmanager/alertmanager.yml (full from project doc).

git add monitoring/alertmanager/
git commit -m "feat(monitoring): add AlertManager routing config with Slack and PagerDuty"
git push origin main

─── STEP 6: Grafana provisioning ────────────────────────────────────────────────
Create monitoring/grafana/datasources/prometheus.yml — Prometheus datasource.
Create monitoring/grafana/datasources/loki.yml — Loki datasource.
Create monitoring/grafana/dashboards/cluster-overview.json — Grafana dashboard JSON with panels:
  - Cluster CPU usage gauge
  - Cluster memory usage gauge
  - Pods running/failed stat panels
  - Node CPU heatmap
  - Top 10 pods by memory table

git add monitoring/grafana/
git commit -m "feat(monitoring): add Grafana datasource provisioning and cluster dashboard"
git push origin main

─── STEP 7: Loki config ─────────────────────────────────────────────────────────
Create monitoring/loki/loki-config.yml:
  - Single-binary mode for local dev
  - Filesystem storage at /loki
  - Retention: 30 days
  - Ingester replication factor: 1

git add monitoring/loki/
git commit -m "feat(monitoring): add Loki log aggregation config"
git push origin main

─── STEP 8: Helm chart templates ────────────────────────────────────────────────
Create all Helm chart templates in charts/kubevision/templates/:
  _helpers.tpl         — fullname, labels, selectorLabels, serviceAccountName helpers
  deployment.yaml      — main app deployment with all env vars from ConfigMap + Secret
  service.yaml         — ClusterIP service
  ingress.yaml         — conditional ingress
  configmap.yaml       — app config (non-sensitive values)
  secret.yaml          — JWT secret, DB URL, Redis URL (base64)
  serviceaccount.yaml  — service account
  clusterrole.yaml     — full RBAC from project doc
  clusterrolebinding.yaml
  hpa.yaml             — HPA with CPU and memory targets
  pdb.yaml             — PodDisruptionBudget minAvailable: 1
  networkpolicy.yaml   — allow ingress from ingress-nginx namespace only
  NOTES.txt            — post-install instructions

Create charts/kubevision/values.yaml (full from project doc).
Create charts/kubevision/values.schema.json (JSON Schema validating all values).
Create charts/kubevision/Chart.yaml (from project doc).

git add charts/kubevision/
git commit -m "feat(helm): add complete KubeVision Helm chart with all templates"
git push origin main

─── STEP 9: Validate Helm chart ──────────────────────────────────────────────────
In terminal:
  helm lint charts/kubevision
  helm template kubevision charts/kubevision --dry-run 2>&1 | head -100

If errors: fix → git add -A && git commit -m "fix(helm): resolve helm lint errors" && git push origin main

git log --oneline -8
```

---

## 📋 PROMPT 13 — CI/CD + Scripts

```
@KUBEVISION_PROJECT.md

RULE: Every file = commit + push.

Phase 13: GitHub Actions pipelines and utility scripts.

─── FILE 1: .github/workflows/ci.yml ────────────────────────────────────────────
Full CI workflow from project doc. Add these extras:
  - golangci-lint step using golangci-lint-action@v4
  - Frontend lint: npm run lint
  - Frontend type-check: npm run type-check
  - Helm chart lint + dry-run against kind cluster using helm/kind-action
  - Upload test coverage to Codecov

git add .github/workflows/ci.yml
git commit -m "ci: add CI workflow with Go tests, frontend checks, Helm lint, security scan"
git push origin main

─── FILE 2: .github/workflows/release.yml ───────────────────────────────────────
Full release workflow from project doc. Add:
  - Cosign image signing after push
  - Generate SBOM with syft
  - Attach SBOM to GitHub release

git add .github/workflows/release.yml
git commit -m "ci: add release workflow with Docker build, Helm package, and SBOM"
git push origin main

─── FILE 3: .github/workflows/helm-release.yml ──────────────────────────────────
Create Helm chart release workflow:
  - Trigger: push to main when charts/** changes
  - Use helm/chart-releaser-action to package + publish to GitHub Pages
  - Update charts/index.yaml and push to gh-pages branch

git add .github/workflows/helm-release.yml
git commit -m "ci: add Helm chart release workflow with GitHub Pages publishing"
git push origin main

─── FILE 4: .github/workflows/pr-checks.yml ─────────────────────────────────────
Create PR-specific workflow:
  - Block merge if coverage drops below 60%
  - Add PR comment with: test results, coverage report, Helm diff (if charts changed)
  - Label PRs automatically based on changed files

git add .github/workflows/pr-checks.yml
git commit -m "ci: add PR checks workflow with coverage gate and auto-labeling"
git push origin main

─── FILE 5: scripts/setup-local.sh ──────────────────────────────────────────────
Full script from project doc with the ASCII art. Make it executable.
chmod +x scripts/setup-local.sh

git add scripts/setup-local.sh
git commit -m "chore(scripts): add one-command local setup script"
git push origin main

─── FILE 6: scripts/package-charts.sh ───────────────────────────────────────────
Full script from project doc. Make executable.
chmod +x scripts/package-charts.sh

git add scripts/package-charts.sh
git commit -m "chore(scripts): add Helm chart packaging script"
git push origin main

─── FILE 7: scripts/seed-demo-data.sh ───────────────────────────────────────────
Create a seed script that:
  - Waits for the API to be healthy (polls /healthz)
  - Creates a demo admin user: admin@kubevision.local / admin
  - Registers the local Kind cluster (reads ~/.kube/config)
  - Creates 3 demo alert rules (PodCrashLooping, NodeHighCPU, NodeHighMemory)
  - Creates a demo team "platform-team" with the admin user

chmod +x scripts/seed-demo-data.sh

git add scripts/seed-demo-data.sh
git commit -m "chore(scripts): add demo data seed script"
git push origin main

─── FILE 8: scripts/generate-certs.sh ───────────────────────────────────────────
Create script that generates self-signed TLS certs for local dev using openssl.
Output: certs/tls.crt and certs/tls.key
Add certs/ to .gitignore.

git add scripts/generate-certs.sh
git commit -m "chore(scripts): add self-signed TLS cert generation script"
git push origin main

─── FILE 9: deploy/kind-cluster.yaml ────────────────────────────────────────────
Full Kind cluster config from project doc.

git add deploy/kind-cluster.yaml
git commit -m "chore(deploy): add Kind cluster config with 1 control-plane + 2 workers"
git push origin main

git log --oneline -12
```

---

## 📋 PROMPT 14 — Tests

```
@KUBEVISION_PROJECT.md

RULE: Every test file = commit + push.

Phase 14: Test suite.

─── FILE 1: Go unit tests ───────────────────────────────────────────────────────
Create these test files:

internal/auth/jwt_test.go:
  - TestIssueToken: valid token, expiry set correctly
  - TestValidateToken: valid token → correct claims, expired token → error, tampered token → error
  - TestRefreshToken: valid refresh, expired token refresh

internal/metrics/collector_test.go:
  - Mock Prometheus HTTP server
  - TestFetchMetricCPU, TestFetchMetricMemory
  - TestGetTopPodsByCPU

internal/alerts/engine_test.go:
  - TestCompare: all comparators (gt, gte, lt, lte, eq)
  - TestEvaluateRule_FiringToNotifying: mock Prometheus, verify notifier called
  - TestEvaluateRule_ResolutionNotification

internal/helm/release_test.go:
  - Mock Helm action client
  - TestListReleases, TestInstall, TestUpgrade, TestRollback

git add internal/**/*_test.go
git commit -m "test(unit): add unit tests for auth, metrics, alerts, and helm packages"
git push origin main

─── FILE 2: API integration tests ───────────────────────────────────────────────
Create internal/api/handlers/health_test.go:
  - TestHealthzAlwaysOK
  - TestReadyzDBDown: mock DB failure → 503
  - TestReadyzRedisDown: mock Redis failure → 503

Create internal/api/handlers/auth_test.go:
  - TestLoginSuccess, TestLoginWrongPassword, TestLoginUserNotFound
  - TestRefreshValidToken, TestRefreshExpiredToken
  - TestLogout

Build a testutil package: internal/testutil/
  - NewTestServer(t) → creates Gin test server with mock dependencies
  - NewTestDB(t) → spins up a real PostgreSQL via testcontainers
  - MakeRequest(t, method, path, body) → returns *httptest.ResponseRecorder

git add internal/api/handlers/*_test.go internal/testutil/
git commit -m "test(integration): add API handler tests with testutil helpers"
git push origin main

─── FILE 3: Frontend tests ──────────────────────────────────────────────────────
Configure frontend/vite.config.ts with Vitest.
Create frontend/src/utils/formatBytes.test.ts — unit tests for all formatter functions.
Create frontend/src/hooks/useWebSocket.test.ts — mock WebSocket, test reconnect logic.
Create frontend/src/design-system/Badge/Badge.test.tsx — render tests for all variants.

git add frontend/src/
git commit -m "test(frontend): add Vitest unit tests for utils, hooks, and components"
git push origin main

─── VERIFY ─────────────────────────────────────────────────────────────────────
go test ./... -v 2>&1 | tail -30
cd frontend && npm run test 2>&1 | tail -20
git log --oneline -6
```

---

## 📋 PROMPT 15 — Final README + Polish + Tag Release

```
@KUBEVISION_PROJECT.md

RULE: Each file = commit + push.

Phase 15: Documentation, polish, and first release.

─── FILE 1: README.md ────────────────────────────────────────────────────────────
Create a world-class README.md:
  - Hero section: project name, badges (build status, coverage, Docker image, Helm version, Go version, license)
  - Animated GIF placeholder comment: <!-- Add dashboard screenshot here -->
  - Feature highlights with icons
  - Quick start (3 commands: clone, make setup-local, open browser)
  - Full install via Helm section
  - Environment variables table (from project doc)
  - Architecture diagram (ASCII art of the system components)
  - Contributing guide link
  - License badge

git add README.md && git commit -m "docs: add comprehensive README with quick start and architecture" && git push origin main

─── FILE 2: CONTRIBUTING.md ─────────────────────────────────────────────────────
Create CONTRIBUTING.md:
  - Local dev setup steps
  - Commit message format (Conventional Commits)
  - PR process
  - Code style guide (Go: gofmt + golangci-lint, TS: ESLint + Prettier)
  - How to add a new API handler (step by step)
  - How to add a new frontend page (step by step)
  - How to add a new Helm chart

git add CONTRIBUTING.md && git commit -m "docs: add contributing guide" && git push origin main

─── FILE 3: CHANGELOG.md ────────────────────────────────────────────────────────
Create CHANGELOG.md with v0.1.0 entry listing all implemented features.

git add CHANGELOG.md && git commit -m "docs: add CHANGELOG for v0.1.0" && git push origin main

─── FILE 4: LICENSE ─────────────────────────────────────────────────────────────
Create Apache 2.0 LICENSE file.

git add LICENSE && git commit -m "chore: add Apache 2.0 license" && git push origin main

─── FILE 5: .golangci.yml ───────────────────────────────────────────────────────
Create golangci-lint config with enabled linters:
  errcheck, gosimple, govet, ineffassign, staticcheck, unused,
  gofmt, goimports, revive, godot, gocritic, exhaustive

git add .golangci.yml && git commit -m "chore(lint): add golangci-lint configuration" && git push origin main

─── FILE 6: frontend ESLint + Prettier ──────────────────────────────────────────
Create frontend/.eslintrc.cjs with TypeScript + React rules.
Create frontend/.prettierrc with 2-space indent, single quotes, 100 char line width.
Add lint and format scripts to frontend/package.json.

git add frontend/.eslintrc.cjs frontend/.prettierrc frontend/package.json
git commit -m "chore(frontend): add ESLint and Prettier configuration"
git push origin main

─── FINAL: Tag the first release ────────────────────────────────────────────────
Run these commands:
  go build ./... 2>&1
  cd frontend && npm run type-check && npm run build && cd ..

  git status  # Should be clean
  git log --oneline -20  # Review all commits

  # Tag v0.1.0
  git tag -a v0.1.0 -m "Release v0.1.0 — Initial KubeVision release

Features:
- Multi-cluster Kubernetes management
- Real-time WebSocket metric streaming
- Custom Helm chart repository
- Prometheus + Grafana + AlertManager + Loki stack
- Pod exec terminal, log streaming
- RBAC + audit logging
- AI log analysis via Claude API
- Full CI/CD pipeline"

  git push origin v0.1.0

  # This will trigger the release GitHub Action which builds Docker image and packages Helm chart
  echo "🚀 v0.1.0 released! Check GitHub Actions for the Docker build."
  
  # Verify
  git log --oneline -5
  git tag -l
```

---

## 🔧 QUICK FIX PROMPT — Use when something breaks

```
Something broke. Here is the error:

[PASTE ERROR HERE]

Before fixing:
1. Show me which file caused the issue
2. Show me the current state of that file (or relevant section)
3. Explain what the fix is and why

After fixing:
git add <fixed files>
git commit -m "fix(<scope>): <description of what was broken and how it was fixed>"
git push origin main
git log --oneline -3
```

---

## 📊 GIT LOG VERIFICATION — Run anytime to see progress

```
Paste this into Cursor terminal to see full project status:

echo "=== Git Log ===" && git log --oneline
echo ""
echo "=== Remote Status ===" && git remote -v && git fetch && git status
echo ""
echo "=== File Count ===" && find . -type f -not -path './.git/*' -not -path './node_modules/*' -not -path './vendor/*' | wc -l
echo ""
echo "=== Go Build ===" && go build ./... && echo "✅ Go build OK"
echo ""
echo "=== Go Tests ===" && go test ./... -count=1 2>&1 | tail -5
```

---

*Every commit tells the story. Every push is a checkpoint. Ship it.*
