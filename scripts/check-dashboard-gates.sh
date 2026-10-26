#!/usr/bin/env bash
# Generate each broken input in a disposable copy, keeping every fixture out of Git.
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
if [[ $# == 0 ]]; then
  for gate in results build revalidation console axe browser all-targets; do
    bash "$root/scripts/check-dashboard-gates.sh" "$gate"
  done
  exit 0
fi
gate=$1
if [[ $gate == results ]]; then
  bash "$root/scripts/check-ci-results.sh"
  exit 0
fi
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
cd "$root"
git ls-files -z --cached --others --exclude-standard | tar --null -T - -cf - | tar -C "$work" -xf -
git -C "$work" init -q
for dir in . apps/dashboard; do
  mkdir -p "$work/$dir/node_modules"
  for item in "$root/$dir/node_modules"/* "$root/$dir/node_modules"/.[!.]*; do
    if [[ -e $item || -L $item ]]; then
      ln -s "$item" "$work/$dir/node_modules/"
    fi
  done
done
cd "$work"
export pnpm_config_verify_deps_before_run=false
GOMAXPROCS=2 GOFLAGS=-p=2 mise trust

# expect_failure requires the intended finding as well as a nonzero exit.
expect_failure() {
  local finding=$1
  shift
  if "$@" >"$work/failure.log" 2>&1; then
    echo 'a gate accepted generated broken input' >&2
    exit 1
  fi
  if ! grep -qF "$finding" "$work/failure.log"; then
    cat "$work/failure.log" >&2
    echo 'gate failed without the expected finding' >&2
    exit 1
  fi
}

if [[ $gate == all-targets ]]; then
  printf '\ntest("reject broken combined proof", () => { expect(false).toBe(true); });\n' \
    >>apps/dashboard/all/targets.spec.ts
  expect_failure 'reject broken combined proof' \
    env GOMAXPROCS=2 GOFLAGS=-p=2 mise run check-all-targets
  echo 'combined check rejects a browser assertion failure'
  exit 0
fi

if [[ $gate == build ]]; then
  GOMAXPROCS=2 GOFLAGS=-p=2 mise run build
  # Keep both the earlier binary and output, then break an input to the web build.
  printf '\nimport "./missing-production-input.ts";\n' >>apps/dashboard/src/main.ts
  expect_failure missing-production-input env GOMAXPROCS=2 GOFLAGS=-p=2 mise run build
  [[ ! -e bin/afakt ]] || {
    echo 'failed web build left a stale binary' >&2
    exit 1
  }
  [[ -f apps/dashboard/dist/index.html ]] || {
    echo 'failure proof lost the earlier output' >&2
    exit 1
  }
  echo 'failed web build removes the previous binary despite earlier output'
  exit 0
fi

# A browser revalidating an old entry must receive the changed page after replacement.
if [[ $gate == revalidation ]]; then
  GOMAXPROCS=2 GOFLAGS=-p=2 mise run build
  bin/afakt server --target=dashboard --listen=127.0.0.1:4184 --shutdown-timeout=5s >server.log 2>&1 &
  child=$!
  curl --fail --silent --retry 20 --retry-connrefused --retry-delay 1 -D before.headers \
    http://127.0.0.1:4184/dashboard/ -o before.html
  etag=
  while IFS= read -r line; do
    if [[ $line == Etag:* || $line == ETag:* || $line == etag:* ]]; then
      etag=${line#*: }
      etag=${etag%$'\r'}
    fi
  done <before.headers
  [[ -n $etag ]] || {
    echo 'entry has no validator' >&2
    exit 1
  }
  kill -TERM "$child"
  wait "$child"
  child=
  printf '\n<!-- replacement entry -->\n' >>apps/dashboard/index.html
  GOMAXPROCS=2 GOFLAGS=-p=2 mise run build
  bin/afakt server --target=dashboard --listen=127.0.0.1:4184 --shutdown-timeout=5s >server.log 2>&1 &
  child=$!
  curl --fail --silent --retry 20 --retry-connrefused --retry-delay 1 -H "If-None-Match: $etag" \
    http://127.0.0.1:4184/dashboard/ -o after.html
  grep -qF 'replacement entry' after.html
  if cmp -s before.html after.html; then
    echo 'replacement reused the old entry' >&2
    exit 1
  fi
  kill -TERM "$child"
  wait "$child"
  child=
  echo 'replacement entry is returned when revalidating the old entry'
  exit 0
fi

# All browser failures use the same production-binary task and pinned browser as CI.
case $gate in
console)
  printf 'console.error("forced console error");\n' >>apps/dashboard/src/app.tsx
  finding='forced console error'
  ;;
axe)
  cat >apps/dashboard/src/app.tsx <<'EOF'
import type { JSX } from "react";
export function App(): JSX.Element {
  return <><main><img src="data:," /></main><p>Outside the main landmark</p></>;
}
EOF
  finding=image-alt
  ;;
browser)
  # mise reapplies its environment, so change the pin in this disposable configuration.
  while IFS= read -r line; do
    if [[ $line == AFAKT_WEB_CHROMIUM\ =* ]]; then
      printf 'AFAKT_WEB_CHROMIUM = "0.0.0"\n'
    else
      printf '%s\n' "$line"
    fi
  done <mise.toml >broken.toml
  mv broken.toml mise.toml
  GOMAXPROCS=2 GOFLAGS=-p=2 mise trust
  finding='Expected: "0.0.0"'
  ;;
*)
  echo 'unknown dashboard failure proof' >&2
  exit 2
  ;;
esac
expect_failure "$finding" env GOMAXPROCS=2 GOFLAGS=-p=2 mise run check-dashboard
if [[ $gate == axe ]]; then grep -qF region "$work/failure.log"; fi
printf '%s failure proof passed\n' "$gate"
