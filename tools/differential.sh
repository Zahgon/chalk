#!/usr/bin/env bash
#
# End-to-end differential harness: real JavaScript oracle vs. real Go binary.
#
# This is oracle tooling, not part of the library. Building, testing and running
# chalk-go requires only the Go toolchain; this script additionally requires
# Node.js and a pristine checkout of the upstream JavaScript repository, and
# exists so a reviewer can independently reproduce the differential result
# recorded in truth.md.
#
# Usage:
#   tools/differential.sh <path-to-upstream-chalk-checkout>
#
# It runs both implementations on identical inputs and compares stdout, stderr
# and the exit code SEPARATELY, byte for byte. It exits non-zero on any
# difference and names every failing case.

set -uo pipefail

ORACLE=${1:-}
if [[ -z $ORACLE || ! -f $ORACLE/source/index.js ]]; then
	echo "usage: tools/differential.sh <path-to-upstream-chalk-checkout>" >&2
	echo "       (the directory containing source/index.js)" >&2
	exit 2
fi
ORACLE=$(cd "$ORACLE" && pwd)
REPO=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

command -v node >/dev/null || { echo "node not found on PATH" >&2; exit 2; }
command -v go >/dev/null || { echo "go not found on PATH" >&2; exit 2; }

WORK=$(mktemp -d "${TMPDIR:-/tmp}/chalk-differential.XXXXXX")
trap 'rm -rf "$WORK"' EXIT

pass=0
fail=0

# compare <case-name> <js-stdout> <js-stderr> <js-rc> <go-stdout> <go-stderr> <go-rc>
compare() {
	local name=$1 jsout=$2 jserr=$3 jsrc=$4 goout=$5 goerr=$6 gorc=$7
	local stream
	for stream in out err; do
		local j=$jsout g=$goout
		[[ $stream == err ]] && { j=$jserr; g=$goerr; }
		if cmp -s "$j" "$g"; then
			pass=$((pass + 1))
		else
			fail=$((fail + 1))
			echo "FAIL  $name  [std$stream]"
			diff <(cat -v "$j") <(cat -v "$g") | head -5
		fi
	done
	if [[ $jsrc == "$gorc" ]]; then
		pass=$((pass + 1))
	else
		fail=$((fail + 1))
		echo "FAIL  $name  [exit code]  js=$jsrc go=$gorc"
	fi
}

echo "oracle: $ORACLE"
echo "target: $REPO"
echo

# ---------------------------------------------------------------------------
# Entry point 1: the screenshot example, across every color level.
# Exercises the whole style table through a real program.
# ---------------------------------------------------------------------------
go build -o "$WORK/screenshot" "$REPO/examples/screenshot" || exit 1
for level in 0 1 2 3; do
	name="screenshot FORCE_COLOR=$level"
	(cd "$ORACLE" && FORCE_COLOR=$level node examples/screenshot.js) \
		>"$WORK/js.out" 2>"$WORK/js.err"
	jsrc=$?
	FORCE_COLOR=$level "$WORK/screenshot" >"$WORK/go.out" 2>"$WORK/go.err"
	gorc=$?
	compare "$name" "$WORK/js.out" "$WORK/js.err" "$jsrc" "$WORK/go.out" "$WORK/go.err" "$gorc"
	echo "checked $name"
done

# ---------------------------------------------------------------------------
# Entry point 2: the test fixture, across the terminal-detection matrix.
# Exercises supports-color end to end, including both the stdout and the
# stderr instance. `env -i` mirrors the upstream test's extendEnv:false.
# ---------------------------------------------------------------------------
go build -o "$WORK/fixture" "$REPO/testdata/fixture" || exit 1
fixture_case() {
	local name=$1
	shift
	env -i PATH="$PATH" HOME="$HOME" "$@" node "$ORACLE/test/_fixture.js" \
		>"$WORK/js.out" 2>"$WORK/js.err"
	local jsrc=$?
	env -i PATH="$PATH" HOME="$HOME" "$@" "$WORK/fixture" \
		>"$WORK/go.out" 2>"$WORK/go.err"
	local gorc=$?
	compare "$name" "$WORK/js.out" "$WORK/js.err" "$jsrc" "$WORK/go.out" "$WORK/go.err" "$gorc"
	echo "checked $name"
}

fixture_case "fixture: no environment"
fixture_case "fixture: FORCE_COLOR=0" FORCE_COLOR=0
fixture_case "fixture: FORCE_COLOR=1" FORCE_COLOR=1
fixture_case "fixture: FORCE_COLOR=2" FORCE_COLOR=2
fixture_case "fixture: FORCE_COLOR=3" FORCE_COLOR=3
fixture_case "fixture: FORCE_COLOR=4 (clamped)" FORCE_COLOR=4
fixture_case "fixture: FORCE_COLOR=true" FORCE_COLOR=true
fixture_case "fixture: FORCE_COLOR=false" FORCE_COLOR=false
fixture_case "fixture: FORCE_COLOR=unicorn" FORCE_COLOR=unicorn
fixture_case "fixture: FORCE_COLOR=1 COLORTERM=truecolor" FORCE_COLOR=1 COLORTERM=truecolor
fixture_case "fixture: FORCE_COLOR=true COLORTERM=truecolor" FORCE_COLOR=true COLORTERM=truecolor
fixture_case "fixture: TF_BUILD+AGENT_NAME" TF_BUILD=1 AGENT_NAME=agent
fixture_case "fixture: TF_BUILD+AGENT_NAME FORCE_COLOR=unicorn" TF_BUILD=1 AGENT_NAME=agent FORCE_COLOR=unicorn
fixture_case "fixture: TF_BUILD+AGENT_NAME FORCE_COLOR=0" TF_BUILD=1 AGENT_NAME=agent FORCE_COLOR=0
fixture_case "fixture: CI+GITHUB_ACTIONS" CI=1 GITHUB_ACTIONS=1
fixture_case "fixture: CI+GITHUB_ACTIONS FORCE_COLOR=1" CI=1 GITHUB_ACTIONS=1 FORCE_COLOR=1
fixture_case "fixture: TERM=xterm-256color" TERM=xterm-256color
fixture_case "fixture: TERM=dumb FORCE_COLOR=3" TERM=dumb FORCE_COLOR=3

echo
echo "differential result: $pass passed, $fail failed"
[[ $fail -eq 0 ]] || exit 1
