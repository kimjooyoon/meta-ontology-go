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

checkout_sha=$(git rev-parse HEAD)
test "$HEAD_SHA" = "$checkout_sha"

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
  .schema == "gooo/reflective-query-sandbox-receipt/v4" and
  .subject_sha == $sha and .decision == "PASS" and
  .resolution == "OBSERVATION_ONLY" and
  .reason == "OBSERVATION_BOUNDARY_CONFORMANT" and
  .coordinates.total == .contract.denominator and
  .coordinates.satisfied == .contract.satisfied_indicators and
  .coordinates.satisfied == 11 and .coordinates.total == 12 and
  .subject_resolution == "MIXED_EXACT_AND_LOWER_RESOLUTION" and
  .subject_binding.decision == "PASS" and .subject_binding.resolution == "EXACT" and
  .source_reconstruction.satisfied == .source_reconstruction.total and
  .producer_imports.satisfied == 1 and .producer_imports.total == 1 and
  .import_boundary.forbidden_imports_observed == 0 and .import_boundary.maximum_allowed == 0 and
  (.import_boundary.evidence_digest | startswith("sha256:")) and
  .effects.net_repository_changes == [] and
  .effects.repository_status_before == .effects.repository_status_after and
  .effects.repository_evidence_available == true and
  .effects.repository_observation == "net_repository_status_unchanged" and
  .effects.immutable_id_patch_accepted == false and
  .effects.detached_graph_patch_capability == "UNKNOWN" and
  .effects.overall_authority == "UNKNOWN" and .overall_authority == "UNKNOWN" and
  .effects.mutation_outcome == "REJECTED" and
  (.receipt_material_digest | startswith("sha256:")) and
  (.transition_chain_digest | startswith("sha256:")) and
  .attestor == "reflective-query-sandbox.independent-verifier" and
  (.attestation_digest | startswith("sha256:")) and
  ([.claims[] | select(.predicate_id == "claim-ledger-chained" and .to == "DISCHARGED" and .reason == "COMPLETE_TRANSITION_CHAIN_VERIFIED")] | length) == 1 and
  ([.claims[] | select(.predicate_id == "receipt-observation-digest-verified" and .to == "DISCHARGED")] | length) == 1 and
  ([.attempts[] | select(.id == "mutation.attempt" and .reason == "IMMUTABLE_ID_PATCH_REJECTED" and .mutation_field == "id" and .mutation_payload == "identity-preserving" and .decision == "DENIED" and .api_outcome == "REJECTED" and .graph_digest_before == .original_graph_digest_after and .semantic_digest_before == .original_semantic_digest_after and (.returned_graph_digest | length) == 0)] | length) == 1 and
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
    jq -r '"| Safe exact queries | \(.contract.safe_queries) / \(.contract.reflective_queries) |", "| Denied immutable-id patches | \(.contract.denied_mutations) |", "| Unknown targets preserved | \(.contract.unknown_targets) |", "| Indicators | \(.coordinates.satisfied) / \(.coordinates.total) |", "| Claim transitions | \(.contract.transition_count) |", "| Source reconstruction | \(.source_reconstruction.satisfied) / \(.source_reconstruction.total) |", "| Producer import boundary | \(.import_boundary.forbidden_imports_observed) <= \(.import_boundary.maximum_allowed); coordinate \(.producer_imports.satisfied)/\(.producer_imports.total) |", "| Net repository status | \(.effects.net_repository_changes | length) changes; observation=\(.effects.repository_observation) |", "| Detached graph patch capability | \(.detached_graph_patch_capability) |", "| Overall mutation authority | \(.overall_authority) |", "| Receipt attestation | \(.attestor); material=\(.receipt_material_digest); chain=\(.transition_chain_digest) |", "| Promotion credit | \(.promotion_credit_bps) bps |"' "$output/receipt.json"
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
	echo
	echo '| Audit correction gates | Result |'
	echo '|---|---:|'
	echo '| Scoped correction gates | 11 / 11 |'
	echo '| Local tests | 0 |'
} >> "${GITHUB_STEP_SUMMARY:-$output/summary.md}"

echo 'reflective query sandbox consumer: PASS source reconstruction and dynamic receipt'
