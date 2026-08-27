#!/usr/bin/env bash
set -euo pipefail

root="${GITHUB_WORKSPACE:-$(pwd)}"
output="${RUNNER_TEMP:-/tmp}/ambiguity-budget"
mkdir -p "$output"
head_sha="${HEAD_SHA:-$(git rev-parse HEAD)}"
contract="$root/examples/ambiguity-budget/contract.json"
source="$root/examples/ambiguity-budget/main.gooo"

forbidden_import='github.com/kimjooyoon/meta-ontology-go/internal/meta/ambiguitybudget"'
forbidden_imports="$(rg -n "$forbidden_import" "$root/internal/meta/ambiguitybudgetjudge" "$root/cmd/ambiguity-budget-verifier" || true)"
if [[ -n "$forbidden_imports" ]]; then
  echo "forbidden producer imports detected:" >&2
  echo "$forbidden_imports" >&2
  exit 1
fi
for suffix in first second; do
  go run ./cmd/ambiguity-budget-witness \
    -head "$head_sha" -contract "$contract" -source "$source" \
    -output "$output/receipt-$suffix.json"
done
cmp "$output/receipt-first.json" "$output/receipt-second.json"

for suffix in first second; do
  go run ./cmd/ambiguity-budget-verifier \
    -contract "$contract" -receipt "$output/receipt-first.json" -source "$source" \
    -output "$output/judge-$suffix.json"
done
cmp "$output/judge-first.json" "$output/judge-second.json"

jq -e '
  .conformance_decision == "PASS" and .conformance_resolution == "EXACT" and
  .subject_decision == "MIXED" and .subject_resolution == "LOWER_RESOLUTION" and
  .summary.cases_total == 4 and .summary.known_cases == 3 and
  .summary.fixed_denominator == 2 and .summary.integer_dimensions == 3 and
  .summary.interventions_total == 2 and .summary.open_claims == 1 and
  .summary.zero_ambiguity_cases == 1 and
  .summary.boundary_cases == 1 and .summary.over_budget_cases == 1 and
  .summary.unknown_cases == 1 and .summary.lower_resolution_cases == 2 and
  .effects.repository_writes == 0 and .effects.mutation_authority == false and
  (.producer | length) > 0 and (.consumer | length) > 0 and
  (.meta_operation | length) > 0 and (.proof_choice | length) > 0
' "$output/receipt-first.json"
jq -e '
  ([.cases[] | select(.id == "zero-ambiguity" and .decision == "PASS" and .resolution == "EXACT")] | length) == 1 and
  ([.cases[] | select(.id == "boundary-ambiguity" and .decision == "PASS" and .resolution == "EXACT")] | length) == 1 and
  ([.cases[] | select(.id == "over-budget-ambiguity" and .decision == "FAIL_CLOSED" and .resolution == "LOWER_RESOLUTION" and .reason == "AMBIGUITY_BUDGET_EXCEEDED")] | length) == 1 and
  ([.cases[] | select(.id == "unknown-ambiguity" and .decision == "UNKNOWN" and .resolution == "LOWER_RESOLUTION" and .coordinate.stage == "ambiguity-budget" and (.coordinate.step | length) > 0 and (.coordinate.reason | length) > 0)] | length) == 1 and
  ([.claims[] | select(.case_id == "over-budget-ambiguity" and .to == "LOWER_RESOLUTION" and (.stage | length) > 0 and (.step | length) > 0 and (.reason | length) > 0 and (.evidence_digest | length) > 0)] | length) == 1 and
  ([.claims[] | select(.case_id == "unknown-ambiguity" and .to == "OPEN" and (.stage | length) > 0 and (.step | length) > 0 and (.reason | length) > 0 and (.evidence_digest | length) > 0)] | length) == 1 and
  ([.claims[] | select((.stage | length) > 0 and (.step | length) > 0 and (.reason | length) > 0 and (.evidence_digest | length) > 0)] | length) == 4 and
  ([.interventions[] | select(.id == "semantic-count-crosses-boundary" and .kind == "SEMANTIC" and .satisfied == true and .counts_before != .counts_after and .semantic_digest_before != .semantic_digest_after and .resolution_after == "LOWER_RESOLUTION")] | length) == 1 and
  ([.interventions[] | select(.id == "nonsemantic-comment-only" and .kind == "NONSEMANTIC" and .satisfied == true and .source_digest_before != .source_digest_after and .semantic_digest_before == .semantic_digest_after and .counts_before == .counts_after)] | length) == 1
' "$output/receipt-first.json"
jq -e '
  .conformance_decision == "PASS" and .conformance_resolution == "EXACT" and
  .subject_decision == "MIXED" and .subject_resolution == "LOWER_RESOLUTION" and
  .fixed_denominator == 2 and .cases_total == 4 and .interventions_total == 2 and
  .forbidden_producer_imports == 0 and .allowed_producer_imports == 0 and
  .repository_writes == 0 and .mutation_authority == false
' "$output/judge-first.json"

cp "$output/receipt-first.json" "$output/receipt.json"
cp "$output/judge-first.json" "$output/judge.json"
echo "ambiguity budget: deterministic replay, boundary descent, and independent verification passed"
