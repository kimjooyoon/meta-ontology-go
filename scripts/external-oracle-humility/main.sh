#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"

output_root="$RUNNER_TEMP/external-oracle-humility"
rm -rf "$output_root"
mkdir -p "$output_root/first" "$output_root/replay"
args=(
  --head "$HEAD_SHA"
  --source examples/external-oracle-humility/main.gooo
  --contract examples/external-oracle-humility/contract.json
  --references examples/external-oracle-humility/references.json
)
go run ./cmd/external-oracle-humility-witness "${args[@]}" --output "$output_root/first"
go run ./cmd/external-oracle-humility-witness "${args[@]}" --output "$output_root/replay"
for name in source-receipt.json agreement-report.json mismatch-report.json absence-report.json suite.json; do
  cmp -s "$output_root/first/$name" "$output_root/replay/$name"
done

jq -e '
  .decision == "REFERENCE_AGREEMENT_OBSERVED" and
  .resolution == "EXACT" and .reference_agreement == "AGREES" and
  .semantic_authority == "GOOO_SOURCE_INTENT" and .authority_grant == "NONE" and
  .enforcement_effect == "NO_EFFECT" and .completed == 12 and .total == 12 and
  .basis_points == 10000 and .official_mutations == 0 and
  .repository_writes == 0 and .promotion_count == 0 and .decision != "PASS" and
  ([.indicators[] | select(.proof_choice == "FOUNDATION")] | length) == 6 and
  ([.indicators[] | select(.proof_choice == "COHERENCE")] | length) == 3 and
  ([.indicators[] | select(.proof_choice == "REGRESSION")] | length) == 3 and
  ([.claim_transitions[] | select(.persisted == true)] | length) == 3
' "$output_root/first/agreement-report.json"

jq -e '
  .decision == "FAIL_CLOSED" and .resolution == "EXACT" and
  .reference_agreement == "DISAGREES" and .enforcement_effect == "BLOCK" and
  .semantic_authority == "GOOO_SOURCE_INTENT" and .authority_grant == "NONE" and
  .decision != "PASS"
' "$output_root/first/mismatch-report.json"

jq -e '
  .decision == "FAIL_CLOSED" and .resolution == "UNKNOWN" and
  .reference_agreement == "UNKNOWN" and .enforcement_effect == "BLOCK" and
  .semantic_authority == "GOOO_SOURCE_INTENT" and .authority_grant == "NONE" and
  .unknown_indicators > 0 and .decision != "PASS"
' "$output_root/first/absence-report.json"

jq -e '
  .decision == "HUMILITY_MODEL_BOUND" and .resolution == "EXACT" and
  .cases_satisfied == 3 and .cases_total == 3 and .coverage_bps == 10000 and
  .official_mutations == 0 and .repository_writes == 0 and .promotion_count == 0
' "$output_root/first/suite.json"

forbidden='internal/(syntax|bidir|languagesourceexecution)'
dependencies=$(go list -deps ./cmd/external-oracle-humility-witness | grep -Ec "$forbidden" || true)
test "$dependencies" -eq 0
jq -n --argjson dependencies "$dependencies" \
  '{schema:"gooo/external-oracle-humility-independence/v1", producer_dependencies:$dependencies, fixed_denominator:12}' \
  > "$output_root/first/independence.json"
test "$(git status --short)" = ""

agreement=$(jq -r '.decision + " / " + .resolution' "$output_root/first/agreement-report.json")
agreement_counts=$(jq -r '"\(.completed)/\(.total) (\(.basis_points) bps)"' "$output_root/first/agreement-report.json")
mismatch=$(jq -r '.decision + " / " + .resolution' "$output_root/first/mismatch-report.json")
absence=$(jq -r '.decision + " / " + .resolution' "$output_root/first/absence-report.json")
{
  echo '### External oracle humility'
  echo
  jq -r --arg agreement "$agreement" --arg agreement_counts "$agreement_counts" \
    --arg mismatch "$mismatch" --arg absence "$absence" \
    '"- agreement report: `\($agreement)`\n- agreement indicators: `\($agreement_counts)`\n- mismatch: `\($mismatch)`\n- absence: `\($absence)`\n- case coverage: `\(.cases_satisfied)/\(.cases_total)` (`\(.coverage_bps)` bps)\n- authority grant: `NONE`\n- writes: `\(.official_mutations)/\(.repository_writes)`; promotions: `\(.promotion_count)`"' \
    "$output_root/first/suite.json"
} >> "$GITHUB_STEP_SUMMARY"
