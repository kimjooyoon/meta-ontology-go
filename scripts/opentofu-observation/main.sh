#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
out="${RUNNER_TEMP:-/tmp}/opentofu-observation"
work="${RUNNER_TEMP:-/tmp}/opentofu-observation-work"
fixture="$root/examples/opentofu-observation/fixture"
mkdir -p "$out/evidence" "$work"
cd "$root"

asset_url="https://github.com/opentofu/opentofu/releases/download/v1.12.6/tofu_1.12.6_linux_amd64.tar.gz"
asset_sha="sha256:50a6106fa4de523d09c87af85f3db1dd47535fc005727fdca6852146476b88ec"
asset_bytes_expected=34646566
sums_sha="sha256:6988e0cb8f4e9ebfa3b0999e44841549741b22d9b38873cb5b89074f1cddcb1c"
release_id="opentofu-v1.12.6"

digest() { printf 'sha256:%s' "$(sha256sum "$1" | cut -d' ' -f1)"; }

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
  jq -n --arg name "$name" --arg cwd_role "$cwd_role" --argjson command "$descriptor" \
    --argjson exit_code "$status" --argjson stdout_bytes "$(wc -c <"$stdout" | tr -d ' ')" \
    --arg stdout_digest "$(digest "$stdout")" --argjson stderr_bytes "$(wc -c <"$stderr" | tr -d ' ')" \
    --arg stderr_digest "$(digest "$stderr")" --argjson wall_ms "$wall_ms" --argjson peak_rss_kib "$peak_rss" \
    '{name:$name,command:$command,cwd_role:$cwd_role,exit_code:$exit_code,stdout_bytes:$stdout_bytes,stdout_digest:$stdout_digest,stderr_bytes:$stderr_bytes,stderr_digest:$stderr_digest,wall_ms:$wall_ms,peak_rss_kib:$peak_rss_kib,executed:true}' >"$receipt"
  printf 'command_observation name=%s exit_code=%s stdout_bytes=%s stderr_bytes=%s\n' \
    "$name" "$status" "$(wc -c <"$stdout" | tr -d ' ')" "$(wc -c <"$stderr" | tr -d ' ')"
  rm -f "$rss"
}

manifest_digest() {
  local directory="$1" manifest="$2"
  : > "$manifest"
  while IFS= read -r -d '' file; do
    printf '%s\t' "${file#"$directory"/}" >> "$manifest"
    sha256sum "$file" | cut -d' ' -f1 >> "$manifest"
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

version_descriptor='["tofu","version","-json"]'
run_timed "tofu-version" "release" "$work" "$work/version.stdout" "$work/version.stderr" "$work/version.json" "$version_descriptor" "$tofu" version -json
jq -e '.exit_code == 0' "$work/version.json"
version_json="$(jq -c . "$work/version.stdout")"
version="$(jq -er '.terraform_version' "$work/version.stdout")"
platform="$(jq -er '.platform' "$work/version.stdout")"
printf 'release_observation version=%s platform=%s\n' "$version" "$platform"
test "$version" = "1.12.6"
test "$platform" = "linux_amd64"

fixture_digest="$(manifest_digest "$fixture" "$work/fixture.manifest")"
mapfile -t fixture_files < <(find "$fixture" -type f -printf '%P\n' | sort)
input_files="$(find "$fixture" -type f | wc -l | tr -d ' ')"
input_lines=0
for file in "${fixture_files[@]}"; do input_lines=$((input_lines + $(wc -l <"$fixture/$file" | tr -d ' '))); done
test "$input_files" -gt 0
test "$input_lines" -gt 0

init_descriptor='["tofu","init","-backend=false"]'
plan_descriptor='["tofu","plan","-refresh=false","-input=false","-out=plan.bin","-json"]'
show_descriptor='["tofu","show","-json","plan.bin"]'
test_descriptor='["tofu","test","-json","-no-color","-test-directory","tests"]'

run_once() {
  local number="$1" directory="$work/run-$1"; shift 2
  cp -R "$fixture" "$directory"
  run_timed "tofu-init" "fixture-run-$number" "$directory" "$directory/init.stdout" "$directory/init.stderr" "$directory/init.json" "$init_descriptor" "$tofu" init -backend=false
  run_timed "tofu-plan" "fixture-run-$number" "$directory" "$directory/plan.stdout" "$directory/plan.stderr" "$directory/plan.json" "$plan_descriptor" "$tofu" plan -refresh=false -input=false -out=plan.bin -json
  run_timed "tofu-show" "fixture-run-$number" "$directory" "$directory/show.stdout" "$directory/show.stderr" "$directory/show.json" "$show_descriptor" "$tofu" show -json plan.bin
  run_timed "tofu-test" "fixture-run-$number" "$directory" "$directory/test.stdout" "$directory/test.stderr" "$directory/test.json" "$test_descriptor" "$tofu" test -json -no-color -test-directory tests
  for receipt in init plan show test; do jq -e '.exit_code == 0' "$directory/$receipt.json"; done
  jq -e 'type == "object" and (.format_version | type) == "string" and has("planned_values")' "$directory/show.stdout"
  : > "$directory/test.normalized"
  while IFS= read -r line; do
    jq -c 'if type == "object" then del(."@timestamp",.timestamp,.elapsed_seconds) else error end' <<<"$line" >> "$directory/test.normalized"
  done < "$directory/test.stdout"
  sort "$directory/test.normalized" -o "$directory/test.normalized"
  test -s "$directory/test.normalized"
  jq -n --argjson index "$number" --arg fixture "$fixture_digest" \
    --arg plan "$(digest "$directory/show.stdout")" --argjson plan_bytes "$(wc -c <"$directory/show.stdout" | tr -d ' ')" \
    --argjson plan_valid true --arg test_events "$(digest "$directory/test.normalized")" \
    --arg test_raw "$(digest "$directory/test.stdout")" --argjson test_count "$(wc -l <"$directory/test.normalized" | tr -d ' ')" \
    --argjson test_valid true --slurpfile init "$directory/init.json" --slurpfile plan_command "$directory/plan.json" \
    --slurpfile show "$directory/show.json" --slurpfile test "$directory/test.json" \
    '{index:$index,fixture_digest:$fixture,plan_json_digest:$plan,plan_json_bytes:$plan_bytes,plan_schema_valid:$plan_valid,test_event_digest:$test_events,test_raw_digest:$test_raw,test_event_count:$test_count,test_events_valid:$test_valid,commands:[$init[0],$plan_command[0],$show[0],$test[0]]}' > "$directory/run.json"
  cp "$directory/show.stdout" "$out/evidence/plan-$number.json"
  cp "$directory/test.stdout" "$out/evidence/test-$number.ndjson"
}

run_once 1
run_once 2

go run ./cmd/gooo check examples/opentofu-observation/main.gooo
go run ./cmd/gooo graph dump examples/opentofu-observation/main.gooo > "$work/gooo-graph.json"
program_digest="$(digest examples/opentofu-observation/main.gooo)"
graph_hash="$(jq -er '.graph_hash' "$work/gooo-graph.json")"
graph_bindings="$(jq -n --slurpfile graph "$work/gooo-graph.json" '
  $graph[0] as $g |
  [{"cell_id":"OPENTOFU_RELEASE_PIN","name":"PinOpenTofuRelease"},{"cell_id":"ASSET_CHECKSUM","name":"VerifyOpenTofuAssetChecksum"},{"cell_id":"CLI_VERSION_JSON","name":"ObserveOpenTofuVersionJSON"},{"cell_id":"FIXTURE_INPUT_DIGEST","name":"PinOpenTofuFixtureInput"},{"cell_id":"PLAN_JSON_SCHEMA","name":"ValidateOpenTofuPlanJSON"},{"cell_id":"TEST_EVENT_INVENTORY","name":"AccountOpenTofuTestEvents"},{"cell_id":"COMMAND_RUNTIME_RECEIPT","name":"RecordOpenTofuCommandRuntime"},{"cell_id":"PEAK_RSS_OBSERVATION","name":"ObserveOpenTofuPeakRSS"},{"cell_id":"DETERMINISTIC_PLAN_REPLAY","name":"ReplayOpenTofuPlanJSON"},{"cell_id":"DETERMINISTIC_TEST_REPLAY","name":"ReplayOpenTofuTestEvents"},{"cell_id":"REUSE_ELIGIBILITY","name":"EvaluateOpenTofuTestReuse"},{"cell_id":"HUMAN_REPORT","name":"PublishOpenTofuObservationReport"}] |
  map(. as $cell | [$g.nodes[] | select(.kind=="Activity" and .name==$cell.name)][0].id as $id |
    {cell_id:$cell.cell_id,activity_id:$id,input_id:("gooo://opentofu-observation/input/" + ($cell.cell_id|ascii_downcase)),output_id:("gooo://opentofu-observation/output/" + ($cell.cell_id|ascii_downcase)),used_edge_count:([$g.relations[] | select(.predicate=="used" and .subject==$id)]|length),generated_edge_count:([$g.relations[] | select(.predicate=="wasGeneratedBy" and .object==$id)]|length)})
')"
jq -e --argjson bindings "$graph_bindings" '($bindings|length)==12 and all($bindings[]; .activity_id != null and .used_edge_count == 1 and .generated_edge_count == 1)' <<<"$graph_bindings"
cp "$work/gooo-graph.json" "$out/evidence/gooo-graph.json"
cp "$work/version.stdout" "$out/evidence/version.json"
cp "$work/SHA256SUMS" "$out/evidence/SHA256SUMS"
output_artifact_files="$(find "$out/evidence" -type f | wc -l | tr -d ' ')"
graph_activity_count="$(jq -er '[.nodes[] | select(.kind=="Activity")] | length' "$work/gooo-graph.json")"
graph_edge_count="$(jq -er '.relations | length' "$work/gooo-graph.json")"
test "$graph_activity_count" -eq 12

runtime="$(jq -n --slurpfile build "$work/consumer-build.json" --slurpfile version "$work/version.json" --slurpfile first "$work/run-1/run.json" --slurpfile second "$work/run-2/run.json" '
  [$build[0],$version[0],$first[0].commands[],$second[0].commands[]] as $all |
  ($first[0].commands | map(select(.name=="tofu-init"))[0]) as $init |
  ($first[0].commands | map(select(.name=="tofu-plan"))[0]) as $plan |
  ($first[0].commands | map(select(.name=="tofu-show"))[0]) as $show |
  ($first[0].commands | map(select(.name=="tofu-test"))[0]) as $test |
  {consumer_build_ms:$build[0].wall_ms,consumer_build_peak_rss_kib:$build[0].peak_rss_kib,tofu_init_ms:$init.wall_ms,tofu_init_peak_rss_kib:$init.peak_rss_kib,tofu_plan_ms:$plan.wall_ms,tofu_plan_peak_rss_kib:$plan.peak_rss_kib,tofu_show_ms:$show.wall_ms,tofu_show_peak_rss_kib:$show.peak_rss_kib,tofu_test_ms:$test.wall_ms,tofu_test_peak_rss_kib:$test.peak_rss_kib,total_wall_ms:([$all[].wall_ms]|add),max_peak_rss_kib:([$all[].peak_rss_kib]|max)}')"

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

jq -n --arg schema "gooo/opentofu-observation/v1" --arg contract_id "opentofu-released-cli-observation-v1" --arg subject "$HEAD_SHA" \
  --arg release_id "$release_id" --arg asset_url "$asset_url" --arg asset_sha "$asset_sha" --argjson asset_bytes "$asset_bytes" --arg sums_sha "$sums_sha" \
  --argjson version_json "$version_json" --arg version_digest "$(digest "$work/version.stdout")" --arg version "$version" --arg platform "$platform" --slurpfile version_command "$work/version.json" \
  --arg fixture_digest "$fixture_digest" --argjson fixture_files "$(printf '%s\n' "${fixture_files[@]}" | jq -R . | jq -s .)" --argjson fixture_lines "$input_lines" \
  --slurpfile first "$work/run-1/run.json" --slurpfile second "$work/run-2/run.json" --argjson reuse_discovered 1 --argjson reuse_executed 1 --argjson reuse 0 --argjson skipped 0 --argjson prior 0 --argjson invalidated 0 \
  --arg source_digest "$source_digest" --arg argument_digest "$argument_digest" --arg environment_digest "$environment_digest" --arg release_digest "$asset_sha" --arg toolchain_digest "$toolchain_digest" --arg dependency_digest "$dependency_digest" --arg expected_digest "$expected_digest" \
  --argjson runtime "$runtime" --argjson input_files "$input_files" --argjson output_files "$output_artifact_files" --arg go_version "$go_version" --arg goversion "$go_env_version" --arg observer_digest "$observer_toolchain_digest" \
  --arg program_digest "$program_digest" --arg graph_hash "$graph_hash" --argjson graph_activity_count "$graph_activity_count" --argjson graph_edge_count "$graph_edge_count" --argjson bindings "$graph_bindings" \
  '{schema:$schema,contract_id:$contract_id,subject_sha:$subject,user_paths:["P1 RELEASE_IDENTITY","P2 PLAN_JSON","P3 TEST_JSON"],release:{release_id:$release_id,asset_url:$asset_url,asset_sha256:$asset_sha,asset_bytes:$asset_bytes,checksums_sha256:$sums_sha,version_json:$version_json,version_json_digest:$version_digest,version:$version,platform:$platform,command:$version_command[0]},fixture_digest:$fixture_digest,fixture_files:$fixture_files,fixture_physical_lines:$fixture_lines,executions:[$first[0],$second[0]],reuse:{discovered:$reuse_discovered,executed:$reuse_executed,reused:$reuse,skipped:$skipped,prior_candidates:$prior,invalidated:$invalidated,decision:"NOT_REUSED_FIRST_RUN",reason:"NO_PRIOR_RECEIPT",source_digest:$source_digest,fixture_digest:$fixture_digest,argument_digest:$argument_digest,environment_allowlist_digest:$environment_digest,release_digest:$asset_sha,observer_toolchain_digest:$toolchain_digest,dependency_graph_digest:$dependency_digest,expected_result_digest:$expected_digest,prior_receipt_digest:""},runtime:$runtime,inventory:{input_regular_files:$input_files,input_physical_lines:$fixture_lines,output_artifact_files:$output_files},observer_go_version:$go_version,observer_go_env_goversion:$goversion,observer_toolchain_digest:$observer_digest,graph:{schema:"gooo/opentofu-observation-graph/v1",program_digest:$program_digest,graph_hash:$graph_hash,activity_count:$graph_activity_count,edge_count:$graph_edge_count,bindings:$bindings},repository_writes:0,local_test_executions:0,human_report_ready:true}' > "$out/observation.json"

"$witness" -contract "$root/examples/opentofu-observation/contract.json" -input "$out/observation.json" -output "$out/report.json"
"$witness" -contract "$root/examples/opentofu-observation/contract.json" -input "$out/observation.json" -output "$out/report.json" -check "$out/report.json"
jq -e '.decision=="PASS" and .resolution=="EXACT" and (.cells|length)==12 and .summary.closed_cells==12 and .summary.foundation_closed==4 and .summary.coherence_closed==4 and .summary.regression_closed==4 and .summary.three_user_paths==3 and .summary.executions==2 and .summary.replay_matches==1 and .repository_writes==0 and .local_test_executions==0 and .promotion_authorized==false' "$out/report.json"
jq -r '"### OpenTofu released-CLI observation\n- decision/resolution: \(.decision)/\(.resolution)\n- exact cells: \(.summary.closed_cells)/\(.summary.cells_total)\n- user paths: \(.summary.three_user_paths)\n- executions/replay matches: \(.summary.executions)/\(.summary.replay_matches)\n- runtime total_ms/max_rss_kib: \(.runtime.total_wall_ms)/\(.runtime.max_peak_rss_kib)\n- consumer build_ms/rss_kib: \(.runtime.consumer_build_ms)/\(.runtime.consumer_build_peak_rss_kib)\n- tofu init/plan/show/test ms: \(.runtime.tofu_init_ms)/\(.runtime.tofu_plan_ms)/\(.runtime.tofu_show_ms)/\(.runtime.tofu_test_ms)\n- input files/lines and output evidence files: \(.inventory.input_regular_files)/\(.inventory.input_physical_lines)/\(.inventory.output_artifact_files)\n- repository writes/local test executions: \(.repository_writes)/\(.local_test_executions)\n- reuse discovered/executed/reused/skipped: \(.reuse.discovered)/\(.reuse.executed)/\(.reuse.reused)/\(.reuse.skipped)\n- reuse decision: \(.reuse.decision)/\(.reuse.reason)\n- Go observer: \(.observer_go_env_goversion)\n\n### Cells\n\(.cells[] | "- \(.id): \(.decision) / \(.observed)/\(.expected) / \(.meta_operation) / \(.proof_choice) / \(.indicator)")\n\n### Counterexamples\n\(.counterexamples[] | "- \(.id): \(.decision) / \(.resolution) / \(.reason)")"' "$out/report.json" >> "$GITHUB_STEP_SUMMARY"
cp "$out/report.json" "$out/evidence/report.json"
cp "$out/observation.json" "$out/evidence/observation.json"
cp "$root/examples/opentofu-observation/contract.json" "$out/evidence/contract.json"
cp "$root/examples/opentofu-observation/main.gooo" "$out/evidence/main.gooo"
