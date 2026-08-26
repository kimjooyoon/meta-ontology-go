#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
work="${RUNNER_TEMP:-/tmp}/language-debug-experiment"
build="${RUNNER_TEMP:-/tmp}/language-debug-build"
binary="$build/gooo"
reducer="$build/language-debug-experiment"
contract="$root/examples/language-debug/contract.json"
mkdir -p "$work" "$build"

go fix ./cmd/gooo ./cmd/language-debug-experiment ./internal/languagedebug ./internal/meta/languagedebugexperiment
git diff --exit-code
gofmt -l cmd/gooo cmd/language-debug-experiment internal/languagedebug internal/meta/languagedebugexperiment | tee "$work/unformatted.txt"
test ! -s "$work/unformatted.txt"
go test ./cmd/gooo ./cmd/language-debug-experiment ./internal/languagedebug ./internal/meta/languagedebugexperiment
go build -trimpath -o "$binary" ./cmd/gooo
go build -trimpath -o "$reducer" ./cmd/language-debug-experiment

"$binary" debug --json --entry PayOrder --break-event SOURCE_PARSED examples/billing/main.gooo > "$work/first.json"
"$binary" debug --json --entry PayOrder --break-event ACTIVITY_INVOKED examples/billing/main.gooo > "$work/second.json"
if "$binary" debug --json --entry PayOrder --break-event MISSING examples/billing/main.gooo > "$work/unknown-breakpoint.json"; then
  echo "unknown debug breakpoint unexpectedly passed" >&2
  exit 1
fi
executable_digest="sha256:$(sha256sum "$binary" | cut -d' ' -f1)"
jq -n --arg subject "$HEAD_SHA" --arg executable "$executable_digest" \
  --slurpfile contract "$contract" --slurpfile first "$work/first.json" \
  --slurpfile second "$work/second.json" --slurpfile unknown "$work/unknown-breakpoint.json" \
  '{subject_sha:$subject,executable_digest:$executable,contract:$contract[0],first:$first[0],second:$second[0],unknown_breakpoint:$unknown[0]}' \
  > "$work/input.json"

"$reducer" -input "$work/input.json" -output "$work/report.json"
"$reducer" -input "$work/input.json" -check "$work/report.json"
jq -e '.decision=="PASS" and .resolution=="EXACT" and .summary.coordinates.satisfied==12 and .summary.coordinates.total==12 and .summary.debug_receipts==2 and .summary.paused_sessions==2 and .summary.breakpoints_reached==2 and .summary.trace_events==4 and .summary.execution_digest_variants==1 and .summary.current_events==2 and .summary.remaining_events==4 and .summary.unknown_breakpoint_rejections==1 and .summary.unknowns==0 and .repository_writes==0 and .mutation_authority==false and (.views[]|select(.audience=="USER")|.satisfied==4 and .total==4) and (.views[]|select(.audience=="TOOL_AUTHOR")|.satisfied==9 and .total==9) and (.views[]|select(.audience=="GOVERNOR")|.satisfied==12 and .total==12)' "$work/report.json"

jq '.first.decision="UNKNOWN"' "$work/input.json" > "$work/unknown-top-input.json"
if "$reducer" -input "$work/unknown-top-input.json" -output "$work/unknown-top-report.json"; then
  echo "unknown debug decision unexpectedly passed" >&2
  exit 1
fi
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .summary.coordinates.satisfied==0 and .summary.coordinates.total==12 and .summary.unknowns==1' "$work/unknown-top-report.json"
git diff --exit-code

jq -r '"### Gooo trace debugger experiment\n- decision: \(.decision) / \(.resolution)\n- indicators: \(.summary.coordinates.satisfied)/\(.summary.coordinates.total)\n- USER: \(.views[0].satisfied)/\(.views[0].total)\n- TOOL_AUTHOR: \(.views[1].satisfied)/\(.views[1].total)\n- GOVERNOR: \(.views[2].satisfied)/\(.views[2].total)\n- paused sessions: \(.summary.paused_sessions)\n- trace events: \(.summary.trace_events)\n- receipt: \(.digest)"' "$work/report.json" >> "$GITHUB_STEP_SUMMARY"
