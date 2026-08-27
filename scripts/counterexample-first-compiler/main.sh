#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

output=counterexample-first-output
compiler=./cmd/counterexample-first-compiler-witness
judge=./cmd/counterexample-first-judge-witness
contract=examples/counterexample-first-compiler/contract.json
source=examples/counterexample-first-compiler/main.gooo
corpus=examples/counterexample-first-compiler/scenarios.json
paths=(cmd/counterexample-first-compiler-witness cmd/counterexample-first-judge-witness internal/meta/counterexamplefirst internal/meta/counterexamplefirstcompiler internal/meta/counterexamplefirstjudge)

mkdir -p "$output"
mapfile -t go_files < <(find "${paths[@]}" -name '*.go' -type f | sort)
gofmt -w "${go_files[@]}"
git diff --exit-code -- "${paths[@]}"
go test ./cmd/counterexample-first-compiler-witness ./cmd/counterexample-first-judge-witness ./internal/meta/counterexamplefirst ./internal/meta/counterexamplefirstcompiler ./internal/meta/counterexamplefirstjudge

go list -deps ./cmd/counterexample-first-judge-witness | sort > "$output/judge-dependencies.txt"
producer_dependencies=$(grep -Ec '^github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirstcompiler$' "$output/judge-dependencies.txt" || true)
jq -n --argjson dependencies "$producer_dependencies" \
  '{schema:"gooo/counterexample-first-independence/v1",producer_dependencies:$dependencies}' \
  > "$output/independence.json"
[[ "$producer_dependencies" == "0" ]]

common=(--head "$HEAD_SHA" --contract "$contract" --source "$source" --corpus "$corpus")
go run "$compiler" "${common[@]}" --out "$output/receipts-a.json"
go run "$compiler" "${common[@]}" --out "$output/receipts-b.json"
cmp -s "$output/receipts-a.json" "$output/receipts-b.json"

judge_common=("${common[@]}" --receipts "$output/receipts-a.json" --independence "$output/independence.json")
go run "$judge" "${judge_common[@]}" --out "$output/report-a.json"
jq -e '.decision == "PASS" and .resolution == "EXACT" and .reason == "COUNTEREXAMPLE_FIRST_CONTRACT_OBSERVED" and .fixed_denominator.version == "counterexample-first-denominator/v1" and .fixed_denominator.cases == 5 and .fixed_denominator.indicators == 10 and .fixed_denominator.claim_transitions == 15' "$output/report-a.json"
jq -e '.summary.cases_satisfied == 5 and .summary.cases_total == 5 and .summary.counterexamples_observed == 4 and .summary.minimal_counterexamples == 3 and .summary.resolutions_observed == 1 and .summary.promotions_after_resolution == 1 and .summary.success_only_blocks == 1 and .summary.unknown_coordinates_preserved == 1 and .summary.claim_transitions_preserved == 15 and .summary.receipts_verified == 5 and .summary.producer_dependencies == 0 and .summary.repository_writes == 0 and .summary.mutation_authority == false' "$output/report-a.json"
jq -e '.indicators | length == 10 and all(.[]; .satisfied == true and .denominator > 0)' "$output/report-a.json"
jq -e '.receipts[] | select(.scenario_id == "success-example-only") | .decision == "FAIL_CLOSED" and .reason == "COUNTEREXAMPLE_REQUIRED" and .coordinate.stage == "COUNTEREXAMPLE"' "$output/report-a.json"
jq -e '.receipts[] | select(.scenario_id == "resolved-minimal-counterexample") | .decision == "PASS" and .decision_input.counterexample_id != "" and .decision_input.resolution_id != "" and (.claim_transitions | length == 3)' "$output/report-a.json"
jq -e '.receipts[] | select(.scenario_id == "unknown-counterexample-coordinate") | .decision == "UNKNOWN" and .coordinate.stage == "UNKNOWN" and .coordinate.step == "UNKNOWN" and .coordinate.reason == "UNKNOWN"' "$output/report-a.json"

go run "$judge" "${common[@]}" --receipts "$output/receipts-b.json" --independence "$output/independence.json" --out "$output/report-b.json"
cmp -s "$output/report-a.json" "$output/report-b.json"

jq '.[0].reason = "COUNTEREXAMPLE_REQUIRED"' "$output/receipts-a.json" > "$output/tampered-receipts.json"
if go run "$judge" "${common[@]}" --receipts "$output/tampered-receipts.json" --independence "$output/independence.json" --out "$output/tampered-report.json"; then
  printf '%s\n' 'independent judge accepted a tampered receipt' >&2
  exit 1
else
  tamper_status=$?
fi
[[ "$tamper_status" == "1" ]]
jq -e '.decision == "FAIL_CLOSED" and .reason == "COUNTEREXAMPLE_RECEIPT_MISMATCH"' "$output/tampered-report.json"

git diff --exit-code

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  jq -r '"### Counterexample-first meta compilation\n- decision: \(.decision) / \(.resolution)\n- cases: \(.summary.cases_satisfied)/\(.summary.cases_total)\n- minimal counterexamples: \(.summary.minimal_counterexamples)\n- promotions after resolution: \(.summary.promotions_after_resolution)\n- success-only blocks: \(.summary.success_only_blocks)\n- unknown coordinates preserved: \(.summary.unknown_coordinates_preserved)\n- claim transitions: \(.summary.claim_transitions_preserved)/15\n- producer dependencies: \(.summary.producer_dependencies)\n- repository writes: \(.summary.repository_writes)\n- receipt: \(.digest)"' "$output/report-a.json" >> "$GITHUB_STEP_SUMMARY"
fi
