#!/usr/bin/env bash
# CI-only observer. Reuse the already-built CLI and existing profile witness.
set -Eeuo pipefail
if [ "$#" -ne 3 ]; then
  printf '%s\n' 'expected: work-directory support-directory exact-head' >&2
  exit 2
fi
work="$1"
support="$2"
head="$3"
graph_dir="$support/graph-dump"
mkdir -p "$graph_dir"
input="$work/input/source.gooo"
witness="$work/first/semantic-ir.json"
input_digest="$(sha256sum "$input" | awk '{print $1}')"
witness_digest="sha256:$(sha256sum "$witness" | awk '{print $1}')"
cli_digest="sha256:$(sha256sum "$work/gooo" | awk '{print $1}')"
unknown_receipt() {
  jq -n --arg head "$head" --arg reason "$1" \
    '{schema:"gooo/entity-fields-graph-support/v1", head_sha:$head, decision:"UNKNOWN", reason:$reason,
      unknown:{stage:"GRAPH_SUPPORT",step:"OBSERVE_PUBLIC_CLI",reason:$reason,unknown_class:"DIRECT_MISSING",
      next_operation:"COMPLETE_CANONICAL_GRAPH_OBSERVATION",blocked_by:[]}}' > "$support/graph-support-receipt.json"
}
unknown_receipt GRAPH_OBSERVATION_INCOMPLETE
env -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE git status --porcelain=v1 --untracked-files=all > "$graph_dir/status-before.txt"
for attempt in first replay; do
  exit_code=0
  /usr/bin/time -f '{"elapsed_seconds":"%e","peak_rss_kib":%M}' -o "$graph_dir/$attempt-runtime-raw.json" \
    "$work/gooo" graph dump "$input" > "$graph_dir/$attempt.json" 2> "$graph_dir/$attempt-stderr.txt" || exit_code=$?
  jq -n --arg attempt "$attempt" --argjson exit_code "$exit_code" \
    '{attempt:$attempt,exit_code:$exit_code}' > "$graph_dir/$attempt-invocation.json"
  if [ "$exit_code" -ne 0 ]; then
    unknown_receipt PUBLIC_GRAPH_DUMP_FAILED
    exit "$exit_code"
  fi
  jq '{wall_ms:((.elapsed_seconds|tonumber)*1000|round),peak_rss_kib,
       observation:"PUBLIC_GRAPH_DUMP",improvement:"UNKNOWN",clock_resolution_ms:10}' \
    "$graph_dir/$attempt-runtime-raw.json" > "$graph_dir/$attempt-runtime.json"
done
replay_equal=false
if cmp -s "$graph_dir/first.json" "$graph_dir/replay.json"; then replay_equal=true; fi
env -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE git status --porcelain=v1 --untracked-files=all > "$graph_dir/status-after.txt"
repository_snapshot_equal=false
if cmp -s "$graph_dir/status-before.txt" "$graph_dir/status-after.txt"; then repository_snapshot_equal=true; fi
: > "$graph_dir/cases.ndjson"
assess_case() {
  local name="$1" expected="$2" observation="$3"
  local assessment="$graph_dir/$name-assessment.json"
  if ! jq -n --slurpfile graph "$observation" --slurpfile witness "$witness" --arg source_digest "$input_digest" \
    -f scripts/entity-fields-graph-support.jq > "$assessment"; then
    jq -n '{schema:"gooo/entity-fields-graph-comparison/v1",decision:"REFUTED",reason:"OBSERVATION_DECODE_OR_SHAPE_INVALID",
      scope:"CANONICAL_FIXTURE_STRUCTURAL_PARITY",mismatches:["decode_or_shape"],unknown:null}' > "$assessment"
  fi
  jq -n --arg name "$name" --arg expected "$expected" --slurpfile assessment "$assessment" \
    '{name:$name,expected_decision:$expected,assessment:$assessment[0]}' >> "$graph_dir/cases.ndjson"
  jq -e --arg expected "$expected" '.decision == $expected and
    (if .decision == "UNKNOWN" then (.unknown | type == "object" and
      has("stage") and has("step") and has("reason") and has("unknown_class") and has("next_operation") and has("blocked_by"))
    else .unknown == null end)' "$assessment" > /dev/null
}
assess_case positive PASS "$graph_dir/first.json"
jq '.source_digest = ("0" * 64)' "$graph_dir/first.json" > "$graph_dir/wrong-source.json"
assess_case wrong-source REFUTED "$graph_dir/wrong-source.json"
jq '.nodes |= map(if .kind == "Entity" then .fields[0].name += "__tampered" else . end)' \
  "$graph_dir/first.json" > "$graph_dir/wrong-field.json"
assess_case wrong-field REFUTED "$graph_dir/wrong-field.json"
jq '.relations[0].object += "#tampered"' "$graph_dir/first.json" > "$graph_dir/wrong-relation.json"
assess_case wrong-relation REFUTED "$graph_dir/wrong-relation.json"
jq '.nodes |= map(select(.kind != "Entity"))' "$graph_dir/first.json" > "$graph_dir/missing-entity.json"
assess_case missing-entity REFUTED "$graph_dir/missing-entity.json"
: > "$graph_dir/missing-observation.json"
assess_case missing-observation UNKNOWN "$graph_dir/missing-observation.json"
jq -n --arg head "$head" --arg source_digest "$input_digest" --arg witness_digest "$witness_digest" \
  --arg cli_digest "$cli_digest" --arg graph_digest "sha256:$(sha256sum "$graph_dir/first.json" | awk '{print $1}')" \
  --arg replay_digest "sha256:$(sha256sum "$graph_dir/replay.json" | awk '{print $1}')" \
  --argjson replay_equal "$replay_equal" --argjson snapshot_equal "$repository_snapshot_equal" \
  --slurpfile comparison "$graph_dir/positive-assessment.json" --slurpfile cases "$graph_dir/cases.ndjson" \
  --slurpfile first_invocation "$graph_dir/first-invocation.json" --slurpfile replay_invocation "$graph_dir/replay-invocation.json" \
  --slurpfile first_runtime "$graph_dir/first-runtime.json" --slurpfile replay_runtime "$graph_dir/replay-runtime.json" \
  '{schema:"gooo/entity-fields-graph-support/v1",head_sha:$head,
    decision:(if $comparison[0].decision == "PASS" and $replay_equal and $snapshot_equal then "PASS" else "REFUTED" end),
    reason:(if ($replay_equal|not) then "GRAPH_REPLAY_MISMATCH" elif ($snapshot_equal|not) then "REPOSITORY_SNAPSHOT_CHANGED"
      else $comparison[0].reason end),unknown:null,
    fixture:"examples/entity-fields-v1/main.gooo",source_digest:$source_digest,witness_digest:$witness_digest,
    cli_digest:$cli_digest,graph_digest:$graph_digest,replay_digest:$replay_digest,
    public_cli_invocations:([$first_invocation[0],$replay_invocation[0]]|length),replay_comparisons:1,replay_equal:$replay_equal,
    repository_snapshot_equal:$snapshot_equal,write_claim:"BEFORE_AFTER_STATUS_ONLY_NOT_TRANSIENT_WRITE_PROOF",
    evidence_origin:"CALLER_OWNED_TEMP_OUTPUT_ONLY",comparison:$comparison[0],boundary_cases:$cases,
    runtime:{first:$first_runtime[0],replay:$replay_runtime[0]},improvement:"UNKNOWN"}' > "$support/graph-support-receipt.json"
jq -e '.decision == "PASS" and (.boundary_cases|length) == 6 and
  all(.boundary_cases[]; .assessment.decision == .expected_decision)' "$support/graph-support-receipt.json" > /dev/null
