#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
work="${RUNNER_TEMP:-/tmp}/language-debug-experiment"
build="${RUNNER_TEMP:-/tmp}/language-debug-build"
binary="$build/gooo"
reducer="$build/language-debug-experiment"
contract="$root/examples/language-debug/contract.json"
source="$root/examples/billing/main.gooo"
mkdir -p "$work" "$build"
cd "$root"

gofmt -l cmd/gooo cmd/language-debug-experiment internal/languagedebug internal/meta/languagedebugexperiment > "$work/unformatted.txt"
test ! -s "$work/unformatted.txt"

measure_command() {
  local name="$1" output="$2" stdout_file="$3" stderr_file="$4"
  shift 4
  local rss_file="$output.rss" start_ns end_ns wall_ns wall_ms peak_rss status
  start_ns="$(date +%s%N)"
  set +e
  /usr/bin/time -q -f '%M' -o "$rss_file" "$@" >"$stdout_file" 2>"$stderr_file"
  status=$?
  set -e
  end_ns="$(date +%s%N)"
  wall_ns=$((end_ns - start_ns))
  wall_ms=$(((wall_ns + 999999) / 1000000))
  if ((wall_ns > 0 && wall_ms == 0)); then wall_ms=1; fi
  peak_rss="$(tail -n 1 "$rss_file")"
  if [[ ! "$peak_rss" =~ ^[0-9]+$ ]]; then
    echo "invalid peak RSS measurement: $rss_file" >&2
    return 1
  fi
  test "$wall_ns" -gt 0
  test "$wall_ms" -gt 0
  test "$peak_rss" -gt 0
  jq -n --arg name "$name" --argjson executed true --argjson wall_ns "$wall_ns" \
    --argjson wall_ms "$wall_ms" --argjson peak_rss_kib "$peak_rss" \
    --arg cache_state "GOCACHE=$(go env GOCACHE);GOMODCACHE=$(go env GOMODCACHE);setup-go-cache=false" \
    '{name:$name,executed:$executed,wall_ns:$wall_ns,wall_ms:$wall_ms,peak_rss_kib:$peak_rss_kib,cache_state:$cache_state}' > "$output"
  if ((status != 0)); then
    echo "measured command failed: $name (exit $status)" >&2
    cat "$stdout_file" "$stderr_file" >&2
  fi
  return "$status"
}

measure_command "debug-relevant-tests" "$work/test.json" "$work/test.stdout" "$work/test.stderr" \
  go test ./cmd/gooo ./cmd/language-debug-experiment ./internal/languagedebug ./internal/meta/languagedebugexperiment
measure_command "debug-producer-build" "$work/build.json" "$work/build.stdout" "$work/build.stderr" \
  go build -trimpath -o "$binary" ./cmd/gooo
measure_command "debug-evaluator-build" "$work/evaluator-build.json" "$work/evaluator-build.stdout" "$work/evaluator-build.stderr" \
  go build -trimpath -o "$reducer" ./cmd/language-debug-experiment

digest() { printf 'sha256:%s' "$(sha256sum "$1" | cut -d' ' -f1)"; }
binary_digest="$(digest "$binary")"
toolchain="$(go version)"
runner="${RUNNER_NAME:-unknown}/${RUNNER_OS:-unknown}/${RUNNER_ARCH:-unknown}"

run_debug() {
  local run="$1" breakpoint="$2" output="$work/$run.json" rss="$work/$run.rss"
  local run_number=1 start_ns end_ns wall_ns wall_ms peak_rss status
  test "$run" = "first" || run_number=2
  local -a args=(debug --json --entry PayOrder --break-event "$breakpoint" "$source")
  start_ns="$(date +%s%N)"
  set +e
  /usr/bin/time -q -f '%M' -o "$rss" "$binary" "${args[@]}" >"$output" 2>"$work/$run.stderr"
  status=$?
  set -e
  end_ns="$(date +%s%N)"
  test "$status" -eq 0
  wall_ns=$((end_ns - start_ns))
  wall_ms=$(((wall_ns + 999999) / 1000000))
  if ((wall_ns > 0 && wall_ms == 0)); then wall_ms=1; fi
  peak_rss="$(tail -n 1 "$rss")"
  if [[ ! "$peak_rss" =~ ^[0-9]+$ ]]; then
    echo "invalid peak RSS measurement: $rss" >&2
    return 1
  fi
  test "$wall_ns" -gt 0
  test "$wall_ms" -gt 0
  test "$peak_rss" -gt 0
  printf '%s\n' "${args[@]}" | jq -R . | jq -s . > "$work/$run.args.json"
  jq -n --argjson run "$run_number" --arg runtime_receipt_schema "gooo/language-debug-runtime-receipt/v1" --arg runner "$runner" --arg toolchain "$toolchain" \
    --arg source_raw_digest "$(jq -r '.source_digest' "$output")" \
    --arg source_semantic_digest "$(jq -r '.semantic_digest' "$output")" \
    --arg binary_digest "$binary_digest" --slurpfile arguments "$work/$run.args.json" \
    --arg subject_sha "$HEAD_SHA" --arg output_digest "$(digest "$output")" \
    --argjson wall_ns "$wall_ns" --argjson wall_ms "$wall_ms" --argjson peak_rss_kib "$peak_rss" \
    '{run:$run,runtime_receipt_schema:$runtime_receipt_schema,runner:$runner,toolchain:$toolchain,source_raw_digest:$source_raw_digest,source_semantic_digest:$source_semantic_digest,binary_digest:$binary_digest,arguments:$arguments[0],subject_sha:$subject_sha,output_digest:$output_digest,wall_ns:$wall_ns,wall_ms:$wall_ms,peak_rss_kib:$peak_rss_kib}' \
    > "$work/$run.runtime.json"
}

# These are two real debug invocations over the same checked-in Gooo source.
# Re-run the same real debug invocation so the semantic receipt comparison is
# based on identical arguments; runtime observations remain separate fields.
run_debug first SOURCE_PARSED
run_debug second SOURCE_PARSED
if "$binary" debug --json --entry PayOrder --break-event MISSING "$source" > "$work/unknown-breakpoint.json" 2>"$work/unknown-breakpoint.stderr"; then
  echo "unknown debug breakpoint unexpectedly passed" >&2
  exit 1
fi

# Generate the contract-defined meta program once from a seed observation whose
# two new debugging cells remain OPEN. The final utility observation below
# binds those same cells to this exact generated program and graph.
seed_observation="$work/seed-observation.json"
jq -n --arg schema 'gooo/language-utility-observation/v1' --arg contract 'gooo-language-utility-v1' \
  --arg head "$HEAD_SHA" --argjson stages '[
    "SOURCE_PRESENT","SYNTAX_ACCEPTED","SEMANTIC_ACCEPTED","OUTCOME_OBSERVED",
    "DETERMINISTIC_REPLAY","RESOURCE_OBSERVED","USER_ARTIFACT_VERIFIED"
  ]' '
  def closed($id;$producer;$file): $stages | map({use_case_id:$id,stage_id:.,state:"CLOSED",
    producer:$producer,step:("VERIFY_"+.),reason:"EVIDENCE_ACCEPTED",evidence_key:($id+".report"),evidence_path:("evidence/"+$file),evidence_digest:"sha256:0000000000000000000000000000000000000000000000000000000000000000"});
  def opened($id;$stage;$reason): {use_case_id:$id,stage_id:$stage,state:"OPEN",producer:("ci:"+$id),step:("COLLECT_"+$stage),reason:$reason};
  {schema:$schema,contract_id:$contract,subject_sha:$head,repository_writes:0,cells:(
    closed("ci-plan-selection";"scripts/ci-plan-usecase";"ci-plan.json") +
    closed("source-execution";"scripts/language-source-execution";"source-execution.json") +
    closed("artifact-emission";"scripts/language-example-experiment";"artifact-emission.json") +
    closed("profiling";"scripts/language-profile-experiment";"profiling.json") +
    closed("debugging";"scripts/language-debug-experiment";"debugging.json") +
    closed("package-execution";"scripts/language-package-execution";"package-execution.json") |
    map(if .use_case_id=="debugging" and (.stage_id=="DETERMINISTIC_REPLAY" or .stage_id=="RESOURCE_OBSERVED")
      then opened("debugging";.stage_id;"DEBUG_EVIDENCE_NOT_BOUND")
      elif .use_case_id=="package-execution" and .stage_id=="RESOURCE_OBSERVED"
      then opened("package-execution";.stage_id;"PACKAGE_RESOURCES_NOT_OBSERVED") else . end))}
' > "$seed_observation"
go run ./cmd/language-utility-witness -contract examples/language-utility/contract.json \
  -observation "$seed_observation" -report "$work/seed-utility-report.json" \
  -program "$work/program.gooo" > "$work/seed.stdout" 2> "$work/seed.stderr"
test -s "$work/program.gooo"
"$binary" graph dump "$work/program.gooo" > "$work/gooo-graph.json"

program_digest="$(digest "$work/program.gooo")"
graph_hash="$(jq -r '.graph_hash' "$work/gooo-graph.json")"
cell_input_id="gooo://meta/language-utility/entity/cell"
cell_output_id="gooo://meta/language-utility/entity/evidence"
replay_activity_id="$(jq -r '.nodes[] | select(.kind=="Activity" and .name=="ObserveDebuggingDeterministicReplay") | .id' "$work/gooo-graph.json")"
resource_activity_id="$(jq -r '.nodes[] | select(.kind=="Activity" and .name=="ObserveDebuggingResourceObserved") | .id' "$work/gooo-graph.json")"
test "$replay_activity_id" != ""
test "$resource_activity_id" != ""

jq -n --arg schema "gooo-graph/v1" --arg program_digest "$program_digest" --arg graph_hash "$graph_hash" \
  --arg replay_activity "$replay_activity_id" --arg resource_activity "$resource_activity_id" \
  --slurpfile graph "$work/gooo-graph.json" '
  ($graph[0]) as $g |
  ([$replay_activity,$resource_activity] | sort) as $debug_activity_ids |
  def is_debug($id): ($debug_activity_ids | index($id)) != null;
  ([$g.nodes[] | select(.kind=="Activity")] | length) as $activity_count |
  ([$g.relations[]] | length) as $edge_count |
  ([$g.nodes[] | select(.kind=="Activity" and (.name=="ObserveDebuggingDeterministicReplay" or .name=="ObserveDebuggingResourceObserved"))] | length) as $debug_activity_count |
  ([$g.relations[] | select(.predicate=="wasGeneratedBy" and is_debug(.object))] | length) as $debug_output_count |
  ([$g.relations[] | select(.predicate=="used" and is_debug(.subject))] | length) as $debug_used_edge_count |
  ([$g.relations[] | select(.predicate=="wasGeneratedBy" and is_debug(.object))] | length) as $debug_generated_edge_count |
  ([$g.relations[] | select((.predicate=="used" and is_debug(.subject)) or (.predicate=="wasGeneratedBy" and is_debug(.object))) | {relation:.predicate,subject:.subject,object:.object}] | sort_by(.relation,.subject,.object)) as $debug_causal_edges |
  {schema:$schema,program_digest:$program_digest,graph_hash:$graph_hash,activity_count:$activity_count,edge_count:$edge_count,debug_activity_count:$debug_activity_count,debug_output_count:$debug_output_count,debug_used_edge_count:$debug_used_edge_count,debug_generated_edge_count:$debug_generated_edge_count,debug_activity_ids:$debug_activity_ids,debug_causal_edges:$debug_causal_edges}
' > "$work/graph-observation.json"
jq -e '.activity_count==44 and .edge_count==88 and .debug_activity_count==2 and .debug_output_count==2 and .debug_used_edge_count==2 and .debug_generated_edge_count==2' "$work/graph-observation.json"

edges_for() {
  local activity_id="$1"
  jq --arg activity "$activity_id" --arg input "$cell_input_id" --arg output "$cell_output_id" \
    '[.relations[] | select((.predicate=="used" and .subject==$activity and .object==$input) or (.predicate=="wasGeneratedBy" and .subject==$output and .object==$activity)) | {relation:.predicate,subject:.subject,object:.object}]' \
    "$work/gooo-graph.json"
}
replay_edges="$(edges_for "$replay_activity_id")"
resource_edges="$(edges_for "$resource_activity_id")"
jq -e 'length==2' <<<"$replay_edges"
jq -e 'length==2' <<<"$resource_edges"

positive_input="$work/positive-input.json"
jq -n --arg subject "$HEAD_SHA" --arg executable "$binary_digest" --slurpfile contract "$contract" \
  --slurpfile first "$work/first.json" --slurpfile second "$work/second.json" --slurpfile unknown "$work/unknown-breakpoint.json" \
  --slurpfile runtime1 "$work/first.runtime.json" --slurpfile runtime2 "$work/second.runtime.json" \
  --slurpfile build "$work/build.json" --slurpfile evaluator_build "$work/evaluator-build.json" --slurpfile test "$work/test.json" --slurpfile graph "$work/graph-observation.json" \
  '{subject_sha:$subject,executable_digest:$executable,contract:$contract[0],first:$first[0],second:$second[0],unknown_breakpoint:$unknown[0],runtime_observations:[$runtime1[0],$runtime2[0]],build:$build[0],evaluator_build:$evaluator_build[0],test:$test[0],graph:$graph[0]}' > "$positive_input"
cp "$positive_input" "$work/input.json"

"$reducer" -input "$positive_input" -output "$work/report.json"
"$reducer" -input "$positive_input" -check "$work/report.json"
jq -e '.decision=="PASS" and .resolution=="EXACT" and .summary.coordinates.satisfied==13 and .summary.coordinates.total==13 and .summary.replay_matches==1 and .summary.resource_observations==2 and .summary.compiler.go127_runtimes==2 and .summary.unknowns==0 and .summary.refuted_cases==0 and .summary.effects.repository_writes==0 and .summary.effects.mutation_authority==false and .replay.equal==true and (.replay.excluded_fields|length)>0 and (.views[]|select(.audience=="USER")|.satisfied==4 and .total==4) and (.views[]|select(.audience=="TOOL_AUTHOR")|.satisfied==9 and .total==9) and (.views[]|select(.audience=="GOVERNOR")|.satisfied==13 and .total==13)' "$work/report.json"

expect_failure() {
  local input="$1" output="$2"
  set +e
  "$reducer" -input "$input" -output "$output" > "$output.stdout" 2> "$output.stderr"
  local code=$?
  set -e
  test "$code" -eq 1
}
jq 'del(.second)' "$positive_input" > "$work/missing-second-input.json"
expect_failure "$work/missing-second-input.json" "$work/missing-second-report.json"
jq '.runtime_observations[1] |= del(.peak_rss_kib)' "$positive_input" > "$work/missing-resource-input.json"
expect_failure "$work/missing-resource-input.json" "$work/missing-resource-report.json"
jq --arg bad "sha256:$(printf 'f%.0s' {1..64})" '.second.semantic_digest=$bad' "$positive_input" > "$work/digest-contradiction-input.json"
expect_failure "$work/digest-contradiction-input.json" "$work/digest-contradiction-report.json"
jq '.first.decision="UNKNOWN"' "$positive_input" > "$work/unknown-top-input.json"
expect_failure "$work/unknown-top-input.json" "$work/unknown-top-report.json"
jq '.first=null' "$positive_input" > "$work/malformed-input.json"
expect_failure "$work/malformed-input.json" "$work/malformed-report.json"
jq '.graph.debug_used_edge_count=1' "$positive_input" > "$work/graph-edge-removal-input.json"
expect_failure "$work/graph-edge-removal-input.json" "$work/graph-edge-removal-report.json"

jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and (.unknown_cases|length)==1 and (.unknown_cases[0] | [.stage,.step,.reason,.unknown_class,.next_operation] | all(. != "")) and (.unknown_cases[0].blocked_by|type)=="array" and (.unknown_cases[0].unknown_class=="DIRECT_MISSING" and (.unknown_cases[0].blocked_by|length)==0)' "$work/missing-second-report.json"
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and (.unknown_cases|length)==1 and (.unknown_cases[0] | [.stage,.step,.reason,.unknown_class,.next_operation] | all(. != "")) and (.unknown_cases[0].blocked_by|type)=="array" and (.unknown_cases[0].unknown_class=="DIRECT_MISSING" and (.unknown_cases[0].blocked_by|length)==0)' "$work/missing-resource-report.json"
jq -e '.decision=="FAIL_CLOSED" and .resolution=="EXACT" and (.unknown_cases|length)==0 and (.refuted_cases[0].reason=="DEBUG_DETERMINISTIC_DIGEST_CONTRADICTION")' "$work/digest-contradiction-report.json"
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and (.unknown_cases|length)==1 and .unknown_cases[0].unknown_class=="UNKNOWN_DECISION" and (.unknown_cases[0].blocked_by|type)=="array" and (.unknown_cases[0].blocked_by|length)==0' "$work/unknown-top-report.json"
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and (.unknown_cases|length)==1 and (.unknown_cases[0].blocked_by|type)=="array" and (.unknown_cases[0].blocked_by|length)==0' "$work/malformed-report.json"
jq -e '.decision=="FAIL_CLOSED" and .resolution=="EXACT" and (.refuted_cases[0].reason=="GOOO_GRAPH_ACTIVITY_OR_EDGE_CONTRADICTION") and (.unknown_cases|length)==0' "$work/graph-edge-removal-report.json"

jq -n --slurpfile missing_second "$work/missing-second-report.json" --slurpfile missing_resource "$work/missing-resource-report.json" \
  --slurpfile digest_contradiction "$work/digest-contradiction-report.json" --slurpfile unknown_top "$work/unknown-top-report.json" \
  --slurpfile malformed "$work/malformed-report.json" --slurpfile graph_edge "$work/graph-edge-removal-report.json" '
  [{id:"missing-second-receipt",expected:"UNKNOWN",decision:$missing_second[0].decision,resolution:$missing_second[0].resolution,unknown:$missing_second[0].unknown_cases[0]},
   {id:"missing-resource-field",expected:"UNKNOWN",decision:$missing_resource[0].decision,resolution:$missing_resource[0].resolution,unknown:$missing_resource[0].unknown_cases[0]},
   {id:"digest-contradiction",expected:"REFUTED",decision:$digest_contradiction[0].decision,resolution:$digest_contradiction[0].resolution,refuted:$digest_contradiction[0].refuted_cases[0]},
   {id:"unknown-top-decision",expected:"FAIL_CLOSED",decision:$unknown_top[0].decision,resolution:$unknown_top[0].resolution,unknown:$unknown_top[0].unknown_cases[0]},
   {id:"malformed-receipt",expected:"FAIL_CLOSED",decision:$malformed[0].decision,resolution:$malformed[0].resolution,unknown:$malformed[0].unknown_cases[0]},
   {id:"graph-edge-removal",expected:"REFUTED",decision:$graph_edge[0].decision,resolution:$graph_edge[0].resolution,refuted:$graph_edge[0].refuted_cases[0]}]
' > "$work/counterexamples.json"

jq -r '"### Gooo trace debugger experiment\n- decision: \(.decision) / \(.resolution)\n- semantic replay: \(.replay.equal) (\(.replay.schema))\n- indicators: \(.summary.coordinates.satisfied)/\(.summary.coordinates.total)\n- USER: \(.views[0].satisfied)/\(.views[0].total)\n- TOOL_AUTHOR: \(.views[1].satisfied)/\(.views[1].total)\n- GOVERNOR: \(.views[2].satisfied)/\(.views[2].total)\n- debug receipts: \(.summary.debug_receipts)\n- resource observations: \(.summary.resource_observations)\n- build wall_ms/RSS KiB: \(.build.wall_ms)/\(.build.peak_rss_kib)\n- test wall_ms/RSS KiB: \(.test.wall_ms)/\(.test.peak_rss_kib)\n- runtime wall_ms/RSS KiB: [\(.runtime_observations[0].wall_ms)/\(.runtime_observations[0].peak_rss_kib), \(.runtime_observations[1].wall_ms)/\(.runtime_observations[1].peak_rss_kib)]\n- Gooo graph activities/edges: \(.graph.activity_count)/\(.graph.edge_count)\n- receipt: \(.digest)"' "$work/report.json" >> "$GITHUB_STEP_SUMMARY"
