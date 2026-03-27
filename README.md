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
