#!/usr/bin/env bash
set -euo pipefail
docker rm -f kubevision-pg kubevision-redis 2>/dev/null || true
echo "Stopped and removed kubevision-pg and kubevision-redis (if they existed)."
