#!/usr/bin/env bash
# Install the committed workspace in a temporary clone so generated files
# cannot change the source checkout or reuse its installed packages.
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

git clone --quiet "$root" "$work/repo"
cd "$work/repo"
mise trust
# Exercise a checkout-local store even when the caller configures a store
# outside the repository. All installed files stay in the temporary clone.
pnpm_config_store_dir="$work/repo/.pnpm-store" mise run install-web

changes=$(git status --porcelain)
if [[ -n $changes ]]; then
  printf 'FAIL install-web left the working tree dirty:\n%s\n' "$changes" >&2
  exit 1
fi
printf 'install-web leaves the temporary clone clean\n'
