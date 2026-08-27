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

producer_imports=$(go list -deps ./scripts/reflective-query-sandbox/consumer | { grep -F '/internal/meta/reflectivequerysandbox' || true; } | wc -l | tr -d ' ')
test "$producer_imports" = 0

go run ./scripts/reflective-query-sandbox/consumer \
	-input "$input/observation.json" -source "$source_path" \
	-subject-sha "$HEAD_SHA" -output "$output/receipt.json"
go run ./scripts/reflective-query-sandbox/consumer \
	-input "$input/observation.json" -source "$source_path" \
	-subject-sha "$HEAD_SHA" -output "$output/replay.json"
cmp -s "$output/receipt.json" "$output/replay.json"

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/reflective-query-sandbox-receipt/v2" and
  .subject_sha == $sha and .decision == "PASS" and
  .resolution == "OBSERVATION_ONLY" and
  .reason == "OBSERVATION_BOUNDARY_CONFORMANT" and
  .coordinates.total == .contract.denominator and
  .coordinates.satisfied == .contract.satisfied_indicators and
  .coordinates.satisfied < .coordinates.total and
  .subject_resolution == "MIXED_EXACT_AND_LOWER_RESOLUTION" and
  .source_reconstruction.satisfied == .source_reconstruction.total and
  .producer_imports.satisfied == 0 and .producer_imports.total == 0 and
  .effects.repository_writes == 0 and .effects.repository_write_set == [] and
  .effects.repository_before == .effects.repository_after and
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
	jq -r '"| Safe exact queries | \(.contract.safe_queries) / \(.contract.reflective_queries) |", "| Denied mutation attempts | \(.contract.denied_mutations) |", "| Unknown targets preserved | \(.contract.unknown_targets) |", "| Indicators | \(.coordinates.satisfied) / \(.coordinates.total) |", "| Claim transitions | \(.contract.transition_count) |", "| Source reconstruction | \(.source_reconstruction.satisfied) / \(.source_reconstruction.total) |", "| Producer imports | \(.producer_imports.satisfied) / \(.producer_imports.total) |", "| Repository writes | \(.effects.repository_writes) |", "| Mutation authority | \(.effects.mutation_authority) |", "| Promotion credit | \(.promotion_credit_bps) bps |"' "$output/receipt.json"
	echo
	echo 'The consumer is an independent verifier; the receipt grants no mutation or promotion authority.'
} >> "${GITHUB_STEP_SUMMARY:-$output/summary.md}"

echo 'reflective query sandbox consumer: PASS source reconstruction and dynamic receipt'
