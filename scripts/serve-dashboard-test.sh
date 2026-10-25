#!/usr/bin/env bash
# Serve a copy of the binary alone, with an empty working directory.
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
./afakt server --target=api,dashboard --listen=127.0.0.1:4183 --shutdown-timeout=5s &
child=$!
wait "$child"
