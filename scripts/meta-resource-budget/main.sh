#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
work="${RUNNER_TEMP:-/tmp}/meta-resource-budget"
build="$work/build"
binary="$build/gooo"
reducer="$build/reducer"
example="examples/meta-resource-budget"
contract="$root/$example/contract.json"
mkdir -p "$work" "$build"

test "$(go env GOVERSION)" = "go1.27.0"
go build -trimpath -o "$binary" ./cmd/gooo
go build -trimpath -o "$reducer" ./cmd/meta-resource-budget-reducer

source_digest="$(sha256sum "$root/$example/activity.gooo" "$root/$example/entities.gooo" | awk '{print $1}' | sha256sum | awk '{print "sha256:"$1}')"
observations="$work/observations.jsonl"
: > "$observations"

measure() {
  local operation="$1" output_kind="$2" sequence="$3"
  shift 3
  local output="$work/$operation-$sequence.json"
  local rss_file="$work/$operation-$sequence.rss"
  local stderr_file="$work/$operation-$sequence.stderr"
  local started finished exit_code wall receipt_bytes generated_bytes
  started="$(date +%s%N)"
  set +e
  (cd "$root" && /usr/bin/time -f '%M' -o "$rss_file" "$binary" "$@" >"$output" 2>"$stderr_file")
  exit_code=$?
  set -e
  finished="$(date +%s%N)"
  wall=$((finished-started))
  receipt_bytes=0
  generated_bytes=0
  if [[ "$output_kind" == "RECEIPT" ]]; then receipt_bytes="$(wc -c <"$output" | tr -d '[:space:]')"; fi
  if [[ "$output_kind" == "GENERATED" ]]; then generated_bytes="$(wc -c <"$output" | tr -d '[:space:]')"; fi
  jq -nc --arg schema "gooo/meta-resource-budget-observation/v1" --arg subject "$HEAD_SHA" \
    --arg producer "scripts/meta-resource-budget" --arg consumer "cmd/meta-resource-budget-reducer" \
    --arg operation "$operation" --argjson sequence "$sequence" --argjson exit_code "$exit_code" \
    --arg stage "$(jq -r --arg id "$operation" '.operations[]|select(.id==$id)|.stage' "$contract")" \
    --arg step "$(jq -r --arg id "$operation" '.operations[]|select(.id==$id)|.step' "$contract")" \
    --arg meta_operation "$(jq -r --arg id "$operation" '.operations[]|select(.id==$id)|.meta_operation' "$contract")" \
    --arg proof_choice "$(jq -r --arg id "$operation" '.operations[]|select(.id==$id)|.proof_choice' "$contract")" \
    --arg reason "RUNNER_RESOURCE_OBSERVED" --argjson wall_time_ns "$wall" \
    --argjson peak_rss_kib "$(tr -d '[:space:]' <"$rss_file")" --argjson receipt_bytes "$receipt_bytes" \
    --argjson generated_bytes "$generated_bytes" --arg output_digest "sha256:$(sha256sum "$output" | cut -d' ' -f1)" \
    '{schema:$schema,subject_sha:$subject,producer:$producer,consumer:$consumer,operation:$operation,stage:$stage,step:$step,meta_operation:$meta_operation,proof_choice:$proof_choice,reason:$reason,sequence:$sequence,exit_code:$exit_code,wall_time_ns:$wall_time_ns,peak_rss_kib:$peak_rss_kib,receipt_bytes:$receipt_bytes,generated_bytes:$generated_bytes,output_digest:$output_digest}' \
    >> "$observations"
  test "$exit_code" = 0
}

for sequence in 1 2 3; do measure source-check RECEIPT "$sequence" check --json "$example/activity.gooo"; done
for sequence in 1 2 3; do measure project-manifest GENERATED "$sequence" emit --kind operation-manifest --entry PayOrder "$example"; done
for sequence in 1 2 3; do measure replay-manifest GENERATED "$sequence" emit --kind operation-manifest --entry PayOrder "$example"; done

observed="$(jq -s . "$observations")"
source_receipt="$work/source-check-1.json"
artifact="$work/project-manifest-1.json"
replay="$work/replay-manifest-1.json"
jq -n --arg schema "gooo/meta-resource-budget-input/v1" --arg head "$HEAD_SHA" \
  --slurpfile contract "$contract" --arg source_receipt_base64 "$(base64 -w0 "$source_receipt")" \
  --arg artifact_base64 "$(base64 -w0 "$artifact")" --arg replay_base64 "$(base64 -w0 "$replay")" \
  --arg source_digest "$source_digest" --argjson observations "$observed" \
  --arg runner_os "$(uname -s)" --arg runner_architecture "$(uname -m)" --arg runner_image "ubuntu-latest" \
  --arg runner_image_version "${ImageVersion:-unknown}" --arg runner_go_version "$(go env GOVERSION)" \
  '{schema:$schema,expected_head:$head,contract:$contract[0],producer:{source_receipt_base64:$source_receipt_base64,artifact_base64:$artifact_base64,replay_base64:$replay_base64,source_digest:$source_digest,source_files:2,go_files:0,runner:{os:$runner_os,architecture:$runner_architecture,image:$runner_image,image_version:$runner_image_version,go_version:$runner_go_version},effects:{repository_writes:0,mutation_authority:false}},observations:$observations}' \
  > "$work/input.json"

"$reducer" -input "$work/input.json" -output "$work/normal-report.json" -case normal
"$reducer" -input "$work/input.json" -check "$work/normal-report.json" -case normal
jq -e '.decision=="PASS" and .resolution=="EXACT" and .summary.coordinates=={satisfied:22,total:22,basis_points:10000} and .summary.operations==3 and .summary.samples==9 and .effects=={repository_writes:0,mutation_authority:false}' "$work/normal-report.json"

jq '.observations[0].wall_time_ns=(.contract.limits.wall_time_ms*1000000+1)' "$work/input.json" > "$work/over-budget-input.json"
set +e
"$reducer" -input "$work/over-budget-input.json" -output "$work/over-budget-report.json" -case over-budget
code=$?
set -e
test "$code" = 1
"$reducer" -input "$work/over-budget-input.json" -check "$work/over-budget-report.json" -case over-budget
jq -e '.decision=="FAIL_CLOSED" and .resolution=="EXACT" and .reason=="RESOURCE_BUDGET_EXCEEDED" and .semantic.decision=="PASS" and .interpretation=="SEMANTIC_EXACT_RESOURCE_CLAIM_REFUTED" and .claim_transitions[1].to=="REFUTED"' "$work/over-budget-report.json"

jq 'del(.observations[-1])' "$work/input.json" > "$work/missing-sample-input.json"
set +e
"$reducer" -input "$work/missing-sample-input.json" -output "$work/missing-sample-report.json" -case missing-sample
code=$?
set -e
test "$code" = 1
"$reducer" -input "$work/missing-sample-input.json" -check "$work/missing-sample-report.json" -case missing-sample
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .reason=="RESOURCE_SAMPLE_MISSING" and .semantic.resolution=="EXACT" and .resource_resolution=="LOWER_RESOLUTION" and .claim_transitions[1].to=="OPEN"' "$work/missing-sample-report.json"

git diff --exit-code
{
  echo '### Meta resource budget experiment'
  echo
  echo '| case | decision | resolution | semantic | resource interpretation |'
  echo '|---|---|---|---|---|'
  jq -r '"| normal | \(.decision) | \(.resolution) | \(.semantic.decision)/\(.semantic.resolution) | \(.interpretation) |"' "$work/normal-report.json"
  jq -r '"| over-budget | \(.decision) | \(.resolution) | \(.semantic.decision)/\(.semantic.resolution) | \(.interpretation) |"' "$work/over-budget-report.json"
  jq -r '"| missing-sample | \(.decision) | \(.resolution) | \(.semantic.decision)/\(.semantic.resolution) | \(.interpretation) |"' "$work/missing-sample-report.json"
  echo
  echo '**Fixed sample set:** 3 operations × 3 samples = 9 observations; 22 coordinates.'
  echo '**Limits:** wall 2,000 ms; peak RSS 131,072 KiB; receipt 8,192 bytes; generated 16,384 bytes.'
  jq -r '"- runner: \(.summary.runner.os)/\(.summary.runner.architecture), image=\(.summary.runner.image) (\(.summary.runner.image_version)), Go=\(.summary.runner.go_version)"' "$work/normal-report.json"
  jq -r '.summary.resources[]|"- \(.operation): samples=\(.samples), wall max=\(.wall_max_ns) ns, peak RSS max=\(.peak_rss_max_kib) KiB, receipt max=\(.receipt_max_bytes) bytes, generated max=\(.generated_max_bytes) bytes"' "$work/normal-report.json"
  echo '- effects: repository writes=0, mutation authority=false'
  echo '- claims: over-budget refutes the resource claim only; missing sample leaves it OPEN and lowers resource resolution'
} >> "$GITHUB_STEP_SUMMARY"
