#!/usr/bin/env bash
# Check the exact goal command without adjacent assets or external services.
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
mkdir "$work/bin"
cp "$root/bin/afakt" "$work/bin/afakt"
cd "$work"

# Capture the same binary's dashboard-only response for byte-for-byte comparison.
bin/afakt server --target=dashboard --listen=127.0.0.1:4187 >dashboard.log 2>&1 &
child=$!
curl --fail --silent --show-error --retry 30 --retry-connrefused --retry-delay 1 \
  http://127.0.0.1:4187/dashboard/ -o dashboard.html
kill -TERM "$child"
wait "$child"
child=

bin/afakt server --target=all --temporal-mode=embedded >server.log 2>&1 &
child=$!

# Start observing before readiness, rather than letting a browser server waiter hide it.
starting=false
end=$((SECONDS + 120))
while ((SECONDS < end)); do
  status=$(curl --silent --max-time 2 -o ready.json -w '%{http_code}' http://127.0.0.1:8080/readyz) || continue
  if [[ $status == 503 ]]; then
    curl --fail --silent --show-error http://127.0.0.1:8080/livez -o live.json
    starting=true
  elif [[ $status == 200 ]]; then
    break
  else
    echo "unexpected readiness status: $status" >&2
    exit 1
  fi
done
if [[ $starting != true || $status != 200 ]]; then
  cat server.log >&2
  echo 'expected readiness 503 while starting, then 200' >&2
  exit 1
fi
echo 'starting: /livez 200, /readyz 503; running: /readyz 200'
curl --fail --silent --show-error http://127.0.0.1:8080/dashboard/ -o combined.html
cmp dashboard.html combined.html
echo 'dashboard: HTTP 200 and identical to dashboard-only response'
cd "$root"
GOMAXPROCS=2 GOFLAGS=-p=2 mise run e2e-all-targets
cd "$work"

# Temporal stops before HTTP, leaving the health endpoints available during drain.
kill -TERM "$child"
draining=false
end=$((SECONDS + 30))
while ((SECONDS < end)); do
  status=$(curl --silent --max-time 2 -o ready.json -w '%{http_code}' http://127.0.0.1:8080/readyz) || break
  if [[ $status == 503 ]]; then
    curl --fail --silent --show-error http://127.0.0.1:8080/livez -o live.json
    draining=true
    break
  fi
done
if [[ $draining != true ]]; then
  cat server.log >&2
  echo 'expected /livez 200 and /readyz 503 while draining' >&2
  exit 1
fi
wait "$child"
child=
echo 'draining: /livez 200, /readyz 503; clean termination'
