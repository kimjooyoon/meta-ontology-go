#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"
root="${GITHUB_WORKSPACE:-$(pwd)}"
out="${RUNNER_TEMP:-/tmp}/language-utility-evidence"
mkdir -p "$out/evidence" "$out/package"
cd "$root"

phase="INITIALIZE"
preflight="$out/preflight.json"
write_preflight() {
  jq -n --arg decision "$1" --arg resolution "$2" --arg reason "$3" \
    --arg phase "$phase" --arg subject "$HEAD_SHA" \
    '{schema:"gooo/language-utility-preflight/v1",decision:$decision,resolution:$resolution,
      reason:$reason,phase:$phase,subject_sha:$subject}' > "$preflight"
}
finish_preflight() {
  code=$?; trap - EXIT
  if test "$code" -ne 0; then
    write_preflight "FAIL_CLOSED" "LOWER_RESOLUTION" "${phase}_FAILED"
  fi
  exit "$code"
}
trap finish_preflight EXIT
write_preflight "UNKNOWN" "LOWER_RESOLUTION" "PREFLIGHT_RUNNING"

phase="GO_FIX"
go fix ./cmd/language-utility-witness ./internal/meta/languageutility
git diff --exit-code -- cmd/language-utility-witness internal/meta/languageutility
phase="GOFMT"
mapfile -t go_files < <(find cmd/language-utility-witness internal/meta/languageutility -type f -name '*.go' -print | sort)
gofmt -w "${go_files[@]}"
git diff --exit-code -- cmd/language-utility-witness internal/meta/languageutility
phase="GO_TEST"
go test ./cmd/language-utility-witness ./internal/meta/languageutility
phase="USE_CASES"
bash scripts/ci-plan-usecase/main.sh
bash scripts/language-source-execution/main.sh
bash scripts/language-example-experiment/main.sh
EXACT_SHA="$HEAD_SHA" bash scripts/language-package-execution/main.sh "$out/package"

phase="EVIDENCE_COPY"
cp "${RUNNER_TEMP:-/tmp}/gooo-ci-plan/scorecard.json" "$out/evidence/ci-plan.json"
cp source-execution-output/artifact.json "$out/evidence/source-execution.json"
cp "${RUNNER_TEMP:-/tmp}/language-example-experiment/report.json" "$out/evidence/artifact-emission.json"
cp "${RUNNER_TEMP:-/tmp}/language-profile-experiment/report.json" "$out/evidence/profiling.json"
cp "${RUNNER_TEMP:-/tmp}/language-debug-experiment/report.json" "$out/evidence/debugging.json"
cp "${RUNNER_TEMP:-/tmp}/language-debug-experiment/counterexamples.json" "$out/debugging-counterexamples.json"
cp "$out/package/report.json" "$out/evidence/package-execution.json"
cp "${RUNNER_TEMP:-/tmp}/language-debug-experiment/gooo-graph.json" "$out/gooo-graph.json"
cp "${RUNNER_TEMP:-/tmp}/language-debug-experiment/graph-observation.json" "$out/graph-observation.json"
cp "${RUNNER_TEMP:-/tmp}/language-debug-experiment/program.gooo" "$out/program.gooo"

digest() { printf 'sha256:%s' "$(sha256sum "$1" | cut -d' ' -f1)"; }
github_curl() {
  if test -n "${GITHUB_TOKEN:-}"; then
    if curl -fsSL -H "Authorization: Bearer ${GITHUB_TOKEN}" -H "Accept: application/vnd.github+json" "$@"; then
      return 0
    fi
  fi
  checkout_header="$(git config --local --get-all http.https://github.com/.extraheader 2>/dev/null | head -n 1 || true)"
  if test -n "$checkout_header"; then
    if curl -fsSL -H "$checkout_header" -H "Accept: application/vnd.github+json" "$@"; then
      return 0
    fi
  fi
  curl -fsSL -H "Accept: application/vnd.github+json" "$@"
}
stages='["SOURCE_PRESENT","SYNTAX_ACCEPTED","SEMANTIC_ACCEPTED","OUTCOME_OBSERVED","DETERMINISTIC_REPLAY","RESOURCE_OBSERVED","USER_ARTIFACT_VERIFIED"]'
phase="OBSERVATION"
debug_report_digest="$(digest "$out/evidence/debugging.json")"
replay_activity_id="$(jq -er '.nodes[] | select(.kind=="Activity" and .name=="ObserveDebuggingDeterministicReplay") | .id' "$out/gooo-graph.json")"
resource_activity_id="$(jq -er '.nodes[] | select(.kind=="Activity" and .name=="ObserveDebuggingResourceObserved") | .id' "$out/gooo-graph.json")"
cell_input_id="gooo://meta/language-utility/entity/cell"
cell_output_id="gooo://meta/language-utility/entity/evidence"
replay_edges="$(jq --arg activity "$replay_activity_id" --arg input "$cell_input_id" --arg output "$cell_output_id" \
  '[.relations[] | select((.predicate=="used" and .subject==$activity and .object==$input) or (.predicate=="wasGeneratedBy" and .subject==$output and .object==$activity)) | {relation:.predicate,subject:.subject,object:.object}]' "$out/gooo-graph.json")"
resource_edges="$(jq --arg activity "$resource_activity_id" --arg input "$cell_input_id" --arg output "$cell_output_id" \
  '[.relations[] | select((.predicate=="used" and .subject==$activity and .object==$input) or (.predicate=="wasGeneratedBy" and .subject==$output and .object==$activity)) | {relation:.predicate,subject:.subject,object:.object}]' "$out/gooo-graph.json")"
jq -e '.activity_count==44 and .edge_count==88 and .debug_activity_count==2 and .debug_output_count==2 and .debug_used_edge_count==2 and .debug_generated_edge_count==2' "$out/graph-observation.json"
jq -e 'length==2' <<<"$replay_edges"
jq -e 'length==2' <<<"$resource_edges"
jq -n --arg schema 'gooo/language-utility-observation/v1' --arg contract 'gooo-language-utility-v1' \
  --arg head "$HEAD_SHA" --argjson stages "$stages" \
  --arg ci "$(digest "$out/evidence/ci-plan.json")" \
  --arg source "$(digest "$out/evidence/source-execution.json")" \
  --arg emit "$(digest "$out/evidence/artifact-emission.json")" \
  --arg profile "$(digest "$out/evidence/profiling.json")" \
  --arg debug "$debug_report_digest" \
  --arg package "$(digest "$out/evidence/package-execution.json")" \
  --arg replay_activity "$replay_activity_id" --arg resource_activity "$resource_activity_id" \
  --argjson replay_edges "$replay_edges" --argjson resource_edges "$resource_edges" \
  --arg input_id "$cell_input_id" --arg output_id "$cell_output_id" \
  --slurpfile graph "$out/graph-observation.json" '
  def closed($id;$producer;$file;$digest): $stages | map({use_case_id:$id,stage_id:.,state:"CLOSED",
    producer:$producer,step:("VERIFY_"+.),reason:"EVIDENCE_ACCEPTED",evidence_key:($id+".report"),
    evidence_path:("evidence/"+$file),evidence_digest:$digest});
  def opened($id;$stage;$reason): {use_case_id:$id,stage_id:$stage,state:"OPEN",producer:("ci:"+$id),
    step:("COLLECT_"+$stage),reason:$reason};
  def debug_closed($stage;$activity;$edges): {use_case_id:"debugging",stage_id:$stage,state:"CLOSED",
    producer:"scripts/language-debug-experiment",step:("VERIFY_"+$stage),reason:"EVIDENCE_ACCEPTED",
    evidence_key:"debugging.report",evidence_path:"evidence/debugging.json",evidence_digest:$debug,
    meta_activity_id:$activity,meta_input_id:$input_id,meta_output_id:$output_id,
    activity_matches:1,output_matches:1,used_edge_matches:1,generated_edge_matches:1,causal_edges:$edges};
  {schema:$schema,contract_id:$contract,subject_sha:$head,repository_writes:0,graph:$graph[0],
   cells: ((closed("ci-plan-selection";"scripts/ci-plan-usecase";"ci-plan.json";$ci) +
    closed("source-execution";"scripts/language-source-execution";"source-execution.json";$source) +
    closed("artifact-emission";"scripts/language-example-experiment";"artifact-emission.json";$emit) +
    closed("profiling";"scripts/language-profile-experiment";"profiling.json";$profile) +
    closed("debugging";"scripts/language-debug-experiment";"debugging.json";$debug) +
    closed("package-execution";"scripts/language-package-execution";"package-execution.json";$package)) |
    map(if .use_case_id=="debugging" and .stage_id=="DETERMINISTIC_REPLAY" then
      debug_closed(.stage_id;$replay_activity;$replay_edges)
    elif .use_case_id=="debugging" and .stage_id=="RESOURCE_OBSERVED" then
      debug_closed(.stage_id;$resource_activity;$resource_edges)
    elif .use_case_id=="package-execution" and .stage_id=="RESOURCE_OBSERVED" then
      opened("package-execution";.stage_id;"PACKAGE_RESOURCES_NOT_OBSERVED") else . end))}
' > "$out/observation.json"

args=(-contract examples/language-utility/contract.json -observation "$out/observation.json" -report "$out/report.json" -program "$out/program.gooo")
phase="REDUCE"
go run ./cmd/language-utility-witness "${args[@]}"
go run ./cmd/language-utility-witness "${args[@]}" -check
phase="PROGRAM_PARSE"
go run ./cmd/gooo check "$out/program.gooo"
go run ./cmd/gooo graph dump "$out/program.gooo" > "$out/final-gooo-graph.json"
jq -e --slurpfile generated "$out/gooo-graph.json" \
  '.graph_hash==$generated[0].graph_hash and .relations==$generated[0].relations and ([.nodes[]|{id,kind,name}]|sort_by(.id))==([$generated[0].nodes[]|{id,kind,name}]|sort_by(.id))' \
  "$out/final-gooo-graph.json"
phase="ASSERT"
jq -e '.decision=="PROGRESS_OBSERVED" and .resolution=="EXACT" and .summary.closed_cells==41 and .summary.open_cells==1 and .summary.cells_total==42 and .summary.complete_use_cases==5 and .summary.use_cases_total==6 and .summary.remaining_cells==1 and .summary.unknown_cells==0 and .summary.refuted_cells==0 and .summary.closed_delta_from_floor==2 and .summary.complete_use_case_floor_delta==1' "$out/report.json"

phase="BASELINE_RECEIPT"
baseline_api="https://api.github.com/repos/${GITHUB_REPOSITORY}/actions/artifacts/9690576734"
github_curl "$baseline_api" > "$out/baseline-artifact.json"
jq -e --arg name "language-utility-evidence-57ac9ec486bbca69e447a8eba94e0ce3cd03ced0" \
  '.id==9690576734 and .name==$name and .digest=="sha256:d491d53556bebbde810fe83ce63aff292c9820474a177677d096c0e8f625ebf5" and .size_in_bytes==6987602' \
  "$out/baseline-artifact.json"
github_curl -L "$baseline_api/zip" -o "$out/baseline-artifact.zip"
baseline_zip_digest="$(digest "$out/baseline-artifact.zip")"
test "$baseline_zip_digest" = "sha256:d491d53556bebbde810fe83ce63aff292c9820474a177677d096c0e8f625ebf5"
baseline_extract="$out/baseline-extract"
mkdir -p "$baseline_extract"
unzip -q "$out/baseline-artifact.zip" -d "$baseline_extract"
baseline_report="$baseline_extract/report.json"
test -f "$baseline_report"
jq -e '.summary.closed_cells==39 and .summary.open_cells==3 and .summary.unknown_cells==0 and .summary.refuted_cells==0 and .summary.cells_total==42 and .summary.complete_use_cases==4 and .summary.use_cases_total==6 and .summary.remaining_cells==3' \
  "$baseline_report"
jq -n --slurpfile before "$baseline_report" --slurpfile after "$out/report.json" \
  --arg artifact_id "9690576734" --arg artifact_name "language-utility-evidence-57ac9ec486bbca69e447a8eba94e0ce3cd03ced0" \
  --arg artifact_digest "sha256:d491d53556bebbde810fe83ce63aff292c9820474a177677d096c0e8f625ebf5" \
  --argjson artifact_size 6987602 --arg zip_digest "$baseline_zip_digest" '
  ($before[0].summary) as $b | ($after[0].summary) as $a |
  {schema:"gooo/language-utility-progress-receipt/v1",
   baseline_artifact:{id:($artifact_id|tonumber),name:$artifact_name,digest:$artifact_digest,size_bytes:$artifact_size,downloaded_zip_digest:$zip_digest},
   contract:{cells_total_before:$b.cells_total,cells_total_after:$a.cells_total},
   before:{closed:$b.closed_cells,open:($b.open_cells+$b.unknown_cells),unknown:$b.unknown_cells,refuted:$b.refuted_cells,complete_use_cases:$b.complete_use_cases,remaining:$b.remaining_cells},
   after:{closed:$a.closed_cells,open:($a.open_cells+$a.unknown_cells),unknown:$a.unknown_cells,refuted:$a.refuted_cells,complete_use_cases:$a.complete_use_cases,remaining:$a.remaining_cells},
   delta_closed:($a.closed_cells-$b.closed_cells),delta_open:(($a.open_cells+$a.unknown_cells)-($b.open_cells+$b.unknown_cells)),
   delta_complete_use_cases:($a.complete_use_cases-$b.complete_use_cases),
   performance_comparison:"NONE_NO_COMPARABLE_PERFORMANCE_PAIR"}' > "$out/progress-receipt.json"
jq -e '.contract.cells_total_before==42 and .contract.cells_total_after==42 and .before=={closed:39,open:3,unknown:0,refuted:0,complete_use_cases:4,remaining:3} and .after=={closed:41,open:1,unknown:0,refuted:0,complete_use_cases:5,remaining:1} and .delta_closed==2 and .delta_open==-2 and .delta_complete_use_cases==1 and .performance_comparison=="NONE_NO_COMPARABLE_PERFORMANCE_PAIR"' "$out/progress-receipt.json"

phase="INVENTORY"
inventory_files=()
while IFS= read -r inventory_path; do
  test "$inventory_path" = "README.md" && continue
  inventory_files+=("$inventory_path")
done < <(git ls-files)
regular_files="${#inventory_files[@]}"
descendant_directories="$(printf '%s\n' "${inventory_files[@]}" | awk -F/ 'NF>1 {for (i=1; i<NF; i++) {d=$1; for (j=2; j<=i; j++) d=d "/" $j; print d}}' | sort -u | wc -l | tr -d '[:space:]')"
physical_lines=0
go_files=0
go_lines=0
gooo_files=0
gooo_lines=0
for inventory_path in "${inventory_files[@]}"; do
  line_count="$(wc -l < "$inventory_path" | tr -d '[:space:]')"
  physical_lines=$((physical_lines + line_count))
  case "$inventory_path" in
    *.go) go_files=$((go_files + 1)); go_lines=$((go_lines + line_count)) ;;
    *.gooo) gooo_files=$((gooo_files + 1)); gooo_lines=$((gooo_lines + line_count)) ;;
  esac
done
jq -n --arg subject "$HEAD_SHA" --argjson regular_files "$regular_files" \
  --argjson descendant_directories "$descendant_directories" --argjson physical_lines "$physical_lines" \
  --argjson go_files "$go_files" --argjson go_lines "$go_lines" \
  --argjson gooo_files "$gooo_files" --argjson gooo_lines "$gooo_lines" '
  {schema:"gooo/repository-inventory/v1",basis:"git ls-files; root README.md excluded",subject_sha:$subject,
   regular_files:$regular_files,descendant_directories:$descendant_directories,physical_lines:$physical_lines,
   go_files:$go_files,go_lines:$go_lines,gooo_files:$gooo_files,gooo_lines:$gooo_lines,
   repository_writes:0,local_test_executions:0,cross_project_required_gates:0,
   mutation_authority:false,generation_authority:false,repair_authority:false,merge_authority:false}' > "$out/inventory.json"
phase="SUMMARY"
{
  echo '## Gooo language utility portfolio'
  jq -r '"- progress: \(.summary.closed_cells)/\(.summary.cells_total) cells (\(.summary.progress_basis_points) bps)\n- complete use cases: \(.summary.complete_use_cases)/\(.summary.use_cases_total)\n- remaining: \(.summary.remaining_cells)\n- observation/utility/promotion complete: \(.summary.observation_complete)/\(.summary.utility_complete)/\(.summary.promotion_complete)\n- decision: \(.decision) / \(.resolution)\n- receipt: \(.digest)"' "$out/report.json"
  echo
  echo '### Exact progress receipt'
  jq -r '"- baseline artifact: \(.baseline_artifact.id) / \(.baseline_artifact.name) / \(.baseline_artifact.digest) / \(.baseline_artifact.size_bytes) bytes\n- closed: \(.before.closed) -> \(.after.closed)\n- open: \(.before.open) -> \(.after.open)\n- complete use cases: \(.before.complete_use_cases) -> \(.after.complete_use_cases)\n- remaining: \(.before.remaining) -> \(.after.remaining)\n- deltas: closed \(.delta_closed), open \(.delta_open), complete use cases \(.delta_complete_use_cases)\n- performance comparison: \(.performance_comparison)"' "$out/progress-receipt.json"
  echo
  echo '### Debugging evidence'
  jq -r '"- replay: \(.replay.equal) / \(.replay.schema)\n- runtime observations: \(.summary.resource_observations)\n- source digests: \(.runtime_observations[0].source_raw_digest), \(.runtime_observations[1].source_raw_digest)\n- semantic digests: \(.runtime_observations[0].source_semantic_digest), \(.runtime_observations[1].source_semantic_digest)\n- binary digest: \(.runtime_observations[0].binary_digest)\n- arguments: \(.runtime_observations[0].arguments | join(" ")) ; \(.runtime_observations[1].arguments | join(" "))\n- subject SHA: \(.runtime_observations[0].subject_sha)\n- output digests: \(.runtime_observations[0].output_digest), \(.runtime_observations[1].output_digest)\n- wall_ns/wall_ms: \(.runtime_observations[0].wall_ns)/\(.runtime_observations[0].wall_ms), \(.runtime_observations[1].wall_ns)/\(.runtime_observations[1].wall_ms)\n- peak RSS KiB: \(.runtime_observations[0].peak_rss_kib), \(.runtime_observations[1].peak_rss_kib)\n- build wall_ms/RSS KiB: \(.build.wall_ms)/\(.build.peak_rss_kib)\n- evaluator build wall_ms/RSS KiB: \(.evaluator_build.wall_ms)/\(.evaluator_build.peak_rss_kib)\n- test wall_ms/RSS KiB: \(.test.wall_ms)/\(.test.peak_rss_kib)\n- cache states: \(.build.cache_state) ; \(.evaluator_build.cache_state) ; \(.test.cache_state)\n- Go runtime receipts: \(.summary.compiler.go127_runtimes)\n- Gooo graph activities/edges: \(.graph.activity_count)/\(.graph.edge_count)\n- Gooo debug activities, outputs, used/generated edges: \(.graph.debug_activity_count)/\(.graph.debug_output_count)/\(.graph.debug_used_edge_count)/\(.graph.debug_generated_edge_count)"' "$out/evidence/debugging.json"
  echo
  echo '### UNKNOWN / REFUTED counterexamples'
  jq -r '.[] | "- \(.id): expected=\(.expected), decision=\(.decision)/\(.resolution)"' "$out/debugging-counterexamples.json"
  echo
  echo '### Repository inventory and authority'
  jq -r '"- files/directories: \(.regular_files)/\(.descendant_directories)\n- physical lines: \(.physical_lines)\n- Go files/lines: \(.go_files)/\(.go_lines)\n- Gooo files/lines: \(.gooo_files)/\(.gooo_lines)\n- repository writes/local tests/cross-project gates: \(.repository_writes)/\(.local_test_executions)/\(.cross_project_required_gates)\n- mutation/generation/repair/merge authority: \(.mutation_authority)/\(.generation_authority)/\(.repair_authority)/\(.merge_authority)"' "$out/inventory.json"
  echo; echo '| Use case | Closed | Total | Complete |'; echo '|---|---:|---:|---|'
  jq -r '.use_cases[] | "| `\(.id)` | \(.closed_cells) | \(.total_cells) | \(.complete) |"' "$out/report.json"
  echo; echo '### Open claim coordinates'
  jq -r '.cells[] | select(.state!="CLOSED") | "- `\(.use_case_id)/\(.stage_id)`: `\(.step)/\(.reason)`"' "$out/report.json"
} >> "$GITHUB_STEP_SUMMARY"
phase="CLOSED"
write_preflight "PASS" "EXACT" "UTILITY_OBSERVATION_CLOSED"
trap - EXIT
