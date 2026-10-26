#!/usr/bin/env bash
# Interim exact-match waivers until tbhb-dev/devtools#732 ships.
set -euo pipefail

format=${1:-text}
case $format in
text) ;;
sarif) [[ $# == 2 ]] || exit 2 ;;
*)
  echo "gomodscan: expected text or sarif" >&2
  exit 2
  ;;
esac

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
report=$tmp/report.sarif
if [[ $format == sarif ]]; then report=$2; fi

# Reuse every upstream completeness and SARIF check. A failed scan never
# reaches waiver matching, even when its partial output contains accepted hits.
bash .devtools/scripts/gomodscan.sh sarif "$report"
jq --slurp --raw-output --exit-status \
  --slurpfile accepted .gomodscan-accepted.json \
  --from-file scripts/gomodscan-findings.jq "$report"
