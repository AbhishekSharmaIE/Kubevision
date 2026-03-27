#!/usr/bin/env bash
set -euo pipefail

if ! docker info >/dev/null 2>&1; then
  echo "Docker is not running or not accessible."
  exit 1
fi

if docker ps -a --format '{{.Names}}' | grep -qx kubevision-pg; then
  docker start kubevision-pg
else
  docker run -d --name kubevision-pg \
    -e POSTGRES_USER=kubevision \
    -e POSTGRES_PASSWORD=kubevision \
    -e POSTGRES_DB=kubevision \
    -p 5432:5432 \
    postgres:16-alpine
fi

if docker ps -a --format '{{.Names}}' | grep -qx kubevision-redis; then
  docker start kubevision-redis
else
  docker run -d --name kubevision-redis -p 6379:6379 redis:7-alpine
fi

echo "Waiting for Postgres..."
for i in {1..30}; do
  if docker exec kubevision-pg pg_isready -U kubevision -d kubevision >/dev/null 2>&1; then
    echo "Postgres is ready."
    exit 0
  fi
  sleep 1
done
echo "Postgres did not become ready in time."
exit 1
