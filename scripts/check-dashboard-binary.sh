#!/usr/bin/env bash
# Compare a binary running alone with the production files built in this invocation.
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
cp "$root/bin/afakt" "$work/afakt"
mkdir "$work/responses"
find "$root/apps/dashboard/dist" -type f -print0 >"$work/files"
for target in dashboard api,dashboard; do
  # The executable is the only file in its working directory.
  mkdir "$work/empty"
  cp "$work/afakt" "$work/empty/afakt"
  (cd "$work/empty" && exec ./afakt server --target="$target" --listen=127.0.0.1:4185 --shutdown-timeout=5s) >"$work/server.log" 2>&1 &
  child=$!
  curl --fail --silent --retry 20 --retry-connrefused --retry-delay 1 http://127.0.0.1:4185/readyz >/dev/null
  while IFS= read -r -d '' file; do
    name=${file#"$root/apps/dashboard/dist/"}
    curl --fail --silent -D "$work/responses/headers" "http://127.0.0.1:4185/dashboard/$name" -o "$work/responses/body"
    cmp "$file" "$work/responses/body"
  done <"$work/files"
  kill -TERM "$child"
  wait "$child"
  child=
  rm "$work/empty/afakt"
  rmdir "$work/empty"
done
echo 'both target configurations serve every production file byte for byte from the binary alone'
