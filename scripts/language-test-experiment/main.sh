#!/usr/bin/env bash
set -euo pipefail

root="${GITHUB_WORKSPACE:-$(pwd)}"
output="${1:-${RUNNER_TEMP:-/tmp}/language-test-experiment}"
gooo="$output/gooo"
evaluator="$output/language-test-experiment"
mkdir -p "$output"

go test ./cmd/gooo ./cmd/language-test-experiment ./internal/languagetest ./internal/meta/languagetestexperiment
gofmt -l cmd/gooo internal/languagetest internal/meta/languagetestexperiment cmd/language-test-experiment \
  | tee "$output/unformatted.txt"
test ! -s "$output/unformatted.txt"
go build -trimpath -o "$gooo" ./cmd/gooo
go build -trimpath -o "$evaluator" ./cmd/language-test-experiment

"$gooo" test --json examples/language-test/main.gooo > "$output/first.json"
"$gooo" test --json examples/language-test/main.gooo > "$output/replay.json"
if "$gooo" test --json examples/language-test/failing.gooo > "$output/assertion-failure.json"; then
  echo "failing language test unexpectedly passed" >&2
  exit 1
fi
if "$gooo" test --json examples/billing/main.gooo > "$output/missing.json"; then
  echo "missing language tests unexpectedly passed" >&2
  exit 1
fi

runtime="$(go env GOVERSION)"
executable_digest="sha256:$(sha256sum "$gooo" | awk '{print $1}')"
jq -n --arg subject "$HEAD_SHA" --arg executable "$executable_digest" --arg runtime "$runtime" \
  --slurpfile contract "$root/examples/language-test/contract.json" \
  --slurpfile first "$output/first.json" --slurpfile replay "$output/replay.json" \
  --slurpfile assertion "$output/assertion-failure.json" --slurpfile missing "$output/missing.json" \
  '{subject_sha:$subject,executable_digest:$executable,contract:$contract[0],first:{runtime:$runtime,receipt:$first[0]},replay:{runtime:$runtime,receipt:$replay[0]},assertion_failure:$assertion[0],missing:$missing[0]}' \
  > "$output/input.json"

"$evaluator" -input "$output/input.json" -output "$output/report.json"
"$evaluator" -input "$output/input.json" -check "$output/report.json"
jq -e '.decision=="PASS" and .resolution=="EXACT" and .summary.coordinates.satisfied==12 and .summary.coordinates.total==12 and (.views[]|select(.audience=="USER")|.satisfied==4 and .total==4) and (.views[]|select(.audience=="TOOL_AUTHOR")|.satisfied==8 and .total==8) and (.views[]|select(.audience=="GOVERNOR")|.satisfied==12 and .total==12) and .summary.passed_tests==2 and .summary.assertion_rejections==1 and .summary.missing_test_rejections==1 and .summary.unknowns==0 and .summary.effects.repository_writes==0 and .summary.effects.mutation_authority==false' "$output/report.json"

jq '.first.receipt.decision="UNKNOWN"' "$output/input.json" > "$output/unknown-top-input.json"
if "$evaluator" -input "$output/unknown-top-input.json" -output "$output/unknown-top-report.json"; then
  echo "unknown top language test decision unexpectedly passed" >&2
  exit 1
fi
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .reason=="LANGUAGE_TEST_DECISION_UNKNOWN" and .summary.coordinates.satisfied==0 and .summary.coordinates.total==12 and .summary.unknowns==1' "$output/unknown-top-report.json"

jq -r '"### Gooo language test experiment\n- decision: \(.decision) / \(.resolution)\n- indicators: \(.summary.coordinates.satisfied)/\(.summary.coordinates.total)\n- USER: \(.views[0].satisfied)/\(.views[0].total)\n- TOOL_AUTHOR: \(.views[1].satisfied)/\(.views[1].total)\n- GOVERNOR: \(.views[2].satisfied)/\(.views[2].total)\n- tests: declared=\(.summary.declared_tests), passed=\(.summary.passed_tests)\n- unknowns: \(.summary.unknowns)\n- receipt: \(.digest)"' "$output/report.json" >> "$GITHUB_STEP_SUMMARY"
