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
  .schema == "gooo/reflective-query-sandbox-intervention/v1" and
  .base.decision == "PASS" and .base.resolution == "EXACT" and
  .semantic_intervention.raw_digest_changed and
  .semantic_intervention.semantic_digest_changed and
  .semantic_intervention.graph_digest_changed and
  .semantic_intervention.query_decision_before == "PASS" and
  .semantic_intervention.query_decision_after == "UNKNOWN" and
  .semantic_intervention.query_resolution_after == "LOWER_RESOLUTION" and
  .semantic_intervention.claim_state_before == "DISCHARGED" and
  .semantic_intervention.claim_state_after == "OPEN" and
  .nonsemantic_intervention.raw_digest_changed and
  (.nonsemantic_intervention.semantic_digest_changed | not) and
  (.nonsemantic_intervention.graph_digest_changed | not) and
  .nonsemantic_intervention.query_decision_before == "PASS" and
  .nonsemantic_intervention.query_decision_after == "PASS" and
  .nonsemantic_intervention.query_resolution_before == "EXACT" and
  .nonsemantic_intervention.query_resolution_after == "EXACT" and
  .nonsemantic_intervention.claim_state_before == "DISCHARGED" and
  .nonsemantic_intervention.claim_state_after == "DISCHARGED"
' "$output" >/dev/null

echo 'reflective query sandbox interventions: PASS semantic relation changed; comment preserved semantic query'
