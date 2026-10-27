#!/usr/bin/env bash
# PlantUML sources are stable across hosts; rendered fonts and pixels are not.
# Check fresh sources and the hashes saved with the images, without re-rendering.
set -euo pipefail
export LC_ALL=C
shopt -s nullglob

root=${WORKSPACE_ROOT:-docs}
dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT
structurizr export -workspace "$root/workspace.dsl" -format plantuml -output "$dir"
sources=("$dir"/structurizr-*.puml)
if [[ ${#sources[@]} == 0 ]]; then
  echo "error: no workspace views were exported" >&2
  exit 1
fi

# Match the repository's end-of-file hook before comparing or hashing exports.
perl -0777 -pi -e 's/\s*\z/\n/' "${sources[@]}"

case "${1:?expected export or check}" in
export)
  # Smetana is bundled with PlantUML and needs no system Graphviz or browser.
  plantuml -Playout=smetana -failfast2 -nometadata -tpng "${sources[@]}"
  plantuml -Playout=smetana -failfast2 -nometadata -tsvg "${sources[@]}"
  perl -0777 -pi -e 's/\s*\z/\n/' "$dir"/structurizr-*.svg
  mkdir -p "$root/diagrams"
  # Only this export's namespace belongs to this task.
  rm -f "$root"/diagrams/structurizr-*.{puml,png,svg}
  cp "$dir"/structurizr-*.{puml,png,svg} "$root/diagrams/"
  ;;
check)
  for source in "${sources[@]}"; do
    if ! diff -u "$root/diagrams/${source##*/}" "$source"; then
      echo "error: diagram sources differ; run 'mise run export-diagrams'" >&2
      exit 1
    fi
  done
  for source in "$root"/diagrams/structurizr-*.puml; do
    if [[ ! -f $dir/${source##*/} ]]; then
      echo "error: stale diagram $source; run 'mise run export-diagrams'" >&2
      exit 1
    fi
  done
  ;;
*)
  echo "error: expected export or check" >&2
  exit 1
  ;;
esac

for source in "${sources[@]}"; do
  name=${source##*/}
  for format in png svg; do
    if [[ ! -s $root/diagrams/${name%.puml}.$format ]]; then
      echo "error: missing $format image for $name; run 'mise run export-diagrams'" >&2
      exit 1
    fi
  done
done
(
  cd "$root"
  shasum -a 256 workspace.dsl diagrams/structurizr-*.{puml,png,svg}
) >"$dir/structurizr.sha256"
if [[ $1 == export ]]; then
  cp "$dir/structurizr.sha256" "$root/diagrams/"
elif ! diff -u "$root/diagrams/structurizr.sha256" "$dir/structurizr.sha256"; then
  echo "error: diagram inputs or images changed; run 'mise run export-diagrams'" >&2
  exit 1
fi
