#!/usr/bin/env bash
# Exercise the release binary without adjacent assets or a separate Temporal service.
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
work=$(mktemp -d)
child=
cleanup() {
  if [[ -n $child ]]; then
    kill -TERM "$child" 2>/dev/null || true
    wait "$child" || true
  fi
  rm -rf "$work"
}
trap cleanup EXIT
trap 'exit 0' TERM INT
cp "$root/bin/afakt" "$work/afakt"
cd "$work"
./afakt server --target=api,dashboard,worker --temporal-mode=embedded \
  --listen=127.0.0.1:4186 --shutdown-timeout=5s &
child=$!
wait "$child"
