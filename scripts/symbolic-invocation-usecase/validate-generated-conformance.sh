#!/usr/bin/env bash
set -euo pipefail

producer=${1:?producer artifact directory is required}
out=${RUNNER_TEMP:?RUNNER_TEMP is required}/symbolic-invocation-usecase
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
artifact=$producer/artifact.json
schema=$producer/schema.json
receipt=$producer/receipt.json
validator=$producer/bin/jv
vectors_dir=$out/generated-conformance
observations=$out/generated-conformance-observations.jsonl
report=$out/generated-conformance-report.json
mkdir -p "$vectors_dir"
: >"$observations"

for required in "$artifact" "$schema" "$receipt" "$validator"; do
  test -f "$required"
done
jq -e '
  .decision == "PASS" and .kind == "symbolic-invocation-schema" and
  .conformance.schema == "gooo/symbolic-invocation-conformance/v1" and
  .conformance.decision == "PASS" and .conformance.resolution == "STRUCTURAL_ONLY" and
  .conformance.generated_vectors == 2 and .conformance.embedded_handwritten_vectors == 0 and
  .conformance.effects.repository_writes == 0 and
  .conformance.effects.mutation_authority == false and
  (.conformance.vectors | length) == 2 and
  ([.conformance.vectors[].id] | unique | length) == 2
' "$artifact" >/dev/null
test "$(jq -r '.subject_sha' "$receipt")" = "${HEAD_SHA:?HEAD_SHA is required}"
test "$(jq -r '.artifact.digest' "$receipt")" = "$(jq -r '.digest' "$artifact")"

validator_digest=sha256:$(sha256sum "$validator" | cut -d' ' -f1)
schema_digest=sha256:$(sha256sum "$schema" | cut -d' ' -f1)
test "$validator_digest" = "$(jq -r '.validation.tool_digest' "$receipt")"
test "$schema_digest" = "$(jq -r '.artifact.json_schema_digest' "$receipt")"

while IFS= read -r vector; do
  id=$(jq -r '.id' <<<"$vector")
  expected=$(jq -r '.expected' <<<"$vector")
  proof=$(jq -r '.proof_choice' <<<"$vector")
  operation=$(jq -r '.meta_operation' <<<"$vector")
  [[ $id =~ ^[a-z0-9-]+$ ]]
  case "$expected" in ACCEPT | REJECT) ;; *) exit 65 ;; esac
  instance=$vectors_dir/$id.json
  validation=$vectors_dir/$id-validation.txt
  jq '.instance' <<<"$vector" >"$instance"
  if "$validator" "$schema" "$instance" >"$validation" 2>&1; then
    observed=ACCEPT
  else
    observed=REJECT
  fi
  test "$observed" = "$expected"
  jq -cn --arg id "$id" --arg expected "$expected" --arg observed "$observed" \
    --arg proof "$proof" --arg operation "$operation" \
    '{id:$id, expected:$expected, observed:$observed, matches:($expected == $observed), proof_choice:$proof, meta_operation:$operation}' \
    >>"$observations"
done < <(jq -c '.conformance.vectors[]' "$artifact")

generated=$(jq '.conformance.vectors | length' "$artifact")
decisions=$(jq -s 'length' "$observations")
matches=$(jq -s 'map(select(.matches)) | length' "$observations")
accepts=$(jq -s 'map(select(.observed == "ACCEPT")) | length' "$observations")
rejects=$(jq -s 'map(select(.observed == "REJECT")) | length' "$observations")
test "$generated" -eq 2
test "$decisions" -eq 2
test "$matches" -eq 2
test "$accepts" -eq 1
test "$rejects" -eq 1

jq -n --slurpfile observations "$observations" --arg subject_sha "$HEAD_SHA" \
  --arg validator_digest "$validator_digest" --arg artifact_digest "$(jq -r '.digest' "$artifact")" \
  --arg json_schema_digest "$schema_digest" --argjson generated_vectors "$generated" \
  --argjson external_decisions "$decisions" --argjson expectation_matches "$matches" \
  --argjson accepted_vectors "$accepts" --argjson rejected_vectors "$rejects" \
  -f "$script_dir/generated-conformance-report.jq" >"$report"
digest=sha256:$(jq -cS . "$report" | sha256sum | cut -d' ' -f1)
jq --arg digest "$digest" '. + {digest:$digest}' "$report" >"$report.sealed"
mv "$report.sealed" "$report"
printf 'generated symbolic conformance: PASS 8/8\n'
