#!/usr/bin/env bats
# Build every scanner response and broken input in the test's temporary directory.

setup() {
	root=$(git -C "$BATS_TEST_DIRNAME" rev-parse --show-toplevel)
	work=$BATS_TEST_TMPDIR/repo
	mkdir -p "$work/scripts" "$work/.devtools/scripts" "$work/vendor" "$work/bin"
	cp "$root/scripts/gomodscan.sh" "$root/scripts/gomodscan-findings.jq" "$work/scripts/"
	cp "$root/.devtools/scripts/gomodscan.sh" "$work/.devtools/scripts/"
	cd "$work" || exit 1
	git init -q
	printf '%s\n' '# example.invalid/module v1.0.0' >vendor/modules.txt
	git add -- vendor/modules.txt
	export SCAN_REPORT=$work/report SCAN_ERRORS=$work/errors SCAN_STATUS=1
	cat >bin/go <<'SH'
#!/usr/bin/env bash
set -euo pipefail
[[ "$*" == 'tool gomodscan -format sarif' ]] || exit 3
cat "$SCAN_REPORT"
cat "$SCAN_ERRORS" >&2
exit "$SCAN_STATUS"
SH
	chmod +x bin/go
	export PATH="$work/bin:$PATH"
	printf '%s\n' 'gomodscan: 1 finding(s) across 1 module(s)' >"$SCAN_ERRORS"
	write_report deprecated example.invalid/module v1.0.0
	cat >.gomodscan-accepted.json <<'JSON'
[{"module":"example.invalid/module","version":"v1.0.0","finding":"deprecated","reason":"Retained for compatibility until the upstream replacement ships."}]
JSON
}

write_report() {
	printf '{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"gomodscan"}},"results":[{"ruleId":"%s","message":{"text":"Finding"},"locations":[{"logicalLocations":[{"fullyQualifiedName":"%s@%s","kind":"module"}]}]}]}]}\n' \
		"$1" "$2" "$3" >"$SCAN_REPORT"
}

@test "accepted finding passes and prints its reason" {
	run bash scripts/gomodscan.sh text
	[ "$status" -eq 0 ]
	[[ $output == *'accepted deprecated: example.invalid/module@v1.0.0; reason: Retained for compatibility'* ]]
}

@test "SARIF retains the complete report and checks the same waivers" {
	run bash scripts/gomodscan.sh sarif output.sarif
	[ "$status" -eq 0 ]
	cmp "$SCAN_REPORT" output.sarif
	[[ $output == *'accepted deprecated:'* ]]
}

@test "unknown finding fails" {
	printf '%s\n' '[]' >.gomodscan-accepted.json
	write_report retracted example.invalid/module v1.0.0
	run bash scripts/gomodscan.sh text
	[ "$status" -ne 0 ]
	[[ $output == *'unaccepted retracted:'* ]]
}

@test "changed module or version fails exact matching" {
	write_report deprecated example.invalid/other v1.0.0
	run bash scripts/gomodscan.sh text
	[ "$status" -ne 0 ]
	write_report deprecated example.invalid/module v1.0.1
	run bash scripts/gomodscan.sh text
	[ "$status" -ne 0 ]
}

@test "stale accepted entry fails after a clean scan" {
	export SCAN_STATUS=0
	printf '%s\n' 'gomodscan: scanned 1 module(s), no findings' >"$SCAN_ERRORS"
	printf '%s\n' '{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"gomodscan"}},"results":[]}]}' >"$SCAN_REPORT"
	run bash scripts/gomodscan.sh text
	[ "$status" -ne 0 ]
	[[ $output == *'stale accepted finding deprecated:'* ]]
}

@test "incomplete scan fails even with an accepted result" {
	printf '%s\n' 'gomodscan: lookup failed' >>"$SCAN_ERRORS"
	run bash scripts/gomodscan.sh text
	[ "$status" -eq 2 ]
	[[ $output == *'incomplete scan'* ]]
	[[ $output != *'accepted deprecated:'* ]]
}

@test "missing completion diagnostic fails" {
	: >"$SCAN_ERRORS"
	run bash scripts/gomodscan.sh text
	[ "$status" -eq 2 ]
}

@test "scanner execution failure fails" {
	export SCAN_STATUS=3
	run bash scripts/gomodscan.sh text
	[ "$status" -eq 3 ]
}

@test "malformed SARIF fails" {
	printf '%s\n' '{"version":"2.1.0"}' >"$SCAN_REPORT"
	run bash scripts/gomodscan.sh text
	[ "$status" -eq 2 ]
}

@test "missing finding identifier fails" {
	write_report '' example.invalid/module v1.0.0
	run bash scripts/gomodscan.sh text
	[ "$status" -ne 0 ]
}

@test "a reason is mandatory" {
	printf '%s\n' '[{"module":"example.invalid/module","version":"v1.0.0","finding":"deprecated","reason":" "}]' >.gomodscan-accepted.json
	run bash scripts/gomodscan.sh text
	[ "$status" -ne 0 ]
}

@test "duplicate accepted entries fail" {
	printf '%s\n' '[{"module":"example.invalid/module","version":"v1.0.0","finding":"deprecated","reason":"Retained"},{"module":"example.invalid/module","version":"v1.0.0","finding":"deprecated","reason":"Retained"}]' >.gomodscan-accepted.json
	run bash scripts/gomodscan.sh text
	[ "$status" -ne 0 ]
}

@test "an unreadable accepted file fails" {
	rm .gomodscan-accepted.json
	run bash scripts/gomodscan.sh text
	[ "$status" -ne 0 ]
}

@test "missing tracked vendor manifest fails" {
	git rm --cached -- vendor/modules.txt
	run bash scripts/gomodscan.sh text
	[ "$status" -eq 2 ]
}

@test "SARIF destination that cannot be opened fails" {
	run bash scripts/gomodscan.sh sarif missing/output.sarif
	[ "$status" -eq 2 ]
}
