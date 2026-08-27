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
	-subject-sha "$HEAD_SHA" -checkout-sha "$checkout_sha" -producer-imports-evidence "$import_evidence" \
	-producer-imports-maximum "$maximum_allowed" -output "$output/receipt.json"
go run ./scripts/reflective-query-sandbox/consumer \
	-input "$input/observation.json" -source "$source_path" \
	-subject-sha "$HEAD_SHA" -checkout-sha "$checkout_sha" -producer-imports-evidence "$import_evidence" \
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
  .subject_binding.format.decision == "PASS" and .subject_binding.format.resolution == "EXACT" and
  .subject_binding.format.reason == "FORMAT_VALID" and
  .subject_binding.checkout.decision == "PASS" and .subject_binding.checkout.resolution == "EXACT" and
  .subject_binding.checkout.reason == "CHECKOUT_BOUND" and .subject_binding.checkout.observed_sha == $sha and
  (.subject_binding.checkout.evidence_digest | length) > 0 and
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

wrong_sha=$(printf '0%.0s' {1..40})
if [ "$wrong_sha" = "$HEAD_SHA" ]; then
	wrong_sha=$(printf 'f%.0s' {1..40})
fi
go run ./scripts/reflective-query-sandbox/consumer \
	-input "$input/regressions/invalid-unobserved.json" -source "$source_path" \
	-subject-sha "" -checkout-sha "" -producer-imports-evidence "$import_evidence" \
	-producer-imports-maximum "$maximum_allowed" -output "$output/invalid-unobserved-receipt.json"
go run ./scripts/reflective-query-sandbox/consumer \
	-input "$input/regressions/exact-mismatch.json" -source "$source_path" \
	-subject-sha "$wrong_sha" -checkout-sha "$checkout_sha" -producer-imports-evidence "$import_evidence" \
	-producer-imports-maximum "$maximum_allowed" -output "$output/exact-mismatch-receipt.json"
jq -e '
  .decision == "UNKNOWN" and .resolution == "LOWER_RESOLUTION" and
  .subject_resolution == "UNKNOWN_SUBJECT_SHA" and
  .subject_binding.format.decision == "UNKNOWN" and
  .subject_binding.format.reason == "FORMAT_INVALID"
' "$output/invalid-unobserved-receipt.json" >/dev/null
jq -e '
  .decision == "REFUTED" and .resolution == "EXACT" and
  .subject_resolution == "REFUTED_SUBJECT_SHA_MISMATCH" and
  .subject_binding.format.decision == "PASS" and
  .subject_binding.checkout.decision == "REFUTED" and
  .subject_binding.checkout.reason == "SUBJECT_SHA_CHECKOUT_MISMATCH"
' "$output/exact-mismatch-receipt.json" >/dev/null
go run ./scripts/reflective-query-sandbox/consumer \
	-input "$input/regressions/mismatch-repository-unknown.json" -source "$source_path" \
	-subject-sha "$wrong_sha" -checkout-sha "$checkout_sha" -producer-imports-evidence "$import_evidence" \
	-producer-imports-maximum "$maximum_allowed" -output "$output/mismatch-repository-unknown-receipt.json"
jq -e '
  .decision == "REFUTED" and .resolution == "EXACT" and
  .subject_resolution == "REFUTED_SUBJECT_SHA_MISMATCH" and
  .subject_binding.checkout.decision == "REFUTED" and
  .subject_binding.checkout.reason == "SUBJECT_SHA_CHECKOUT_MISMATCH" and
  .effects.repository_evidence_available == false and
  .effects.repository_observation == "UNOBSERVED" and
  ([.claims[] | select(.predicate_id == "net-repository-status-unchanged" and .to == "OPEN" and .reason == "REPOSITORY_EVIDENCE_MISSING")] | length) == 1
' "$output/mismatch-repository-unknown-receipt.json" >/dev/null

for proposal in "$input"/proposals/*.gooo; do
	go run ./cmd/gooo check "$proposal"
done
proposal_import_evidence="$output/proposal-consumer-imports-go-list.txt"
go list -deps ./scripts/reflective-query-sandbox/proposal-consumer > "$proposal_import_evidence"
proposal_producer_imports=$(grep -F '/scripts/reflective-query-sandbox/producer' "$proposal_import_evidence" | wc -l | tr -d ' ' || true)
test "$proposal_producer_imports" -eq 0
go run ./scripts/reflective-query-sandbox/proposal-consumer \
	-input "$input" -source "$source_path" -output "$output/proposal-receipt.json" \
	-producer-imports-evidence "$proposal_import_evidence" -producer-imports-maximum 0
jq -e '
  .schema == "gooo/reflective-query-sandbox/proposal-verification/v1" and
  .decision == "PROPOSAL_ONLY" and .resolution == "EXACT_RECONSTRUCTION" and
  .reason == "PROPOSAL_EMITTED_WITHOUT_PROMOTION" and .authority == "NONE" and
  .repository_writes == 0 and .mutation_authority == false and
  (.source_raw_digest | length) == 64 and (.source_semantic_digest | length) == 64 and
  (.query_receipt_digest | length) == 64 and
  .emitted.satisfied == 1 and .emitted.total == 3 and
  .rejected.satisfied == 1 and .rejected.total == 3 and
  .generated_artifact_count.satisfied == 3 and .generated_artifact_count.total == 3 and
  .generated_reconsumption.satisfied == 1 and .generated_reconsumption.total == 1 and
  .open_claim_transition.satisfied == 1 and .open_claim_transition.total == 1 and
  ([.cases[] | select(.case_id == "exact-observation" and .outcome == "NOT_EMITTED" and .proposal_emitted == false and .proposal_coordinate.satisfied == 0 and .proposal_coordinate.total == 1)] | length) == 1 and
  ([.cases[] | select(.case_id == "unknown-observation" and .outcome == "EMITTED" and .proposal_emitted == true and .proposal_coordinate.satisfied == 1 and .proposal_coordinate.total == 1 and .claim_transition == "OPEN->OPEN" and .proposal_transition == "OPEN->OPEN")] | length) == 1 and
  ([.cases[] | select(.case_id == "mutation-request" and .outcome == "REJECTED" and .rejection_coordinate.satisfied == 1 and .rejection_coordinate.total == 1 and .proposal_emitted == false)] | length) == 1 and
  .counterexample_denominator == 5 and
  ([.counterexamples[] | select(.decision == "REFUTED" and .resolution == "EXACT" and (.stage | length) > 0 and (.step | length) > 0 and (.reason | length) > 0)] | length) == 5 and
  .directional_import_boundary.forbidden_imports_observed == 0 and
  .directional_import_boundary.maximum_allowed == 0 and
  .directional_import_boundary.conformance.satisfied == 1 and
  .directional_import_boundary.conformance.total == 1 and
  (.directional_import_boundary.evidence_digest | length) == 64
' "$output/proposal-receipt.json" >/dev/null

{
	echo '## Reflective query sandbox consumer'
	echo
	echo '| Observation | Exact value |'
	echo '|---|---:|'
    jq -r '"| Safe exact queries | \(.contract.safe_queries) / \(.contract.reflective_queries) |", "| Denied immutable-id patches | \(.contract.denied_mutations) |", "| Unknown targets preserved | \(.contract.unknown_targets) |", "| Indicators | \(.coordinates.satisfied) / \(.coordinates.total) |", "| Claim transitions | \(.contract.transition_count) |", "| Source reconstruction | \(.source_reconstruction.satisfied) / \(.source_reconstruction.total) |", "| Producer import boundary | \(.import_boundary.forbidden_imports_observed) <= \(.import_boundary.maximum_allowed); coordinate \(.producer_imports.satisfied)/\(.producer_imports.total) |", "| Net repository status | \(.effects.net_repository_changes | length) changes; observation=\(.effects.repository_observation) |", "| Detached graph patch capability | \(.detached_graph_patch_capability) |", "| Overall mutation authority | \(.overall_authority) |", "| Receipt attestation | \(.attestor); material=\(.receipt_material_digest); chain=\(.transition_chain_digest) |", "| Promotion credit | \(.promotion_credit_bps) bps |"' "$output/receipt.json"
	jq -r '"| Subject binding | format=\(.subject_binding.format.reason); checkout=\(.subject_binding.checkout.reason); evidence=\(.subject_binding.checkout.evidence_digest) |"' "$output/receipt.json"
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
	echo '| Fixed regression matrix | 3 / 3 |'
	echo '| Refinement proposal cases (exact / unknown / mutation) | 0/1, 1/1, 1/1 |'
	echo '| Generated proposal re-consumption | 1 / 1 |'
	echo '| Proposal counterexamples | 5 / 5 |'
	echo '| Directional producer import boundary | observed=0; allowed=0; conformance=1/1 |'
	echo '| Local tests | 0 |'
} >> "${GITHUB_STEP_SUMMARY:-$output/summary.md}"

echo 'reflective query sandbox consumer: PASS source reconstruction and dynamic receipt'
