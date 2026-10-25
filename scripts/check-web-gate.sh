#!/usr/bin/env bash
# Usage: check-web-gate.sh GATE
#
# Show that one web gate fails on broken input. The check copies the
# workspace's tracked files into a fresh temporary directory, links each
# installed package into it, writes the broken input there, and runs the gate
# with the same command and configuration as its mise task. It requires a
# non-zero exit plus the finding that names each broken rule, and it removes
# the directory afterwards. Nothing is written into the checkout.
#
# GATE is one of: biome, types, eslint, cover, property, dup, deadcode,
# audit, mutate, e2e-console, e2e-axe, e2e-browser. The audit check needs no copy: it builds a one-package
# project of its own in the directory.
set -euo pipefail

gate=${1:?usage: check-web-gate.sh GATE}
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
export NO_COLOR=1 FORCE_COLOR=0
# pnpm exec would otherwise install into the copy before each run, and the
# copy's linked node_modules is not a real install.
export pnpm_config_verify_deps_before_run=false

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
app="$work/apps/dashboard"

# Each node_modules entry becomes a link to the installed package. The
# directories themselves stay real, so tool caches land in the copy.
link_modules() {
  local from=$1 to=$2 entry
  mkdir -p "$to"
  for entry in "$from"/* "$from"/.[!.]*; do
    if [[ -e $entry || -L $entry ]]; then
      ln -s "$entry" "$to/"
    fi
  done
}

# Copy the root files and apps/ as git sees them, untracked files included,
# then give the copy a repository of its own so the tools that read
# .gitignore behave as they do in the checkout.
copy_workspace() {
  git -C "$root" ls-files -z --cached --others --exclude-standard -- ':(glob)*' ':(glob)apps/**' |
    tar -C "$root" --null -T - -cf - | tar -C "$work" -xf -
  git -C "$work" init -q
  link_modules "$root/node_modules" "$work/node_modules"
  link_modules "$root/apps/dashboard/node_modules" "$app/node_modules"
}

if [[ $gate != audit ]]; then
  copy_workspace
fi

output=""
failures=0

# run DIR COMMAND... runs the gate in DIR of the copy and requires it to fail.
run() {
  local dir=$1 status=0
  shift
  output=$(cd "$dir" && "$@" 2>&1) || status=$?
  if ((status == 0)); then
    printf 'FAIL %s: the gate passed on broken input\n%s\n' "$gate" "$output" >&2
    failures=$((failures + 1))
  else
    printf 'ok   %s exits %d\n' "$gate" "$status"
  fi
}

# finding PATTERN requires the gate's output to match PATTERN.
finding() {
  if grep -qE -- "$1" <<<"$output"; then
    printf 'ok   %s reports %s\n' "$gate" "$1"
  else
    printf 'FAIL %s: no finding matches %s\n%s\n' "$gate" "$1" "$output" >&2
    failures=$((failures + 1))
  fi
}

case $gate in
biome)
  printf 'var answer = 42;\nexport default answer;\n' >"$app/src/lint.ts"
  printf 'export const answer   =   42\n' >"$app/src/format.ts"
  # Biome ignores NO_COLOR; plain output keeps the findings matchable.
  run "$work" biome ci --colors=off .
  finding 'src/lint\.ts.*lint/suspicious/noVar'
  finding 'src/lint\.ts.*lint/style/noDefaultExport'
  finding 'src/format\.ts format'
  ;;
types)
  src="$app/src"
  printf 'export function double(value) {\n  return value * 2;\n}\n' >"$src/implicit-any.ts"
  printf 'export function size(text: string | null): number {\n  return text.length;\n}\n' >"$src/strict-null.ts"
  printf 'const values: number[] = [];\nexport const first: number = values[0];\n' >"$src/unchecked-index.ts"
  printf 'export interface Options {\n  label?: string;\n}\nexport const options: Options = { label: undefined };\n' >"$src/exact-optional.ts"
  printf 'class Base {\n  run(): void {}\n}\nexport class Derived extends Base {\n  run(): void {}\n}\n' >"$src/implicit-override.ts"
  printf 'const env: Record<string, string> = {};\nexport const home = env.home;\n' >"$src/index-signature.ts"
  printf 'import { Options } from "./exact-optional.ts";\nexport type Settings = Options;\n' >"$src/type-import.ts"
  printf 'export enum Color {\n  Red,\n}\n' >"$src/erasable.ts"
  printf 'export function noop(): void {\n  const unused = 1;\n}\n' >"$src/unused-local.ts"
  printf 'export function sign(n: number): string | undefined {\n  if (n > 0) {\n    return "+";\n  }\n}\n' >"$src/implicit-return.ts"
  printf 'export function name(n: number): string {\n  let out = "";\n  switch (n) {\n    case 0:\n      out = "zero";\n    case 1:\n      out = "one";\n      break;\n  }\n  return out;\n}\n' >"$src/fallthrough.ts"
  run "$work" pnpm exec tsc -p apps/dashboard/tsconfig.json
  finding 'implicit-any\.ts.*TS7006'
  finding 'strict-null\.ts.*TS18047'
  finding 'unchecked-index\.ts.*TS2322'
  finding 'exact-optional\.ts.*TS2375'
  finding 'implicit-override\.ts.*TS4114'
  finding 'index-signature\.ts.*TS4111'
  finding 'type-import\.ts.*TS1484'
  finding 'erasable\.ts.*TS1294'
  finding 'unused-local\.ts.*TS6133'
  finding 'implicit-return\.ts.*TS7030'
  finding 'fallthrough\.ts.*TS7029'
  ;;
eslint)
  printf 'export const value: number = JSON.parse("1");\n' >"$app/src/unsafe.ts"
  run "$app" pnpm exec eslint --max-warnings 0 .
  finding 'src/unsafe\.ts'
  finding '@typescript-eslint/no-unsafe-assignment'
  ;;
cover)
  # A source file that no test imports.
  printf 'export function sign(n: number): string {\n  return n < 0 ? "-" : "+";\n}\n' >"$app/src/untested.ts"
  run "$app" pnpm exec vitest run --coverage
  finding 'untested\.ts'
  finding 'does not meet'
  ;;
property)
  cat >"$app/src/counterexample.test.ts" <<-'EOF'
		import { test } from "@fast-check/vitest";
		import fc from "fast-check";

		test.prop([fc.integer()])("an integer is never negative", (n) => n >= 0);
	EOF
  run "$app" pnpm exec vitest run
  finding 'Counterexample'
  ;;
dup)
  for name in first second; do
    cat >"$app/src/$name.ts" <<-EOF
			export function $name(values: readonly number[]): number {
			  let total = 0;
			  for (const value of values) {
			    if (value > 0) {
			      total += value * 2;
			    } else {
			      total -= value * 3;
			    }
			  }
			  return total;
			}
		EOF
  done
  run "$work" pnpm exec jscpd --config .jscpd.json
  finding 'dashboard/src/first\.ts'
  finding 'dashboard/src/second\.ts'
  finding 'Found 1 clones'
  ;;
deadcode)
  # An unused file, an export nothing imports, a declared dependency
  # nothing imports, and an import of a package the manifest omits.
  printf 'export function orphan(): number {\n  return 1;\n}\n' >"$app/src/orphan.ts"
  printf '\nexport const unusedExport = 1;\n' >>"$app/src/app.tsx"
  printf 'import "scheduler";\n' >>"$app/src/main.ts"
  # Declare a package that no file imports, next to the vite entry.
  manifest=$(<"$app/package.json")
  printf '%s\n' "${manifest/\"vite\": /\"is-unused-fixture\": \"1.0.0\",
    \"vite\": }" >"$app/package.json"
  run "$work" pnpm exec knip
  finding 'Unused files'
  finding 'apps/dashboard/src/orphan\.ts'
  finding 'Unused exports'
  finding 'unusedExport'
  finding 'Unused devDependencies'
  finding 'is-unused-fixture'
  finding 'Unlisted dependencies'
  finding 'scheduler'
  ;;
audit)
  # A project whose lockfile resolves a release with a published critical
  # advisory, then the same audit against a registry that does not answer,
  # which must count as a failed scan.
  printf '{\n  "name": "audit-check",\n  "private": true,\n  "dependencies": { "minimist": "1.2.5" }\n}\n' >"$work/package.json"
  (cd "$work" && pnpm install --lockfile-only --ignore-workspace >/dev/null)
  run "$work" pnpm audit --audit-level high --ignore-workspace
  finding 'Severity: 1 critical'
  run "$work" pnpm audit --audit-level high --ignore-workspace --registry http://127.0.0.1:9/
  finding 'ERR_PNPM_AUDIT_BAD_RESPONSE'
  ;;
mutate)
  # A test that runs the code but checks nothing that a mutant changes.
  printf 'export function sign(n: number): string {\n  return n < 0 ? "-" : "+";\n}\n' >"$app/src/sign.ts"
  cat >"$app/src/sign.test.ts" <<-'EOF'
		import { expect, test } from "vitest";
		import { sign } from "./sign.ts";

		test("sign returns a string", () => {
		  expect(sign(-1)).toBeTypeOf("string");
		});
	EOF
  run "$app" pnpm exec stryker run ../../stryker.config.json
  finding 'src/sign\.ts'
  finding 'under breaking threshold 100'
  ;;
e2e-console)
  # The page logs a console error and is otherwise unchanged.
  cat >"$app/src/app.tsx" <<-'EOF'
		import type { JSX } from "react";

		export function App(): JSX.Element {
		  console.error("forced console error");
		  return <main />;
		}
	EOF
  run "$app" sh -c 'pnpm exec vite build && pnpm exec playwright test'
  finding 'the blank page loads without errors'
  finding 'forced console error'
  ;;
e2e-axe)
  # Check missing image text and content outside landmarks, a best-practice violation.
  cat >"$app/src/app.tsx" <<-'EOF'
		import type { JSX } from "react";

		export function App(): JSX.Element {
		  return (
		    <>
		      <main><img src="data:," /></main>
		      <p>Outside the main landmark</p>
		    </>
		  );
		}
	EOF
  run "$app" sh -c 'pnpm exec vite build && pnpm exec playwright test'
  finding 'the blank page passes the accessibility scan'
  finding 'image-alt'
  finding 'region'
  ;;
e2e-browser)
  # The pinned browser version names a build the run does not use.
  run "$app" env AFAKT_WEB_CHROMIUM=0.0.0 sh -c 'pnpm exec vite build && pnpm exec playwright test'
  finding 'Expected: "0\.0\.0"'
  ;;
*)
  printf 'unknown gate: %s\n' "$gate" >&2
  exit 2
  ;;
esac

# The links lead into the checkout's installed packages, and an install run
# inside the copy would rewrite the bin shims there to name the copy.
if [[ $gate != audit ]] && grep -rlqF -- "$work" "$root/node_modules/.bin" "$root/apps/dashboard/node_modules/.bin"; then
  printf 'FAIL %s: the run wrote into the installed packages; run pnpm install again\n' "$gate" >&2
  failures=$((failures + 1))
fi

if ((failures > 0)); then
  exit 1
fi
printf '%s fails on broken input\n' "$gate"
