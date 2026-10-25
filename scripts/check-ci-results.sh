#!/usr/bin/env bash
# Exercise failure, cancellation and skip for each dependency of the required result.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
bash scripts/require-ci-results.sh success success success success
for position in 0 1 2 3; do
  for result in failure cancelled skipped; do
    results=(success success success success)
    results[position]=$result
    if bash scripts/require-ci-results.sh "${results[@]}"; then
      printf 'required result passed with dependency %s %s\n' "$position" "$result" >&2
      exit 1
    fi
  done
done
echo 'every dependency rejects failure, cancellation and skip'
