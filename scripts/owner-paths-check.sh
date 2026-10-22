#!/usr/bin/env bash
# owner-paths-check — keep .github/CODEOWNERS inside owner_paths.
#
# Fails when a CODEOWNERS pattern matches no tracked file, or when a
# tracked file that CODEOWNERS assigns is not matched by owner_paths in
# .agents/agent-skills.conf. The two lists need not be equal: owner_paths
# also holds the workflows.
set -euo pipefail

export LC_ALL=C

cd "$(dirname "${BASH_SOURCE[0]}")/.."

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

status=0
owned=
# Without its leading slash, a CODEOWNERS pattern is a git path pattern, and a
# trailing slash names the whole directory.
while read -r pattern _; do
  case $pattern in '' | '#'*) continue ;; esac
  found=$(git ls-files -- "${pattern#/}")
  if [[ -z $found ]]; then
    echo "owner-paths-check: no tracked file matches $pattern" >&2
    status=1
  else
    owned+=$found$'\n'
  fi
done <.github/CODEOWNERS

regex=$(sed -n 's/^owner_paths = //p' .agents/agent-skills.conf)
if [[ -z $regex ]]; then
  echo "owner-paths-check: no owner_paths in .agents/agent-skills.conf" >&2
  exit 1
fi

# grep exits 0 on a match and 1 on none; any other status is an error.
tracked=$(git ls-files)
rc=0
held=$(grep -E -e "$regex" <<<"$tracked") || rc=$?
if ((rc > 1)); then
  echo "owner-paths-check: grep failed on owner_paths (status $rc)" >&2
  exit 1
fi

# comm -23 keeps the lines that only CODEOWNERS has.
sort -u <<<"$owned" >"$tmp/owned"
sort -u <<<"$held" >"$tmp/held"
missing=$(comm -23 "$tmp/owned" "$tmp/held")
if [[ -n $missing ]]; then
  echo "owner-paths-check: CODEOWNERS assigns files that owner_paths misses:" >&2
  printf '  %s\n' "$missing" >&2
  status=1
fi

exit "$status"
