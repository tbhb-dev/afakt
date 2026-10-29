#!/usr/bin/env bash
# Prove prose lint rejects findings in Markdown and Go comments without changing the checkout.
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

git -C "$root" ls-files -z --cached --others --exclude-standard |
  tar -C "$root" --null -T - -cf - | tar -C "$work" -xf -
# Include the ignored style packages that bootstrap downloads.
cp -R "$root/.vale/." "$work/.vale/"
git -C "$work" init -q
cd "$work"
mise trust

# A clean baseline prevents an unrelated finding from proving the gate works.
mise run lint-prose
for file in prose-gate.md docs/prose-gate/nested.md .claude/prose-gate/nested.md \
  .agents/prose-gate/nested.md apps/prose-gate/nested.md .github/prose-gate/nested.md \
  prose-gate.go cmd/prose-gate/nested.go internal/prose-gate/nested.go; do
  mkdir -p "$(dirname "$file")"
  case "$file" in
  *.go)
    printf 'package prose\n\nvar message = "Send e-mail."\n' >"$file"
    # A string literal passes before the same wording in a comment fails.
    mise run lint-prose
    printf '\n// Send e-mail.\n' >>"$file"
    ;;
  *) printf 'Send e-mail.\n' >"$file" ;;
  esac
  if mise run lint-prose >"$work/failure.log" 2>&1; then
    echo "prose gate accepted a planted finding in $file" >&2
    exit 1
  fi
  if ! grep -qF "FILE: $file" "$work/failure.log" ||
    ! grep -qF 'project.WordList match="e-mail"' "$work/failure.log"; then
    cat "$work/failure.log" >&2
    echo "prose gate failed without the planted finding in $file" >&2
    exit 1
  fi
  rm "$file"
  echo "prose gate rejects a planted finding in $file"
done
