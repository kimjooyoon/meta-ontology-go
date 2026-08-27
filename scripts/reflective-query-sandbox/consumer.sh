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

go run ./scripts/reflective-query-sandbox/consumer \
	-input "$input/observation.json" -source "$source_path" \
	-subject-sha "$HEAD_SHA" -output "$output/receipt.json"
go run ./scripts/reflective-query-sandbox/consumer \
	-input "$input/observation.json" -source "$source_path" \
	-subject-sha "$HEAD_SHA" -output "$output/replay.json"
cmp -s "$output/receipt.json" "$output/replay.json"

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/reflective-query-sandbox-receipt/v1" and
  .subject_sha == $sha and .decision == "PASS" and
  .resolution == "OBSERVATION_ONLY" and
  .reason == "READ_ONLY_REFLECTION_BOUNDARY_PROVED" and
  .coordinates == {satisfied:12,total:12,basis_points:10000} and
  .classes == [{name:"OUTCOME",satisfied:4,total:4},{name:"DRIVER",satisfied:4,total:4},{name:"GUARDRAIL",satisfied:4,total:4}] and
  .proofs == [{name:"FOUNDATION",satisfied:4,total:4},{name:"COHERENCE",satisfied:4,total:4},{name:"REGRESSION",satisfied:4,total:4}] and
  .effects == {repository_writes:0,mutation_authority:false} and
  .repository_writes == 0 and .mutation_authority == false and
  .promotion_credit_bps == 0 and
  (.not_claimed | length) == 5
' "$output/receipt.json" >/dev/null

{
	echo '## Reflective query sandbox consumer'
	echo
	echo '| Observation | Exact value |'
	echo '|---|---:|'
	jq -r '"| Safe exact queries | \([.attempts[] | select(.decision == "PASS")] | length) / 3 |", "| Denied mutation attempts | \([.attempts[] | select(.decision == "DENIED")] | length) / 1 |", "| Unknown targets preserved | \([.attempts[] | select(.decision == "UNKNOWN")] | length) / 1 |", "| Fixed indicators | \(.coordinates.satisfied) / \(.coordinates.total) |", "| Claim transitions | \(.claims | length) / 24 |", "| Repository writes | \(.effects.repository_writes) |", "| Mutation authority | \(.effects.mutation_authority) |", "| Promotion credit | \(.promotion_credit_bps) bps |"' "$output/receipt.json"
	echo
	echo 'The consumer is an independent verifier; the receipt grants no mutation or promotion authority.'
} >> "${GITHUB_STEP_SUMMARY:-$output/summary.md}"

echo 'reflective query sandbox consumer: PASS 12/12 indicators'
