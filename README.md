# KubeVision

Kubernetes cluster management dashboard: real-time metrics, Helm, and a Go API (see [KUBEVISION_PROJECT.md](./KUBEVISION_PROJECT.md) for the full roadmap).

**Engineering guide (setup, phases, curl walkthroughs):** [Documentation.md](./Documentation.md).

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

### Phase 2 & 3 — Users, teams, RBAC, JWT auth

Use **JWT access tokens** (`Authorization: Bearer <jwt>`). Tokens are stored in **Redis** until logout or refresh. Set `JWT_SECRET` (≥32 chars) in production—see `.env.example`.

```bash
# 1) Bootstrap admin when users table is empty (no token)
curl -s -X POST http://127.0.0.1:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@local","name":"Admin","password":"changeme12","isAdmin":true}'

# 2) Login → copy .data.token from JSON
curl -s -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@local","password":"changeme12"}'

# 3) Call APIs (replace TOKEN)
curl -s http://127.0.0.1:8080/api/v1/teams \
  -H "Authorization: Bearer TOKEN"
```

Refresh: `POST /api/v1/auth/refresh` with JSON `{"token":"<expired-or-current-jwt>"}` or the same `Authorization` header. Logout: `POST /api/v1/auth/logout` with a valid bearer token.

RBAC: `GET /api/v1/permissions/check?...` and `GET /api/v1/rbac/probe/:cluster_id/:namespace` (requires read on that namespace).

Details: [Documentation.md](./Documentation.md).

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
