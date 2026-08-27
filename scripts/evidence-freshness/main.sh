#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
output="${EVIDENCE_FRESHNESS_OUTPUT:-evidence-freshness-output}"
source_path="examples/evidence-freshness/main.gooo"
mkdir -p "$output"

status_snapshot() {
  local status count digest
  status="$(git status --porcelain=v1 --untracked-files=all)"
  count="$(printf '%s\n' "$status" | awk 'NF {n++} END {print n+0}')"
  digest="$(printf '%s' "$status" | sha256sum | awk '{print "sha256:"$1}')"
  jq -n --arg digest "$digest" --argjson count "$count" '{digest:$digest,count:$count}'
}

before_snapshot="$(status_snapshot)"
forbidden_dependency_count=$(go list -deps ./cmd/evidence-freshness-decider | grep -Ec 'internal/meta/evidencefreshness/producer|internal/meta/evidencefreshness$' || true)
after_snapshot="$(status_snapshot)"
jq -n --argjson before "$before_snapshot" --argjson after "$after_snapshot" \
  '{before_digest:$before.digest,after_digest:$after.digest,before_count:$before.count,after_count:$after.count}' \
  > "$output/write-set.json"
jq -n --argjson forbidden_dependency_count "$forbidden_dependency_count" \
  '{schema:"gooo/evidence-freshness-independence/v1", forbidden_dependency_count:$forbidden_dependency_count, independence_contract:{numerator:1, denominator:1}}' \
  > "$output/independence.json"
test "$forbidden_dependency_count" = 0

common=(
  -contract examples/evidence-freshness/contract.json
  -source "$source_path"
  -head "$HEAD_SHA"
  -independence "$output/independence.json"
  -write-set "$output/write-set.json"
  -emit-dir "$output"
)
go run ./cmd/evidence-freshness "${common[@]}" -output "$output/report.json"
go run ./cmd/evidence-freshness "${common[@]}" -output "$output/replay.json"
cmp -s "$output/report.json" "$output/replay.json"

{
  printf '%s\n' '// CI intervention: presentation-only comment'
  cat "$source_path"
} > "$output/comment-only.gooo"
sed 's/claim-subject"/claim-subject-semantic-change"/' "$source_path" > "$output/semantic-change.gooo"

go run ./cmd/evidence-freshness-decider \
  -source "$source_path" -receipt "$output/receipt.json" \
  -context "$output/fresh-context.json" -output "$output/standalone-current.json"
go run ./cmd/evidence-freshness-decider \
  -source "$output/comment-only.gooo" -receipt "$output/receipt.json" \
  -context "$output/synthetic-comment-only-context.json" -output "$output/standalone-comment.json"
go run ./cmd/evidence-freshness-decider \
  -source "$output/semantic-change.gooo" -receipt "$output/receipt.json" \
  -context "$output/synthetic-semantic-change-context.json" -output "$output/standalone-semantic.json"
go run ./cmd/evidence-freshness-decider \
  -receipt "$output/receipt.json" -context "$output/synthetic-source-unavailable-context.json" \
  -output "$output/standalone-unavailable.json"
go run ./cmd/evidence-freshness -check "$output/report.json"

jq -e '
  .schema == "gooo/evidence-freshness-report/v2" and
  .scope == "CLAIM_JUSTIFICATION_BOUNDARY_ONLY" and
  .decision == "PASS" and .resolution == "EXACT" and
  .summary.cases_observed == 10 and .summary.current_evidence_cases == 1 and
  .summary.synthetic_counterexample_cases == 9 and .summary.axis_changes_observed == 6 and
  .summary.fixed_axis_denominator == 6 and
  .summary.raw_fresh_cases == 7 and .summary.raw_stale_cases == 2 and .summary.raw_unknown_cases == 1 and
  .summary.semantic_fresh_cases == 8 and .summary.semantic_stale_cases == 1 and .summary.semantic_unknown_cases == 1 and
  .summary.claim_fresh_cases == 2 and .summary.claim_stale_cases == 7 and .summary.claim_unknown_cases == 1 and
  .summary.raw_stale_by_stage.MATERIAL_CLOSURE == 2 and
  .summary.stale_by_stage.SUBJECT_BINDING == 1 and .summary.stale_by_stage.MATERIAL_CLOSURE == 1 and
  .summary.stale_by_stage.RECIPE_RESOLUTION == 1 and .summary.stale_by_stage.ENVIRONMENT_CAPTURE == 1 and
  .summary.stale_by_stage.RUNNER_EXECUTION == 1 and .summary.stale_by_stage.VERIFIER_JUDGMENT == 2 and
  .summary.unknown_by_stage.SUBJECT_BINDING == 1 and
  .summary.freshness_transitions == 10 and .summary.claim_ledger_entries == 10 and
  .summary.claim_discharged == 2 and .summary.claim_open_preserved == 8 and .summary.claim_refuted == 0 and
  .summary.source_reconstructed_cases == 9 and .summary.source_unavailable_cases == 1 and
  .summary.forbidden_dependency_count == 0 and
  .summary.independence_contract.numerator == 1 and .summary.independence_contract.denominator == 1 and
  .summary.read_only_before_count == 0 and .summary.read_only_after_count == 0 and
  .summary.read_only_write_set_stable == true and
  .independence.forbidden_dependency_count == 0 and
  .independence.independence_contract.numerator == 1 and
  .independence.independence_contract.denominator == 1 and
  .receipt.independence == .independence and .receipt.write_set == .write_set and
  ([.cases[] | select(.observation_class == "CURRENT_EVIDENCE")]|length) == 1 and
  ([.cases[] | select(.observation_class == "SYNTHETIC_COUNTEREXAMPLE")]|length) == 9 and
  ([.indicators[] | select(.satisfied != true)]|length) == 0 and
  .claim_ledger[0].prior_state == "OPEN" and .claim_ledger[0].next_state == "DISCHARGED" and
  ([.claim_ledger[] | select(.next_state == "REFUTED")]|length) == 0
' "$output/report.json"

jq -e '
  .schema == "gooo/evidence-freshness-verdict/v2" and .state == "FRESH" and .decision == "PASS" and
  .raw_freshness == "FRESH" and .semantic_freshness == "FRESH" and .transition.to == "CLAIM_PRESERVED"
' "$output/standalone-current.json"
jq -e '
  .state == "FRESH" and .decision == "PASS" and .raw_freshness == "STALE" and
  .semantic_freshness == "FRESH" and .reason == "RAW_MATERIAL_CHANGED_SEMANTIC_PRESERVED"
' "$output/standalone-comment.json"
jq -e '
  .state == "STALE" and .decision == "FAIL_CLOSED" and .raw_freshness == "STALE" and
  .semantic_freshness == "STALE" and .reason == "SEMANTIC_DIGEST_CHANGED" and .transition.to == "CLAIM_STALE"
' "$output/standalone-semantic.json"
jq -e '
  .state == "UNKNOWN" and .decision == "FAIL_CLOSED" and .resolution == "LOWER_RESOLUTION" and
  .coordinate.stage == "SUBJECT_BINDING" and .coordinate.step == "reconstruct-source" and
  .reason == "SOURCE_UNAVAILABLE" and .transition.to == "CLAIM_UNKNOWN"
' "$output/standalone-unavailable.json"

final_snapshot="$(status_snapshot)"
test "$(jq -r .digest <<<"$final_snapshot")" = "$(jq -r .after_digest "$output/write-set.json")"
test "$(jq -r .count <<<"$final_snapshot")" = "$(jq -r .after_count "$output/write-set.json")"
git diff --exit-code

{
  echo '## Evidence freshness'
  jq -r '"- decision: \(.decision) / \(.resolution)\n- current/synthetic: \(.summary.current_evidence_cases)/\(.summary.synthetic_counterexample_cases)\n- raw freshness fresh/stale/unknown: \(.summary.raw_fresh_cases)/\(.summary.raw_stale_cases)/\(.summary.raw_unknown_cases)\n- semantic freshness fresh/stale/unknown: \(.summary.semantic_fresh_cases)/\(.summary.semantic_stale_cases)/\(.summary.semantic_unknown_cases)\n- claim transitions: \(.summary.freshness_transitions) / ledger entries: \(.summary.claim_ledger_entries)\n- claim ledger discharged/open/refuted: \(.summary.claim_discharged)/\(.summary.claim_open_preserved)/\(.summary.claim_refuted)\n- source reconstruction: \(.summary.source_reconstructed_cases)/9\n- source unavailable: \(.summary.source_unavailable_cases)/1\n- producer forbidden dependency count (guardrail): \(.summary.forbidden_dependency_count)\n- producer import contract: \(.independence.independence_contract.numerator)/\(.independence.independence_contract.denominator)\n- write set before/after: \(.summary.read_only_before_count)/\(.summary.read_only_after_count)\n- receipt: \(.receipt_digest)"' "$output/report.json"
} >> "${GITHUB_STEP_SUMMARY:-/dev/null}"
