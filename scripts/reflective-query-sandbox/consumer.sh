#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
source_path="examples/reflective-query-sandbox/main.gooo"
input="${1:?producer directory is required}"
output="$RUNNER_TEMP/reflective-query-sandbox/consumer"
mkdir -p "$output"
cd "$root"

import_evidence="$output/producer-imports-go-list.txt"
go list -deps ./scripts/reflective-query-sandbox/consumer > "$import_evidence"
producer_imports=$(grep -F '/internal/meta/reflectivequerysandbox' "$import_evidence" | wc -l | tr -d ' ' || true)
maximum_allowed=0
test "$producer_imports" -le "$maximum_allowed"

go run ./scripts/reflective-query-sandbox/consumer \
	-input "$input/observation.json" -source "$source_path" \
	-subject-sha "$HEAD_SHA" -producer-imports-evidence "$import_evidence" \
	-producer-imports-maximum "$maximum_allowed" -output "$output/receipt.json"
go run ./scripts/reflective-query-sandbox/consumer \
	-input "$input/observation.json" -source "$source_path" \
	-subject-sha "$HEAD_SHA" -producer-imports-evidence "$import_evidence" \
	-producer-imports-maximum "$maximum_allowed" -output "$output/replay.json"
cmp -s "$output/receipt.json" "$output/replay.json"

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/reflective-query-sandbox-receipt/v3" and
  .subject_sha == $sha and .decision == "PASS" and
  .resolution == "OBSERVATION_ONLY" and
  .reason == "OBSERVATION_BOUNDARY_CONFORMANT" and
  .coordinates.total == .contract.denominator and
  .coordinates.satisfied == .contract.satisfied_indicators and
  .coordinates.satisfied < .coordinates.total and
  .subject_resolution == "MIXED_EXACT_AND_LOWER_RESOLUTION" and
  .source_reconstruction.satisfied == .source_reconstruction.total and
  .producer_imports.satisfied == 1 and .producer_imports.total == 1 and
  .import_boundary.forbidden_imports_observed == 0 and .import_boundary.maximum_allowed == 0 and
  (.import_boundary.evidence_digest | startswith("sha256:")) and
  .effects.net_repository_changes == [] and
  .effects.repository_status_before == .effects.repository_status_after and
  .effects.mutation_authority == false and .effects.mutation_outcome == "REJECTED" and
  ([.attempts[] | select(.operation == "query" and .decision == "PASS" and .resolution == "EXACT")] | length) == .contract.safe_queries and
  ([.attempts[] | select(.operation == "mutate" and .decision == "DENIED" and .api_outcome == "REJECTED")] | length) == .contract.denied_mutations and
  ([.attempts[] | select(.decision == "UNKNOWN" and .resolution == "LOWER_RESOLUTION" and .stage == "UNKNOWN" and .step == "resolve-unknown-subject" and .reason == "UNKNOWN_TARGET")] | length) == .contract.unknown_targets and
  ([.claims[] | select(.evidence_attempt == "unknown.target" and .from == "OPEN" and .to == "OPEN" and .reason == "UNKNOWN_PRESERVED")] | length) == 1 and
  .promotion_credit_bps == 0 and
  (.not_claimed | length) == 5
' "$output/receipt.json" >/dev/null

{
	echo '## Reflective query sandbox consumer'
	echo
	echo '| Observation | Exact value |'
	echo '|---|---:|'
	jq -r '"| Safe exact queries | \(.contract.safe_queries) / \(.contract.reflective_queries) |", "| Denied mutation attempts | \(.contract.denied_mutations) |", "| Unknown targets preserved | \(.contract.unknown_targets) |", "| Indicators | \(.coordinates.satisfied) / \(.coordinates.total) |", "| Claim transitions | \(.contract.transition_count) |", "| Source reconstruction | \(.source_reconstruction.satisfied) / \(.source_reconstruction.total) |", "| Producer import boundary | \(.import_boundary.forbidden_imports_observed) <= \(.import_boundary.maximum_allowed); coordinate \(.producer_imports.satisfied)/\(.producer_imports.total) |", "| Net repository changes | \(.effects.net_repository_changes | length) |", "| Mutation authority | \(.effects.mutation_authority) |", "| Promotion credit | \(.promotion_credit_bps) bps |"' "$output/receipt.json"
	echo
	echo '### Mutation boundary evidence'
	echo
	jq -r '.attempts[] | select(.id == "mutation.attempt") | "| Field | API outcome | Decision | Before graph | Original after | Returned graph |", "|---|---|---|---|---|---|", "| source-declared | \(.api_outcome) | \(.decision)/\(.resolution) | \(.graph_digest_before) | \(.original_graph_digest_after) | \(.returned_graph_digest // "") |"' "$output/receipt.json"
	echo
	echo '### Claim predicates'
	echo
	echo '| Claim | Predicate | From→to | Stage/step/reason | Material digest |'
	echo '|---|---|---|---|---|'
	jq -r '.claims[] | select(.sequence % 2 == 0) | "| \(.claim_id) | \(.predicate_id) | \(.from)→\(.to) | \(.stage)/\(.step)/\(.reason) | \(.observed_material_digest // "") |"' "$output/receipt.json"
	echo
	echo 'The consumer is an independent verifier; the receipt grants no mutation or promotion authority.'
} >> "${GITHUB_STEP_SUMMARY:-$output/summary.md}"

echo 'reflective query sandbox consumer: PASS source reconstruction and dynamic receipt'
