#!/usr/bin/env bash
# Execute only in GitHub Actions. Output and executable are caller-owned.
set -euo pipefail
cli="${1:?path to the CI-built gooo executable}"
out="${2:?new external observation directory}"
repo="$(pwd -P)"
out="$(realpath -m "$out")"
case "$out" in "$repo"|"$repo"/*) printf 'output must be outside the repository\n' >&2; exit 1;; esac
test ! -e "$out"
mkdir -p "$out"
source_path=examples/domain-observation/main.gooo
input_path=examples/domain-observation/input.json
sha256sum "$source_path" "$input_path" examples/domain-observation/definition.gooo > "$out/source-before.sha256"

# A concrete execution, not generation of a binding-free substitute.
start_ns="$(date +%s%N)"
/usr/bin/time -f '%M' -o "$out/first-peak-rss-kib.txt" \
  "$cli" run --json --entry ObserveBudget --input "$input_path" "$source_path" > "$out/first.json"
wall_ms=$(( ($(date +%s%N) - start_ns) / 1000000 ))
"$cli" run --json --entry ObserveBudget --input "$input_path" "$source_path" > "$out/replay.json"
"$cli" compare --json "$out/first.json" "$out/replay.json" > "$out/replay-comparison.json"
jq -e '.state == "CLOSED" and .reason == "DETERMINISTIC_REPLAY"' "$out/replay-comparison.json" > /dev/null
jq -e '.decision == "PASS" and .execution.apply_calls == 3 and .execution.deliveries == 2 and
  .execution.activities == ["ObserveBudget","ConsumeAttempt","PublishRemaining"] and
  .execution.results.ObserveBudget.value == 3 and .execution.results.ConsumeAttempt.value == 2 and
  .execution.results.PublishRemaining.value == 2' "$out/first.json" > /dev/null

# The next invocation consumes the prior observed result, not a second literal.
# This is explicit caller orchestration, not an autonomous adoption capability.
jq '{value:.execution.results.PublishRemaining.value}' "$out/first.json" > "$out/next-input.json"
"$cli" run --json --entry ObserveBudget --input "$out/next-input.json" "$source_path" > "$out/next.json"
jq -e --slurpfile previous "$out/first.json" '.decision == "PASS" and
  .source_digest == $previous[0].source_digest and .semantic_fingerprint == $previous[0].semantic_fingerprint and
  .execution.plan_digest == $previous[0].execution.plan_digest and
  .execution.results.ObserveBudget.value == $previous[0].execution.results.PublishRemaining.value and
  .execution.results.PublishRemaining.value == 1 and .execution.apply_calls == 3 and .execution.deliveries == 2' \
  "$out/next.json" > /dev/null
if "$cli" compare --json "$out/first.json" "$out/next.json" > "$out/changed-input-comparison.json"; then
  printf 'different input incorrectly accepted as the same replay\n' >&2; exit 1
fi
jq -e '.state == "UNKNOWN" and .reason == "REPLAY_SCOPE_MISMATCH" and
  .same_plan == true and .same_input == false' "$out/changed-input-comparison.json" > /dev/null
if "$cli" propose-repair "$out/changed-input-comparison.json" --out "$out/unknown-repair" > "$out/unknown-repair.log" 2>&1; then
  printf 'UNKNOWN incorrectly became a repair candidate\n' >&2; exit 1
fi
test ! -e "$out/unknown-repair/repair-candidate.json"

# Preserve a real arithmetic failure and its partial execution, then recover.
printf '{"value":-9223372036854775808}\n' > "$out/underflow-input.json"
if "$cli" run --json --entry ObserveBudget --input "$out/underflow-input.json" "$source_path" > "$out/underflow.json"; then
  printf 'integer underflow unexpectedly succeeded\n' >&2; exit 1
fi
jq -e '.decision == "FAIL_CLOSED" and .reason == "VALUE_INTEGER_OVERFLOW" and
  .failure.stage == "EXECUTE" and .failure.step == "apply-int-add" and
  .execution.apply_calls == 2 and .execution.deliveries == 1 and
  .execution.activities == ["ObserveBudget","ConsumeAttempt"] and
  (.execution.results | keys) == ["ObserveBudget"]' "$out/underflow.json" > /dev/null
"$cli" run --json --entry ObserveBudget --input "$input_path" "$source_path" > "$out/after-failure.json"
"$cli" compare --json "$out/first.json" "$out/after-failure.json" > "$out/recovery-comparison.json"
jq -e '.state == "CLOSED" and .reason == "DETERMINISTIC_REPLAY"' "$out/recovery-comparison.json" > /dev/null

# Synthetic corruption demonstrates refutation, not a discovered runtime bug.
jq '.execution.results.PublishRemaining.value += 1' "$out/first.json" > "$out/synthetic-tampered.json"
if "$cli" compare --json "$out/first.json" "$out/synthetic-tampered.json" > "$out/synthetic-refutation.json"; then
  printf 'tampered receipt unexpectedly accepted\n' >&2; exit 1
fi
jq -e '.state == "REFUTED" and .reason == "REPLAY_RECEIPT_DIGEST_INVALID"' "$out/synthetic-refutation.json" > /dev/null
"$cli" propose-repair "$out/synthetic-refutation.json" --out "$out/synthetic-candidate" > "$out/synthetic-candidate.log"
jq -e '.trigger_state == "REFUTED" and .execution_allowed == false and .repository_writes == 0' \
  "$out/synthetic-candidate/repair-candidate.json" > /dev/null
sha256sum -c "$out/source-before.sha256" > "$out/source-after-check.txt"

jq -n --slurpfile first "$out/first.json" --slurpfile next "$out/next.json" \
  --slurpfile failure "$out/underflow.json" --slurpfile replay "$out/replay-comparison.json" \
  --slurpfile changed "$out/changed-input-comparison.json" --slurpfile recovery "$out/recovery-comparison.json" \
  --slurpfile refuted "$out/synthetic-refutation.json" \
  --argjson wall_ms "$wall_ms" --argjson peak_rss_kib "$(cat "$out/first-peak-rss-kib.txt")" \
  '{schema:"gooo/domain-budget-observation/v1",source_digest:$first[0].source_digest,
    semantic_fingerprint:$first[0].semantic_fingerprint,
    first:{input:$first[0].execution.results.ObserveBudget.value,
      output:$first[0].execution.results.PublishRemaining.value,
      apply_calls:$first[0].execution.apply_calls,deliveries:$first[0].execution.deliveries,
      wall_ms:$wall_ms,peak_rss_kib:$peak_rss_kib},
    next:{input:$next[0].execution.results.ObserveBudget.value,
      output:$next[0].execution.results.PublishRemaining.value,
      apply_calls:$next[0].execution.apply_calls,deliveries:$next[0].execution.deliveries},
    failure:{reason:$failure[0].reason,stage:$failure[0].failure.stage,step:$failure[0].failure.step,
      apply_calls:$failure[0].execution.apply_calls,deliveries:$failure[0].execution.deliveries},
    replay:$replay[0].state,changed_input:$changed[0].state,recovery:$recovery[0].state,
    synthetic_corruption:$refuted[0].state,synthetic_candidate_execution_allowed:false,
    fixture_files_unchanged:3,runtime_mode:"source-interpreter",
    generator_binding_support:"UNSUPPORTED",orchestration:"explicit-caller",
    external_utility:"UNKNOWN",source_repair_adoption:"NOT_IMPLEMENTED",improvement:"UNKNOWN"}' \
  > "$out/observation.json"
jq -r '"### Executed Gooo budget domain\n- first: \(.first.input) -> \(.first.output); applies=\(.first.apply_calls), deliveries=\(.first.deliveries)\n- next consumes prior output: \(.next.input) -> \(.next.output)\n- real failure: \(.failure.reason), \(.failure.stage)/\(.failure.step)\n- replay/recovery: \(.replay)/\(.recovery); changed-input comparison: \(.changed_input)\n- synthetic corruption: \(.synthetic_corruption); candidate cannot execute\n- first process: \(.first.wall_ms) ms, peak RSS \(.first.peak_rss_kib) KiB\n- utility/improvement: UNKNOWN; source-repair adoption: NOT_IMPLEMENTED"' \
  "$out/observation.json" > "$out/report.md"
