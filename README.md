# KubeVision

Kubernetes cluster management dashboard: real-time metrics, Helm, and a Go API (see [KUBEVISION_PROJECT.md](./KUBEVISION_PROJECT.md) for the full roadmap).

## Prerequisites

- **Go 1.22+**
- **Docker** (for local Postgres and Redis)

## Run the API locally

1. Start dependencies (Postgres + Redis):

   ```bash
   make deps-up
   ```

   If you use Docker Compose v2 instead, you can run:  
   `docker compose -f deploy/docker-compose.deps.yml up -d`

2. Copy env and adjust if needed:

   ```bash
   cp .env.example .env
   ```

3. Run the server:

   ```bash
   make run-dev
   ```

4. Check health:

   ```bash
   curl -s http://127.0.0.1:8080/healthz
   curl -s http://127.0.0.1:8080/readyz
   curl -s http://127.0.0.1:8080/api/v1/clusters
   ```

### Phase 2 — Users, teams, permissions (API)

Identity is interim: send `Authorization: Bearer <user-uuid>` (the row id from `users`). The first `POST /api/v1/users` call creates the bootstrap admin **without** a bearer token; further user creation requires an admin bearer.

Examples:

```bash
# Bootstrap admin (empty DB only)
curl -s -X POST http://127.0.0.1:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@local","name":"Admin","password":"changeme12","isAdmin":true}'

# Authenticated call (replace USER_ID)
curl -s http://127.0.0.1:8080/api/v1/teams \
  -H "Authorization: Bearer USER_ID"
```

RBAC evaluation: `GET /api/v1/permissions/check?cluster_id=...&namespace=...&permission=read|write|admin`. Example enforced route: `GET /api/v1/rbac/probe/:cluster_id/:namespace` (requires read).

Stop containers: `make deps-down`

## Build

```bash
make build
./bin/kubevision
```

## Module path

Go module: `github.com/AbhishekSharmaIE/Kubevision`

## Cursor workflow

See [CURSOR_GIT_WORKFLOW.md](./CURSOR_GIT_WORKFLOW.md) for phased prompts and commit conventions.
