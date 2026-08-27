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
cp "$out/package/report.json" "$out/evidence/package-execution.json"

digest() { printf 'sha256:%s' "$(sha256sum "$1" | cut -d' ' -f1)"; }
stages='["SOURCE_PRESENT","SYNTAX_ACCEPTED","SEMANTIC_ACCEPTED","OUTCOME_OBSERVED","DETERMINISTIC_REPLAY","RESOURCE_OBSERVED","USER_ARTIFACT_VERIFIED"]'
phase="OBSERVATION"
jq -n --arg schema 'gooo/language-utility-observation/v1' --arg contract 'gooo-language-utility-v1' \
  --arg head "$HEAD_SHA" --argjson stages "$stages" \
  --arg ci "$(digest "$out/evidence/ci-plan.json")" \
  --arg source "$(digest "$out/evidence/source-execution.json")" \
  --arg emit "$(digest "$out/evidence/artifact-emission.json")" \
  --arg profile "$(digest "$out/evidence/profiling.json")" \
  --arg debug "$(digest "$out/evidence/debugging.json")" \
  --arg package "$(digest "$out/evidence/package-execution.json")" '
  def closed($id;$producer;$file;$digest): $stages | map({use_case_id:$id,stage_id:.,state:"CLOSED",
    producer:$producer,step:("VERIFY_"+.),reason:"EVIDENCE_ACCEPTED",evidence_key:($id+".report"),
    evidence_path:("evidence/"+$file),evidence_digest:$digest});
  def opened($id;$stage;$reason): {use_case_id:$id,stage_id:$stage,state:"OPEN",producer:("ci:"+$id),
    step:("COLLECT_"+$stage),reason:$reason};
  {schema:$schema,contract_id:$contract,subject_sha:$head,repository_writes:0,
   cells: ((closed("ci-plan-selection";"scripts/ci-plan-usecase";"ci-plan.json";$ci) +
    closed("source-execution";"scripts/language-source-execution";"source-execution.json";$source) +
    closed("artifact-emission";"scripts/language-example-experiment";"artifact-emission.json";$emit) +
    closed("profiling";"scripts/language-profile-experiment";"profiling.json";$profile) +
    closed("debugging";"scripts/language-debug-experiment";"debugging.json";$debug) +
    closed("package-execution";"scripts/language-package-execution";"package-execution.json";$package)) |
    map(if .use_case_id=="debugging" and .stage_id=="DETERMINISTIC_REPLAY" then
      opened("debugging";.stage_id;"DEBUG_REPLAY_NOT_EXECUTED")
    elif .use_case_id=="debugging" and .stage_id=="RESOURCE_OBSERVED" then
      opened("debugging";.stage_id;"DEBUG_RESOURCES_NOT_OBSERVED")
    elif .use_case_id=="package-execution" and .stage_id=="RESOURCE_OBSERVED" then
      opened("package-execution";.stage_id;"PACKAGE_RESOURCES_NOT_OBSERVED") else . end))}
' > "$out/observation.json"

args=(-contract examples/language-utility/contract.json -observation "$out/observation.json" -report "$out/report.json" -program "$out/program.gooo")
phase="REDUCE"
go run ./cmd/language-utility-witness "${args[@]}"
go run ./cmd/language-utility-witness "${args[@]}" -check
phase="PROGRAM_PARSE"
go run ./cmd/gooo check "$out/program.gooo"
phase="ASSERT"
jq -e '.decision=="PROGRESS_OBSERVED" and .resolution=="EXACT" and .summary.closed_cells==39 and .summary.cells_total==42 and .summary.complete_use_cases==4 and .summary.use_cases_total==6 and .summary.remaining_cells==3 and .summary.unknown_cells==0 and .summary.refuted_cells==0' "$out/report.json"
phase="SUMMARY"
{
  echo '## Gooo language utility portfolio'
  jq -r '"- progress: \(.summary.closed_cells)/\(.summary.cells_total) cells (\(.summary.progress_basis_points) bps)\n- complete use cases: \(.summary.complete_use_cases)/\(.summary.use_cases_total)\n- remaining: \(.summary.remaining_cells)\n- observation/utility/promotion complete: \(.summary.observation_complete)/\(.summary.utility_complete)/\(.summary.promotion_complete)\n- decision: \(.decision) / \(.resolution)\n- receipt: \(.digest)"' "$out/report.json"
  echo; echo '| Use case | Closed | Total | Complete |'; echo '|---|---:|---:|---|'
  jq -r '.use_cases[] | "| `\(.id)` | \(.closed_cells) | \(.total_cells) | \(.complete) |"' "$out/report.json"
  echo; echo '### Open claim coordinates'
  jq -r '.cells[] | select(.state!="CLOSED") | "- `\(.use_case_id)/\(.stage_id)`: `\(.step)/\(.reason)`"' "$out/report.json"
} >> "$GITHUB_STEP_SUMMARY"
phase="CLOSED"
write_preflight "PASS" "EXACT" "UTILITY_OBSERVATION_CLOSED"
trap - EXIT
