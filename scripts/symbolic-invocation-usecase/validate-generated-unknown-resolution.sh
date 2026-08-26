#!/usr/bin/env bash
set -euo pipefail

producer_dir=${1:?producer artifact directory is required}
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
: "${GITHUB_WORKSPACE:?GITHUB_WORKSPACE is required}"
: "${HEAD_SHA:?HEAD_SHA is required}"

script_dir=$(cd "$(dirname "$0")" && pwd)
output_dir="$RUNNER_TEMP/symbolic-invocation-usecase"
source_report="$output_dir/generated-conformance-report.json"
producer_artifact="$producer_dir/artifact.json"
producer_receipt="$producer_dir/receipt.json"
evaluator="$script_dir/generated-unknown-resolution.jq"
integrity_evaluator="$script_dir/generated-unknown-integrity.jq"
counterfactual="$output_dir/generated-unknown-counterfactual.json"
report="$output_dir/generated-unknown-resolution-report.json"
integrity_receipt="$output_dir/generated-unknown-integrity-receipt.json"
manifest="$output_dir/generated-unknown-manifest.sha256"

for required in "$source_report" "$producer_artifact" "$producer_receipt" "$evaluator" "$integrity_evaluator"; do
  test -f "$required"
done

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/generated-symbolic-conformance-report/v1"
  and .subject_sha == $sha
  and .decision == "PASS"
  and .resolution == "EXACT"
  and .summary.generated_vectors == 2
  and .summary.external_decisions == 2
  and .summary.expectation_matches == 2
  and .summary.unknowns == 0
  and .repository_writes == 0
  and .mutation_authority == false
  and (.digest | startswith("sha256:"))
  and ([.observations[]
    | select(
        .id == "accept-exact"
        and .expected == "ACCEPT"
        and .observed == "ACCEPT"
        and .matches == true
      )] | length) == 1
' "$source_report" >/dev/null

artifact_digest=$(jq -er '.digest' "$producer_artifact")
receipt_artifact_digest=$(jq -er '.artifact.digest' "$producer_receipt")
receipt_validator_digest=$(jq -er '.validation.tool_digest' "$producer_receipt")
receipt_subject_sha=$(jq -er '.subject_sha' "$producer_receipt")
report_artifact_digest=$(jq -er '.summary.artifact_digest' "$source_report")
report_validator_digest=$(jq -er '.summary.validator_digest' "$source_report")

test "$artifact_digest" = "$receipt_artifact_digest"
test "$artifact_digest" = "$report_artifact_digest"
test "$receipt_validator_digest" = "$report_validator_digest"
test "$receipt_subject_sha" = "$HEAD_SHA"

repository_writes=0
for output_path in "$counterfactual" "$report"; do
  case "$output_path" in
    "$GITHUB_WORKSPACE" | "$GITHUB_WORKSPACE"/*)
      repository_writes=$((repository_writes + 1))
      ;;
  esac
done

counterfactual_tmp=$(mktemp "$output_dir/.generated-unknown-counterfactual.XXXXXX")
report_tmp=$(mktemp "$output_dir/.generated-unknown-resolution-report.XXXXXX")
sealed_tmp=$(mktemp "$output_dir/.generated-unknown-resolution-sealed.XXXXXX")
trap 'rm -f "$counterfactual_tmp" "$report_tmp" "$sealed_tmp"' EXIT

jq -c \
  --arg sha "$HEAD_SHA" \
  --argjson repository_writes "$repository_writes" '
  . as $report
  | ($report.observations[] | select(.id == "accept-exact")) as $observation
  | {
      schema: "gooo/generated-verdict-counterfactual/v1",
      subject_sha: $sha,
      source_report_schema: $report.schema,
      source_report_digest: $report.digest,
      source_artifact_digest: $report.summary.artifact_digest,
      source_validator_digest: $report.summary.validator_digest,
      source_external_decisions: $report.summary.external_decisions,
      source_observation: $observation,
      injected_observed: "UNKNOWN",
      effects: {
        repository_writes: $repository_writes,
        mutation_authority: false
      }
    }
' "$source_report" > "$counterfactual_tmp"

mv "$counterfactual_tmp" "$counterfactual"
jq -f "$evaluator" "$counterfactual" > "$report_tmp"

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/generated-unknown-resolution-report/v1"
  and .subject_sha == $sha
  and .decision == "FAIL_CLOSED"
  and .resolution == "LOWER_RESOLUTION"
  and .reason == "GENERATED_CONFORMANCE_DECISION_UNKNOWN"
  and .counterfactual.injected_observed == "UNKNOWN"
  and .counterfactual.observed_known == false
  and .conformance.decision == "PASS"
  and .conformance.coordinates.satisfied == 7
  and .conformance.coordinates.total == 7
  and .conformance.coordinates.basis_points == 10000
  and ([.indicators[] | select(.satisfied == false)] | length) == 0
  and ([.indicators[]
    | select(
        .id == "guardrail.false-fixed-points"
        and .observed == 0
        and .expected == 0
        and .satisfied == true
      )] | length) == 1
  and (.views == [
    {
      "audience": "USER",
      "resolution": "USER_VISIBLE",
      "satisfied": 3,
      "total": 3,
      "basis_points": 10000
    },
    {
      "audience": "TOOL_AUTHOR",
      "resolution": "TOOL_CONTRACT",
      "satisfied": 5,
      "total": 5,
      "basis_points": 10000
    },
    {
      "audience": "GOVERNOR",
      "resolution": "FULL_RECEIPT",
      "satisfied": 7,
      "total": 7,
      "basis_points": 10000
    }
  ])
  and .promotion_credit_bps == 0
  and .repository_writes == 0
  and .mutation_authority == false
' "$report_tmp" >/dev/null

canonical=$(jq -S -c 'del(.digest)' "$report_tmp")
digest=$(printf '%s' "$canonical" | sha256sum | awk '{print $1}')
jq --arg digest "sha256:$digest" '. + {digest: $digest}' "$report_tmp" > "$sealed_tmp"
mv "$sealed_tmp" "$report"

counterfactual_file_digest=$(sha256sum "$counterfactual" | awk '{print $1}')
report_file_digest=$(sha256sum "$report" | awk '{print $1}')
integrity_input_tmp=$(mktemp "$output_dir/.generated-unknown-integrity-input.XXXXXX")
integrity_report_tmp=$(mktemp "$output_dir/.generated-unknown-integrity-report.XXXXXX")
integrity_sealed_tmp=$(mktemp "$output_dir/.generated-unknown-integrity-sealed.XXXXXX")
trap 'rm -f "$counterfactual_tmp" "$report_tmp" "$sealed_tmp" "$integrity_input_tmp" "$integrity_report_tmp" "$integrity_sealed_tmp"' EXIT

jq -n \
  --arg sha "$HEAD_SHA" \
  --arg counterfactual_digest "sha256:$counterfactual_file_digest" \
  --arg report_digest "sha256:$report_file_digest" \
  --argjson repository_writes "$repository_writes" '
  {
    schema: "gooo/generated-unknown-integrity-input/v1",
    subject_sha: $sha,
    manifest_path: "generated-unknown-manifest.sha256",
    files: [
      {
        path: "generated-unknown-counterfactual.json",
        digest: $counterfactual_digest
      },
      {
        path: "generated-unknown-resolution-report.json",
        digest: $report_digest
      }
    ],
    effects: {
      repository_writes: $repository_writes,
      mutation_authority: false
    }
  }
' > "$integrity_input_tmp"

jq -f "$integrity_evaluator" "$integrity_input_tmp" > "$integrity_report_tmp"
jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/generated-unknown-integrity-receipt/v1"
  and .subject_sha == $sha
  and .decision == "PASS"
  and .resolution == "EXACT"
  and .reason == "GENERATED_UNKNOWN_PAYLOADS_BOUND"
  and .manifest.payload_bindings == 2
  and .coordinates.satisfied == 3
  and .coordinates.total == 3
  and .coordinates.basis_points == 10000
  and ([.indicators[] | select(.satisfied == false)] | length) == 0
  and .promotion_credit_bps == 0
  and .repository_writes == 0
  and .mutation_authority == false
' "$integrity_report_tmp" >/dev/null

integrity_canonical=$(jq -S -c 'del(.digest)' "$integrity_report_tmp")
integrity_digest=$(printf '%s' "$integrity_canonical" | sha256sum | awk '{print $1}')
jq --arg digest "sha256:$integrity_digest" '. + {digest: $digest}' "$integrity_report_tmp" > "$integrity_sealed_tmp"
mv "$integrity_sealed_tmp" "$integrity_receipt"

(
  cd "$output_dir"
  sha256sum \
    generated-unknown-counterfactual.json \
    generated-unknown-resolution-report.json > generated-unknown-manifest.sha256
  test "$(wc -l < generated-unknown-manifest.sha256)" -eq 2
  sha256sum -c generated-unknown-manifest.sha256
)

printf 'generated unknown resolution: %s %s %s %s/%s\n' \
  "$(jq -r '.decision' "$report")" \
  "$(jq -r '.resolution' "$report")" \
  "$(jq -r '.conformance.decision' "$report")" \
  "$(jq -r '.conformance.coordinates.satisfied' "$report")" \
  "$(jq -r '.conformance.coordinates.total' "$report")"
printf 'generated unknown integrity: %s %s/%s\n' \
  "$(jq -r '.decision' "$integrity_receipt")" \
  "$(jq -r '.coordinates.satisfied' "$integrity_receipt")" \
  "$(jq -r '.coordinates.total' "$integrity_receipt")"
