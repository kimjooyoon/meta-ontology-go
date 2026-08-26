#!/usr/bin/env bash
set -euo pipefail

producer_dir=${1:?producer artifact directory is required}
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
: "${GITHUB_WORKSPACE:?GITHUB_WORKSPACE is required}"
: "${HEAD_SHA:?HEAD_SHA is required}"

script_dir=$(cd "$(dirname "$0")" && pwd)
output_dir="$RUNNER_TEMP/symbolic-invocation-usecase"
producer_artifact="$producer_dir/artifact.json"
producer_receipt="$producer_dir/receipt.json"
external_report="$output_dir/generated-conformance-report.json"
projector="$script_dir/project-generated-value-envelope.jq"
consumer="$script_dir/consume-generated-value-envelope.jq"
report_evaluator="$script_dir/generated-value-projection-report.jq"
integrity_evaluator="$script_dir/generated-value-integrity.jq"

ready_envelope="$output_dir/generated-value-ready-envelope.json"
failed_envelope="$output_dir/generated-value-fail-closed-envelope.json"
ready_observation="$output_dir/generated-value-ready-observation.json"
failed_observation="$output_dir/generated-value-fail-closed-observation.json"
projection_report="$output_dir/generated-value-projection-report.json"
integrity_receipt="$output_dir/generated-value-integrity-receipt.json"
manifest="$output_dir/generated-value-manifest.sha256"

for required in \
  "$producer_artifact" \
  "$producer_receipt" \
  "$external_report" \
  "$projector" \
  "$consumer" \
  "$report_evaluator" \
  "$integrity_evaluator"; do
  test -f "$required"
done

jq -e '
  .schema == "gooo/symbolic-invocation-schema-artifact/v1"
  and .decision == "PASS"
  and .conformance.schema == "gooo/symbolic-invocation-conformance/v1"
  and .conformance.resolution == "STRUCTURAL_ONLY"
  and .conformance.generated_vectors == 2
  and .conformance.effects.repository_writes == 0
  and .conformance.effects.mutation_authority == false
  and ([.conformance.vectors[]
    | select(
        .id == "accept-exact"
        and .expected == "ACCEPT"
        and .proof_choice == "FOUNDATION"
      )] | length) == 1
  and ([.conformance.vectors[]
    | select(
        .id == "reject-missing-activity"
        and .expected == "REJECT"
        and .proof_choice == "REGRESSION"
      )] | length) == 1
' "$producer_artifact" >/dev/null

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/symbolic-invocation-schema-receipt/v1"
  and .subject_sha == $sha
  and .decision == "PASS"
  and .resolution == "EXACT"
  and .source.gooo_files == 2
  and .source.go_files == 0
  and .source.gooo_lines == 10
  and .source.files == 5
  and .source.directories == 0
  and .effects.repository_writes == 0
  and .effects.mutation_authority == false
' "$producer_receipt" >/dev/null

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/generated-symbolic-conformance-report/v1"
  and .subject_sha == $sha
  and .decision == "PASS"
  and .resolution == "EXACT"
  and .summary.generated_vectors == 2
  and .summary.external_decisions == 2
  and .summary.expectation_matches == 2
  and .summary.unknowns == 0
  and ([.observations[]
    | select(.id == "accept-exact" and .expected == "ACCEPT" and .observed == "ACCEPT")] | length) == 1
  and ([.observations[]
    | select(.id == "reject-missing-activity" and .expected == "REJECT" and .observed == "REJECT")] | length) == 1
' "$external_report" >/dev/null

artifact_digest=$(jq -er '.digest' "$producer_artifact")
receipt_artifact_digest=$(jq -er '.artifact.digest' "$producer_receipt")
external_artifact_digest=$(jq -er '.summary.artifact_digest' "$external_report")
validator_digest=$(jq -er '.validation.tool_digest' "$producer_receipt")
external_validator_digest=$(jq -er '.summary.validator_digest' "$external_report")
external_report_digest=$(jq -er '.digest' "$external_report")

test "$artifact_digest" = "$receipt_artifact_digest"
test "$artifact_digest" = "$external_artifact_digest"
test "$validator_digest" = "$external_validator_digest"

repository_writes=0
for output_path in \
  "$ready_envelope" \
  "$failed_envelope" \
  "$ready_observation" \
  "$failed_observation" \
  "$projection_report" \
  "$integrity_receipt" \
  "$manifest"; do
  case "$output_path" in
    "$GITHUB_WORKSPACE" | "$GITHUB_WORKSPACE"/*)
      repository_writes=$((repository_writes + 1))
      ;;
  esac
done

tmp_dir=$(mktemp -d "$RUNNER_TEMP/generated-value-projection.XXXXXX")
ready_input="$tmp_dir/ready-input.json"
failed_input="$tmp_dir/failed-input.json"

make_projection_input() {
  local vector_id=$1
  local destination=$2
  jq -c \
    --arg sha "$HEAD_SHA" \
    --arg vector_id "$vector_id" \
    --arg artifact_digest "$artifact_digest" \
    --arg external_report_digest "$external_report_digest" \
    --arg validator_digest "$validator_digest" \
    --argjson repository_writes "$repository_writes" '
    . as $artifact
    | ($artifact.conformance.vectors[] | select(.id == $vector_id)) as $vector
    | {
        schema: "gooo/generated-value-projection-input/v1",
        subject_sha: $sha,
        source: {
          artifact_digest: $artifact_digest,
          external_report_digest: $external_report_digest,
          validator_digest: $validator_digest
        },
        vector: $vector,
        effects: {
          repository_writes: $repository_writes,
          mutation_authority: false
        }
      }
  ' "$producer_artifact" > "$destination"
}

seal_json() {
  local source_file=$1
  local destination=$2
  local canonical
  local digest
  canonical=$(jq -S -c 'del(.digest)' "$source_file")
  digest=$(printf '%s' "$canonical" | sha256sum | awk '{print $1}')
  jq -S --arg digest "sha256:$digest" '. + {digest: $digest}' "$source_file" > "$destination"
}

verify_sealed_json() {
  local source_file=$1
  local expected
  local canonical
  local observed
  expected=$(jq -er '.digest' "$source_file")
  canonical=$(jq -S -c 'del(.digest)' "$source_file")
  observed="sha256:$(printf '%s' "$canonical" | sha256sum | awk '{print $1}')"
  test "$expected" = "$observed"
}

make_projection_input "accept-exact" "$ready_input"
make_projection_input "reject-missing-activity" "$failed_input"

jq -S -c -f "$projector" "$ready_input" > "$tmp_dir/ready-1.json"
jq -S -c -f "$projector" "$ready_input" > "$tmp_dir/ready-2.json"
jq -S -c -f "$projector" "$failed_input" > "$tmp_dir/failed-1.json"
jq -S -c -f "$projector" "$failed_input" > "$tmp_dir/failed-2.json"
cmp -s "$tmp_dir/ready-1.json" "$tmp_dir/ready-2.json"
cmp -s "$tmp_dir/failed-1.json" "$tmp_dir/failed-2.json"
seal_json "$tmp_dir/ready-1.json" "$ready_envelope"
seal_json "$tmp_dir/failed-1.json" "$failed_envelope"
verify_sealed_json "$ready_envelope"
verify_sealed_json "$failed_envelope"

jq -e '
  .schema == "gooo/symbolic-invocation-value-envelope/v1"
  and .language == "gooo"
  and .source.vector_id == "accept-exact"
  and .source.expected == "ACCEPT"
  and .decision == "READY"
  and .resolution == "VALUE_EXACT"
  and .reason == "SYMBOLIC_INVOCATION_VALUE_PROJECTED"
  and .invocation.activity == "Checkout"
  and .invocation.inputs == [
    "urn:gooo:checkout:cart",
    "urn:gooo:checkout:payment-method"
  ]
  and .invocation.input_count == 2
  and .diagnostics == []
  and .effects.requested_invocations == 1
  and .effects.executed_invocations == 0
  and .effects.repository_writes == 0
  and .effects.mutation_authority == false
' "$ready_envelope" >/dev/null

jq -e '
  .schema == "gooo/symbolic-invocation-value-envelope/v1"
  and .language == "gooo"
  and .source.vector_id == "reject-missing-activity"
  and .source.expected == "REJECT"
  and .decision == "FAIL_CLOSED"
  and .resolution == "LOWER_RESOLUTION"
  and .reason == "SYMBOLIC_INVOCATION_VALUE_INCOMPLETE"
  and (.diagnostics | index("activity-required")) != null
  and .effects.requested_invocations == 0
  and .effects.executed_invocations == 0
  and .effects.repository_writes == 0
  and .effects.mutation_authority == false
' "$failed_envelope" >/dev/null

jq -S -c -f "$consumer" "$ready_envelope" > "$tmp_dir/ready-observation-1.json"
jq -S -c -f "$consumer" "$ready_envelope" > "$tmp_dir/ready-observation-2.json"
jq -S -c -f "$consumer" "$failed_envelope" > "$tmp_dir/failed-observation-1.json"
jq -S -c -f "$consumer" "$failed_envelope" > "$tmp_dir/failed-observation-2.json"
cmp -s "$tmp_dir/ready-observation-1.json" "$tmp_dir/ready-observation-2.json"
cmp -s "$tmp_dir/failed-observation-1.json" "$tmp_dir/failed-observation-2.json"
seal_json "$tmp_dir/ready-observation-1.json" "$ready_observation"
seal_json "$tmp_dir/failed-observation-1.json" "$failed_observation"
verify_sealed_json "$ready_observation"
verify_sealed_json "$failed_observation"

jq -e '
  .decision == "OBSERVED_READY"
  and .resolution == "USER_VALUE"
  and .reason == "READY_INVOCATION_OBSERVED"
  and .value.activity == "Checkout"
  and .value.input_count == 2
  and .effects.executed_invocations == 0
' "$ready_observation" >/dev/null

jq -e '
  .decision == "OBSERVED_FAIL_CLOSED"
  and .resolution == "USER_GUARDRAIL"
  and .reason == "INCOMPLETE_INVOCATION_BLOCKED"
  and (.value.diagnostics | index("activity-required")) != null
  and .value.input_count == 0
  and .effects.executed_invocations == 0
' "$failed_observation" >/dev/null

jq -n \
  --arg sha "$HEAD_SHA" \
  --slurpfile receipt "$producer_receipt" \
  --slurpfile external "$external_report" \
  --slurpfile ready "$ready_envelope" \
  --slurpfile failed "$failed_envelope" \
  --slurpfile ready_observation "$ready_observation" \
  --slurpfile failed_observation "$failed_observation" '
  {
    schema: "gooo/generated-value-projection-report-input/v1",
    subject_sha: $sha,
    source: {
      gooo_files: $receipt[0].source.gooo_files,
      go_files: $receipt[0].source.go_files,
      gooo_lines: $receipt[0].source.gooo_lines,
      files: $receipt[0].source.files,
      directories: $receipt[0].source.directories,
      generated_vectors: $external[0].summary.generated_vectors,
      external_decisions: $external[0].summary.external_decisions,
      artifact_digest: $external[0].summary.artifact_digest,
      external_report_digest: $external[0].digest,
      validator_digest: $external[0].summary.validator_digest
    },
    envelopes: [$ready[0], $failed[0]],
    observations: [$ready_observation[0], $failed_observation[0]],
    replay_matches: [true, true, true, true]
  }
' > "$tmp_dir/report-input.json"

jq -f "$report_evaluator" "$tmp_dir/report-input.json" > "$tmp_dir/projection-report.json"
jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/generated-value-projection-report/v1"
  and .subject_sha == $sha
  and .decision == "PASS"
  and .resolution == "VALUE_PROJECTION_ONLY"
  and .reason == "GENERATED_VALUES_PROJECTED_AND_OBSERVED"
  and .coordinates.satisfied == 14
  and .coordinates.total == 14
  and .coordinates.basis_points == 10000
  and .classes == [
    {"class":"OUTCOME","satisfied":4,"total":4},
    {"class":"DRIVER","satisfied":6,"total":6},
    {"class":"GUARDRAIL","satisfied":4,"total":4}
  ]
  and .views == [
    {"audience":"USER","resolution":"USER_VISIBLE","satisfied":6,"total":6,"basis_points":10000},
    {"audience":"TOOL_AUTHOR","resolution":"TOOL_CONTRACT","satisfied":12,"total":12,"basis_points":10000},
    {"audience":"GOVERNOR","resolution":"FULL_RECEIPT","satisfied":14,"total":14,"basis_points":10000}
  ]
  and .proofs == [
    {"proof_choice":"FOUNDATION","satisfied":6,"total":6},
    {"proof_choice":"COHERENCE","satisfied":5,"total":5},
    {"proof_choice":"REGRESSION","satisfied":3,"total":3}
  ]
  and .effects.requested_invocations == 1
  and .effects.executed_invocations == 0
  and .repository_writes == 0
  and .mutation_authority == false
  and .promotion_credit_bps == 0
' "$tmp_dir/projection-report.json" >/dev/null
seal_json "$tmp_dir/projection-report.json" "$projection_report"
verify_sealed_json "$projection_report"

ready_envelope_digest=$(sha256sum "$ready_envelope" | awk '{print $1}')
failed_envelope_digest=$(sha256sum "$failed_envelope" | awk '{print $1}')
ready_observation_digest=$(sha256sum "$ready_observation" | awk '{print $1}')
failed_observation_digest=$(sha256sum "$failed_observation" | awk '{print $1}')
projection_report_digest=$(sha256sum "$projection_report" | awk '{print $1}')

jq -n \
  --arg sha "$HEAD_SHA" \
  --arg ready_envelope_digest "sha256:$ready_envelope_digest" \
  --arg failed_envelope_digest "sha256:$failed_envelope_digest" \
  --arg ready_observation_digest "sha256:$ready_observation_digest" \
  --arg failed_observation_digest "sha256:$failed_observation_digest" \
  --arg projection_report_digest "sha256:$projection_report_digest" \
  --argjson repository_writes "$repository_writes" '
  {
    schema: "gooo/generated-value-integrity-input/v1",
    subject_sha: $sha,
    manifest_path: "generated-value-manifest.sha256",
    files: [
      {"path":"generated-value-ready-envelope.json","digest":$ready_envelope_digest},
      {"path":"generated-value-fail-closed-envelope.json","digest":$failed_envelope_digest},
      {"path":"generated-value-ready-observation.json","digest":$ready_observation_digest},
      {"path":"generated-value-fail-closed-observation.json","digest":$failed_observation_digest},
      {"path":"generated-value-projection-report.json","digest":$projection_report_digest}
    ],
    effects: {
      repository_writes: $repository_writes,
      mutation_authority: false
    }
  }
' > "$tmp_dir/integrity-input.json"

jq -f "$integrity_evaluator" "$tmp_dir/integrity-input.json" > "$tmp_dir/integrity-receipt.json"
jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/generated-value-integrity-receipt/v1"
  and .subject_sha == $sha
  and .decision == "PASS"
  and .resolution == "EXACT"
  and .reason == "GENERATED_VALUE_PAYLOADS_BOUND"
  and .manifest.payload_bindings == 5
  and .coordinates.satisfied == 6
  and .coordinates.total == 6
  and .coordinates.basis_points == 10000
  and ([.indicators[] | select(.satisfied == false)] | length) == 0
  and .promotion_credit_bps == 0
  and .repository_writes == 0
  and .mutation_authority == false
' "$tmp_dir/integrity-receipt.json" >/dev/null
seal_json "$tmp_dir/integrity-receipt.json" "$integrity_receipt"
verify_sealed_json "$integrity_receipt"

(
  cd "$output_dir"
  sha256sum \
    generated-value-ready-envelope.json \
    generated-value-fail-closed-envelope.json \
    generated-value-ready-observation.json \
    generated-value-fail-closed-observation.json \
    generated-value-projection-report.json > generated-value-manifest.sha256
  test "$(wc -l < generated-value-manifest.sha256)" -eq 5
  sha256sum -c generated-value-manifest.sha256
)

printf 'generated value projection: %s %s/%s ready=%s fail_closed=%s observed=%s replay=%s\n' \
  "$(jq -r '.decision' "$projection_report")" \
  "$(jq -r '.coordinates.satisfied' "$projection_report")" \
  "$(jq -r '.coordinates.total' "$projection_report")" \
  "$(jq '[.envelopes[] | select(.decision == "READY")] | length' "$projection_report")" \
  "$(jq '[.envelopes[] | select(.decision == "FAIL_CLOSED")] | length' "$projection_report")" \
  "$(jq '.observations | length' "$projection_report")" \
  "$(jq -r '.indicators[] | select(.id == "tool.deterministic-replay-matches") | .observed' "$projection_report")"
printf 'generated value integrity: %s %s/%s payloads=%s\n' \
  "$(jq -r '.decision' "$integrity_receipt")" \
  "$(jq -r '.coordinates.satisfied' "$integrity_receipt")" \
  "$(jq -r '.coordinates.total' "$integrity_receipt")" \
  "$(jq -r '.manifest.payload_bindings' "$integrity_receipt")"
