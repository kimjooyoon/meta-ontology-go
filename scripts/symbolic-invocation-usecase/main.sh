#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
producer=${1:?producer artifact directory is required}
root="${GITHUB_WORKSPACE:-$(pwd)}"
work="${RUNNER_TEMP:-/tmp}/symbolic-invocation-usecase"
build="${RUNNER_TEMP:-/tmp}/symbolic-invocation-usecase-build"
reducer="$build/symbolic-invocation-usecase"
contract="$root/examples/symbolic-invocation-usecase/contract.json"
paths=(cmd/symbolic-invocation-usecase internal/meta/symbolicinvocationusecase)
mkdir -p "$work" "$build"

go fix ./cmd/symbolic-invocation-usecase ./internal/meta/symbolicinvocationusecase
git diff --exit-code -- "${paths[@]}"
mapfile -t go_files < <(find "${paths[@]}" -name '*.go' -type f | sort)
gofmt -l "${go_files[@]}" | tee "$work/unformatted.txt"
test ! -s "$work/unformatted.txt"
go test ./cmd/symbolic-invocation-usecase ./internal/meta/symbolicinvocationusecase
go build -trimpath -o "$reducer" ./cmd/symbolic-invocation-usecase

for file in receipt.json artifact.json schema.json generated-accepted.json bin/jv; do
  test -f "$producer/$file"
done
chmod +x "$producer/bin/jv"
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
  .summary.source=={gooo_files:2,go_files:0,gooo_lines:10,files:5,directories:0} and
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

sha256sum "$work/report.json" "$work/unknown-report.json" "$work/link-mismatch-report.json" "$work/generated-link-mismatch-report.json" > "$work/manifest.sha256"
git diff --exit-code -- "${paths[@]}"
jq -r '"### Symbolic invocation user use case\n- decision: \(.decision) / \(.resolution)\n- indicators: \(.summary.coordinates.satisfied)/\(.summary.coordinates.total)\n- generated invocation / golden match: \(.summary.generated_instances)/\(.summary.generated_golden_matches)\n- user decisions: \(.summary.user_decisions)\n- accepted: \(.summary.accepted_instances)\n- rejected: \(.summary.rejected_instances)\n- runner samples: \(.summary.resources.samples)\n- max wall: \(.summary.resources.max_wall_ms) ms\n- max RSS: \(.summary.resources.max_rss_kib) KiB\n- repository writes: \(.repository_writes)\n- promotion credit: \(.promotion_credit_bps) bps\n- receipt: \(.digest)"' "$work/report.json" >> "$GITHUB_STEP_SUMMARY"
