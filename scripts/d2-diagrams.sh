#!/usr/bin/env bash
# Render each D2 diagram as x.d2.svg and x.d2.png beside its x.d2 source, and
# record the hashes of all three in x.d2.sha256. D2 output is stable across
# hosts, so check re-renders the SVG; rasterized pixels may not be, so the PNG
# is checked against its recorded hash without re-rendering.
set -euo pipefail
export LC_ALL=C

dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT

# sources lists the given D2 files, or every D2 file git doesn't ignore.
sources() {
  if [[ $# -gt 0 ]]; then
    printf '%s\0' "$@"
  else
    git ls-files -z --cached --others --exclude-standard -- '*.d2'
  fi
}

# outputs lists every rendered file git doesn't ignore.
outputs() {
  git ls-files -z --cached --others --exclude-standard -- '*.d2.svg' '*.d2.png' '*.d2.sha256'
}

# render_svg writes the optimized SVG for a source, ending in one newline.
render_svg() {
  d2 "$1" "$dir/raw.svg" 2>"$dir/d2.log" || {
    cat "$dir/d2.log" >&2
    return 1
  }
  svgo --quiet --input "$dir/raw.svg" --output "$2"
  local last
  last=$(tail -c 1 "$2")
  [[ -z $last ]] || echo >>"$2"
}

# manifest prints the hashes of a source and its renders, relative to their directory.
manifest() {
  local parent name
  parent=$(dirname "$1")
  name=$(basename "$1")
  (
    cd "$parent"
    shasum -a 256 "$name" "$name.svg" "$name.png"
  )
}

# render renders a source's SVG and PNG and records their hashes.
render() {
  render_svg "$1" "$1.svg"
  # resvg skips @font-face, so it reads the fonts D2 embeds from an installed
  # copy of the same family instead.
  sed -E \
    -e 's/font-family: "d2-[0-9]+-font-regular";/font-family: "Source Sans Pro";/' \
    -e 's/font-family: "d2-[0-9]+-font-bold";/font-family: "Source Sans Pro"; font-weight: bold;/' \
    -e 's/font-family: "d2-[0-9]+-font-italic";/font-family: "Source Sans Pro"; font-style: italic;/' \
    "$dir/raw.svg" >"$dir/png.svg"
  resvg --quiet --skip-system-fonts --use-fonts-dir "$fonts" --zoom 2 "$dir/png.svg" "$1.png"
  oxipng --quiet --opt 4 --strip safe "$1.png"
  manifest "$1" >"$1.sha256"
  echo "rendered $1"
}

# stale reports whether a source's recorded hash differs from its contents.
stale() {
  [[ -f $1.sha256 && -f $1.svg && -f $1.png ]] || return 0
  local recorded current
  recorded=$(head -n 1 "$1.sha256")
  current=$(manifest "$1")
  [[ $recorded != "${current%%$'\n'*}" ]]
}

# orphan reports whether a rendered file has lost its source.
orphan() {
  [[ ! -f ${1%.*} ]]
}

# remove_orphans deletes rendered files whose source is gone.
remove_orphans() {
  outputs >"$dir/outputs"
  while IFS= read -r -d '' output; do
    if orphan "$output"; then
      rm -f "$output"
      echo "removed $output"
    fi
  done <"$dir/outputs"
}

case "${1:?expected render, check, or refresh}" in
render | refresh)
  fonts=$(mise where http:source-sans-pro)/TTF
  command=$1
  shift
  sources "$@" >"$dir/sources"
  while IFS= read -r -d '' source; do
    if [[ $command == render ]] || stale "$source"; then
      render "$source"
    fi
  done <"$dir/sources"
  [[ $# -gt 0 ]] || remove_orphans
  ;;
check)
  status=0
  sources >"$dir/sources"
  while IFS= read -r -d '' source; do
    if [[ ! -f $source.sha256 || ! -f $source.svg || ! -f $source.png ]]; then
      echo "error: $source has no rendered diagrams; run 'mise run render-d2 $source'" >&2
      status=1
      continue
    fi
    render_svg "$source" "$dir/check.svg"
    manifest "$source" >"$dir/check.sha256"
    if ! cmp -s "$source.svg" "$dir/check.svg"; then
      echo "error: $source.svg differs from its source; run 'mise run render-d2 $source'" >&2
      status=1
    elif ! diff -u "$source.sha256" "$dir/check.sha256" >&2; then
      echo "error: $source or its renders changed; run 'mise run render-d2 $source'" >&2
      status=1
    fi
  done <"$dir/sources"
  outputs >"$dir/outputs"
  while IFS= read -r -d '' output; do
    if orphan "$output"; then
      echo "error: $output has no source; run 'mise run render-d2'" >&2
      status=1
    fi
  done <"$dir/outputs"
  exit "$status"
  ;;
*)
  echo "error: expected render, check, or refresh" >&2
  exit 1
  ;;
esac
