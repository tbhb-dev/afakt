#!/usr/bin/env bash
# Refuse a false Go-only proof before the ordinary gates run, without excluding any test.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
for tool in node pnpm npm bun; do
  if command -v "$tool" >/dev/null; then
    printf 'Go-only proof requires an environment without %s\n' "$tool" >&2
    exit 1
  fi
done
if [[ -e apps/dashboard/dist || -e internal/dashboard/assets || -e node_modules ]]; then
  echo 'Go-only proof requires a clean tree without web dependencies or output' >&2
  exit 1
fi
