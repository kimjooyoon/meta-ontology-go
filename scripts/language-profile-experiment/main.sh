#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
work="${RUNNER_TEMP:-/tmp}/language-profile-experiment"
build="${RUNNER_TEMP:-/tmp}/language-profile-build"
binary="$build/gooo"
reducer="$build/language-profile-experiment"
contract="$root/examples/language-profile/contract.json"
mkdir -p "$work" "$build"

preflight="$work/preflight.json"
phase="INITIALIZE"
write_preflight() {
  jq -n --arg decision "$1" --arg resolution "$2" --arg reason "$3" --arg phase "$phase" --arg subject "$HEAD_SHA" \
    '{schema:"gooo/language-profile-preflight/v1",decision:$decision,resolution:$resolution,reason:$reason,phase:$phase,subject_sha:$subject}' > "$preflight"
}
finish_preflight() {
  code=$?
  trap - EXIT
  if test "$code" -ne 0; then
    write_preflight "FAIL_CLOSED" "LOWER_RESOLUTION" "${phase}_FAILED"
  fi
  exit "$code"
}
trap finish_preflight EXIT
write_preflight "UNKNOWN" "LOWER_RESOLUTION" "PREFLIGHT_RUNNING"

phase="GO_FIX"
go fix ./cmd/gooo ./cmd/language-profile-experiment ./internal/languageprofile ./internal/meta/languageprofileexperiment
git diff --exit-code
phase="GOFMT"
gofmt -l cmd/gooo cmd/language-profile-experiment internal/languageprofile internal/meta/languageprofileexperiment | tee "$work/unformatted.txt"
test ! -s "$work/unformatted.txt"
phase="GO_TEST"
go test ./cmd/gooo ./cmd/language-profile-experiment ./internal/languageprofile ./internal/meta/languageprofileexperiment
phase="GO_BUILD"
go build -trimpath -o "$binary" ./cmd/gooo
go build -trimpath -o "$reducer" ./cmd/language-profile-experiment

phase="PROFILE"
"$binary" profile --json --samples 5 --entry PayOrder examples/billing/main.gooo > "$work/first.json"
"$binary" profile --json --samples 5 --entry PayOrder examples/billing/main.gooo > "$work/replay.json"
if "$binary" profile --json --samples 5 --entry Missing examples/billing/main.gooo > "$work/unknown-entry.json"; then
  echo "unknown profile entry unexpectedly passed" >&2
  exit 1
fi
executable_digest="sha256:$(sha256sum "$binary" | cut -d' ' -f1)"
jq -n --arg subject "$HEAD_SHA" --arg executable "$executable_digest" \
  --slurpfile contract "$contract" --slurpfile first "$work/first.json" \
  --slurpfile replay "$work/replay.json" --slurpfile unknown "$work/unknown-entry.json" \
  '{subject_sha:$subject,executable_digest:$executable,contract:$contract[0],first:$first[0],replay:$replay[0],unknown_entry:$unknown[0]}' \
  > "$work/input.json"

phase="REDUCE"
"$reducer" -input "$work/input.json" -output "$work/report.json"
"$reducer" -input "$work/input.json" -check "$work/report.json"
jq -e '.decision=="PASS" and .resolution=="EXACT" and .summary.coordinates.satisfied==13 and .summary.coordinates.total==13 and .summary.profiles==2 and .summary.samples==10 and .summary.successful_executions==10 and .summary.execution_digest_variants==1 and .summary.resources.wall_observations==10 and .summary.resources.allocation_observations==10 and .summary.unknown_entry_rejections==1 and .summary.unknowns==0 and .repository_writes==0 and .mutation_authority==false and (.views[]|select(.audience=="USER")|.satisfied==5 and .total==5) and (.views[]|select(.audience=="TOOL_AUTHOR")|.satisfied==10 and .total==10) and (.views[]|select(.audience=="GOVERNOR")|.satisfied==13 and .total==13)' "$work/report.json"

phase="COUNTERFACTUAL"
jq '.first.decision="UNKNOWN"' "$work/input.json" > "$work/unknown-top-input.json"
if "$reducer" -input "$work/unknown-top-input.json" -output "$work/unknown-top-report.json"; then
  echo "unknown profile decision unexpectedly passed" >&2
  exit 1
fi
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .summary.coordinates.satisfied==0 and .summary.coordinates.total==13 and .summary.unknowns==1' "$work/unknown-top-report.json"
git diff --exit-code

phase="CLOSED"
write_preflight "PASS" "EXACT" "EXPERIMENT_CLOSED"
trap - EXIT
jq -r '"### Gooo language profile experiment\n- decision: \(.decision) / \(.resolution)\n- indicators: \(.summary.coordinates.satisfied)/\(.summary.coordinates.total)\n- USER: \(.views[0].satisfied)/\(.views[0].total)\n- TOOL_AUTHOR: \(.views[1].satisfied)/\(.views[1].total)\n- GOVERNOR: \(.views[2].satisfied)/\(.views[2].total)\n- profiles/samples: \(.summary.profiles)/\(.summary.samples)\n- wall ns min/median/max: \(.summary.resources.wall_min_nanoseconds)/\(.summary.resources.wall_median_nanoseconds)/\(.summary.resources.wall_max_nanoseconds)\n- TotalAlloc bytes min/median/max: \(.summary.resources.total_alloc_min_bytes)/\(.summary.resources.total_alloc_median_bytes)/\(.summary.resources.total_alloc_max_bytes)\n- receipt: \(.digest)"' "$work/report.json" >> "$GITHUB_STEP_SUMMARY"
