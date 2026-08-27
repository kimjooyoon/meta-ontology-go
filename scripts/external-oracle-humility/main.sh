#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"

output_root="$RUNNER_TEMP/external-oracle-humility"
rm -rf "$output_root"
mkdir -p "$output_root/first" "$output_root/replay" "$output_root/current" "$output_root/snapshots"

source_path="examples/external-oracle-humility/main.gooo"
contract_path="examples/external-oracle-humility/contract.json"
capsule_path="examples/external-oracle-humility/references.json"
mismatch_path="examples/external-oracle-humility/conformance/mismatch-capsule.json"
absence_path="examples/external-oracle-humility/conformance/absence-capsule.json"
intervention_path="examples/external-oracle-humility/interventions/authority-policy.gooo"
comment_path="examples/external-oracle-humility/interventions/comment-only.gooo"
current_path="$output_root/current/observations.json"
effects_path="$output_root/effects.json"
independence_path="$output_root/independence.json"

retrieve_reference() {
  local reference_id="$1"
  local reference_url="$2"
  local raw_path="$output_root/current/$reference_id.raw"
  local http_status
  if http_status=$(curl --fail --silent --show-error --location --retry 2 --connect-timeout 20 --max-time 90 --output "$raw_path" --write-out '%{http_code}' "$reference_url"); then
    :
  else
    http_status=0
    : > "$raw_path"
  fi
  local content_sha256="sha256:$(sha256sum "$raw_path" | awk '{print $1}')"
  local byte_count
  byte_count=$(wc -c < "$raw_path" | tr -d '[:space:]')
  jq -n --arg id "$reference_id" --arg url "$reference_url" --arg captured_at "actions-head:$HEAD_SHA" \
    --arg digest "$content_sha256" --argjson status "$http_status" --argjson bytes "$byte_count" \
    '{id:$id,url:$url,http_status:$status,bytes:$bytes,content_sha256:$digest,origin:"ACTIONS_RETRIEVAL",captured_at:$captured_at}' \
    > "$output_root/current/$reference_id.json"
}

while IFS=$'\t' read -r reference_id reference_url; do
  retrieve_reference "$reference_id" "$reference_url"
done < <(jq -r '.references[] | [.id, .url] | @tsv' "$contract_path")
jq -s '{schema:"gooo/external-oracle-current-observations/v1",observation_state:"ACTIONS_RETRIEVAL",references:.}' \
  "$output_root/current/gomacro-readme.json" "$output_root/current/racket-syntax-model.json" \
  "$output_root/current/reproducible-builds-definition.json" > "$current_path"

git status --porcelain=v1 | sort > "$output_root/snapshots/before.status"
head_before=$(git rev-parse HEAD)

producer_to_consumer=$(go list -deps ./internal/meta/externaloraclehumilityproducer | grep -F 'internal/meta/externaloraclehumilityconsumer' | wc -l | tr -d '[:space:]' || true)
consumer_to_producer=$(go list -deps ./internal/meta/externaloraclehumilityconsumer | grep -F 'internal/meta/externaloraclehumilityproducer' | wc -l | tr -d '[:space:]' || true)
jq -n --argjson producer "$producer_to_consumer" --argjson consumer "$consumer_to_producer" \
  '{schema:"gooo/external-oracle-independence/v2",producer_to_consumer:$producer,consumer_to_producer:$consumer}' > "$independence_path"
test "$producer_to_consumer" -eq 0
test "$consumer_to_producer" -eq 0

args=(
  --head "$HEAD_SHA"
  --source "$source_path"
  --contract "$contract_path"
  --references "$capsule_path"
  --current "$current_path"
  --mismatch-capsule "$mismatch_path"
  --absence-capsule "$absence_path"
  --intervention-source "$intervention_path"
  --comment-source "$comment_path"
)
go run ./cmd/external-oracle-humility-witness "${args[@]}" --output "$output_root/first"

git status --porcelain=v1 | sort > "$output_root/snapshots/after.status"
head_after=$(git rev-parse HEAD)
if cmp -s "$output_root/snapshots/before.status" "$output_root/snapshots/after.status"; then
  repository_writes=0
  official_mutations=0
else
  repository_writes=$(comm -13 "$output_root/snapshots/before.status" "$output_root/snapshots/after.status" | sed '/^$/d' | wc -l | tr -d '[:space:]')
  official_mutations=$(comm -13 "$output_root/snapshots/before.status" "$output_root/snapshots/after.status" | grep -E 'examples/external-oracle-humility/(main\.gooo|contract\.json|references\.json)$' | wc -l | tr -d '[:space:]' || true)
fi
promotion_count=0
if test "$head_before" != "$head_after"; then promotion_count=1; fi
jq -n --arg before "$(cat "$output_root/snapshots/before.status")" --arg after "$(cat "$output_root/snapshots/after.status")" \
  --arg head_before "$head_before" --arg head_after "$head_after" \
  --argjson official "$official_mutations" --argjson writes "$repository_writes" --argjson promotion "$promotion_count" \
  '{before_status:$before,after_status:$after,head_before:$head_before,head_after:$head_after,official_mutations:$official,repository_writes:$writes,promotion_count:$promotion}' > "$effects_path"

args+=(--effects "$effects_path" --independence "$independence_path")
go run ./cmd/external-oracle-humility-witness "${args[@]}" --output "$output_root/first"
go run ./cmd/external-oracle-humility-witness "${args[@]}" --output "$output_root/replay"
for name in source-receipt.json agreement-report.json mismatch-report.json absence-report.json intervention-report.json comment-report.json suite.json; do
  cmp -s "$output_root/first/$name" "$output_root/replay/$name"
done
cp "$effects_path" "$output_root/first/effects.json"
cp "$independence_path" "$output_root/first/independence.json"

jq -e '
  .decision == "REFERENCE_AGREEMENT_OBSERVED" and
  (.resolution == "EXACT" or .resolution == "LOWER_RESOLUTION") and
  .reference_agreement == "AGREES" and .semantic_authority == "GOOO_SOURCE_INTENT" and
  .authority_grant == "NONE" and .enforcement_effect == "NO_EFFECT" and
  .total == 12 and .decision != "PASS" and .read_only == true and
  .official_mutations == 0 and .repository_writes == 0 and .promotion_count == 0 and
  .fixed_denominator.source_policy.completed == 1 and .fixed_denominator.source_policy.total == 1 and
  .fixed_denominator.producer_imports.completed == 0 and .fixed_denominator.producer_imports.total == 0 and
  .fixed_denominator.producer_imports.satisfied == true and
  (.fixed_denominator.current_reference_observations.completed >= 0) and
  .fixed_denominator.current_reference_observations.total == 3 and
  .fixed_denominator.historical_fixtures.completed == 3 and .fixed_denominator.historical_fixtures.total == 3 and
  .fixed_denominator.semantic_causality.completed == 1 and .fixed_denominator.semantic_causality.total == 1 and
  .fixed_denominator.nonsemantic_preservation.completed == 1 and .fixed_denominator.nonsemantic_preservation.total == 1 and
  .current_reference_observations == .fixed_denominator.current_reference_observations.completed and
  ([.indicators[]] | length) == 12 and ([.claim_transitions[] | select(.persisted == true)] | length) >= 4 and
  ([.historical_references[] | select(.state == "SATISFIED")] | length) == 3 and
  ([.persistent_claims[] | select(.status == "REFUTED" and .evidence_digest != "" and .provenance != "" and .stage != "" and .step != "" and .reason != "")] | length) >= 1
' "$output_root/first/agreement-report.json"

jq -e '.decision == "FAIL_CLOSED" and .resolution == "EXACT" and .reference_agreement == "DISAGREES" and .enforcement_effect == "BLOCK" and .semantic_authority == "GOOO_SOURCE_INTENT" and .authority_grant == "NONE" and .decision != "PASS"' "$output_root/first/mismatch-report.json"
jq -e '.decision == "FAIL_CLOSED" and .resolution == "UNKNOWN" and .reference_agreement == "UNKNOWN" and .enforcement_effect == "BLOCK" and .semantic_authority == "GOOO_SOURCE_INTENT" and .authority_grant == "NONE" and .decision != "PASS"' "$output_root/first/absence-report.json"
jq -e '.decision == "FAIL_CLOSED" and .semantic_authority == "EXTERNAL_REFERENCE_AUTHORITY" and .authority_grant == "NONE" and ([.claim_transitions[] | select(.claim_id == "source-intent-authority" and .after == "REFUTED")] | length) == 1' "$output_root/first/intervention-report.json"
jq -e '.decision == "REFERENCE_AGREEMENT_OBSERVED" and .authority_grant == "NONE"' "$output_root/first/comment-report.json"
jq -e '.decision == "HUMILITY_MODEL_BOUND" and .resolution == "EXACT" and .cases_satisfied == 3 and .cases_total == 3 and .coverage_bps == 10000 and .semantic_causality.completed == 1 and .semantic_causality.total == 1 and .nonsemantic_preservation.completed == 1 and .nonsemantic_preservation.total == 1' "$output_root/first/suite.json"

test "$(jq -r '.producer_to_consumer' "$independence_path")" -eq 0
test "$(jq -r '.consumer_to_producer' "$independence_path")" -eq 0
test "$(git status --short)" = ""

agreement=$(jq -r '.decision + " / " + .resolution' "$output_root/first/agreement-report.json")
agreement_counts=$(jq -r '"\(.completed)/\(.total) (\(.basis_points) bps); current=\(.current_reference_observations)/\(.current_reference_total)"' "$output_root/first/agreement-report.json")
current_observations=$(jq -r '.current_reference_observations' "$output_root/first/agreement-report.json")
mismatch=$(jq -r '.decision + " / " + .resolution' "$output_root/first/mismatch-report.json")
absence=$(jq -r '.decision + " / " + .resolution' "$output_root/first/absence-report.json")
{
  echo '### External oracle humility'
  echo
  jq -r --arg agreement "$agreement" --arg agreement_counts "$agreement_counts" --arg mismatch "$mismatch" --arg absence "$absence" --arg current "$current_observations" \
    '"- subject agreement: `\($agreement)`\n- indicators: `\($agreement_counts)`\n- mismatch: `\($mismatch)`\n- absence: `\($absence)`\n- cases: `\(.cases_satisfied)/\(.cases_total)` (`\(.coverage_bps)` bps)\n- fixed denominator: source_policy=1/1; producer_imports=0/0; current_reference_observations=\($current)/3; historical_fixtures=3/3; semantic_causality=\(.semantic_causality.completed)/1; nonsemantic_preservation=\(.nonsemantic_preservation.completed)/1\n- authority grant: `NONE`; writes: `\(.repository_writes)`; promotions: `\(.promotion_count)`"' \
    "$output_root/first/suite.json"
} >> "$GITHUB_STEP_SUMMARY"
