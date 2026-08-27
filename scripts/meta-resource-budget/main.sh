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

if rg -n '"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageresourcebudget"|"github.com/kimjooyoon/meta-ontology-go/cmd/meta-resource-budget-reducer"' internal/meta/languageresourcebudgetconsumer cmd/meta-resource-budget-consumer; then
  echo 'independent consumer imports producer or reducer implementation' >&2
  exit 1
fi

source_hashes="$work/source-hashes.txt"
sha256sum "$root/$example/activity.gooo" "$root/$example/entities.gooo" | awk '{print $1}' > "$source_hashes"
source_digest="sha256:$(sha256sum "$source_hashes" | awk '{print $1}')"
source_files_json="$(jq -nc \
  --arg activity_path "$example/activity.gooo" --arg activity "$(base64 -w0 "$root/$example/activity.gooo")" \
  --arg entities_path "$example/entities.gooo" --arg entities "$(base64 -w0 "$root/$example/entities.gooo")" \
  '[{filename:$activity_path,content_base64:$activity},{filename:$entities_path,content_base64:$entities}]')"

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
    --arg producer "scripts/meta-resource-budget" --arg consumer "cmd/meta-resource-budget-consumer" \
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

entry="$($consumer -source-dir "$root/$example")"
for sequence in 1 2 3; do measure source-check RECEIPT "$sequence" check --json "$example/activity.gooo"; done
for sequence in 1 2 3; do measure project-manifest GENERATED "$sequence" emit --kind operation-manifest --entry "$entry" "$example"; done
for sequence in 1 2 3; do measure replay-manifest GENERATED "$sequence" emit --kind operation-manifest --entry "$entry" "$example"; done

before_tree_digest="$(cd "$root" && git rev-parse HEAD^{tree})"
status_file="$work/write-set-status.txt"
(cd "$root" && git status --porcelain=v1 --untracked-files=all) > "$status_file"
changed_paths_json="$(sed -E 's/^.. //' "$status_file" | jq -R -s 'split("\n") | map(select(length > 0))')"
untracked_file_count="$(cd "$root" && git ls-files --others --exclude-standard | wc -l | tr -d '[:space:]')"
diff_exit_code=0
(cd "$root" && git diff --no-ext-diff --quiet --exit-code) || diff_exit_code=$?
after_tree_digest="$(cd "$root" && git rev-parse HEAD^{tree})"
write_set_digest="sha256:$(sha256sum "$status_file" | awk '{print $1}')"
write_set_json="$(jq -nc --arg schema "gooo/meta-resource-budget-write-set/v1" --arg producer "scripts/meta-resource-budget" --arg consumer "cmd/meta-resource-budget-consumer" --arg before "$before_tree_digest" --arg after "$after_tree_digest" --arg write_set "$write_set_digest" --argjson changed "$changed_paths_json" --argjson diff_exit "$diff_exit_code" --argjson untracked "$untracked_file_count" --arg reason "GIT_DIFF_EXIT_0_AND_WRITE_SET_EMPTY" '{schema:$schema,producer:$producer,consumer:$consumer,before_tree_digest:$before,after_tree_digest:$after,write_set_digest:$write_set,changed_paths:$changed,diff_exit_code:$diff_exit,untracked_file_count:$untracked,repository_writes:($changed|length),mutation_authority:false,reason:$reason}')"

observed="$(jq -s . "$observations")"
source_receipt="$work/source-check-1.json"
artifact="$work/project-manifest-1.json"
replay="$work/replay-manifest-1.json"
effects_json="$(jq -c '.effects' "$artifact")"
jq -n --arg schema "gooo/meta-resource-budget-input/v1" --arg head "$HEAD_SHA" --arg evidence_class "CURRENT_EVIDENCE" \
  --slurpfile contract "$contract" --argjson source_files "$source_files_json" \
  --arg source_receipt_base64 "$(base64 -w0 "$source_receipt")" --arg artifact_base64 "$(base64 -w0 "$artifact")" --arg replay_base64 "$(base64 -w0 "$replay")" \
  --arg source_digest "$source_digest" --argjson observations "$observed" --argjson effects "$effects_json" --argjson write_set "$write_set_json" \
  --arg runner_os "$(uname -s)" --arg runner_architecture "$(uname -m)" --arg runner_image "ubuntu-latest" \
  --arg runner_image_version "${ImageVersion:-unknown}" --arg runner_go_version "$(go env GOVERSION)" --arg entry "$entry" \
  '{schema:$schema,expected_head:$head,evidence_class:$evidence_class,contract:$contract[0],producer:{source_receipt_base64:$source_receipt_base64,artifact_base64:$artifact_base64,replay_base64:$replay_base64,source_digest:$source_digest,source_files:$source_files,source_file_count:($source_files|length),go_files:0,runner:{os:$runner_os,architecture:$runner_architecture,image:$runner_image,image_version:$runner_image_version,go_version:$runner_go_version},effects:$effects,write_set:$write_set},observations:$observations,derived_entry:$entry}' \
  > "$work/input.json"

"$reducer" -input "$work/input.json" -output "$work/normal-report.json" -case normal
"$reducer" -input "$work/input.json" -check "$work/normal-report.json" -case normal
"$consumer" -input "$work/input.json" -output "$work/current-consumer-report.json" -label CURRENT_EVIDENCE
jq -e '.decision=="PASS" and .resolution=="EXACT" and .summary.coordinates=={satisfied:22,total:22,basis_points:10000} and .summary.operations==3 and .summary.samples==9 and .effects=={repository_writes:0,mutation_authority:false} and .claim_transitions[2].to=="DISCHARGED"' "$work/normal-report.json"
jq -e '.decision=="PASS" and .evidence_class=="CURRENT_EVIDENCE" and .semantic_decision=="PASS" and .resource.samples==9 and .resource.expected_samples==9 and .imports.numerator==2 and .imports.denominator==2 and .imports.independent==true and .claim_transitions[2].to=="DISCHARGED"' "$work/current-consumer-report.json"
test "$diff_exit_code" = 0

jq --argjson limit_ns "$(jq '.contract.limits.wall_time_ms*1000000+1' "$work/input.json")" '.evidence_class="SYNTHETIC_COUNTEREXAMPLE" | .observations[0].wall_time_ns=$limit_ns' "$work/input.json" > "$work/over-budget-input.json"
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
jq -e '.decision=="FAIL_CLOSED" and .resolution=="EXACT" and .reason=="RESOURCE_BUDGET_EXCEEDED" and .evidence_class=="SYNTHETIC_COUNTEREXAMPLE" and .semantic.decision=="PASS" and .interpretation=="SEMANTIC_EXACT_RESOURCE_CLAIM_REFUTED" and .claim_transitions[1].to=="REFUTED" and .claim_transitions[2].to=="DISCHARGED"' "$work/over-budget-report.json"
jq -e '.decision=="FAIL_CLOSED" and .evidence_class=="SYNTHETIC_COUNTEREXAMPLE" and .semantic_decision=="PASS" and .resource.samples==9 and .claim_transitions[1].to=="REFUTED" and .claim_transitions[2].to=="DISCHARGED"' "$work/over-budget-consumer-report.json"

jq '.evidence_class="SYNTHETIC_COUNTEREXAMPLE" | del(.observations[-1])' "$work/input.json" > "$work/missing-sample-input.json"
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
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .reason=="RESOURCE_SAMPLE_MISSING" and .evidence_class=="SYNTHETIC_COUNTEREXAMPLE" and .semantic.decision=="PASS" and .semantic.resolution=="EXACT" and .resource_resolution=="LOWER_RESOLUTION" and .claim_transitions[1].to=="OPEN" and .claim_transitions[2].to=="DISCHARGED"' "$work/missing-sample-report.json"
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .evidence_class=="SYNTHETIC_COUNTEREXAMPLE" and .semantic_decision=="PASS" and .resource.samples==8 and .claim_transitions[1].to=="OPEN" and .claim_transitions[2].to=="DISCHARGED"' "$work/missing-sample-consumer-report.json"

write_intervention_input() {
  local label="$1" directory="$2"
  local source_receipt_path="$work/$label-source-check.json"
  local artifact_path="$work/$label-project-manifest.json"
  local replay_path="$work/$label-replay-manifest.json"
  local hashes_path="$work/$label-source-hashes.txt"
  local source_digest_value source_files_value intervention_observations intervention_entry intervention_effects
  (cd "$work" && "$binary" check --json "$directory/activity.gooo" > "$source_receipt_path")
  intervention_entry="$($consumer -source-dir "$work/$directory")"
  (cd "$work" && "$binary" emit --kind operation-manifest --entry "$intervention_entry" "$directory" > "$artifact_path")
  (cd "$work" && "$binary" emit --kind operation-manifest --entry "$intervention_entry" "$directory" > "$replay_path")
  sha256sum "$work/$directory/activity.gooo" "$work/$directory/entities.gooo" | awk '{print $1}' > "$hashes_path"
  source_digest_value="sha256:$(sha256sum "$hashes_path" | awk '{print $1}')"
  source_files_value="$(jq -nc --arg activity_path "$directory/activity.gooo" --arg activity "$(base64 -w0 "$work/$directory/activity.gooo")" --arg entities_path "$directory/entities.gooo" --arg entities "$(base64 -w0 "$work/$directory/entities.gooo")" '[{filename:$activity_path,content_base64:$activity},{filename:$entities_path,content_base64:$entities}]')"
  intervention_observations="$(jq --arg source "sha256:$(sha256sum "$source_receipt_path" | cut -d' ' -f1)" --arg artifact "sha256:$(sha256sum "$artifact_path" | cut -d' ' -f1)" --arg replay "sha256:$(sha256sum "$replay_path" | cut -d' ' -f1)" '.observations | map(if .operation=="source-check" and .sequence==1 then .output_digest=$source elif .operation=="project-manifest" and .sequence==1 then .output_digest=$artifact elif .operation=="replay-manifest" and .sequence==1 then .output_digest=$replay else . end)' "$work/input.json")"
  intervention_effects="$(jq -c '.effects' "$artifact_path")"
  jq -n --arg schema "gooo/meta-resource-budget-input/v1" --arg head "$HEAD_SHA" --arg evidence_class "INTERVENTION" --slurpfile contract "$contract" --argjson source_files "$source_files_value" --arg source_receipt_base64 "$(base64 -w0 "$source_receipt_path")" --arg artifact_base64 "$(base64 -w0 "$artifact_path")" --arg replay_base64 "$(base64 -w0 "$replay_path")" --arg source_digest "$source_digest_value" --argjson observations "$intervention_observations" --argjson effects "$intervention_effects" --argjson write_set "$write_set_json" --arg runner_os "$(uname -s)" --arg runner_architecture "$(uname -m)" --arg runner_image "ubuntu-latest" --arg runner_image_version "${ImageVersion:-unknown}" --arg runner_go_version "$(go env GOVERSION)" --arg entry "$intervention_entry" '{schema:$schema,expected_head:$head,evidence_class:$evidence_class,contract:$contract[0],producer:{source_receipt_base64:$source_receipt_base64,artifact_base64:$artifact_base64,replay_base64:$replay_base64,source_digest:$source_digest,source_files:$source_files,source_file_count:($source_files|length),go_files:0,runner:{os:$runner_os,architecture:$runner_architecture,image:$runner_image,image_version:$runner_image_version,go_version:$runner_go_version},effects:$effects,write_set:$write_set},observations:$observations,derived_entry:$entry}' > "$work/$label-input.json"
}

semantic_dir="intervention-semantic"
nonsemantic_dir="intervention-nonsemantic"
mkdir -p "$work/$semantic_dir" "$work/$nonsemantic_dir"
cp "$root/$example/activity.gooo" "$work/$semantic_dir/activity.gooo"
cp "$root/$example/entities.gooo" "$work/$semantic_dir/entities.gooo"
cp "$root/$example/activity.gooo" "$work/$nonsemantic_dir/activity.gooo"
cp "$root/$example/entities.gooo" "$work/$nonsemantic_dir/entities.gooo"
sed -i 's/activity PayOrder/activity RefundOrder/' "$work/$semantic_dir/activity.gooo"
printf '%s\n' '// presentation-only intervention' >> "$work/$nonsemantic_dir/activity.gooo"
write_intervention_input semantic-intervention "$semantic_dir"
write_intervention_input nonsemantic-intervention "$nonsemantic_dir"
"$consumer" -input "$work/semantic-intervention-input.json" -output "$work/semantic-intervention-report.json" -label SEMANTIC_INTERVENTION
"$consumer" -input "$work/nonsemantic-intervention-input.json" -output "$work/nonsemantic-intervention-report.json" -label NONSEMANTIC_INTERVENTION
jq -e --slurpfile baseline "$work/current-consumer-report.json" '.source.activity=="RefundOrder" and .artifact.activity=="RefundOrder" and .source.semantic_digest != $baseline[0].source.semantic_digest and .resource.samples == $baseline[0].resource.samples and .resource.expected_samples == $baseline[0].resource.expected_samples and .claim_transitions[0].digest != $baseline[0].claim_transitions[0].digest' "$work/semantic-intervention-report.json"
jq -e --slurpfile baseline "$work/current-consumer-report.json" '.source.source_digest != $baseline[0].source.source_digest and .source.semantic_digest == $baseline[0].source.semantic_digest and .source.activity == $baseline[0].source.activity and .artifact.activity == $baseline[0].artifact.activity and .resource.samples == $baseline[0].resource.samples and .resource.expected_samples == $baseline[0].resource.expected_samples and .resource.resolution == $baseline[0].resource.resolution and .claim_transitions[0].digest == $baseline[0].claim_transitions[0].digest' "$work/nonsemantic-intervention-report.json"
jq -n --slurpfile current "$work/current-consumer-report.json" --slurpfile semantic "$work/semantic-intervention-report.json" --slurpfile nonsemantic "$work/nonsemantic-intervention-report.json" '{schema:"gooo/meta-resource-budget-interventions/v1",current_evidence:{source_digest:$current[0].source.source_digest,semantic_digest:$current[0].source.semantic_digest,activity:$current[0].source.activity,operation_decision:$current[0].artifact.decision,samples:$current[0].resource.samples,resource_resolution:$current[0].resource.resolution,claim_transition_digest:$current[0].claim_transitions[0].digest},semantic_intervention:{source_digest:$semantic[0].source.source_digest,semantic_digest:$semantic[0].source.semantic_digest,activity:$semantic[0].source.activity,operation_decision:$semantic[0].artifact.decision,target_changed:($semantic[0].source.activity != $current[0].source.activity),semantic_digest_changed:($semantic[0].source.semantic_digest != $current[0].source.semantic_digest),claim_transition_changed:($semantic[0].claim_transitions[0].digest != $current[0].claim_transitions[0].digest),samples:$semantic[0].resource.samples,resource_resolution:$semantic[0].resource.resolution,claim_transition_digest:$semantic[0].claim_transitions[0].digest},nonsemantic_intervention:{source_digest:$nonsemantic[0].source.source_digest,semantic_digest:$nonsemantic[0].source.semantic_digest,activity:$nonsemantic[0].source.activity,operation_decision:$nonsemantic[0].artifact.decision,raw_digest_changed:($nonsemantic[0].source.source_digest != $current[0].source.source_digest),semantic_digest_preserved:($nonsemantic[0].source.semantic_digest == $current[0].source.semantic_digest),operation_preserved:($nonsemantic[0].source.activity == $current[0].source.activity),claim_transition_preserved:($nonsemantic[0].claim_transitions[0].digest == $current[0].claim_transitions[0].digest),samples:$nonsemantic[0].resource.samples,resource_resolution:$nonsemantic[0].resource.resolution,claim_transition_digest:$nonsemantic[0].claim_transitions[0].digest}}' > "$work/interventions.json"

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
  echo '**Current evidence:** 3 operations × 3 samples = 9 raw observations; 22 fixed coordinates.'
  echo '**Synthetic counterexamples:** over-budget and missing-sample inputs are derived with jq and are not counted as current runner measurements.'
  echo '**Limits:** wall 2,000 ms; peak RSS 131,072 KiB; receipt bytes 8,192; generated bytes 16,384.'
  jq -r '"- runner: \(.summary.runner.os)/\(.summary.runner.architecture), image=\(.summary.runner.image) (\(.summary.runner.image_version)), Go=\(.summary.runner.go_version)"' "$work/normal-report.json"
  jq -r '.summary.resources[]|"- \(.operation): samples=\(.samples), wall max=\(.wall_max_ns) ns, peak RSS max=\(.peak_rss_max_kib) KiB, receipt max=\(.receipt_max) bytes, generated max=\(.generated_max) bytes"' "$work/normal-report.json"
  jq -r '"- write-set: before=\(.write_set.before_tree_digest), after=\(.write_set.after_tree_digest), diff_exit=\(.write_set.diff_exit_code), changed_paths=\(.write_set.changed_paths|length), digest=\(.write_set.write_set_digest), repository writes=\(.write_set.repository_writes), mutation authority=\(.write_set.mutation_authority)"' "$work/normal-report.json"
  echo '- claims: semantic/source replay, runner resource envelope, and read-only observation each carry evidence and chained transition digests.'
  echo '- independent consumer: raw sources=2, raw operation outputs=3, raw resource observations=9, producer/reducer imports=2/2 forbidden imports absent.'
  jq -r '"- semantic intervention: activity \(.semantic_intervention.activity), semantic digest changed=\(.semantic_intervention.semantic_digest_changed), target changed=\(.semantic_intervention.target_changed), claim transition changed=\(.semantic_intervention.claim_transition_changed), transition=\(.semantic_intervention.claim_transition_digest)"' "$work/interventions.json"
  jq -r '"- nonsemantic intervention: raw digest changed=\(.nonsemantic_intervention.raw_digest_changed), semantic digest preserved=\(.nonsemantic_intervention.semantic_digest_preserved), operation preserved=\(.nonsemantic_intervention.operation_preserved), claim transition preserved=\(.nonsemantic_intervention.claim_transition_preserved), samples=\(.nonsemantic_intervention.samples)"' "$work/interventions.json"
} >> "$GITHUB_STEP_SUMMARY"
