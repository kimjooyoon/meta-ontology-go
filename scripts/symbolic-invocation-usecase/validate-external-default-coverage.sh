#!/usr/bin/env bash
set -euo pipefail

producer_dir=${1:?producer artifact directory is required}
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
: "${GITHUB_WORKSPACE:?GITHUB_WORKSPACE is required}"
: "${HEAD_SHA:?HEAD_SHA is required}"

script_dir=$(cd "$(dirname "$0")" && pwd)
output_dir="$RUNNER_TEMP/symbolic-invocation-usecase"
user_source="$GITHUB_WORKSPACE/examples/symbolic-invocation-usecase/default-unmatched-input.json"
external_input="$output_dir/external-default-input.json"
validation="$output_dir/external-default-structural-validation.txt"
envelope="$output_dir/external-default-envelope.json"
observation="$output_dir/external-default-observation.json"
reachability_copy="$output_dir/external-default-compiler-reachability.json"
report="$output_dir/external-default-coverage-report.json"
manifest="$output_dir/external-default-manifest.sha256"
integrity_receipt="$output_dir/external-default-integrity-receipt.json"

schema="$producer_dir/schema.json"
validator="$producer_dir/bin/jv"
artifact="$producer_dir/artifact.json"
contract="$producer_dir/symbolic-value-contract.json"
contract_manifest="$producer_dir/symbolic-value-contract-manifest.sha256"
reachability="$producer_dir/symbolic-value-reachability.json"
reachability_manifest="$producer_dir/symbolic-value-reachability-manifest.sha256"
external_report="$output_dir/generated-conformance-report.json"
projector="$script_dir/project-generated-value-envelope.jq"
consumer="$script_dir/consume-generated-value-envelope.jq"
report_evaluator="$script_dir/external-default-coverage-report.jq"
integrity_evaluator="$script_dir/external-default-integrity.jq"

for required in "$user_source" "$schema" "$validator" "$artifact" "$contract" "$contract_manifest" \
  "$reachability" "$reachability_manifest" \
  "$external_report" "$projector" "$consumer" "$report_evaluator" "$integrity_evaluator"; do
  test -f "$required"
done

(
  cd "$producer_dir"
  test "$(wc -l < symbolic-value-contract-manifest.sha256)" -eq 1
  sha256sum -c symbolic-value-contract-manifest.sha256
  test "$(wc -l < symbolic-value-reachability-manifest.sha256)" -eq 1
  sha256sum -c symbolic-value-reachability-manifest.sha256
)

jq -e '
  .activity == "Checkout"
  and .inputs == []
' "$user_source" >/dev/null

repository_writes=0
for output_path in "$external_input" "$validation" "$envelope" "$observation" "$reachability_copy" "$report" "$manifest" "$integrity_receipt"; do
  case "$output_path" in
    "$GITHUB_WORKSPACE" | "$GITHUB_WORKSPACE"/*)
      repository_writes=$((repository_writes + 1))
      ;;
  esac
done

cp "$user_source" "$external_input"
cp "$reachability" "$reachability_copy"
cmp -s "$reachability" "$reachability_copy"
validator_exit_code=0
"$validator" "$schema" "$external_input" > "$validation" 2>&1 || validator_exit_code=$?
test "$validator_exit_code" -eq 1
test -s "$validation"
printf 'external default structural validation: REJECT exit=%s\n' "$validator_exit_code"

artifact_digest=$(jq -er '.digest' "$artifact")
schema_digest="sha256:$(sha256sum "$schema" | awk '{print $1}')"
contract_digest=$(jq -er '.digest' "$contract")
contract_artifact_digest=$(jq -er '.source_artifact_digest' "$contract")
reachability_digest=$(jq -er '.digest' "$reachability_copy")
reachability_file_digest="sha256:$(sha256sum "$reachability_copy" | awk '{print $1}')"
validator_digest=$(jq -er '.summary.validator_digest' "$external_report")
external_report_digest=$(jq -er '.digest' "$external_report")
test "$artifact_digest" = "$contract_artifact_digest"
jq -e \
  --arg sha "$HEAD_SHA" \
  --arg artifact_digest "$artifact_digest" \
  --arg contract_digest "$contract_digest" '
  .schema == "gooo/symbolic-invocation-value-reachability/v1"
  and .subject_sha == $sha
  and .metric_id == "gooo.metric.compiler.symbolic-value-reachability.v1"
  and .decision == "PASS"
  and .resolution == "SCHEMA_VALUE_REACHABILITY_ONLY"
  and .source.artifact_digest == $artifact_digest
  and .source.contract_digest == $contract_digest
  and .summary.policy_branches == 3
  and .summary.reachable_rules == 1
  and .summary.defense_only_rules == 1
  and .summary.reachable_defaults == 0
  and .summary.defense_only_defaults == 1
  and .summary.unknown_policy_branches == 0
  and (.digest | test("^sha256:[0-9a-f]{64}$"))
' "$reachability_copy" >/dev/null

tmp_dir=$(mktemp -d "$RUNNER_TEMP/external-default-coverage.XXXXXX")

seal_json() {
  local source_file=$1
  local destination=$2
  local canonical
  local digest
  canonical=$(jq -S -c 'del(.digest)' "$source_file")
  digest=$(printf '%s' "$canonical" | sha256sum | awk '{print $1}')
  jq -S --arg digest "sha256:$digest" '. + {digest: $digest}' "$source_file" > "$destination"
}

jq -n \
  --arg sha "$HEAD_SHA" \
  --arg artifact_digest "$artifact_digest" \
  --arg external_report_digest "$external_report_digest" \
  --arg validator_digest "$validator_digest" \
  --argjson repository_writes "$repository_writes" \
  --slurpfile instance "$external_input" \
  --slurpfile contract "$contract" '
  {
    schema: "gooo/external-default-projection-input/v1",
    subject_sha: $sha,
    source: {
      artifact_digest: $artifact_digest,
      external_report_digest: $external_report_digest,
      validator_digest: $validator_digest
    },
    vector: {
      id: "external-default-empty-inputs",
      expected: "REJECT",
      proof_choice: "REGRESSION",
      meta_operation: "exercise-compiler-default-defense-counterfactually",
      instance: $instance[0]
    },
    contract: $contract[0],
    effects: {repository_writes: $repository_writes, mutation_authority: false}
  }
' > "$tmp_dir/projection-input.json"

jq -S -c -f "$projector" "$tmp_dir/projection-input.json" > "$tmp_dir/envelope-1.json"
jq -S -c -f "$projector" "$tmp_dir/projection-input.json" > "$tmp_dir/envelope-2.json"
cmp -s "$tmp_dir/envelope-1.json" "$tmp_dir/envelope-2.json"
seal_json "$tmp_dir/envelope-1.json" "$envelope"

jq -e --arg contract_digest "$contract_digest" '
  .source.vector_id == "external-default-empty-inputs"
  and .source.expected == "REJECT"
  and .source.contract_digest == $contract_digest
  and .source.rule_id == "default"
  and .decision == "FAIL_CLOSED"
  and .resolution == "LOWER_RESOLUTION"
  and .reason == "SYMBOLIC_INVOCATION_VALUE_UNMATCHED"
  and (.diagnostics | index("inputs-required")) != null
  and .effects.requested_invocations == 0
  and .effects.executed_invocations == 0
  and .effects.repository_writes == 0
  and .effects.mutation_authority == false
' "$envelope" >/dev/null

jq -S -c -f "$consumer" "$envelope" > "$tmp_dir/observation-1.json"
jq -S -c -f "$consumer" "$envelope" > "$tmp_dir/observation-2.json"
cmp -s "$tmp_dir/observation-1.json" "$tmp_dir/observation-2.json"
seal_json "$tmp_dir/observation-1.json" "$observation"

jq -e '
  .decision == "OBSERVED_FAIL_CLOSED"
  and .resolution == "USER_GUARDRAIL"
  and .reason == "INCOMPLETE_INVOCATION_BLOCKED"
  and (.value.diagnostics | index("inputs-required")) != null
  and .effects.executed_invocations == 0
  and .effects.repository_writes == 0
  and .effects.mutation_authority == false
' "$observation" >/dev/null

input_digest="sha256:$(sha256sum "$external_input" | awk '{print $1}')"
validation_digest="sha256:$(sha256sum "$validation" | awk '{print $1}')"

jq -n \
  --arg sha "$HEAD_SHA" \
  --arg input_digest "$input_digest" \
  --arg validation_digest "$validation_digest" \
  --arg artifact_digest "$artifact_digest" \
  --arg schema_digest "$schema_digest" \
  --arg contract_digest "$contract_digest" \
  --arg reachability_digest "$reachability_digest" \
  --arg reachability_file_digest "$reachability_file_digest" \
  --arg validator_digest "$validator_digest" \
  --argjson validator_exit_code "$validator_exit_code" \
  --argjson repository_writes "$repository_writes" \
  --slurpfile reachability "$reachability_copy" \
  --slurpfile envelope "$envelope" \
  --slurpfile observation "$observation" '
  {
    schema: "gooo/external-default-coverage-report-input/v2",
    subject_sha: $sha,
    source: {
      input_digest: $input_digest,
      validation_digest: $validation_digest,
      structural_decision: "REJECT",
      validator_exit_code: $validator_exit_code,
      projection_mode: "COUNTERFACTUAL_AFTER_STRUCTURAL_REJECT",
      artifact_digest: $artifact_digest,
      schema_digest: $schema_digest,
      contract_digest: $contract_digest,
      compiler_reachability_digest: $reachability_digest,
      compiler_reachability_file_digest: $reachability_file_digest,
      validator_digest: $validator_digest
    },
    compiler_reachability: $reachability[0],
    envelope: $envelope[0],
    observation: $observation[0],
    replay_matches: [true, true],
    effects: {repository_writes: $repository_writes, mutation_authority: false}
  }
' > "$tmp_dir/report-input.json"

jq -f "$report_evaluator" "$tmp_dir/report-input.json" > "$tmp_dir/report.json"
jq '.compiler_reachability.digest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"' \
  "$tmp_dir/report-input.json" > "$tmp_dir/report-input-digest-mismatch.json"
jq -f "$report_evaluator" "$tmp_dir/report-input-digest-mismatch.json" > "$tmp_dir/report-digest-mismatch.json"
jq -e '
  .decision == "FAIL_CLOSED"
  and .resolution == "INVARIANT_ONLY"
  and ([.indicators[] | select(.id == "source.compiler-reachability-bindings" and .satisfied == false)] | length) == 1
  and .contrast.default_reachable_after_structural_gate == true
' "$tmp_dir/report-digest-mismatch.json" >/dev/null
jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/external-default-coverage-report/v2"
  and .subject_sha == $sha
  and .metric_id == "gooo.metric.user.external-default-defense.v2"
  and .decision == "PASS"
  and .resolution == "COUNTERFACTUAL_DEFENSE_ONLY"
  and .reason == "COMPILER_REACHABILITY_BOUND_COUNTERFACTUAL_DEFAULT_FAIL_CLOSED"
  and .contrast.structural_decision == "REJECT"
  and .contrast.value_decision == "FAIL_CLOSED"
  and .contrast.selected_rule == "default"
  and .contrast.user_decision == "OBSERVED_FAIL_CLOSED"
  and .contrast.projection_mode == "COUNTERFACTUAL_AFTER_STRUCTURAL_REJECT"
  and .contrast.reachability_source == "COMPILER_SIDECAR"
  and .contrast.reachability_resolution == "SCHEMA_VALUE_REACHABILITY_ONLY"
  and .contrast.schema_entails_ready_rule == true
  and .contrast.default_reachable_after_structural_gate == false
  and .compiler_reachability.digest == .source.compiler_reachability_digest
  and .compiler_reachability.summary.unknown_policy_branches == 0
  and .coordinates.satisfied == 14
  and .coordinates.total == 14
  and .coordinates.basis_points == 10000
  and .classes == [
    {"class":"OUTCOME","satisfied":3,"total":3},
    {"class":"DRIVER","satisfied":6,"total":6},
    {"class":"GUARDRAIL","satisfied":5,"total":5}
  ]
  and .views == [
    {"audience":"USER","resolution":"USER_VISIBLE","satisfied":8,"total":8,"basis_points":10000},
    {"audience":"TOOL_AUTHOR","resolution":"TOOL_CONTRACT","satisfied":12,"total":12,"basis_points":10000},
    {"audience":"GOVERNOR","resolution":"FULL_RECEIPT","satisfied":14,"total":14,"basis_points":10000}
  ]
  and .proofs == [
    {"proof_choice":"FOUNDATION","satisfied":6,"total":6},
    {"proof_choice":"COHERENCE","satisfied":4,"total":4},
    {"proof_choice":"REGRESSION","satisfied":4,"total":4}
  ]
  and .effects.executed_invocations == 0
  and .repository_writes == 0
  and .mutation_authority == false
  and .promotion_credit_bps == 0
' "$tmp_dir/report.json" >/dev/null
seal_json "$tmp_dir/report.json" "$report"

external_input_digest=$(sha256sum "$external_input" | awk '{print $1}')
validation_file_digest=$(sha256sum "$validation" | awk '{print $1}')
envelope_digest=$(sha256sum "$envelope" | awk '{print $1}')
observation_digest=$(sha256sum "$observation" | awk '{print $1}')
reachability_payload_digest=$(sha256sum "$reachability_copy" | awk '{print $1}')
report_digest=$(sha256sum "$report" | awk '{print $1}')

jq -n \
  --arg sha "$HEAD_SHA" \
  --arg external_input_digest "sha256:$external_input_digest" \
  --arg validation_file_digest "sha256:$validation_file_digest" \
  --arg envelope_digest "sha256:$envelope_digest" \
  --arg observation_digest "sha256:$observation_digest" \
  --arg reachability_payload_digest "sha256:$reachability_payload_digest" \
  --arg report_digest "sha256:$report_digest" \
  --argjson repository_writes "$repository_writes" '
  {
    schema: "gooo/external-default-integrity-input/v1",
    subject_sha: $sha,
    manifest_path: "external-default-manifest.sha256",
    files: [
      {"path":"external-default-input.json","digest":$external_input_digest},
      {"path":"external-default-structural-validation.txt","digest":$validation_file_digest},
      {"path":"external-default-envelope.json","digest":$envelope_digest},
      {"path":"external-default-observation.json","digest":$observation_digest},
      {"path":"external-default-compiler-reachability.json","digest":$reachability_payload_digest},
      {"path":"external-default-coverage-report.json","digest":$report_digest}
    ],
    effects: {repository_writes: $repository_writes, mutation_authority: false}
  }
' > "$tmp_dir/integrity-input.json"

jq -f "$integrity_evaluator" "$tmp_dir/integrity-input.json" > "$tmp_dir/integrity-receipt.json"
jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/external-default-integrity-receipt/v2"
  and .subject_sha == $sha
  and .metric_id == "gooo.metric.guardrail.external-default-defense-integrity.v2"
  and .decision == "PASS"
  and .resolution == "EXACT"
  and .reason == "EXTERNAL_DEFAULT_DEFENSE_PAYLOADS_BOUND"
  and .manifest.payload_bindings == 6
  and .coordinates.satisfied == 7
  and .coordinates.total == 7
  and .coordinates.basis_points == 10000
  and ([.indicators[] | select(.satisfied == false)] | length) == 0
  and .promotion_credit_bps == 0
  and .repository_writes == 0
  and .mutation_authority == false
' "$tmp_dir/integrity-receipt.json" >/dev/null
seal_json "$tmp_dir/integrity-receipt.json" "$integrity_receipt"

(
  cd "$output_dir"
  sha256sum \
    external-default-input.json \
    external-default-structural-validation.txt \
    external-default-envelope.json \
    external-default-observation.json \
    external-default-compiler-reachability.json \
    external-default-coverage-report.json > external-default-manifest.sha256
  test "$(wc -l < external-default-manifest.sha256)" -eq 6
  sha256sum -c external-default-manifest.sha256
)

printf 'external default defense: %s %s/%s structural=%s value=%s selected=%s reachable=%s false_ready=%s\n' \
  "$(jq -r '.decision' "$report")" \
  "$(jq -r '.coordinates.satisfied' "$report")" \
  "$(jq -r '.coordinates.total' "$report")" \
  "$(jq -r '.contrast.structural_decision' "$report")" \
  "$(jq -r '.contrast.value_decision' "$report")" \
  "$(jq -r '.contrast.selected_rule' "$report")" \
  "$(jq -r '.contrast.default_reachable_after_structural_gate' "$report")" \
  "$(jq -r '.indicators[] | select(.id == "guardrail.false-ready-envelopes") | .observed' "$report")"
printf 'external default integrity: %s %s/%s payloads=%s\n' \
  "$(jq -r '.decision' "$integrity_receipt")" \
  "$(jq -r '.coordinates.satisfied' "$integrity_receipt")" \
  "$(jq -r '.coordinates.total' "$integrity_receipt")" \
  "$(jq -r '.manifest.payload_bindings' "$integrity_receipt")"
