#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
output="${EVIDENCE_FRESHNESS_OUTPUT:-evidence-freshness-output}"
mkdir -p "$output"

forbidden_dependency_count=$(go list -deps ./cmd/evidence-freshness-decider | grep -Ec 'internal/meta/evidencefreshness/(producer|evidencefreshness$)' || true)
jq -n --argjson forbidden_dependency_count "$forbidden_dependency_count" \
  '{schema:"gooo/evidence-freshness-independence/v1", forbidden_dependency_count:$forbidden_dependency_count, independence_contract:{numerator:1, denominator:1}}' \
  > "$output/independence.json"
test "$forbidden_dependency_count" = 0

common=(
  -contract examples/evidence-freshness/contract.json
  -source examples/evidence-freshness/main.gooo
  -head "$HEAD_SHA"
  -independence "$output/independence.json"
  -emit-dir "$output"
)
go run ./cmd/evidence-freshness "${common[@]}" -output "$output/report.json"
go run ./cmd/evidence-freshness "${common[@]}" -output "$output/replay.json"
cmp -s "$output/report.json" "$output/replay.json"

go run ./cmd/evidence-freshness-decider \
  -receipt "$output/receipt.json" \
  -context "$output/fresh-context.json" \
  -output "$output/standalone-verdict.json"
go run ./cmd/evidence-freshness -check "$output/report.json"

jq -e '
  .schema == "gooo/evidence-freshness-report/v1" and
  .scope == "CLAIM_JUSTIFICATION_BOUNDARY_ONLY" and
  .decision == "PASS" and .resolution == "EXACT" and
  .summary.cases_satisfied == 10 and .summary.cases_total == 10 and
  .summary.fresh_cases == 1 and .summary.stale_cases == 7 and .summary.unknown_cases == 2 and
  .summary.fixed_axis_denominator == 6 and .summary.axis_changes_observed == 6 and
  .summary.preservation_transitions == 10 and .summary.temporal_boundary_cases == 1 and
  .summary.read_only_cases == 1 and
  .summary.stale_by_stage.SUBJECT_BINDING == 1 and
  .summary.stale_by_stage.MATERIAL_CLOSURE == 1 and
  .summary.stale_by_stage.RECIPE_RESOLUTION == 1 and
  .summary.stale_by_stage.ENVIRONMENT_CAPTURE == 1 and
  .summary.stale_by_stage.RUNNER_EXECUTION == 1 and
  .summary.stale_by_stage.VERIFIER_JUDGMENT == 2 and
  .summary.unknown_by_stage.SUBJECT_BINDING == 1 and
  .summary.unknown_by_stage.VERIFIER_JUDGMENT == 1 and
  .summary.forbidden_dependency_count == 0 and
  .summary.independence_contract.numerator == 1 and
  .summary.independence_contract.denominator == 1 and
  .independence.forbidden_dependency_count == 0 and
  .independence.independence_contract.numerator == 1 and
  .independence.independence_contract.denominator == 1 and
  .receipt.independence == .independence and
  ([.cases[] | select(.status == "SATISFIED")]|length) == 10 and
  ([.indicators[] | select(.satisfied != true)]|length) == 0 and
  .repository_writes == 0 and .mutation_authority == false
' "$output/report.json"

jq -e '
  .schema == "gooo/evidence-freshness-verdict/v1" and
  .state == "FRESH" and .decision == "PASS" and
  .coordinate.stage == "VERIFIER_JUDGMENT" and
  .reason == "TUPLE_EXACT_AND_BOUNDARY_CURRENT" and
  .transition.to == "CLAIM_PRESERVED"
' "$output/standalone-verdict.json"

git diff --exit-code

{
  echo '## Evidence freshness'
  jq -r '"- decision: \(.decision) / \(.resolution)\n- cases: \(.summary.cases_satisfied)/\(.summary.cases_total)\n- fresh/stale/unknown: \(.summary.fresh_cases)/\(.summary.stale_cases)/\(.summary.unknown_cases)\n- coupling axes: \(.summary.axis_changes_observed)/\(.summary.fixed_axis_denominator)\n- claim transitions: \(.summary.preservation_transitions)\n- stale stages: subject \(.summary.stale_by_stage.SUBJECT_BINDING), material \(.summary.stale_by_stage.MATERIAL_CLOSURE), recipe \(.summary.stale_by_stage.RECIPE_RESOLUTION), environment \(.summary.stale_by_stage.ENVIRONMENT_CAPTURE), runner \(.summary.stale_by_stage.RUNNER_EXECUTION), verifier \(.summary.stale_by_stage.VERIFIER_JUDGMENT)\n- forbidden dependency count (guardrail): \(.summary.forbidden_dependency_count)\n- independence contract: \(.summary.independence_contract.numerator)/\(.summary.independence_contract.denominator)\n- receipt: \(.receipt_digest)"' "$output/report.json"
} >> "${GITHUB_STEP_SUMMARY:-/dev/null}"
