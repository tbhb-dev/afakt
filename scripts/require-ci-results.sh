#!/usr/bin/env bash
# Only success from every dependency permits the required test result to pass.
set -euo pipefail
[[ $# == 4 ]] || {
  echo 'expected every required dependency result' >&2
  exit 1
}
for result in "$@"; do
  if [[ $result != success ]]; then
    printf 'required dependency finished as %s\n' "$result" >&2
    exit 1
  fi
done
