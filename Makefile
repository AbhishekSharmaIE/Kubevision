.PHONY: dev build test lint docker-build helm-package setup-local deps-up deps-down run run-dev git-status

# Start full local stack (when compose files exist)
dev:
	docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.dev.yml up --build

build:
	go build -ldflags="-s -w -X main.Version=$(shell git describe --tags --always 2>/dev/null || echo dev)" -o bin/kubevision ./cmd/server

test:
	go test ./... -v -race -coverprofile=coverage.out

lint:
	golangci-lint run ./...

docker-build:
	docker build -t kubevision:$(shell git describe --tags --always 2>/dev/null || echo dev) .

helm-package:
	bash scripts/package-charts.sh

setup-local:
	bash scripts/setup-local.sh

deps-up:
	bash scripts/deps-up.sh

deps-down:
	bash scripts/deps-down.sh

run: build
	./bin/kubevision

run-dev:
	go run -ldflags="-X main.Version=dev" ./cmd/server

git-status:
	@git status
	@git log --oneline -5
