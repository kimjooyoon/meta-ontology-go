#!/usr/bin/env bash
set -euo pipefail

: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
: "${HEAD_SHA:?HEAD_SHA is required}"

output_dir="$RUNNER_TEMP/symbolic-invocation-schema"
artifact="$output_dir/artifact.json"
contract="$output_dir/symbolic-value-contract.json"
manifest="$output_dir/symbolic-value-contract-manifest.sha256"

test -f "$artifact"
go run ./scripts/symbolic-invocation-schema/valuecontract "$artifact" "$contract" "$HEAD_SHA"

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/symbolic-invocation-value-contract/v1"
  and .subject_sha == $sha
  and .metric_id == "gooo.metric.compiler.symbolic-value-contract.v1"
  and .decision == "PASS"
  and .resolution == "VALUE_CONTRACT_ONLY"
  and .reason == "SYMBOLIC_VALUE_SEMANTICS_COMPILED"
  and .coordinates.satisfied == 8
  and .coordinates.total == 8
  and .coordinates.basis_points == 10000
  and .classes == [
    {"class":"OUTCOME","satisfied":2,"total":2},
    {"class":"DRIVER","satisfied":3,"total":3},
    {"class":"GUARDRAIL","satisfied":3,"total":3}
  ]
  and .views == [
    {"audience":"USER","resolution":"USER_VISIBLE","satisfied":2,"total":2,"basis_points":10000},
    {"audience":"TOOL_AUTHOR","resolution":"TOOL_CONTRACT","satisfied":6,"total":6,"basis_points":10000},
    {"audience":"GOVERNOR","resolution":"FULL_RECEIPT","satisfied":8,"total":8,"basis_points":10000}
  ]
  and .proofs == [
    {"proof_choice":"FOUNDATION","satisfied":4,"total":4},
    {"proof_choice":"COHERENCE","satisfied":2,"total":2},
    {"proof_choice":"REGRESSION","satisfied":2,"total":2}
  ]
  and ([.rules[]
    | select(
        .id == "complete-symbolic-invocation"
        and .match.activity == "NON_EMPTY"
        and .match.inputs == "NON_EMPTY"
        and .decision == "READY"
        and .resolution == "VALUE_EXACT"
      )] | length) == 1
  and ([.rules[]
    | select(
        .id == "missing-activity"
        and .match.activity == "MISSING_OR_EMPTY"
        and .match.inputs == "ANY"
        and .decision == "FAIL_CLOSED"
        and .resolution == "LOWER_RESOLUTION"
      )] | length) == 1
  and .default.decision == "FAIL_CLOSED"
  and .default.resolution == "LOWER_RESOLUTION"
  and .effects.repository_writes == 0
  and .effects.mutation_authority == false
  and .promotion_credit_bps == 0
' "$contract" >/dev/null

artifact_digest=$(jq -er '.digest' "$artifact")
contract_artifact_digest=$(jq -er '.source_artifact_digest' "$contract")
test "$artifact_digest" = "$contract_artifact_digest"

expected_digest=$(jq -er '.digest' "$contract")
canonical=$(jq -S -c 'del(.digest)' "$contract")
observed_digest="sha256:$(printf '%s' "$canonical" | sha256sum | awk '{print $1}')"
test "$expected_digest" = "$observed_digest"

(
  cd "$output_dir"
  sha256sum symbolic-value-contract.json > symbolic-value-contract-manifest.sha256
  test "$(wc -l < symbolic-value-contract-manifest.sha256)" -eq 1
  sha256sum -c symbolic-value-contract-manifest.sha256
)

printf 'compiler symbolic value contract manifest: PASS 1/1\n'
