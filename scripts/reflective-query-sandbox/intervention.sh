#!/usr/bin/env bash
set -euo pipefail

: "${RUNNER_TEMP:?RUNNER_TEMP is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
source_path="examples/reflective-query-sandbox/main.gooo"
output="$RUNNER_TEMP/reflective-query-sandbox/intervention.json"
mkdir -p "$(dirname "$output")"
cd "$root"

go run ./scripts/reflective-query-sandbox/intervention -source "$source_path" -output "$output"

jq -e '
  .schema == "gooo/reflective-query-sandbox-intervention/v2" and
  .base.mutation_decision == "DENIED" and .base.mutation_resolution == "EXACT_REJECTION" and
  .base.mutation_api_outcome == "REJECTED" and (.base.mutation_authority | not) and
  .semantic_intervention.raw_digest_changed and
  .semantic_intervention.semantic_digest_changed and
  .semantic_intervention.graph_digest_changed and
  .semantic_intervention.mutation_field_before == "id" and
  .semantic_intervention.mutation_field_after == "name" and
  .semantic_intervention.mutation_decision_before == "DENIED" and
  .semantic_intervention.mutation_decision_after == "REFUTED" and
  .semantic_intervention.mutation_api_outcome_before == "REJECTED" and
  .semantic_intervention.mutation_api_outcome_after == "ACCEPTED" and
  (.semantic_intervention.mutation_authority_before | not) and
  .semantic_intervention.mutation_authority_after and
  .semantic_intervention.graph_digest_before == .semantic_intervention.original_graph_digest_after and
  (.semantic_intervention.returned_graph_digest_after | length) > 0 and
  .semantic_intervention.claim_state_before == "DISCHARGED" and
  .semantic_intervention.claim_state_after == "REFUTED" and
  .nonsemantic_intervention.raw_digest_changed and
  (.nonsemantic_intervention.semantic_digest_changed | not) and
  (.nonsemantic_intervention.graph_digest_changed | not) and
  .nonsemantic_intervention.mutation_decision_before == "DENIED" and
  .nonsemantic_intervention.mutation_decision_after == "DENIED" and
  .nonsemantic_intervention.mutation_api_outcome_before == "REJECTED" and
  .nonsemantic_intervention.mutation_api_outcome_after == "REJECTED" and
  (.nonsemantic_intervention.mutation_authority_before | not) and
  (.nonsemantic_intervention.mutation_authority_after | not) and
  .nonsemantic_intervention.claim_state_before == "DISCHARGED" and
  .nonsemantic_intervention.claim_state_after == "DISCHARGED"
' "$output" >/dev/null

echo 'reflective query sandbox interventions: PASS semantic relation changed; comment preserved semantic query'
