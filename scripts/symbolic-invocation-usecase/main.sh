#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
producer=${1:?producer artifact directory is required}
root="${GITHUB_WORKSPACE:-$(pwd)}"
work="${RUNNER_TEMP:-/tmp}/symbolic-invocation-usecase"
build="${RUNNER_TEMP:-/tmp}/symbolic-invocation-usecase-build"
reducer="$build/symbolic-invocation-usecase"
reader_observer="$build/symbolic-reader-request-observer"
contract="$root/examples/symbolic-invocation-usecase/contract.json"
paths=(cmd/symbolic-invocation-usecase internal/meta/symbolicinvocationusecase scripts/symbolic-invocation-usecase/reader-observation scripts/symbolic-invocation-usecase/claim-ledger)
mkdir -p "$work" "$build"

go fix ./cmd/symbolic-invocation-usecase ./internal/meta/symbolicinvocationusecase ./internal/meta/symbolicinvocationusecase/claimledger ./scripts/symbolic-invocation-usecase/reader-observation ./scripts/symbolic-invocation-usecase/claim-ledger
git diff --exit-code -- "${paths[@]}"
mapfile -t go_files < <(find "${paths[@]}" -name '*.go' -type f | sort)
gofmt -l "${go_files[@]}" | tee "$work/unformatted.txt"
test ! -s "$work/unformatted.txt"
go test ./cmd/symbolic-invocation-usecase ./internal/meta/symbolicinvocationusecase ./internal/meta/symbolicinvocationusecase/claimledger ./scripts/symbolic-invocation-usecase/reader-observation ./scripts/symbolic-invocation-usecase/claim-ledger
go build -trimpath -o "$reducer" ./cmd/symbolic-invocation-usecase
go build -trimpath -o "$reader_observer" ./scripts/symbolic-invocation-usecase/reader-observation

for file in receipt.json artifact.json schema.json generated-accepted.json symbolic-reader-request-result.json bin/jv; do
  test -f "$producer/$file"
done
chmod +x "$producer/bin/jv"

reader_observation="$work/symbolic-reader-request-user-observation.json"
"$reader_observer" \
  -input "$producer/symbolic-reader-request-result.json" \
  -output "$reader_observation" \
  -expected-subject-sha "$HEAD_SHA"
jq -e --arg subject "$HEAD_SHA" '
  .schema=="gooo/symbolic-reader-request-user-observation/v1" and
  .metric_id=="gooo.metric.user.symbolic-reader-request-observation.v1" and
  .subject_sha==$subject and .source.subject_sha==$subject and
  .decision=="PASS" and .resolution=="USER_OBSERVATION_ONLY" and
  .reason=="CANONICAL_READER_REQUEST_OBSERVED" and
  .coordinates=={satisfied:10,total:10,basis_points:10000} and
  .classes==[
    {class:"OUTCOME",satisfied:3,total:3},
    {class:"DRIVER",satisfied:3,total:3},
    {class:"GUARDRAIL",satisfied:4,total:4}
  ] and
  .proofs==[
    {proof_choice:"FOUNDATION",satisfied:4,total:4},
    {proof_choice:"COHERENCE",satisfied:3,total:3},
    {proof_choice:"REGRESSION",satisfied:3,total:3}
  ] and
  .effects=={repository_writes:0,mutation_authority:false} and
  .promotion_credit_bps==0 and
  (.source.artifact_digest|startswith("sha256:")) and
  (.source.file_digest|startswith("sha256:")) and
  (.digest|startswith("sha256:")) and
  (.not_claimed|index("exact-head cross-job artifact transport")!=null)
' "$reader_observation"

jq -S . "$root/examples/symbolic-invocation-schema/accepted.json" > "$work/accepted-golden.json"
cmp -s "$producer/generated-accepted.json" "$work/accepted-golden.json"
"$producer/bin/jv" "$producer/schema.json" "$producer/generated-accepted.json" > "$work/accepted-validation.txt"
if "$producer/bin/jv" "$producer/schema.json" "$root/examples/symbolic-invocation-schema/rejected.json" > "$work/rejected-validation.txt" 2>&1; then
  echo "rejected user instance unexpectedly passed" >&2
  exit 1
fi

artifact_digest=$(jq -er '.digest' "$producer/artifact.json")
schema_digest="sha256:$(sha256sum "$producer/schema.json" | cut -d' ' -f1)"
tool_digest="sha256:$(sha256sum "$producer/bin/jv" | cut -d' ' -f1)"
generated_instance_digest="sha256:$(sha256sum "$producer/generated-accepted.json" | cut -d' ' -f1)"
jq -n --arg subject "$HEAD_SHA" --arg artifact "$artifact_digest" --arg schema "$schema_digest" \
  --arg tool "$tool_digest" --arg generated "$generated_instance_digest" '{
  schema:"gooo/symbolic-invocation-usecase-observation/v1",
  decision:"PASS",
  resolution:"EXACT",
  reason:"EXTERNAL_USER_VALIDATION_REPLAYED",
  subject_sha:$subject,
  artifact_digest:$artifact,
  json_schema_digest:$schema,
  tool_digest:$tool,
  accepted_instances:1,
  rejected_instances:1,
  generated_instances:1,
  generated_golden_matches:1,
  generated_instance_digest:$generated,
  effects:{repository_writes:0,mutation_authority:false}
}' > "$work/observation.json"

jq -n --arg subject "$HEAD_SHA" --slurpfile contract "$contract" \
  --slurpfile receipt "$producer/receipt.json" --slurpfile artifact "$producer/artifact.json" \
  --slurpfile observation "$work/observation.json" '{
    subject_sha:$subject,
    contract:$contract[0],
    producer_receipt:$receipt[0],
    producer_artifact:$artifact[0],
    observation:$observation[0]
  }' > "$work/input.json"

"$reducer" -input "$work/input.json" -output "$work/report.json"
"$reducer" -input "$work/input.json" -output "$work/replay.json"
cmp -s "$work/report.json" "$work/replay.json"
"$reducer" -input "$work/input.json" -check "$work/report.json"
jq -e '
  .decision=="PASS" and .resolution=="EXACT" and .reason=="SYMBOLIC_INVOCATION_USECASE_OBSERVED" and
  .summary.coordinates=={satisfied:8,total:8,basis_points:10000} and
  .summary.user_decisions==2 and .summary.accepted_instances==1 and .summary.rejected_instances==1 and
  .summary.generated_instances==1 and .summary.generated_golden_matches==1 and
  .summary.deterministic_replays==1 and .summary.unknowns==0 and
  .summary.source=={gooo_files:3,go_files:0,gooo_lines:16,files:6,directories:0} and
  .summary.producer.registered_emitters==3 and .summary.resources.mode=="RUNNER_SCOPED_NONDETERMINISTIC" and
  .summary.resources.measurement_replay_authority==false and .summary.resources.samples==5 and
  .promotion_credit_bps==0 and .repository_writes==0 and .mutation_authority==false and
  ([.indicators[]|select(.class=="OUTCOME")]|length)==2 and
  ([.indicators[]|select(.class=="DRIVER")]|length)==3 and
  ([.indicators[]|select(.class=="GUARDRAIL")]|length)==3 and
  ([.indicators[]|select(.proof_choice=="FOUNDATION")]|length)==4 and
  ([.indicators[]|select(.proof_choice=="COHERENCE")]|length)==3 and
  ([.indicators[]|select(.proof_choice=="REGRESSION")]|length)==1 and
  [.views[]|[.audience,.satisfied,.total]]==[["USER",5,5],["TOOL_AUTHOR",6,6],["GOVERNOR",8,8]] and
  (.not_claimed|length)==4
' "$work/report.json"

jq '.producer_receipt.decision="UNKNOWN"' "$work/input.json" > "$work/unknown-input.json"
if "$reducer" -input "$work/unknown-input.json" -output "$work/unknown-report.json"; then
  echo "unknown producer decision unexpectedly passed" >&2
  exit 1
fi
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .reason=="SYMBOLIC_INVOCATION_USECASE_DECISION_UNKNOWN" and .summary.coordinates.satisfied==0 and .summary.coordinates.total==8 and .summary.unknowns==1' "$work/unknown-report.json"

jq '.producer_artifact.digest="sha256:0000000000000000000000000000000000000000000000000000000000000000"' "$work/input.json" > "$work/link-mismatch-input.json"
if "$reducer" -input "$work/link-mismatch-input.json" -output "$work/link-mismatch-report.json"; then
  echo "artifact link mismatch unexpectedly passed" >&2
  exit 1
fi
jq -e '.decision=="FAIL_CLOSED" and .resolution=="INVARIANT_ONLY" and .reason=="SYMBOLIC_INVOCATION_USECASE_LINK_MISMATCH" and .summary.coordinates.satisfied==0 and .summary.coordinates.total==8' "$work/link-mismatch-report.json"

jq '.observation.generated_instance_digest="sha256:0000000000000000000000000000000000000000000000000000000000000000"' "$work/input.json" > "$work/generated-link-mismatch-input.json"
if "$reducer" -input "$work/generated-link-mismatch-input.json" -output "$work/generated-link-mismatch-report.json"; then
  echo "generated invocation link mismatch unexpectedly passed" >&2
  exit 1
fi
jq -e '.decision=="FAIL_CLOSED" and .resolution=="INVARIANT_ONLY" and .reason=="SYMBOLIC_INVOCATION_USECASE_LINK_MISMATCH" and .summary.coordinates.satisfied==0 and .summary.coordinates.total==8' "$work/generated-link-mismatch-report.json"

sha256sum "$work/report.json" "$work/unknown-report.json" "$work/link-mismatch-report.json" "$work/generated-link-mismatch-report.json" "$reader_observation" > "$work/manifest.sha256"
git diff --exit-code -- "${paths[@]}"
jq -r '"### Symbolic invocation user use case\n- decision: \(.decision) / \(.resolution)\n- indicators: \(.summary.coordinates.satisfied)/\(.summary.coordinates.total)\n- generated invocation / golden match: \(.summary.generated_instances)/\(.summary.generated_golden_matches)\n- user decisions: \(.summary.user_decisions)\n- accepted: \(.summary.accepted_instances)\n- rejected: \(.summary.rejected_instances)\n- runner samples: \(.summary.resources.samples)\n- max wall: \(.summary.resources.max_wall_ms) ms\n- max RSS: \(.summary.resources.max_rss_kib) KiB\n- repository writes: \(.repository_writes)\n- promotion credit: \(.promotion_credit_bps) bps\n- receipt: \(.digest)"' "$work/report.json" >> "$GITHUB_STEP_SUMMARY"
jq -r '"### Compiled reader request: user observation\n- decision: \(.decision) / \(.resolution)\n- indicators: \(.coordinates.satisfied)/\(.coordinates.total)\n- classes: \([.classes[]|"\(.class)=\(.satisfied)/\(.total)"]|join(", "))\n- proofs: \([.proofs[]|"\(.proof_choice)=\(.satisfied)/\(.total)"]|join(", "))\n- source subject: \(.source.subject_sha)\n- repository writes: \(.effects.repository_writes)\n- promotion credit: \(.promotion_credit_bps) bps\n- receipt: \(.digest)\n- cross-job transport claim: excluded"' "$reader_observation" >> "$GITHUB_STEP_SUMMARY"

bash "$(dirname "$0")/validate-generated-conformance.sh" "$1"
bash "$(dirname "$0")/validate-generated-unknown-resolution.sh" "$1"
bash "$(dirname "$0")/validate-generated-value-projection.sh" "$1"
bash "$(dirname "$0")/validate-external-default-coverage.sh" "$1"

# Project every user-visible assertion into a fixed, append-only claim ledger.
claim_ledger_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
claim_ledger_candidates="$({
  find "${RUNNER_TEMP:?RUNNER_TEMP is required}" "${claim_ledger_root}" \
    -type f -name 'symbolic-reader-request-user-observation.json' -print
} | sort -u)"
claim_ledger_candidate_count="$(printf '%s\n' "${claim_ledger_candidates}" | sed '/^$/d' | wc -l | tr -d ' ')"
if [[ "${claim_ledger_candidate_count}" != "1" ]]; then
  printf 'claim-ledger: expected one reader observation, found %s\n' "${claim_ledger_candidate_count}" >&2
  exit 1
fi
claim_ledger_observation="${claim_ledger_candidates}"
claim_ledger_output="$(dirname "${claim_ledger_observation}")/symbolic-reader-request-claim-ledger.json"
claim_ledger_binary="${RUNNER_TEMP}/gooo-symbolic-reader-claim-ledger"
claim_ledger_contract="${claim_ledger_root}/examples/symbolic-invocation-usecase/claim-ledger-contract.json"
(
  cd "${claim_ledger_root}"
  go build -trimpath -o "${claim_ledger_binary}" ./scripts/symbolic-invocation-usecase/claim-ledger
)
"${claim_ledger_binary}" \
  -contract "${claim_ledger_contract}" \
  -observation "${claim_ledger_observation}" \
  -subject "${HEAD_SHA:?HEAD_SHA is required}" \
  -out "${claim_ledger_output}"

jq -e '
  .schema == "gooo/claim-ledger/v1" and
  (.contract_digest | startswith("sha256:")) and
  (.observation_digest | startswith("sha256:")) and
  .conformance.decision == "PASS" and
  .claim_set.decision == "FAIL_CLOSED" and
  .claim_set.resolution == "STAGE_LOCAL" and
  .metrics.fixed_claim_total == 12 and
  .metrics.in_scope_claim_total == 8 and
  .metrics.discharged_total == 4 and
  .metrics.unknown_total == 4 and
  .metrics.excluded_total == 4 and
  .metrics.open_claim_total == 4 and
  .metrics.discharge_basis_points == 5000 and
  .metrics.false_promotion_count == 0 and
  .metrics.proof_routes == {"foundation": 4, "coherence": 4, "regression": 4} and
  ([.claims[] | select(.status == "UNKNOWN") |
    select((.coordinate.stage | length) > 0 and (.coordinate.step | length) > 0 and (.reason | length) > 0)] | length) == 4 and
  ([.claims[] | select(.status == "DISCHARGED")] | length) == 4
' "${claim_ledger_output}" >/dev/null

claim_ledger_manifest="$(dirname "${claim_ledger_observation}")/symbolic-reader-request-claim-ledger-manifest.json"
claim_ledger_report_digest="sha256:$(sha256sum "${claim_ledger_output}" | cut -d' ' -f1)"
jq -n --arg subject "${HEAD_SHA}" --arg report_digest "${claim_ledger_report_digest}" \
  --arg contract_digest "$(jq -er '.contract_digest' "${claim_ledger_output}")" \
  --arg observation_digest "$(jq -er '.observation_digest' "${claim_ledger_output}")" '{
    schema:"gooo/symbolic-reader-request-claim-ledger-manifest/v1",
    subject_sha:$subject,
    metric_id:"gooo.metric.user.symbolic-reader-request-claim-ledger.v1",
    report_file:"symbolic-reader-request-claim-ledger.json",
    report_digest:$report_digest,
    contract_digest:$contract_digest,
    observation_digest:$observation_digest,
    conformance_decision:"PASS",
    claim_set_decision:"FAIL_CLOSED",
    resolution:"STAGE_LOCAL",
    fixed_claim_total:12,
    in_scope_claim_total:8,
    discharged_total:4,
    unknown_total:4,
    excluded_total:4,
    false_promotion_count:0
  }' >"${claim_ledger_manifest}"

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    printf '### Gooo claim ledger\n\n'
    printf -- '- claim set: `%s` at `%s` resolution\n' \
      "$(jq -r '.claim_set.decision' "${claim_ledger_output}")" \
      "$(jq -r '.claim_set.resolution' "${claim_ledger_output}")"
    printf -- '- discharged: `%s/%s` scoped claims (`%s` basis points)\n' \
      "$(jq -r '.metrics.discharged_total' "${claim_ledger_output}")" \
      "$(jq -r '.metrics.in_scope_claim_total' "${claim_ledger_output}")" \
      "$(jq -r '.metrics.discharge_basis_points' "${claim_ledger_output}")"
    printf -- '- unknown: `%s`, excluded: `%s`, false promotions: `%s`\n' \
      "$(jq -r '.metrics.unknown_total' "${claim_ledger_output}")" \
      "$(jq -r '.metrics.excluded_total' "${claim_ledger_output}")" \
      "$(jq -r '.metrics.false_promotion_count' "${claim_ledger_output}")"
  } >>"${GITHUB_STEP_SUMMARY}"
fi
