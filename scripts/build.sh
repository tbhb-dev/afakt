#!/usr/bin/env bash
# Remove the previous release before any step can fail, then embed a fresh web build.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
rm -f bin/afakt
rm -rf internal/dashboard/assets
GOMAXPROCS=2 GOFLAGS=-p=2 mise run build-web
cp -R apps/dashboard/dist internal/dashboard/assets
GOMAXPROCS=2 GOFLAGS='-p=2 -tags=dashboard_production' mise run devtools:build
