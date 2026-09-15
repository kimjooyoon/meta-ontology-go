#!/usr/bin/env bash
set -euo pipefail

: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
: "${HEAD_SHA:?HEAD_SHA is required}"

output_dir="$RUNNER_TEMP/symbolic-invocation-schema"
artifact="$output_dir/artifact.json"
contract="$output_dir/symbolic-value-contract.json"
reachability="$output_dir/symbolic-value-reachability.json"

test -f "$artifact"
test -f "$contract"
go run ./scripts/symbolic-invocation-schema/valuereachability "$artifact" "$contract" "$reachability" "$HEAD_SHA"

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/symbolic-invocation-value-reachability/v1"
  and .subject_sha == $sha
  and .metric_id == "gooo.metric.compiler.symbolic-value-reachability.v1"
  and .decision == "PASS"
  and .resolution == "SCHEMA_VALUE_REACHABILITY_ONLY"
  and .reason == "SYMBOLIC_VALUE_REACHABILITY_COMPILED"
  and .summary == {
    "policy_branches":3,
    "reachable_rules":1,
    "defense_only_rules":1,
    "reachable_defaults":0,
    "defense_only_defaults":1,
    "unknown_policy_branches":0
  }
  and .coordinates == {"satisfied":11,"total":11,"basis_points":10000}
  and .classes == [
    {"class":"OUTCOME","satisfied":4,"total":4},
    {"class":"DRIVER","satisfied":3,"total":3},
    {"class":"GUARDRAIL","satisfied":4,"total":4}
  ]
  and .views == [
    {"audience":"USER","resolution":"USER_VISIBLE","satisfied":5,"total":5,"basis_points":10000},
    {"audience":"TOOL_AUTHOR","resolution":"TOOL_CONTRACT","satisfied":9,"total":9,"basis_points":10000},
    {"audience":"GOVERNOR","resolution":"FULL_RECEIPT","satisfied":11,"total":11,"basis_points":10000}
  ]
  and .proofs == [
    {"proof_choice":"FOUNDATION","satisfied":5,"total":5},
    {"proof_choice":"COHERENCE","satisfied":3,"total":3},
    {"proof_choice":"REGRESSION","satisfied":3,"total":3}
  ]
  and .rules == [
    {
      "id":"complete-symbolic-invocation",
      "reachability":"REACHABLE",
      "reachable_after_structural_gate":true,
      "role":"NORMAL_PATH",
      "reason":"GENERATED_SCHEMA_PROVIDES_READY_WITNESS",
      "proof_choice":"FOUNDATION",
      "meta_operation":"project-exact-symbolic-invocation"
    },
    {
      "id":"missing-activity",
      "reachability":"UNREACHABLE",
      "reachable_after_structural_gate":false,
      "role":"DEFENSE_IN_DEPTH",
      "reason":"GENERATED_SCHEMA_REQUIRES_NON_EMPTY_ACTIVITY",
      "proof_choice":"REGRESSION",
      "meta_operation":"remove-required-activity"
    }
  ]
  and .default.reachability == "UNREACHABLE"
  and .default.reachable_after_structural_gate == false
  and .default.role == "DEFENSE_IN_DEPTH"
  and .default.reason == "GENERATED_SCHEMA_ENTAILS_COMPLETE_READY_RULE"
  and ([.indicators[] | select(.satisfied == false)] | length) == 0
  and .effects == {"repository_writes":0,"mutation_authority":false}
  and .promotion_credit_bps == 0
' "$reachability" >/dev/null

artifact_digest=$(jq -er '.digest' "$artifact")
contract_digest=$(jq -er '.digest' "$contract")
test "$(jq -er '.source.artifact_digest' "$reachability")" = "$artifact_digest"
test "$(jq -er '.source.contract_digest' "$reachability")" = "$contract_digest"

expected_digest=$(jq -er '.digest' "$reachability")
canonical=$(jq -S -c 'del(.digest)' "$reachability")
observed_digest="sha256:$(printf '%s' "$canonical" | sha256sum | awk '{print $1}')"
test "$expected_digest" = "$observed_digest"

(
  cd "$output_dir"
  sha256sum symbolic-value-reachability.json > symbolic-value-reachability-manifest.sha256
  test "$(wc -l < symbolic-value-reachability-manifest.sha256)" -eq 1
  sha256sum -c symbolic-value-reachability-manifest.sha256
)

printf 'compiler symbolic value reachability manifest: PASS 1/1\n'
