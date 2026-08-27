#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
source_path="examples/reflective-query-sandbox/main.gooo"
output="$RUNNER_TEMP/reflective-query-sandbox/producer"
mkdir -p "$output"
cd "$root"

before_status="$output/repository-before.txt"
after_status="$output/repository-after.txt"
git status --porcelain=v1 --untracked-files=all > "$before_status"
go run ./cmd/gooo check "$source_path"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "$HEAD_SHA" -repository-before "$before_status" -repository-after "$before_status" -output "$output/warmup.json"
git status --porcelain=v1 --untracked-files=all > "$after_status"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "$HEAD_SHA" -repository-before "$before_status" -repository-after "$after_status" -output "$output/observation.json"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "$HEAD_SHA" -repository-before "$before_status" -repository-after "$after_status" -output "$output/replay.json"
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

echo 'producer validation snapshot:'
jq '{contract: {source_nodes, source_facts, claim_count, denominator, attempt_count, reflective_queries, safe_queries, denied_mutations, unknown_targets, refuted_attempts, transition_count, satisfied_indicators}, attempts: [.attempts[] | {id, operation, decision, resolution, reason, api_outcome}], effects, claims: [.claims[] | {claim_id, from, to, reason, evidence_attempt}]}' "$output/observation.json"
jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/reflective-query-sandbox-observation/v2" and
  .subject_sha == $sha and
  .source.path == "examples/reflective-query-sandbox/main.gooo" and
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
  ([.attempts[] | select(.id == "mutation.attempt" and .decision == "DENIED" and .api_outcome == "REJECTED" and .reason == "MUTATION_REQUEST_REJECTED")] | length) == 1 and
  ([.attempts[] | select(.id == "unknown.target" and .decision == "UNKNOWN" and .resolution == "LOWER_RESOLUTION" and .reason == "UNKNOWN_TARGET" and .stage == "UNKNOWN" and .step == "resolve-unknown-subject")] | length) == 1 and
  .effects.repository_writes == (.effects.repository_write_set | length) and
  .effects.repository_before == .effects.repository_after and
  .effects.repository_write_set == [] and
  .effects.mutation_authority == false
' "$output/observation.json" >/dev/null

echo 'reflective query sandbox producer: PASS source-derived contract and boundary observations'
