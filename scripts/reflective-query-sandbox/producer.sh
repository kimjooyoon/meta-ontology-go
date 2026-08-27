#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
source_path="examples/reflective-query-sandbox/main.gooo"
output="$RUNNER_TEMP/reflective-query-sandbox/producer"
mkdir -p "$output"
cd "$root"

checkout_sha=$(git rev-parse HEAD)
test "$HEAD_SHA" = "$checkout_sha"

before_status="$output/repository-before.txt"
after_status="$output/repository-after.txt"
subject_evidence="$output/subject-checkout.txt"
printf 'subject_sha=%s\ncheckout_head=%s\nsubject_matches_checkout=%s\n' "$HEAD_SHA" "$checkout_sha" true > "$subject_evidence"
git status --porcelain=v1 --untracked-files=all > "$before_status"
go run ./cmd/gooo check "$source_path"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "$HEAD_SHA" -subject-checkout-evidence "$subject_evidence" \
	-repository-before "$before_status" -repository-after "$before_status" -output "$output/warmup.json"
git status --porcelain=v1 --untracked-files=all > "$after_status"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "$HEAD_SHA" -subject-checkout-evidence "$subject_evidence" \
	-repository-before "$before_status" -repository-after "$after_status" -output "$output/observation.json" \
	-proposal-dir "$output/proposals" -proposal-receipt "$output/proposal-receipt.json"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "$HEAD_SHA" -subject-checkout-evidence "$subject_evidence" \
	-repository-before "$before_status" -repository-after "$after_status" -output "$output/replay.json"
if ! cmp -s "$output/observation.json" "$output/replay.json"; then
	echo 'producer replay mismatch:' >&2
	diff -u "$output/observation.json" "$output/replay.json" >&2 || true
	exit 1
fi
after=$(git status --porcelain=v1 --untracked-files=all)
if [ "$(cat "$before_status")" != "$after" ]; then
	echo 'repository write-set changed during producer run:' >&2
	echo 'before:' >&2
	cat "$before_status" >&2
	echo 'after:' >&2
	echo "$after" >&2
	exit 1
fi

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/reflective-query-sandbox-observation/v4" and
  .subject_sha == $sha and
  .source.path == "examples/reflective-query-sandbox/main.gooo" and
  .subject_binding.format.decision == "PASS" and
  .subject_binding.format.resolution == "EXACT" and
  .subject_binding.format.reason == "FORMAT_VALID" and
  .subject_binding.checkout.decision == "PASS" and
  .subject_binding.checkout.resolution == "EXACT" and
  .subject_binding.checkout.reason == "CHECKOUT_BOUND" and
  .subject_binding.checkout.observed_sha == $sha and
  (.subject_binding.checkout.evidence_digest | length) > 0 and
  .provisional == true and
  (.digest | length) == 0 and
  (.provisional_digest | startswith("sha256:")) and
  (.receipt_material_digest | length) == 0 and
  .contract.source_nodes == .source.node_count and
  .contract.source_facts == .source.fact_count and
  .contract.claim_count == ((.claims | length) / 2) and
  .contract.denominator == .contract.claim_count and
  .contract.attempt_count == (.attempts | length) and
  .contract.reflective_queries == ([.attempts[] | select(.operation == "query")] | length) and
  .contract.safe_queries == ([.attempts[] | select(.operation == "query" and .decision == "PASS" and .resolution == "EXACT")] | length) and
  .contract.denied_mutations == ([.attempts[] | select(.operation == "mutate" and .decision == "DENIED")] | length) and
  .contract.unknown_targets == ([.attempts[] | select(.decision == "UNKNOWN")] | length) and
  .contract.refuted_attempts == ([.attempts[] | select(.decision == "REFUTED")] | length) and
  .contract.transition_count == (.claims | length) and
  .contract.satisfied_indicators == ([.claims[] | select(.to == "DISCHARGED" and .from != .to)] | length) and
  .contract.satisfied_indicators == 10 and
  ([.claims[] | select(.predicate_id == "receipt-observation-digest-verified" and .to == "OPEN" and (.observed_material_digest | length) == 0)] | length) == 1 and
  ([.attempts[] | select(.id == "mutation.attempt" and .decision == "DENIED" and .resolution == "EXACT_REJECTION" and .api_outcome == "REJECTED" and .api_error_code == "immutable_field" and .reason == "IMMUTABLE_ID_PATCH_REJECTED" and .mutation_field == "id" and .mutation_payload == "identity-preserving" and .graph_digest_before == .original_graph_digest_after and .semantic_digest_before == .original_semantic_digest_after and (.returned_graph_digest | length) == 0)] | length) == 1 and
  ([.attempts[] | select(.id == "unknown.target" and .decision == "UNKNOWN" and .resolution == "LOWER_RESOLUTION" and .reason == "UNKNOWN_TARGET" and .stage == "UNKNOWN" and .step == "resolve-unknown-subject")] | length) == 1 and
  .effects.repository_status_before == .effects.repository_status_after and
  .effects.net_repository_changes == [] and
  .effects.repository_evidence_available == true and
  .effects.repository_observation == "net_repository_status_unchanged" and
  .effects.immutable_id_patch_accepted == false and
  .effects.detached_graph_patch_capability == "UNKNOWN" and
  .effects.overall_authority == "UNKNOWN" and
  .effects.mutation_outcome == "REJECTED" and
  ([.claims[] | select(.predicate_id == "claim-ledger-chained" and .to == "DISCHARGED" and .reason == "COMPLETE_TRANSITION_CHAIN_VERIFIED")] | length) == 1
' "$output/observation.json" >/dev/null

jq -e '
  .schema == "gooo/reflective-query-sandbox/proposal-receipt/v1" and
  .source_path == "examples/reflective-query-sandbox/main.gooo" and
  (.source_raw_digest | length) == 64 and
  (.source_semantic_digest | length) == 64 and
  (.query_receipt_digest | length) == 64 and
  .authority == "NONE" and .repository_writes == 0 and .mutation_authority == false and
  .proposal_count == 3 and .emitted.satisfied == 1 and .emitted.total == 3 and
  .rejected.satisfied == 1 and .rejected.total == 3 and
  ([.cases[] | select(.case_id == "exact-observation" and .outcome == "NOT_EMITTED" and .proposal_emitted == false and .proposal_coordinate.satisfied == 0 and .proposal_coordinate.total == 1 and .observation_kind == "EXACT")] | length) == 1 and
  ([.cases[] | select(.case_id == "unknown-observation" and .outcome == "EMITTED" and .proposal_emitted == true and .proposal_coordinate.satisfied == 1 and .proposal_coordinate.total == 1 and .unknown_stage == "UNKNOWN" and .unknown_step == "resolve-unknown-subject" and .unknown_reason == "UNKNOWN_TARGET" and .observed_claim_from == "OPEN" and .observed_claim_to == "OPEN")] | length) == 1 and
  ([.cases[] | select(.case_id == "mutation-request" and .outcome == "REJECTED" and .rejection_coordinate.satisfied == 1 and .rejection_coordinate.total == 1 and .proposal_emitted == false and .observation_kind == "MUTATION_REQUEST")] | length) == 1 and
  ([.cases[] | select((.path | startswith("proposals/")) and (.bytes > 0) and (.bytes_digest | length) == 64 and (.semantic_digest | length) == 64 and (.claim_id | length) > 0 and (.target_semantic_address | length) > 0 and (.requested_refinement | length) > 0 and (.proof_choice | length) > 0 and (.meta_operation | length) > 0 and .authority == "NONE")] | length) == 3 and
  (.portable_receipt_digest | length) == 64
' "$output/proposal-receipt.json" >/dev/null

regressions="$output/regressions"
mkdir -p "$regressions"
wrong_sha=$(printf '0%.0s' {1..40})
if [ "$wrong_sha" = "$HEAD_SHA" ]; then
	wrong_sha=$(printf 'f%.0s' {1..40})
fi
wrong_evidence="$regressions/wrong-sha-checkout.txt"
printf 'subject_sha=%s\ncheckout_head=%s\nsubject_matches_checkout=%s\n' "$wrong_sha" "$checkout_sha" false > "$wrong_evidence"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "" -repository-before "$before_status" -repository-after "$before_status" -output "$regressions/invalid-unobserved.json"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "$wrong_sha" -subject-checkout-evidence "$wrong_evidence" \
	-repository-before "$before_status" -repository-after "$before_status" -output "$regressions/exact-mismatch.json"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "$wrong_sha" -subject-checkout-evidence "$wrong_evidence" \
	-repository-before "" -repository-after "" -output "$regressions/mismatch-repository-unknown.json"

jq -e '
  .subject_binding.format.decision == "UNKNOWN" and
  .subject_binding.format.resolution == "LOWER_RESOLUTION" and
  .subject_binding.format.reason == "FORMAT_INVALID" and
  .subject_binding.checkout.decision == "UNKNOWN" and
  .subject_binding.checkout.reason == "SUBJECT_SHA_CHECKOUT_UNOBSERVED"
' "$regressions/invalid-unobserved.json" >/dev/null
jq -e --arg sha "$wrong_sha" --arg checkout "$checkout_sha" '
  .subject_binding.format.decision == "PASS" and
  .subject_binding.format.reason == "FORMAT_VALID" and
  .subject_binding.checkout.decision == "REFUTED" and
  .subject_binding.checkout.resolution == "EXACT" and
  .subject_binding.checkout.reason == "SUBJECT_SHA_CHECKOUT_MISMATCH" and
  .subject_binding.checkout.observed_sha == $checkout and
  .subject_sha == $sha
' "$regressions/exact-mismatch.json" >/dev/null
jq -e '
  .effects.repository_evidence_available == false and
  .effects.repository_observation == "UNOBSERVED" and
  .effects.repository_observation_stage == "REPOSITORY" and
  .effects.repository_observation_step == "read-status" and
  (.effects.repository_observation_reason | startswith("REPOSITORY_EVIDENCE_")) and
  .effects.repository_status_before == null and
  .effects.repository_status_after == null and
  .effects.net_repository_changes == null and
  ([.attempts[] | select(.id == "repository.net-status-unchanged" and .decision == "UNKNOWN" and .resolution == "LOWER_RESOLUTION" and .stage == "REPOSITORY" and .step == "read-status")] | length) == 1
' "$regressions/mismatch-repository-unknown.json" >/dev/null

if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
	{
		echo '| Fixed regression matrix | Passed / fixed denominator |'
		echo '|---|---:|'
		echo '| Invalid/unobserved / exact mismatch / mismatch + repository unknown | 3 / 3 |'
	} >> "$GITHUB_STEP_SUMMARY"
fi

if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
	{
		echo '### Correction audit'
		echo
		echo '| Scoped correction gates | Local tests |'
		echo '|---:|---:|'
		echo '| 11 / 11 | 0 |'
		echo '| Subject binding | FORMAT_VALID + CHECKOUT_BOUND |'
	} >> "$GITHUB_STEP_SUMMARY"
fi

echo 'reflective query sandbox producer: PASS source-derived contract and boundary observations'
