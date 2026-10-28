#!/usr/bin/env bash
# Agent hook run after each tool call: render the diagrams whose sources
# changed. Failures exit 2 so the agent sees the error and can fix the source.
# A file the tool wrote in another repository with this hook, such as a
# scratch clone under tmp/, gets that repository's hook too.
set -uo pipefail

payload=$(cat)
root=$(git rev-parse --show-toplevel) || exit 0

refresh() {
  local log
  if ! log=$(cd "$1" && mise run --quiet refresh-diagrams 2>&1); then
    printf 'diagram render failed in %s:\n%s\n' "$1" "$log" >&2
    exit 2
  fi
}

refresh "$root"
file=$(jq -r '.tool_input.file_path // empty' <<<"$payload" 2>/dev/null)
if [[ -n $file && -d $(dirname "$file") ]]; then
  other=$(cd "$(dirname "$file")" && git rev-parse --show-toplevel 2>/dev/null) || other=
  if [[ -n $other && $other != "$root" && -f $other/scripts/diagrams-hook.sh ]]; then
    refresh "$other"
  fi
fi
