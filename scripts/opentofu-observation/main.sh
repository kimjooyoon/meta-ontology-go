#!/usr/bin/env bash
set -Eeuo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
out="${RUNNER_TEMP:-/tmp}/opentofu-observation"
work="${RUNNER_TEMP:-/tmp}/opentofu-observation-work"
fixture="$root/examples/opentofu-observation/fixture"
mkdir -p "$out/evidence" "$work"
cd "$root"
phase="release-download"
record_failure() {
  local status="$1" line="$2" command="$3"
  trap - ERR
  mkdir -p "$out/evidence"
  jq -n --arg phase "$phase" --arg step "$phase" --arg line "$line" --arg command "$command" --argjson exit_code "$status" \
    '{schema:"gooo/opentofu-observation-failure/v1",stage:"observe-released-cli",step:$step,phase:$phase,line:($line|tonumber),command:$command,reason:"OBSERVATION_COMMAND_FAILED",exit_code:$exit_code}' \
    > "$out/evidence/failure.json"
  printf 'observation_failure phase=%s line=%s command=%s status=%s\n' "$phase" "$line" "$command" "$status" >&2
  return "$status"
}
trap 'record_failure "$?" "$LINENO" "$BASH_COMMAND"' ERR

asset_url="https://github.com/opentofu/opentofu/releases/download/v1.12.6/tofu_1.12.6_linux_amd64.tar.gz"
asset_sha="sha256:50a6106fa4de523d09c87af85f3db1dd47535fc005727fdca6852146476b88ec"
asset_bytes_expected=34646566
sums_sha="sha256:6988e0cb8f4e9ebfa3b0999e44841549741b22d9b38873cb5b89074f1cddcb1c"
release_id="opentofu-v1.12.6"

digest() { printf 'sha256:%s' "$(sha256sum "$1" | cut -d' ' -f1)"; }

projection_digest() {
  printf '%s' "$1" | sha256sum | cut -d' ' -f1 | sed 's/^/sha256:/'
}

run_timed() {
  local name="$1" cwd_role="$2" cwd="$3" stdout="$4" stderr="$5" receipt="$6" descriptor="$7"
  shift 7
  local rss="${receipt}.rss" start_ns end_ns wall_ns wall_ms peak_rss status
  start_ns="$(date +%s%N)"
  set +e
  (cd "$cwd" && /usr/bin/time -q -f '%M' -o "$rss" "$@") >"$stdout" 2>"$stderr"
  status=$?
  set -e
  end_ns="$(date +%s%N)"
  wall_ns=$((end_ns - start_ns))
  wall_ms=$(((wall_ns + 999999) / 1000000))
  if ((wall_ns > 0 && wall_ms == 0)); then wall_ms=1; fi
  peak_rss="$(tail -n 1 "$rss" 2>/dev/null || true)"
  test "$wall_ns" -gt 0
  test "$wall_ms" -gt 0
  [[ "$peak_rss" =~ ^[0-9]+$ ]]
  jq -n --arg name "$name" --arg phase "$phase" --arg cwd_role "$cwd_role" --argjson command "$descriptor" \
    --argjson exit_code "$status" --argjson stdout_bytes "$(wc -c <"$stdout" | tr -d ' ')" \
    --arg stdout_digest "$(digest "$stdout")" --argjson stderr_bytes "$(wc -c <"$stderr" | tr -d ' ')" \
    --arg stderr_digest "$(digest "$stderr")" --argjson wall_ms "$wall_ms" --argjson peak_rss_kib "$peak_rss" \
    '{name:$name,phase:$phase,command:$command,cwd_role:$cwd_role,exit_code:$exit_code,stdout_bytes:$stdout_bytes,stdout_digest:$stdout_digest,stderr_bytes:$stderr_bytes,stderr_digest:$stderr_digest,wall_ms:$wall_ms,peak_rss_kib:$peak_rss_kib,executed:true}' >"$receipt"
  printf 'command_observation name=%s exit_code=%s stdout_bytes=%s stderr_bytes=%s\n' \
    "$name" "$status" "$(wc -c <"$stdout" | tr -d ' ')" "$(wc -c <"$stderr" | tr -d ' ')"
  rm -f "$rss"
}

manifest_digest() {
  local directory="$1" manifest="$2"
  : > "$manifest"
	while IFS= read -r -d '' file; do
		printf '%s\t' "${file#"$directory"/}" >> "$manifest"
		printf '%s\n' "$(digest "$file")" >> "$manifest"
	done < <(find "$directory" -type f -print0 | sort -z)
  digest "$manifest"
}

curl -fsSL "$asset_url" -o "$work/tofu.tar.gz"
curl -fsSL "${asset_url%/*}/tofu_1.12.6_SHA256SUMS" -o "$work/SHA256SUMS"
test "$(digest "$work/tofu.tar.gz")" = "$asset_sha"
test "$(digest "$work/SHA256SUMS")" = "$sums_sha"
grep -F "50a6106fa4de523d09c87af85f3db1dd47535fc005727fdca6852146476b88ec  tofu_1.12.6_linux_amd64.tar.gz" "$work/SHA256SUMS"
asset_bytes="$(wc -c <"$work/tofu.tar.gz" | tr -d ' ')"
test "$asset_bytes" -eq "$asset_bytes_expected"
tar -xzf "$work/tofu.tar.gz" -C "$work"
tofu="$work/tofu"
chmod +x "$tofu"

go_version="$(go version)"
go_env_version="$(go env GOVERSION)"
test "$go_env_version" = "go1.27.0"
printf '%s\n%s\n' "$go_version" "$go_env_version" > "$work/observer-toolchain.txt"
observer_toolchain_digest="$(digest "$work/observer-toolchain.txt")"

build_descriptor='["go","build","-trimpath","-o","<temp-output>","./cmd/opentofu-observation-witness"]'
run_timed "consumer-build" "repository" "$root" "$work/consumer-build.stdout" "$work/consumer-build.stderr" "$work/consumer-build.json" "$build_descriptor" \
  go build -trimpath -o "$work/opentofu-observation-witness" ./cmd/opentofu-observation-witness
jq -e '.exit_code == 0' "$work/consumer-build.json"
witness="$work/opentofu-observation-witness"

phase="release-identity"
version_descriptor='["tofu","version","-json"]'
run_timed "tofu-version" "release" "$work" "$work/version.stdout" "$work/version.stderr" "$work/version.json" "$version_descriptor" "$tofu" version -json
jq -e '.exit_code == 0' "$work/version.json"
version_json="$(jq -c . "$work/version.stdout")"
version_projection="$(jq -cS . "$work/version.stdout")"
version_digest="$(projection_digest "$version_projection")"
version="$(jq -er '.terraform_version' "$work/version.stdout")"
platform="$(jq -er '.platform' "$work/version.stdout")"
printf 'release_observation version=%s platform=%s\n' "$version" "$platform"
test "$version" = "1.12.6"
test "$platform" = "linux_amd64"

phase="fixture-input"
fixture_digest="$(manifest_digest "$fixture" "$work/fixture.manifest")"
mapfile -t fixture_files < <(find "$fixture" -type f -printf '%P\n' | sort)
input_files="$(find "$fixture" -type f | wc -l | tr -d ' ')"
input_lines=0
for file in "${fixture_files[@]}"; do input_lines=$((input_lines + $(wc -l <"$fixture/$file" | tr -d ' '))); done
test "$input_files" -gt 0
test "$input_lines" -gt 0

init_descriptor='["tofu","init","-backend=false"]'
plan_descriptor='["tofu","plan","-refresh=false","-input=false","-out=plan.bin","-json"]'
show_descriptor='["tofu","show","-json","-plan=plan.bin"]'
test_descriptor='["tofu","test","-json-into=test-events.jsonl","-no-color","-test-directory","tests"]'

run_once() {
  local number="$1" directory="$work/run-$1"
  cp -R "$fixture" "$directory"
  phase="fixture-run-$number"
  printf 'observation_stage=fixture-run-%s\n' "$number"
  run_timed "tofu-init" "fixture-run-$number" "$directory" "$directory/init.stdout" "$directory/init.stderr" "$directory/init.json" "$init_descriptor" "$tofu" init -backend=false
  run_timed "tofu-plan" "fixture-run-$number" "$directory" "$directory/plan.stdout" "$directory/plan.stderr" "$directory/plan.json" "$plan_descriptor" "$tofu" plan -refresh=false -input=false -out=plan.bin -json
  run_timed "tofu-show" "fixture-run-$number" "$directory" "$directory/show.stdout" "$directory/show.stderr" "$directory/show.json" "$show_descriptor" "$tofu" show -json -plan=plan.bin
  run_timed "tofu-test" "fixture-run-$number" "$directory" "$directory/test.stdout" "$directory/test.stderr" "$directory/test.json" "$test_descriptor" "$tofu" test -json-into=test-events.jsonl -no-color -test-directory tests
  for receipt in init plan show test; do jq -e '.exit_code == 0' "$directory/$receipt.json"; done
  jq -e 'type == "object" and (.format_version | type) == "string" and has("planned_values")' "$directory/show.stdout"
  raw_plan_digest="$(digest "$directory/show.stdout")"
  jq -cS 'del(.timestamp)' "$directory/show.stdout" > "$directory/plan.canonical.json"
  plan_digest="$(digest "$directory/plan.canonical.json")"
  test_types="$(jq -cs 'map(.type) | sort | group_by(.) | map({key:.[0],value:length}) | from_entries' "$directory/test-events.jsonl")"
  abstract_discovered="$(jq -cs '[.[] | select(.type=="test_abstract") | .test_abstract | to_entries[] | (.value | length)] | add // 0' "$directory/test-events.jsonl")"
  run_executed="$(jq -cs 'map(select(.type=="test_run" and (.test_run | type)=="object")) | length' "$directory/test-events.jsonl")"
  summary_passed="$(jq -cs 'map(select(.type=="test_summary") | .test_summary)[0].passed // 0' "$directory/test-events.jsonl")"
  summary_failed="$(jq -cs 'map(select(.type=="test_summary") | .test_summary)[0].failed // 0' "$directory/test-events.jsonl")"
  summary_errored="$(jq -cs 'map(select(.type=="test_summary") | .test_summary)[0].errored // 0' "$directory/test-events.jsonl")"
  summary_skipped="$(jq -cs 'map(select(.type=="test_summary") | .test_summary)[0].skipped // 0' "$directory/test-events.jsonl")"
  test_count="$(jq -cs 'length' "$directory/test-events.jsonl")"
  test -n "$test_types"
  test "$test_count" -eq 5
  test "$abstract_discovered" -eq 1
  test "$run_executed" -eq 1
  test "$summary_passed" -eq 1
  test "$summary_failed" -eq 0
  test "$summary_errored" -eq 0
  test "$summary_skipped" -eq 0
  test -s "$directory/test-events.jsonl"
  : > "$directory/test.normalized"
  while IFS= read -r line; do
	    jq -cS 'if type == "object" then del(."@timestamp",.timestamp,.elapsed_seconds) else error end' <<<"$line" >> "$directory/test.normalized"
  done < "$directory/test-events.jsonl"
  sort "$directory/test.normalized" -o "$directory/test.normalized"
  test -s "$directory/test.normalized"
  jq -n --argjson index "$number" --arg fixture "$fixture_digest" \
    --arg plan "$plan_digest" --arg plan_raw "$raw_plan_digest" --arg plan_canonicalizer "opentofu-plan-json/v1" --arg plan_canonicalizer_digest "$(projection_digest 'opentofu-plan-json/v1')" --argjson plan_volatile_fields '["timestamp"]' \
    --argjson plan_bytes "$(wc -c <"$directory/show.stdout" | tr -d ' ')" \
    --argjson plan_valid true --arg test_events "$(digest "$directory/test.normalized")" \
    --arg test_raw "$(digest "$directory/test-events.jsonl")" --argjson test_count "$test_count" --argjson test_types "$test_types" \
    --argjson abstract_discovered "$abstract_discovered" --argjson run_executed "$run_executed" --argjson summary_passed "$summary_passed" \
    --argjson summary_failed "$summary_failed" --argjson summary_errored "$summary_errored" --argjson summary_skipped "$summary_skipped" \
    --argjson test_valid true --slurpfile init "$directory/init.json" --slurpfile plan_command "$directory/plan.json" \
    --slurpfile show "$directory/show.json" --slurpfile test "$directory/test.json" \
    '{index:$index,fixture_digest:$fixture,plan_json_digest:$plan,plan_raw_digest:$plan_raw,plan_canonicalizer:$plan_canonicalizer,plan_canonicalizer_digest:$plan_canonicalizer_digest,plan_volatile_fields:$plan_volatile_fields,plan_json_bytes:$plan_bytes,plan_schema_valid:$plan_valid,test_event_digest:$test_events,test_raw_digest:$test_raw,test_event_count:$test_count,test_type_counts:$test_types,test_abstract_discovered:$abstract_discovered,test_run_executed:$run_executed,test_summary_passed:$summary_passed,test_summary_failed:$summary_failed,test_summary_errored:$summary_errored,test_summary_skipped:$summary_skipped,test_events_valid:$test_valid,commands:[$init[0],$plan_command[0],$show[0],$test[0]]}' > "$directory/run.json"
  cp "$directory/show.stdout" "$out/evidence/plan-$number.json"
  cp "$directory/plan.canonical.json" "$out/evidence/plan-$number.canonical.json"
  cp "$directory/test-events.jsonl" "$out/evidence/test-$number.ndjson"
}

run_once 1
run_once 2

cli_receipt_ledger="$(jq -s '.' "$work/version.json" "$work/run-1/init.json" "$work/run-1/plan.json" "$work/run-1/show.json" "$work/run-1/test.json" "$work/run-2/init.json" "$work/run-2/plan.json" "$work/run-2/show.json" "$work/run-2/test.json")"
jq -e 'all(.[]; .executed == true and (.phase | type) == "string" and (.phase | length) > 0)' <<<"$cli_receipt_ledger"
baseline_physical_command_executions="$(jq -er '[.[] | select(.executed == true and (.name | startswith("tofu-")) and ((.phase | startswith("reuse")) | not))] | length' <<<"$cli_receipt_ledger")"
baseline_physical_test_executions="$(jq -er '[.[] | select(.executed == true and .name == "tofu-test" and ((.phase | startswith("reuse")) | not))] | length' <<<"$cli_receipt_ledger")"
reuse_physical_command_executions="$(jq -er '[.[] | select(.executed == true and (.name | startswith("tofu-")) and (.phase | startswith("reuse")))] | length' <<<"$cli_receipt_ledger")"
reuse_physical_test_executions="$(jq -er '[.[] | select(.executed == true and .name == "tofu-test" and (.phase | startswith("reuse")))] | length' <<<"$cli_receipt_ledger")"
jq -e --argjson commands "$baseline_physical_command_executions" --argjson tests "$baseline_physical_test_executions" --argjson reuse_commands "$reuse_physical_command_executions" --argjson reuse_tests "$reuse_physical_test_executions" '($commands == 9 and $tests == 2 and $reuse_commands == 0 and $reuse_tests == 0)' <<<"$cli_receipt_ledger"

go run ./cmd/gooo check examples/opentofu-observation/main.gooo
go run ./cmd/gooo graph dump examples/opentofu-observation/main.gooo > "$work/gooo-graph.json"
program_digest="sha256:$(jq -er '.source_digest' "$work/gooo-graph.json")"
graph_hash="$(jq -er '.graph_hash' "$work/gooo-graph.json")"
graph_bindings="$(jq -n --slurpfile graph "$work/gooo-graph.json" '
  $graph[0] as $g |
  [{"cell_id":"OPENTOFU_RELEASE_PIN","name":"PinOpenTofuRelease"},{"cell_id":"ASSET_CHECKSUM","name":"VerifyOpenTofuAssetChecksum"},{"cell_id":"CLI_VERSION_JSON","name":"ObserveOpenTofuVersionJSON"},{"cell_id":"FIXTURE_INPUT_DIGEST","name":"PinOpenTofuFixtureInput"},{"cell_id":"PLAN_JSON_SCHEMA","name":"ValidateOpenTofuPlanJSON"},{"cell_id":"TEST_EVENT_INVENTORY","name":"AccountOpenTofuTestEvents"},{"cell_id":"COMMAND_RUNTIME_RECEIPT","name":"RecordOpenTofuCommandRuntime"},{"cell_id":"PEAK_RSS_OBSERVATION","name":"ObserveOpenTofuPeakRSS"},{"cell_id":"DETERMINISTIC_PLAN_REPLAY","name":"ReplayOpenTofuPlanJSON"},{"cell_id":"DETERMINISTIC_TEST_REPLAY","name":"ReplayOpenTofuTestEvents"},{"cell_id":"REUSE_ELIGIBILITY","name":"EvaluateOpenTofuTestReuse"},{"cell_id":"HUMAN_REPORT","name":"PublishOpenTofuObservationReport"}] |
  map(. as $cell | [$g.nodes[] | select(.kind=="Activity" and .name==$cell.name)] as $activity |
    [$g.relations[] | select(.predicate=="used" and .subject==$activity[0].id)] as $used |
    [$g.relations[] | select(.predicate=="wasGeneratedBy" and .object==$activity[0].id)] as $generated |
    {cell_id:$cell.cell_id,activity_id:$activity[0].id,input_id:$used[0].object,output_id:$generated[0].subject,used_edge_count:($used|length),generated_edge_count:($generated|length)})
')"
jq -e --argjson bindings "$graph_bindings" '($bindings|length)==12 and all($bindings[]; .activity_id != null and .input_id != null and .output_id != null and .used_edge_count == 1 and .generated_edge_count == 1)' <<<"$graph_bindings"
cp "$work/gooo-graph.json" "$out/evidence/gooo-graph.json"
cp "$work/version.stdout" "$out/evidence/version.json"
cp "$work/SHA256SUMS" "$out/evidence/SHA256SUMS"
evidence_files_before_reports="$(find "$out/evidence" -type f | wc -l | tr -d ' ')"
artifact_root_members=("report.json" "observation.json")
planned_evidence_members=("baseline-observation.json" "baseline-report.json" "contract.json" "main.gooo" "report.json" "observation.json")
pending_evidence_files=0
for member in "${planned_evidence_members[@]}"; do
  if [[ ! -e "$out/evidence/$member" ]]; then pending_evidence_files=$((pending_evidence_files + 1)); fi
done
output_artifact_files=$((evidence_files_before_reports + pending_evidence_files + ${#artifact_root_members[@]}))
graph_activity_count="$(jq -er '[.nodes[] | select(.kind=="Activity")] | length' "$work/gooo-graph.json")"
graph_edge_count="$(jq -er '.relations | length' "$work/gooo-graph.json")"
test "$graph_activity_count" -eq 12

runtime="$(jq -n --slurpfile build "$work/consumer-build.json" --slurpfile version "$work/version.json" --slurpfile first "$work/run-1/run.json" --slurpfile second "$work/run-2/run.json" '
  [$build[0],$version[0],$first[0].commands[],$second[0].commands[]] as $all |
  ($first[0].commands | map(select(.name=="tofu-init"))[0]) as $init |
  ($first[0].commands | map(select(.name=="tofu-plan"))[0]) as $plan |
  ($first[0].commands | map(select(.name=="tofu-show"))[0]) as $show |
  ($first[0].commands | map(select(.name=="tofu-test"))[0]) as $test |
  {consumer_build_ms:$build[0].wall_ms,consumer_build_peak_rss_kib:$build[0].peak_rss_kib,tofu_init_ms:$init.wall_ms,tofu_init_peak_rss_kib:$init.peak_rss_kib,tofu_plan_ms:$plan.wall_ms,tofu_plan_peak_rss_kib:$plan.peak_rss_kib,tofu_show_ms:$show.wall_ms,tofu_show_peak_rss_kib:$show.peak_rss_kib,tofu_test_ms:$test.wall_ms,tofu_test_peak_rss_kib:$test.peak_rss_kib,tofu_test_executions:([$all[] | select(.name=="tofu-test")]|length),total_wall_ms:([$all[].wall_ms]|add),max_peak_rss_kib:([$all[].peak_rss_kib]|max)}')"

source_digest="$(printf '%s\n' "$(git rev-parse HEAD^{tree})" | sha256sum | cut -d' ' -f1)"
source_digest="sha256:$source_digest"
argument_digest="$(printf '%s\n' "$init_descriptor" "$plan_descriptor" "$show_descriptor" "$test_descriptor" | sha256sum | cut -d' ' -f1)"
argument_digest="sha256:$argument_digest"
environment_digest="$(printf '%s\n' 'GOTOOLCHAIN=go1.27.0' 'GOWORK=off' 'GOFLAGS=-mod=readonly' 'TF_IN_AUTOMATION=1' | sha256sum | cut -d' ' -f1)"
environment_digest="sha256:$environment_digest"
dependency_digest="$(digest "$work/run-1/show.stdout")"
expected_digest="$(printf '%s\n' 'init=0' 'plan=0' 'show=0' 'test=0' 'plan_events=valid' | sha256sum | cut -d' ' -f1)"
expected_digest="sha256:$expected_digest"
toolchain_digest="$observer_toolchain_digest"

first_plan_digest="$(jq -er '.plan_json_digest' "$work/run-1/run.json")"
second_plan_digest="$(jq -er '.plan_json_digest' "$work/run-2/run.json")"
first_test_digest="$(jq -er '.test_event_digest' "$work/run-1/run.json")"
second_test_digest="$(jq -er '.test_event_digest' "$work/run-2/run.json")"
release_pin_projection="release_id=$release_id;asset_url=$asset_url;asset_sha=$asset_sha;asset_bytes=$asset_bytes;sums_sha=$sums_sha"
asset_checksum_projection="asset_sha=$asset_sha;sums_sha=$sums_sha;asset_bytes=$asset_bytes"
fixture_projection="fixture_digest=$fixture_digest;fixture_files=$(printf '%s,' "${fixture_files[@]}");fixture_lines=$input_lines"
plan_schema_projection="plan_digest=$first_plan_digest;format_version=$(jq -er '.format_version' "$work/run-1/show.stdout");planned_values=$(jq -cS '.planned_values' "$work/run-1/show.stdout")"
test_inventory_projection="$(jq -cS '{test_type_counts,test_abstract_discovered,test_run_executed,test_summary_passed,test_summary_failed,test_summary_errored,test_summary_skipped}' "$work/run-1/run.json")"
command_runtime_projection="$(jq -cS '.commands' "$work/run-1/run.json")"
peak_rss_projection="$(jq -cS '[.commands[] | {name,peak_rss_kib}]' "$work/run-1/run.json")"
plan_replay_projection="plan=$first_plan_digest|$second_plan_digest"
test_replay_projection="test=$first_test_digest|$second_test_digest"
reuse_projection="decision=NOT_REUSED_FIRST_RUN;discovered=1;executed=1;reused=0;prior=0;invalidated=0;release=$asset_sha;toolchain=$toolchain_digest"
human_report_projection="head=$HEAD_SHA;paths=3;cells=12;runtime=$(jq -cS . <<<"$runtime")"
cell_evidence="$(jq -n \
  --arg release "$release_pin_projection" --arg checksum "$asset_checksum_projection" --arg version "$version_projection" \
  --arg fixture "$fixture_projection" --arg plan_schema "$plan_schema_projection" --arg test_inventory "$test_inventory_projection" \
  --arg runtime "$command_runtime_projection" --arg rss "$peak_rss_projection" --arg plan_replay "$plan_replay_projection" \
  --arg test_replay "$test_replay_projection" --arg reuse "$reuse_projection" --arg report "$human_report_projection" \
  '{OPENTOFU_RELEASE_PIN:$release,ASSET_CHECKSUM:$checksum,CLI_VERSION_JSON:$version,FIXTURE_INPUT_DIGEST:$fixture,PLAN_JSON_SCHEMA:$plan_schema,TEST_EVENT_INVENTORY:$test_inventory,COMMAND_RUNTIME_RECEIPT:$runtime,PEAK_RSS_OBSERVATION:$rss,DETERMINISTIC_PLAN_REPLAY:$plan_replay,DETERMINISTIC_TEST_REPLAY:$test_replay,REUSE_ELIGIBILITY:$reuse,HUMAN_REPORT:$report}')"
release_cell_digest="$(projection_digest "$release_pin_projection")"
checksum_cell_digest="$(projection_digest "$asset_checksum_projection")"
version_cell_digest="$(projection_digest "$version_projection")"
fixture_cell_digest="$(projection_digest "$fixture_projection")"
plan_schema_cell_digest="$(projection_digest "$plan_schema_projection")"
test_inventory_cell_digest="$(projection_digest "$test_inventory_projection")"
runtime_cell_digest="$(projection_digest "$command_runtime_projection")"
rss_cell_digest="$(projection_digest "$peak_rss_projection")"
plan_replay_cell_digest="$(projection_digest "$plan_replay_projection")"
test_replay_cell_digest="$(projection_digest "$test_replay_projection")"
reuse_cell_digest="$(projection_digest "$reuse_projection")"
human_report_cell_digest="$(projection_digest "$human_report_projection")"
cell_evidence_digests="$(jq -n \
  --arg release "$release_cell_digest" --arg checksum "$checksum_cell_digest" --arg version "$version_cell_digest" \
  --arg fixture "$fixture_cell_digest" --arg plan_schema "$plan_schema_cell_digest" --arg test_inventory "$test_inventory_cell_digest" \
  --arg runtime "$runtime_cell_digest" --arg rss "$rss_cell_digest" --arg plan_replay "$plan_replay_cell_digest" \
  --arg test_replay "$test_replay_cell_digest" --arg reuse "$reuse_cell_digest" --arg report "$human_report_cell_digest" \
	  '{OPENTOFU_RELEASE_PIN:$release,ASSET_CHECKSUM:$checksum,CLI_VERSION_JSON:$version,FIXTURE_INPUT_DIGEST:$fixture,PLAN_JSON_SCHEMA:$plan_schema,TEST_EVENT_INVENTORY:$test_inventory,COMMAND_RUNTIME_RECEIPT:$runtime,PEAK_RSS_OBSERVATION:$rss,DETERMINISTIC_PLAN_REPLAY:$plan_replay,DETERMINISTIC_TEST_REPLAY:$test_replay,REUSE_ELIGIBILITY:$reuse,HUMAN_REPORT:$report}')"

decision_projection_descriptor='["jq","-n","opentofu-reuse-decision-projection"]'
phase="baseline-decision"
run_timed "baseline-decision-projection" "caller-owned-observation" "$work" "$work/baseline-decision.stdout" "$work/baseline-decision.stderr" "$work/baseline-decision.json" "$decision_projection_descriptor" \
	jq -n --argjson commands "$baseline_physical_command_executions" --argjson tests "$baseline_physical_test_executions" \
		'{physical_command_executions:$commands,physical_test_executions:$tests}'
jq -e --argjson commands "$baseline_physical_command_executions" --argjson tests "$baseline_physical_test_executions" \
		'.physical_command_executions == $commands and .physical_test_executions == $tests' "$work/baseline-decision.stdout"
baseline_decision_wall_ms="$(jq -er '.wall_ms' "$work/baseline-decision.json")"
baseline_decision_peak_rss="$(jq -er '.peak_rss_kib' "$work/baseline-decision.json")"

jq -n --arg schema "gooo/opentofu-observation/v1" --arg contract_id "opentofu-released-cli-observation-v1" --arg subject "$HEAD_SHA" \
  --arg release_id "$release_id" --arg asset_url "$asset_url" --arg asset_sha "$asset_sha" --argjson asset_bytes "$asset_bytes" --arg sums_sha "$sums_sha" \
  --argjson version_json "$version_json" --arg version_digest "$version_digest" --arg version "$version" --arg platform "$platform" --slurpfile version_command "$work/version.json" \
  --arg fixture_digest "$fixture_digest" --argjson fixture_files "$(printf '%s\n' "${fixture_files[@]}" | jq -R . | jq -s .)" --argjson fixture_lines "$input_lines" \
  --slurpfile first "$work/run-1/run.json" --slurpfile second "$work/run-2/run.json" --arg source_digest "$source_digest" --arg argument_digest "$argument_digest" --arg environment_digest "$environment_digest" --arg release_digest "$asset_sha" --arg toolchain_digest "$toolchain_digest" --arg dependency_digest "$dependency_digest" --arg expected_digest "$expected_digest" \
	--argjson runtime "$runtime" --argjson input_files "$input_files" --argjson output_files "$output_artifact_files" --argjson reusable_files 0 --arg go_version "$go_version" --arg goversion "$go_env_version" --arg observer_digest "$observer_toolchain_digest" \
	--argjson baseline_commands "$baseline_physical_command_executions" --argjson baseline_tests "$baseline_physical_test_executions" --argjson reuse_commands "$reuse_physical_command_executions" --argjson reuse_tests "$reuse_physical_test_executions" --argjson decision_wall_ms "$baseline_decision_wall_ms" --argjson decision_peak_rss "$baseline_decision_peak_rss" \
	--arg program_digest "$program_digest" --arg graph_hash "$graph_hash" --argjson graph_activity_count "$graph_activity_count" --argjson graph_edge_count "$graph_edge_count" --argjson graph_nodes "$(jq -c '.nodes' "$work/gooo-graph.json")" --argjson graph_relations "$(jq -c '.relations' "$work/gooo-graph.json")" --argjson bindings "$graph_bindings" --argjson cell_evidence_projections "$cell_evidence" --argjson cell_evidence_digests "$cell_evidence_digests" \
	'{schema:$schema,contract_id:$contract_id,subject_sha:$subject,user_paths:["P1 RELEASE_IDENTITY","P2 PLAN_JSON","P3 TEST_JSON"],release:{release_id:$release_id,asset_url:$asset_url,asset_sha256:$asset_sha,asset_bytes:$asset_bytes,checksums_sha256:$sums_sha,version_json:$version_json,version_json_digest:$version_digest,version:$version,platform:$platform,command:$version_command[0]},fixture_digest:$fixture_digest,fixture_files:$fixture_files,fixture_physical_lines:$fixture_lines,executions:[$first[0],$second[0]],reuse:{request_mode:"BASELINE",requests:1,discovered:1,executed:1,reused:0,skipped:0,prior_candidates:0,invalidated:0,decision:"EXECUTE",reason:"NO_PRIOR_RECEIPT",source_digest:$source_digest,fixture_digest:$fixture_digest,argument_digest:$argument_digest,environment_allowlist_digest:$environment_digest,release_digest:$asset_sha,observer_toolchain_digest:$toolchain_digest,dependency_graph_digest:$dependency_digest,expected_result_digest:$expected_digest,prior_receipt_digest:"",prior_receipt_file_digest:"",prior_artifact_manifest_digest:"",baseline_physical_command_executions:$baseline_commands,baseline_physical_test_executions:$baseline_tests,reuse_physical_command_executions:$reuse_commands,reuse_physical_test_executions:$reuse_tests,prior_receipts_valid:0,reused_artifact_files:[],decision_wall_ms:$decision_wall_ms,decision_peak_rss_kib:$decision_peak_rss,requires_execution:true},runtime:$runtime,inventory:{input_regular_files:$input_files,input_physical_lines:$fixture_lines,output_artifact_files:$output_files,reusable_artifact_files:$reusable_files},observer_go_version:$go_version,observer_go_env_goversion:$goversion,observer_toolchain_digest:$observer_digest,cell_evidence_projections:$cell_evidence_projections,cell_evidence_digests:$cell_evidence_digests,graph:{schema:"gooo-graph/v1",program_digest:$program_digest,graph_hash:$graph_hash,activity_count:$graph_activity_count,edge_count:$graph_edge_count,nodes:$graph_nodes,relations:$graph_relations,bindings:$bindings},repository_writes:0,local_test_executions:0,release_binary_builds:0,release_binary_build_reason:"NOT_EXECUTED_RELEASE_BINARY_BOUNDARY",human_report_ready:true}' > "$work/baseline-observation.json"

"$witness" -contract "$root/examples/opentofu-observation/contract.json" -input "$work/baseline-observation.json" -output "$work/baseline-report.json"
"$witness" -contract "$root/examples/opentofu-observation/contract.json" -input "$work/baseline-observation.json" -output "$work/baseline-report.json" -check "$work/baseline-report.json"
jq -e '.decision=="PASS" and .resolution=="EXACT" and (.cells|length)==12 and .summary.closed_cells==12 and .summary.replay_matches==1 and .reuse.request_mode=="BASELINE" and .reuse.decision=="EXECUTE" and .reuse.prior_candidates==0' "$work/baseline-report.json"
cp "$work/baseline-observation.json" "$out/evidence/baseline-observation.json"
cp "$work/baseline-report.json" "$out/evidence/baseline-report.json"
cp "$root/examples/opentofu-observation/contract.json" "$out/evidence/contract.json"
cp "$root/examples/opentofu-observation/main.gooo" "$out/evidence/main.gooo"

prior_artifact_manifest="$work/prior-artifact.manifest"
prior_artifact_manifest_digest="$(manifest_digest "$out/evidence" "$prior_artifact_manifest")"
prior_artifact_files_json="$(find "$out/evidence" -type f -printf '%P\n' | sort | jq -R . | jq -s .)"
prior_artifact_digests_json="$(while IFS= read -r file; do printf '%s\t%s\n' "$file" "$(digest "$out/evidence/$file")"; done < <(find "$out/evidence" -type f -printf '%P\n' | sort) | jq -Rn 'reduce inputs as $line ({}; ($line | split("\t")) as $parts | .[$parts[0]]=$parts[1])')"
prior_receipt_file_digest="$(digest "$work/baseline-report.json")"
baseline_report_digest="$(jq -er '.report_digest' "$work/baseline-report.json")"
reuse_cell_projection="request=REUSE;decision=REUSED;prior_receipt=$baseline_report_digest;artifact_manifest=$prior_artifact_manifest_digest;digests=$(jq -cS '.reuse | {source_digest,fixture_digest,argument_digest,environment_allowlist_digest,release_digest,observer_toolchain_digest,dependency_graph_digest,expected_result_digest}' "$work/baseline-observation.json")"
reuse_cell_digest="$(projection_digest "$reuse_cell_projection")"
reuse_human_projection="head=$HEAD_SHA;paths=3;cells=12;reuse=REUSED;prior_receipt=$baseline_report_digest"
reuse_human_digest="$(projection_digest "$reuse_human_projection")"

phase="reuse-decision"
run_timed "reuse-decision-projection" "caller-owned-observation" "$work" "$work/reuse-decision.stdout" "$work/reuse-decision.stderr" "$work/reuse-decision.json" "$decision_projection_descriptor" \
	jq -n --slurpfile baseline "$work/baseline-observation.json" --slurpfile prior "$work/baseline-report.json" \
		--arg receipt_file_digest "$prior_receipt_file_digest" --arg artifact_manifest_digest "$prior_artifact_manifest_digest" --arg reuse_cell_projection "$reuse_cell_projection" --arg reuse_cell_digest "$reuse_cell_digest" --arg reuse_human_projection "$reuse_human_projection" --arg reuse_human_digest "$reuse_human_digest" --argjson artifact_files "$prior_artifact_files_json" --argjson artifact_digests "$prior_artifact_digests_json" \
		--argjson baseline_commands "$baseline_physical_command_executions" --argjson baseline_tests "$baseline_physical_test_executions" --argjson reuse_commands "$reuse_physical_command_executions" --argjson reuse_tests "$reuse_physical_test_executions" \
		'$baseline[0] as $baseline_observation | $prior[0] as $prior_report | $baseline_observation | .reuse = (.reuse + {request_mode:"REUSE",requests:2,discovered:2,executed:1,reused:1,skipped:0,prior_candidates:1,invalidated:0,decision:"REUSED",reason:"EXACT_PRIOR_RECEIPT",prior_receipt_digest:$prior_report.report_digest,prior_receipt_file_digest:$receipt_file_digest,prior_artifact_manifest_digest:$artifact_manifest_digest,baseline_physical_command_executions:$baseline_commands,baseline_physical_test_executions:$baseline_tests,reuse_physical_command_executions:$reuse_commands,reuse_physical_test_executions:$reuse_tests,prior_receipts_valid:1,reused_artifact_files:$artifact_files,requires_execution:false}) | .inventory.reusable_artifact_files=($artifact_files | length) | .cell_evidence_projections.REUSE_ELIGIBILITY=$reuse_cell_projection | .cell_evidence_digests.REUSE_ELIGIBILITY=$reuse_cell_digest | .cell_evidence_projections.HUMAN_REPORT=$reuse_human_projection | .cell_evidence_digests.HUMAN_REPORT=$reuse_human_digest | .prior_receipt={report:$prior_report,receipt_file_digest:$receipt_file_digest,artifact_files:$artifact_files,artifact_digests:$artifact_digests,artifact_manifest_digest:$artifact_manifest_digest}'
jq -e '.exit_code == 0' "$work/reuse-decision.json"
reuse_decision_wall_ms="$(jq -er '.wall_ms' "$work/reuse-decision.json")"
reuse_decision_peak_rss="$(jq -er '.peak_rss_kib' "$work/reuse-decision.json")"
jq --argjson wall_ms "$reuse_decision_wall_ms" --argjson peak_rss "$reuse_decision_peak_rss" '.reuse.decision_wall_ms=$wall_ms | .reuse.decision_peak_rss_kib=$peak_rss' "$work/reuse-decision.stdout" > "$out/observation.json"

"$witness" -contract "$root/examples/opentofu-observation/contract.json" -input "$out/observation.json" -output "$out/report.json"
"$witness" -contract "$root/examples/opentofu-observation/contract.json" -input "$out/observation.json" -output "$out/report.json" -check "$out/report.json"
jq -e --argjson baseline_commands "$baseline_physical_command_executions" --argjson baseline_tests "$baseline_physical_test_executions" --argjson reuse_commands "$reuse_physical_command_executions" --argjson reuse_tests "$reuse_physical_test_executions" '.decision=="PASS" and .resolution=="EXACT" and (.cells|length)==12 and .summary.closed_cells==12 and .summary.foundation_closed==4 and .summary.coherence_closed==4 and .summary.regression_closed==4 and .summary.three_user_paths==3 and .summary.executions==2 and .summary.replay_matches==1 and .reuse.request_mode=="REUSE" and .reuse.requests==2 and .reuse.baseline_physical_command_executions==$baseline_commands and .reuse.baseline_physical_test_executions==$baseline_tests and .reuse.reuse_physical_command_executions==$reuse_commands and .reuse.reuse_physical_test_executions==$reuse_tests and .reuse.prior_candidates==1 and .reuse.prior_receipts_valid==1 and (.reuse.reused_artifact_files|length)==13 and .reuse.reused==1 and .reuse.invalidated==0 and .reuse.skipped==0 and .reuse.requires_execution==false and .repository_writes==0 and .local_test_executions==0 and .promotion_authorized==false' "$out/report.json"
cp "$out/report.json" "$out/evidence/report.json"
cp "$out/observation.json" "$out/evidence/observation.json"
cp "$root/examples/opentofu-observation/contract.json" "$out/evidence/contract.json"
cp "$root/examples/opentofu-observation/main.gooo" "$out/evidence/main.gooo"
uploaded_artifact_files="$(find "$out" -type f | wc -l | tr -d ' ')"
jq -e --argjson uploaded_files "$uploaded_artifact_files" '.inventory.output_artifact_files == $uploaded_files and .inventory.reusable_artifact_files == (.reuse.reused_artifact_files | length)' "$out/report.json"
jq -r '"### OpenTofu released-CLI observation\n- decision/resolution: \(.decision)/\(.resolution)\n- exact cells: \(.summary.closed_cells)/\(.summary.cells_total)\n- user paths: \(.summary.three_user_paths)\n- executions/replay matches: \(.summary.executions)/\(.summary.replay_matches)\n- runtime total_ms/max_rss_kib: \(.runtime.total_wall_ms)/\(.runtime.max_peak_rss_kib)\n- consumer build_ms/rss_kib: \(.runtime.consumer_build_ms)/\(.runtime.consumer_build_peak_rss_kib)\n- tofu init/plan/show/test ms: \(.runtime.tofu_init_ms)/\(.runtime.tofu_plan_ms)/\(.runtime.tofu_show_ms)/\(.runtime.tofu_test_ms)\n- tofu test executions: \(.runtime.tofu_test_executions)\n- input files/lines, uploaded files, reusable prior files: \(.inventory.input_regular_files)/\(.inventory.input_physical_lines)/\(.inventory.output_artifact_files)/\(.inventory.reusable_artifact_files)\n- repository writes/local test executions: \(.repository_writes)/\(.local_test_executions)\n- reuse discovered/executed/reused/skipped: \(.reuse.discovered)/\(.reuse.executed)/\(.reuse.reused)/\(.reuse.skipped)\n- reuse decision: \(.reuse.decision)/\(.reuse.reason)\n- Go observer: \(.observer_go_env_goversion)\n\n### Cells\n\(.cells[] | "- \(.id): \(.decision) / \(.observed)/\(.expected) / \(.meta_operation) / \(.proof_choice) / \(.indicator)")\n\n### Counterexamples\n\(.counterexamples[] | "- \(.id): \(.decision) / \(.resolution) / \(.reason)")"' "$out/report.json" >> "$GITHUB_STEP_SUMMARY"
jq -r '"- plan raw digests: \(.executions[0].plan_raw_digest), \(.executions[1].plan_raw_digest)\n- plan canonical digests: \(.executions[0].plan_json_digest), \(.executions[1].plan_json_digest)\n- plan canonicalizer/removed fields: \(.executions[0].plan_canonicalizer)/\(.executions[0].plan_volatile_fields | join(","))\n- test type counts: \(.executions[0].test_type_counts | to_entries | sort_by(.key) | map("\(.key)=\(.value)") | join(","))\n- test abstract discovered/run executed/summary pass-fail-error-skip: \(.executions[0].test_abstract_discovered)/\(.executions[0].test_run_executed)/\(.executions[0].test_summary_passed)-\(.executions[0].test_summary_failed)-\(.executions[0].test_summary_errored)-\(.executions[0].test_summary_skipped)\n- release binary builds: \(.release_binary_builds) / \(.release_binary_build_reason)"' "$out/report.json" >> "$GITHUB_STEP_SUMMARY"
jq -r '"- reuse exact indicators: requests=\(.reuse.requests), baseline_physical_command_executions=\(.reuse.baseline_physical_command_executions), baseline_physical_test_executions=\(.reuse.baseline_physical_test_executions), reuse_physical_command_executions=\(.reuse.reuse_physical_command_executions), reuse_physical_test_executions=\(.reuse.reuse_physical_test_executions), prior_candidates=\(.reuse.prior_candidates), prior_receipts_valid=\(.reuse.prior_receipts_valid), reused=\(.reuse.reused), reused_artifact_files=\(.reuse.reused_artifact_files | length), uploaded_artifact_files=\(.inventory.output_artifact_files), invalidated=\(.reuse.invalidated), skipped=\(.reuse.skipped), decision_wall_ms=\(.reuse.decision_wall_ms), decision_peak_rss_kib=\(.reuse.decision_peak_rss_kib), requires_execution=\(.reuse.requires_execution)"' "$out/report.json" >> "$GITHUB_STEP_SUMMARY"
