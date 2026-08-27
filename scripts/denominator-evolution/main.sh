#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

work="${RUNNER_TEMP:-/tmp}/denominator-evolution-${HEAD_SHA}"
mkdir -p "$work"

go test ./cmd/denominator-evolution-witness ./cmd/denominator-evolution-verify \
  ./internal/meta/denominatorevolution ./internal/meta/denominatorevolutionverify
go run ./cmd/gooo check examples/denominator-evolution/main.gooo

go list -deps ./cmd/denominator-evolution-verify > "$work/consumer-dependencies.txt"
if grep -q '^github.com/kimjooyoon/meta-ontology-go/internal/meta/denominatorevolution$' "$work/consumer-dependencies.txt"; then
  echo "independent consumer imports producer" >&2
  exit 1
fi

common=(--head "$HEAD_SHA" --contract examples/denominator-evolution/contract.json --source examples/denominator-evolution/main.gooo)
go run ./cmd/denominator-evolution-witness "${common[@]}" --out "$work/report-a.json"
go run ./cmd/denominator-evolution-witness "${common[@]}" --out "$work/report-b.json"
cmp -s "$work/report-a.json" "$work/report-b.json"

verify_common=("${common[@]}" --report "$work/report-a.json")
go run ./cmd/denominator-evolution-verify "${verify_common[@]}" --out "$work/verification-a.json"
go run ./cmd/denominator-evolution-verify "${verify_common[@]}" --out "$work/verification-b.json"
cmp -s "$work/verification-a.json" "$work/verification-b.json"

jq -e '
  .decision == "PASS" and
  .resolution == "EXACT" and
  .summary.cases_satisfied == 3 and .summary.cases_total == 3 and
  .denominator.version == "gooo/measurement-denominator/v1" and
  (.denominator.obligations | length) == 5 and
  .summary.fixed_denominator_numerator == 5 and .summary.fixed_denominator_denominator == 5 and
  .summary.legal_advance_numerator == 1 and .summary.legal_advance_denominator == 1 and
  .summary.unauthorized_rejection_numerator == 1 and .summary.unauthorized_rejection_denominator == 1 and
  .summary.unknown_predecessor_numerator == 1 and .summary.unknown_predecessor_denominator == 1 and
  .summary.addition_reason_numerator == 1 and .summary.addition_reason_denominator == 1 and
  .summary.deletion_reason_numerator == 1 and .summary.deletion_reason_denominator == 1 and
  .summary.forbidden_estimate_numerator == 0 and .summary.forbidden_estimate_denominator == 1 and
  .repository_writes == 0 and .mutation_authority == false and
  ([.cases[] | select(.status != "SATISFIED")] | length) == 0 and
  ([.indicators[] | select(.satisfied != true)] | length) == 0
' "$work/report-a.json"
jq -e '.decision == "PASS" and .resolution == "EXACT" and .repository_writes == 0 and .mutation_authority == false and ([.checks[] | select(.status != "PASS")] | length) == 0' "$work/verification-a.json"

git diff --exit-code

{
  echo "## Denominator evolution experiment"
  jq -r '"- fixed denominator: \(.denominator.version) / \(.summary.fixed_denominator_numerator)/\(.summary.fixed_denominator_denominator)\n- legal advance: \(.summary.legal_advance_numerator)/\(.summary.legal_advance_denominator)\n- unauthorized changes rejected: \(.summary.unauthorized_rejection_numerator)/\(.summary.unauthorized_rejection_denominator)\n- unknown predecessors rejected: \(.summary.unknown_predecessor_numerator)/\(.summary.unknown_predecessor_denominator)\n- aggregate estimate claims: \(.summary.forbidden_estimate_numerator)/\(.summary.forbidden_estimate_denominator)\n- producer receipt: \(.digest)"' "$work/report-a.json"
  jq -r '"- independent decision: \(.decision) / \(.resolution)\n- consumer receipt: \(.digest)"' "$work/verification-a.json"
} >> "$GITHUB_STEP_SUMMARY"
