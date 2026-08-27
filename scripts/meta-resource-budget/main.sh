#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
work="${RUNNER_TEMP:-/tmp}/meta-resource-budget"
build="$work/build"
binary="$build/gooo"
reducer="$build/reducer"
consumer="$build/consumer"
example="examples/meta-resource-budget"
contract="$root/$example/contract.json"
mkdir -p "$work" "$build"

test "$(go env GOVERSION)" = "go1.27.0"
go build -trimpath -o "$binary" ./cmd/gooo
go build -trimpath -o "$reducer" ./cmd/meta-resource-budget-reducer
go build -trimpath -o "$consumer" ./cmd/meta-resource-budget-consumer

package_imported=false
command_imported=false
if rg -q 'internal/meta/languageresourcebudget"' "$root/internal/meta/languageresourcebudgetconsumer"; then package_imported=true; fi
if rg -q 'internal/meta/languageresourcebudget"' "$root/cmd/meta-resource-budget-consumer"; then command_imported=true; fi
package_files_scanned="$(rg --files "$root/internal/meta/languageresourcebudgetconsumer" -g '*.go' | wc -l | tr -d '[:space:]')"
command_files_scanned="$(rg --files "$root/cmd/meta-resource-budget-consumer" -g '*.go' | wc -l | tr -d '[:space:]')"
import_numerator=2
if [[ "$package_imported" == true ]]; then import_numerator=$((import_numerator - 1)); fi
if [[ "$command_imported" == true ]]; then import_numerator=$((import_numerator - 1)); fi
import_base="$(jq -nc --arg schema "gooo/meta-resource-budget-import-scan/v1" --argjson package_imported "$package_imported" --argjson command_imported "$command_imported" --argjson package_files "$package_files_scanned" --argjson command_files "$command_files_scanned" --argjson numerator "$import_numerator" '{schema:$schema,consumer_package_reducer_imported:$package_imported,consumer_command_reducer_imported:$command_imported,consumer_package_files_scanned:$package_files,consumer_command_files_scanned:$command_files,numerator:$numerator,denominator:2}')"
import_digest="sha256:$(printf '%s' "$import_base" | sha256sum | awk '{print $1}')"
import_scan="$(jq --arg digest "$import_digest" '. + {digest:$digest}' <<<"$import_base")"
printf '%s\n' "$import_scan" > "$work/import-scan.json"

contract_canonical="$(jq -c '{schema:.schema,id:.id,source_paths:.source_paths,samples_per_operation:.samples_per_operation,indicators:.indicators,operations:.operations,limits:.limits,not_claimed:.not_claimed,references:.references}' "$contract")"
contract_digest="sha256:$(printf '%s' "$contract_canonical" | sha256sum | awk '{print $1}')"

observation_for() {
  jq -r --arg id "$1" '.operations[] | select(.id == $id)' "$contract"
}

measure_one() {
  local label="$1" directory="$2" entry="$3" operation="$4" output_kind="$5" sequence="$6"
  local output="$work/$label-$operation-$sequence.json"
  local rss_file="$work/$label-$operation-$sequence.rss"
  local stderr_file="$work/$label-$operation-$sequence.stderr"
  local started finished exit_code wall receipt_bytes generated_bytes
  local stage step meta_operation proof_choice
  stage="$(observation_for "$operation" | jq -r '.stage')"
  step="$(observation_for "$operation" | jq -r '.step')"
  meta_operation="$(observation_for "$operation" | jq -r '.meta_operation')"
  proof_choice="$(observation_for "$operation" | jq -r '.proof_choice')"
  started="$(date +%s%N)"
  set +e
  if [[ "$operation" == "source-check" ]]; then
    (cd "$root" && /usr/bin/time -f '%M' -o "$rss_file" "$binary" check --json "$directory/activity.gooo" >"$output" 2>"$stderr_file")
  else
    (cd "$root" && /usr/bin/time -f '%M' -o "$rss_file" "$binary" emit --kind operation-manifest --entry "$entry" "$directory" >"$output" 2>"$stderr_file")
  fi
  exit_code=$?
  set -e
  finished="$(date +%s%N)"
  wall=$((finished-started))
  receipt_bytes=0
  generated_bytes=0
  if [[ "$output_kind" == "RECEIPT" ]]; then receipt_bytes="$(wc -c <"$output" | tr -d '[:space:]')"; fi
  if [[ "$output_kind" == "GENERATED" ]]; then generated_bytes="$(wc -c <"$output" | tr -d '[:space:]')"; fi
  jq -nc --arg schema "gooo/meta-resource-budget-observation/v1" --arg subject "$HEAD_SHA" --arg producer "scripts/meta-resource-budget" --arg consumer "cmd/meta-resource-budget-consumer" --arg operation "$operation" --argjson sequence "$sequence" --argjson exit_code "$exit_code" --arg stage "$stage" --arg step "$step" --arg meta_operation "$meta_operation" --arg proof_choice "$proof_choice" --arg reason "RUNNER_RESOURCE_OBSERVED" --argjson wall_time_ns "$wall" --argjson peak_rss_kib "$(tr -d '[:space:]' <"$rss_file")" --argjson receipt_bytes "$receipt_bytes" --argjson generated_bytes "$generated_bytes" --arg output_digest "sha256:$(sha256sum "$output" | awk '{print $1}')" --arg source_raw_digest "$SOURCE_RAW_DIGEST" --arg source_semantic_digest "$SOURCE_SEMANTIC_DIGEST" --arg entry_digest "$TARGET_DIGEST" --arg target_digest "$TARGET_DIGEST" '{schema:$schema,subject_sha:$subject,producer:$producer,consumer:$consumer,operation:$operation,stage:$stage,step:$step,meta_operation:$meta_operation,proof_choice:$proof_choice,reason:$reason,sequence:$sequence,exit_code:$exit_code,wall_time_ns:$wall_time_ns,peak_rss_kib:$peak_rss_kib,receipt_bytes:$receipt_bytes,generated_bytes:$generated_bytes,output_digest:$output_digest,source_raw_digest:$source_raw_digest,source_semantic_digest:$source_semantic_digest,entry_digest:$entry_digest,target_digest:$target_digest}' >> "$OBSERVATIONS"
  if [[ "$sequence" == "1" ]]; then
    jq -nc --arg operation "$operation" --argjson sequence 1 --arg kind "$output_kind" --arg payload_base64 "$(base64 -w0 "$output")" '{operation:$operation,sequence:$sequence,kind:$kind,payload_base64:$payload_base64}' >> "$RAW_OUTPUTS"
  fi
  test "$exit_code" = 0
}

capture_write_set() {
  local label="$1" operation="$2"
  working_tree_digest > "$work/$label-$operation-before.tree"
  (cd "$root" && git status --porcelain=v1 --untracked-files=all) > "$work/$label-$operation-before.status"
}

working_tree_digest() {
  local status_file="$work/working-tree-status.tmp" diff_file="$work/working-tree-diff.tmp"
  (cd "$root" && git status --porcelain=v1 --untracked-files=all) > "$status_file"
  (cd "$root" && git diff --no-ext-diff --binary HEAD) > "$diff_file"
  { cat "$status_file"; printf '\0'; cat "$diff_file"; } | sha256sum | awk '{print $1}'
}

finish_write_set() {
  local label="$1" operation="$2"
  local before_status="$work/$label-$operation-before.status"
  local after_status="$work/$label-$operation-after.status"
  local before_tree after_tree changed_paths_json untracked_file_count diff_exit_code write_set_digest
  working_tree_digest > "$work/$label-$operation-after.tree"
  (cd "$root" && git status --porcelain=v1 --untracked-files=all) > "$after_status"
  before_tree="$(cat "$work/$label-$operation-before.tree")"
  after_tree="$(cat "$work/$label-$operation-after.tree")"
  changed_paths_json="$(sed -E 's/^.. //' "$after_status" | jq -R -s 'split("\n") | map(select(length > 0))')"
  untracked_file_count="$(cd "$root" && git ls-files --others --exclude-standard | wc -l | tr -d '[:space:]')"
  diff_exit_code=0
  (cd "$root" && git diff --no-ext-diff --quiet --exit-code) || diff_exit_code=$?
  write_set_digest="sha256:$( { cat "$before_status"; printf '\0'; cat "$after_status"; } | sha256sum | awk '{print $1}')"
  jq -nc --arg operation "$operation" --arg schema "gooo/meta-resource-budget-write-set/v1" --arg producer "scripts/meta-resource-budget" --arg consumer "cmd/meta-resource-budget-consumer" --arg before "$before_tree" --arg after "$after_tree" --arg before_status "$(base64 -w0 "$before_status")" --arg after_status "$(base64 -w0 "$after_status")" --arg write_set "$write_set_digest" --argjson changed "$changed_paths_json" --argjson diff_exit "$diff_exit_code" --argjson untracked "$untracked_file_count" '{operation:$operation,schema:$schema,producer:$producer,consumer:$consumer,before_tree_digest:$before,after_tree_digest:$after,before_status_base64:$before_status,after_status_base64:$after_status,before_status_observed:true,after_status_observed:true,write_set_digest:$write_set,changed_paths:$changed,diff_exit_code:$diff_exit,untracked_file_count:$untracked,repository_writes:($changed|length),mutation_authority:false,authority_observed:true,sample_start:1,sample_end:3,reason:"NET_REPOSITORY_STATE_UNCHANGED_ACROSS_OPERATION_WINDOW"}' >> "$WRITE_SETS"
}

measure_group() {
  local label="$1" directory="$2" discovery="$3"
  local entry operation output_kind
  entry="$(jq -r '.activity' <<<"$discovery")"
  SOURCE_RAW_DIGEST="$(jq -r '.source_raw_digest' <<<"$discovery")"
  SOURCE_SEMANTIC_DIGEST="$(jq -r '.source_semantic_digest' <<<"$discovery")"
  TARGET_DIGEST="$(jq -r '.target_digest' <<<"$discovery")"
  OBSERVATIONS="$work/$label.observations.jsonl"
  RAW_OUTPUTS="$work/$label.raw.jsonl"
  WRITE_SETS="$work/$label.write-sets.jsonl"
  : > "$OBSERVATIONS"
  : > "$RAW_OUTPUTS"
  : > "$WRITE_SETS"
  while IFS= read -r operation; do
    output_kind="$(observation_for "$operation" | jq -r '.output')"
    capture_write_set "$label" "$operation"
    for sequence in 1 2 3; do measure_one "$label" "$directory" "$entry" "$operation" "$output_kind" "$sequence"; done
    finish_write_set "$label" "$operation"
  done < <(jq -r '.operations[].id' "$contract")
}

source_files_for() {
  local directory="$1"
  jq -nc --arg activity_path "$directory/activity.gooo" --arg activity "$(base64 -w0 "$directory/activity.gooo")" --arg entities_path "$directory/entities.gooo" --arg entities "$(base64 -w0 "$directory/entities.gooo")" '[{filename:$activity_path,content_base64:$activity},{filename:$entities_path,content_base64:$entities}]'
}

make_input() {
  local label="$1" evidence_class="$2" directory="$3" output="$4"
  local source_files observed raw_outputs write_sets source_receipt artifact replay effects runner_image_version
  source_files="$(source_files_for "$directory")"
  observed="$(jq -s . "$work/$label.observations.jsonl")"
  raw_outputs="$(jq -s . "$work/$label.raw.jsonl")"
  write_sets="$(jq -s . "$work/$label.write-sets.jsonl")"
  source_receipt="$work/$label-source-check-1.json"
  artifact="$work/$label-project-manifest-1.json"
  replay="$work/$label-replay-manifest-1.json"
  effects="$(jq -c '.effects // {repository_writes:0,mutation_authority:false}' "$artifact")"
  runner_image_version="${ImageVersion:-github-hosted-runner}"
  jq -n --arg schema "gooo/meta-resource-budget-input/v1" --arg head "$HEAD_SHA" --arg evidence_class "$evidence_class" --arg contract_digest "$contract_digest" --slurpfile contract "$contract" --argjson source_files "$source_files" --argjson raw_outputs "$raw_outputs" --argjson observations "$observed" --argjson write_sets "$write_sets" --argjson import_scan "$import_scan" --arg source_receipt_base64 "$(base64 -w0 "$source_receipt")" --arg artifact_base64 "$(base64 -w0 "$artifact")" --arg replay_base64 "$(base64 -w0 "$replay")" --arg source_digest "$SOURCE_RAW_DIGEST" --argjson effects "$effects" --arg runner_os "$(uname -s)" --arg runner_architecture "$(uname -m)" --arg runner_image "ubuntu-latest" --arg runner_image_version "$runner_image_version" --arg runner_go_version "$(go env GOVERSION)" '{schema:$schema,expected_head:$head,evidence_class:$evidence_class,contract_digest:$contract_digest,contract:$contract[0],producer:{source_receipt_base64:$source_receipt_base64,artifact_base64:$artifact_base64,replay_base64:$replay_base64,source_digest:$source_digest,source_files:$source_files,raw_outputs:$raw_outputs,source_file_count:($source_files|length),go_files:0,runner:{os:$runner_os,architecture:$runner_architecture,image:$runner_image,image_version:$runner_image_version,go_version:$runner_go_version},effects:$effects,write_sets:$write_sets,import_scan:$import_scan},observations:$observations}' > "$output"
}

discovery="$(cd "$root" && "$consumer" -source-dir "$example" -discover-json)"
measure_group current "$example" "$discovery"
make_input current CURRENT_EVIDENCE "$example" "$work/input.json"

"$reducer" -input "$work/input.json" -output "$work/normal-report.json" -case normal
"$reducer" -input "$work/input.json" -check "$work/normal-report.json" -case normal
"$consumer" -input "$work/input.json" -output "$work/current-consumer-report.json" -label CURRENT_EVIDENCE
jq -e '.decision=="PASS" and .resolution=="EXACT" and .summary.coordinates.satisfied==19 and .summary.coordinates.total==19 and .summary.coordinates.refuted==0 and .summary.coordinates.unknown==0 and .summary.operations==3 and .summary.samples==9 and .effects=={repository_writes:0,mutation_authority:false} and .claim_transitions[2].to=="DISCHARGED"' "$work/normal-report.json"
jq -e '.decision=="PASS" and .evidence_class=="CURRENT_EVIDENCE" and .semantic_decision=="PASS" and .resource.samples==9 and .resource.expected_samples==9 and .imports.numerator==2 and .imports.denominator==2 and .claim_transitions[2].to=="DISCHARGED"' "$work/current-consumer-report.json"

jq --argjson limit_ns "$(jq '.contract.limits.wall_time_ms*1000000+1' "$work/input.json")" '.evidence_class="SYNTHETIC_COUNTEREXAMPLE" | .observations |= map(if .operation=="source-check" and .sequence==1 then .wall_time_ns=$limit_ns else . end)' "$work/input.json" > "$work/over-budget-input.json"
set +e
"$reducer" -input "$work/over-budget-input.json" -output "$work/over-budget-report.json" -case over-budget
code=$?
set -e
test "$code" = 1
"$reducer" -input "$work/over-budget-input.json" -check "$work/over-budget-report.json" -case over-budget
set +e
"$consumer" -input "$work/over-budget-input.json" -output "$work/over-budget-consumer-report.json" -label SYNTHETIC_COUNTEREXAMPLE
code=$?
set -e
test "$code" = 1
jq -e '.decision=="FAIL_CLOSED" and .resolution=="EXACT" and .reason=="RESOURCE_BUDGET_EXCEEDED" and .evidence_class=="SYNTHETIC_COUNTEREXAMPLE" and .semantic.claim_state=="DISCHARGED" and .claim_transitions[1].to=="REFUTED" and .claim_transitions[2].to=="DISCHARGED"' "$work/over-budget-report.json"
jq -e '.decision=="FAIL_CLOSED" and .evidence_class=="SYNTHETIC_COUNTEREXAMPLE" and .semantic_decision=="PASS" and .resource.samples==9 and .claim_transitions[1].to=="REFUTED" and .claim_transitions[2].to=="DISCHARGED"' "$work/over-budget-consumer-report.json"

jq '.evidence_class="SYNTHETIC_COUNTEREXAMPLE" | .observations |= map(select(.operation!="replay-manifest" or .sequence!=3))' "$work/input.json" > "$work/missing-sample-input.json"
set +e
"$reducer" -input "$work/missing-sample-input.json" -output "$work/missing-sample-report.json" -case missing-sample
code=$?
set -e
test "$code" = 1
"$reducer" -input "$work/missing-sample-input.json" -check "$work/missing-sample-report.json" -case missing-sample
set +e
"$consumer" -input "$work/missing-sample-input.json" -output "$work/missing-sample-consumer-report.json" -label SYNTHETIC_COUNTEREXAMPLE
code=$?
set -e
test "$code" = 1
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .reason=="RESOURCE_SAMPLE_MISSING" and .evidence_class=="SYNTHETIC_COUNTEREXAMPLE" and .semantic.claim_state=="DISCHARGED" and ([.summary.resources[]|select(.operation=="source-check")][0].missing_samples==0)' "$work/missing-sample-report.json"
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .evidence_class=="SYNTHETIC_COUNTEREXAMPLE" and .semantic_decision=="PASS" and .resource.samples==8 and ([.resource.per_operation[]|select(.operation=="replay-manifest")][0].missing_samples==1)' "$work/missing-sample-consumer-report.json"
jq -e '[.indicators[] | select(.id|startswith("resource.source-check")) | .status] | all(. == "SATISFIED")' "$work/missing-sample-report.json"
jq -e '[.indicators[] | select(.id|startswith("resource.project-manifest")) | .status] | all(. == "SATISFIED")' "$work/missing-sample-report.json"
jq -e '[.indicators[] | select(.id|startswith("resource.replay-manifest")) | .status] | any(. == "UNKNOWN")' "$work/missing-sample-report.json"

semantic_dir="$work/intervention-semantic"
nonsemantic_dir="$work/intervention-nonsemantic"
mkdir -p "$semantic_dir" "$nonsemantic_dir"
cp "$root/$example/activity.gooo" "$semantic_dir/activity.gooo"
cp "$root/$example/entities.gooo" "$semantic_dir/entities.gooo"
cp "$root/$example/activity.gooo" "$nonsemantic_dir/activity.gooo"
cp "$root/$example/entities.gooo" "$nonsemantic_dir/entities.gooo"
sed -i 's/activity PayOrder/activity RefundOrder/' "$semantic_dir/activity.gooo"
printf '%s\n' '// presentation-only intervention' >> "$nonsemantic_dir/activity.gooo"
semantic_discovery="$("$consumer" -source-dir "$semantic_dir" -discover-json)"
nonsemantic_discovery="$("$consumer" -source-dir "$nonsemantic_dir" -discover-json)"
measure_group semantic-intervention "$semantic_dir" "$semantic_discovery"
make_input semantic-intervention INTERVENTION "$semantic_dir" "$work/semantic-intervention-input.json"
measure_group nonsemantic-intervention "$nonsemantic_dir" "$nonsemantic_discovery"
make_input nonsemantic-intervention INTERVENTION "$nonsemantic_dir" "$work/nonsemantic-intervention-input.json"
set +e
"$consumer" -input "$work/semantic-intervention-input.json" -output "$work/semantic-intervention-report.json" -label SEMANTIC_INTERVENTION
semantic_code=$?
"$consumer" -input "$work/nonsemantic-intervention-input.json" -output "$work/nonsemantic-intervention-report.json" -label NONSEMANTIC_INTERVENTION
nonsemantic_code=$?
set -e
test "$semantic_code" = 0 -o "$semantic_code" = 1
test "$nonsemantic_code" = 0 -o "$nonsemantic_code" = 1
jq -e --slurpfile baseline "$work/current-consumer-report.json" '.source.activity=="RefundOrder" and .artifact.activity=="RefundOrder" and .source.semantic_digest != $baseline[0].source.semantic_digest and .resource.samples==9 and .resource.expected_samples==9' "$work/semantic-intervention-report.json"
jq -e --slurpfile baseline "$work/current-consumer-report.json" '.source.source_digest != $baseline[0].source.source_digest and .source.semantic_digest == $baseline[0].source.semantic_digest and .source.activity == $baseline[0].source.activity and .artifact.activity == $baseline[0].artifact.activity and .resource.samples == 9 and .resource.expected_samples == 9 and .resource.resolution == $baseline[0].resource.resolution' "$work/nonsemantic-intervention-report.json"
jq -n --slurpfile current "$work/current-consumer-report.json" --slurpfile semantic "$work/semantic-intervention-report.json" --slurpfile nonsemantic "$work/nonsemantic-intervention-report.json" '{schema:"gooo/meta-resource-budget-interventions/v2",current_evidence:{source_raw_digest:$current[0].source.source_digest,source_semantic_digest:$current[0].source.semantic_digest,activity:$current[0].source.activity,operation_decision:$current[0].artifact.decision,samples:$current[0].resource.samples,resource_resolution:$current[0].resource.resolution},semantic_intervention:{source_raw_digest:$semantic[0].source.source_digest,source_semantic_digest:$semantic[0].source.semantic_digest,activity:$semantic[0].source.activity,operation_decision:$semantic[0].artifact.decision,semantic_digest_changed:($semantic[0].source.semantic_digest != $current[0].source.semantic_digest),target_digest_changed:($semantic[0].source.target_digest != $current[0].source.target_digest),samples:$semantic[0].resource.samples,resource_resolution:$semantic[0].resource.resolution},nonsemantic_intervention:{source_raw_digest:$nonsemantic[0].source.source_digest,source_semantic_digest:$nonsemantic[0].source.semantic_digest,activity:$nonsemantic[0].source.activity,operation_decision:$nonsemantic[0].artifact.decision,raw_digest_changed:($nonsemantic[0].source.source_digest != $current[0].source.source_digest),semantic_digest_preserved:($nonsemantic[0].source.semantic_digest == $current[0].source.semantic_digest),operation_preserved:($nonsemantic[0].source.activity == $current[0].source.activity),samples:$nonsemantic[0].resource.samples,resource_resolution:$nonsemantic[0].resource.resolution}}' > "$work/interventions.json"

(cd "$root" && git diff --exit-code)
{
  echo '### Meta resource budget experiment'
  echo
  echo '| evidence class | case | decision | resolution | semantic | resource interpretation |'
  echo '|---|---|---|---|---|---|'
  jq -r '"| \(.evidence_class) | normal | \(.decision) | \(.resolution) | \(.semantic.decision)/\(.semantic.resolution) | \(.interpretation) |"' "$work/normal-report.json"
  jq -r '"| \(.evidence_class) | over-budget | \(.decision) | \(.resolution) | \(.semantic.decision)/\(.semantic.resolution) | \(.interpretation) |"' "$work/over-budget-report.json"
  jq -r '"| \(.evidence_class) | missing-sample | \(.decision) | \(.resolution) | \(.semantic.decision)/\(.semantic.resolution) | \(.interpretation) |"' "$work/missing-sample-report.json"
  echo
  echo '**Current evidence:** 3 operations × 3 samples = 9 raw resource observations; 3 raw operation output payloads, each bound by all three repeated observations.'
  echo '**Synthetic counterexamples:** over-budget and missing-sample are derived without rerunning and excluded from current evidence.'
  echo '**Fixed denominator:** 9 applicable resource coordinates (wall/RSS for each operation, receipt only for source-check, generated only for project/replay) + 10 semantic/effect/binding/sample coordinates = 19; N/A axes are excluded.'
  echo '**Limits:** wall 2,000 ms; peak RSS 131,072 KiB; receipt bytes 8,192; generated bytes 16,384.'
  jq -r '"- runner: \(.summary.runner.os)/\(.summary.runner.architecture), image=\(.summary.runner.image) (\(.summary.runner.image_version)), Go=\(.summary.runner.go_version)"' "$work/normal-report.json"
  jq -r '.summary.resources[]|"- \(.operation): samples=\(.samples), missing=\(.missing_samples), invalid=\(.invalid_samples), wall max=\(.wall_max_ns) ns, peak RSS max=\(.peak_rss_max_kib) KiB, receipt max=\(.receipt_max_bytes) bytes, generated max=\(.generated_max_bytes) bytes"' "$work/normal-report.json"
  jq -r '.write_sets[]|"- write-set/\(.operation): before=\(.before_tree_digest), after=\(.after_tree_digest), diff_exit=\(.diff_exit_code), changed_paths=\(.changed_paths|length), digest=\(.write_set_digest), repository writes=\(.repository_writes), mutation authority=\(.mutation_authority), snapshots=\(.sample_start)-\(.sample_end)"' "$work/normal-report.json"
  echo '- claims: semantic meaning, runner resource envelope, and net repository state each carry stage/step/reason, evidence digest, and chained transition digest.'
  echo '- independent consumer: raw sources=2, raw operation outputs=3, raw resource observations=9; import independence is computed from import-scan.json rather than hardcoded.'
  jq -r '"- semantic intervention: activity=\(.semantic_intervention.activity), raw digest changed=\(.semantic_intervention.source_raw_digest != .current_evidence.source_raw_digest), semantic digest changed=\(.semantic_intervention.semantic_digest_changed), target changed=\(.semantic_intervention.target_digest_changed), samples=\(.semantic_intervention.samples)"' "$work/interventions.json"
  jq -r '"- nonsemantic intervention: raw digest changed=\(.nonsemantic_intervention.raw_digest_changed), semantic digest preserved=\(.nonsemantic_intervention.semantic_digest_preserved), operation preserved=\(.nonsemantic_intervention.operation_preserved), samples=\(.nonsemantic_intervention.samples)"' "$work/interventions.json"
} >> "$GITHUB_STEP_SUMMARY"
