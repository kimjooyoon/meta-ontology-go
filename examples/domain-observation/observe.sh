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
repair_source=examples/domain-observation/repair.gooo
repair_input=examples/domain-observation/repair-input.json
repair_digest="sha256:$(sha256sum "$repair_source" | awk '{print $1}')"
sha256sum "$source_path" "$input_path" examples/domain-observation/definition.gooo "$repair_source" "$repair_input" > "$out/source-before.sha256"

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

# Gooo owns the feedback edge. The caller supplies only the initial value and bound.
"$cli" run --json --entry ObserveBudget --input "$input_path" --iterations 2 "$source_path" > "$out/continuation.json"
jq -e '.decision == "PASS" and .continuation.iterations_requested == 2 and
  .continuation.iterations_completed == 2 and .continuation.feedback_deliveries == 1 and
  (.continuation.executions | length) == 2 and .continuation.failure == null and
  .continuation.executions[0].results.PublishRemaining.value == 2 and
  .continuation.executions[1].results.ObserveBudget.value == 2 and
  .continuation.executions[1].results.PublishRemaining.value == 1' "$out/continuation.json" > /dev/null
# Project the already executed second receipt for the existing replay comparator.
# This projection does not choose input, perform an execution or grant authority.
jq '{schema:"gooo/value-execution-plan/v1",decision,source_path,source_digest,semantic_fingerprint,
  entry:"ObserveBudget",execution:.continuation.executions[1]}' "$out/continuation.json" > "$out/next.json"
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

# The continuation runtime stops on the second iteration's real failure.
printf '{"value":-9223372036854775807}\n' > "$out/late-underflow-input.json"
if "$cli" run --json --entry ObserveBudget --input "$out/late-underflow-input.json" --iterations 3 "$source_path" > "$out/failed-continuation.json"; then
  printf 'failed continuation unexpectedly reached its limit\n' >&2; exit 1
fi
jq -e '.decision == "FAIL_CLOSED" and .continuation.iterations_requested == 3 and
  .continuation.iterations_completed == 1 and .continuation.feedback_deliveries == 1 and
  (.continuation.executions | length) == 2 and .continuation.executions[1].apply_calls == 2 and
  .continuation.failure.code == "VALUE_INTEGER_OVERFLOW" and
  .continuation.failure.stage == "EXECUTE" and .continuation.failure.step == "apply-int-add"' \
  "$out/failed-continuation.json" > /dev/null

# Synthetic corruption demonstrates refutation, not a discovered runtime bug.
jq '.execution.results.PublishRemaining.value += 1' "$out/first.json" > "$out/synthetic-tampered.json"
if "$cli" compare --json "$out/first.json" "$out/synthetic-tampered.json" > "$out/synthetic-refutation.json"; then
  printf 'tampered receipt unexpectedly accepted\n' >&2; exit 1
fi
jq -e '.state == "REFUTED" and .reason == "REPLAY_RECEIPT_DIGEST_INVALID"' "$out/synthetic-refutation.json" > /dev/null
"$cli" propose-repair "$out/synthetic-refutation.json" --out "$out/synthetic-candidate" > "$out/synthetic-candidate.log"
jq -e '.trigger_state == "REFUTED" and .execution_allowed == false and .repository_writes == 0' \
  "$out/synthetic-candidate/repair-candidate.json" > /dev/null

# A source revision is an exact external candidate. The compiler does not edit
# repair.gooo, and the candidate is independently evaluated before contract
# preservation is checked through the typed CLI boundary.
"$cli" revise-source "$repair_source" --source-digest "$repair_digest" \
  --activity ObserveRepair --expected 'int.add:1' --replace 'int.add:0' \
  --reason VALUE_INTEGER_OVERFLOW --out "$out/source-revision" > "$out/source-revision.log"
jq -e '.execution_allowed == false and .repository_writes == 0 and
  .source_digest == "'"$repair_digest"'" and
  .next_operation == "EVALUATE_SOURCE_REVISION_INDEPENDENTLY"' \
  "$out/source-revision/revision.json" > /dev/null
"$cli" evaluate-revision "$repair_source" "$out/source-revision/candidate.gooo" \
  --revision "$out/source-revision/revision.json" --activity ObserveRepair \
  --input "$repair_input" --out "$out/source-revision-evaluation" > "$out/source-revision-evaluation.log"
jq -e '.state == "CLOSED" and .reason == "SOURCE_REVISION_COUNTEREXAMPLE_RECOVERED" and
	.accepted == false and .candidate_executed == true and .repository_writes == 0 and
	.next_operation == "VERIFY_SOURCE_REVISION_CONTRACT" and
	.baseline_failure.code == "VALUE_INTEGER_OVERFLOW"' \
	"$out/source-revision-evaluation/evaluation.json" > /dev/null
# Verify the broader contract explicitly. This candidate repairs one overflow
# counterexample but changes ordinary outputs, so the language must refute it
# and keep adoption blocked rather than silently accepting it.
printf '[0,1,2,3]\n' > "$out/source-revision-contract-inputs.json"
if "$cli" verify-revision-contract "$repair_source" "$out/source-revision/candidate.gooo" \
	--revision "$out/source-revision/revision.json" \
	--evaluation "$out/source-revision-evaluation/evaluation.json" \
	--activity ObserveRepair --inputs "$out/source-revision-contract-inputs.json" \
	--out "$out/source-revision-contract" > "$out/source-revision-contract.log" 2>&1; then
	printf 'contract-violating candidate unexpectedly passed verification\n' >&2
	exit 1
fi
jq -e '.state == "REFUTED" and .reason == "SOURCE_REVISION_CONTRACT_OUTPUT_CHANGED" and
	.contract_preservation == false and .accepted == false and
	.adoption_authorized == false and .repository_writes == 0' \
	"$out/source-revision-contract/evaluation.json" > /dev/null
# The rejected candidate remains an external input. Generate it, then reverse-
# observe its Go output without granting source-adoption authority.
candidate_generation="$out/source-revision-candidate-generation"
mkdir -p "$candidate_generation"
"$cli" generate "$out/source-revision/candidate.gooo" --out "$candidate_generation" > "$out/candidate-source-generation.log"
"$cli" analyze "$out/source-revision/candidate.gooo" --go "$candidate_generation/semantic.gooo.go" > "$candidate_generation/analyze.json"
jq -e '
	.schema_version == "analyzer-semantic-delta/v1" and
	.semantic_equal == true and
	.authority_semantic_digest == .observed_semantic_digest and
	.write_effect == "no-write" and
	(.digest | length == 64)
' "$candidate_generation/analyze.json" > /dev/null
sha256sum -c "$out/source-before.sha256" > "$out/source-after-check.txt"

jq -n --slurpfile first "$out/first.json" --slurpfile next "$out/next.json" \
  --slurpfile failure "$out/underflow.json" --slurpfile replay "$out/replay-comparison.json" \
  --slurpfile changed "$out/changed-input-comparison.json" --slurpfile recovery "$out/recovery-comparison.json" \
  --slurpfile continuation "$out/continuation.json" --slurpfile failed_continuation "$out/failed-continuation.json" \
  --slurpfile refuted "$out/synthetic-refutation.json" \
  --slurpfile source_revision "$out/source-revision/revision.json" \
	--slurpfile source_revision_evaluation "$out/source-revision-contract/evaluation.json" \
	--slurpfile candidate_generation "$out/source-revision-candidate-generation/analyze.json" \
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
    source_revision:{state:$source_revision_evaluation[0].state,reason:$source_revision_evaluation[0].reason,
      candidate_id:$source_revision[0].candidate_id,source_digest:$source_revision[0].source_digest,
      candidate_source_digest:$source_revision[0].candidate_source_digest,execution_allowed:$source_revision[0].execution_allowed,
      repository_writes:$source_revision_evaluation[0].repository_writes,
	  	contract_verification:{state:$source_revision_evaluation[0].state,reason:$source_revision_evaluation[0].reason,
	  	  contract_preservation:$source_revision_evaluation[0].contract_preservation,
	  	  accepted:$source_revision_evaluation[0].accepted,
	  	  adoption_authorized:$source_revision_evaluation[0].adoption_authorized,
	  	  next_operation:$source_revision_evaluation[0].next_operation},
	  	candidate_generated_go_reverse_observation:{semantic_equal:$candidate_generation[0].semantic_equal,
	  	  authority_semantic_digest:$candidate_generation[0].authority_semantic_digest,
	  	  observed_semantic_digest:$candidate_generation[0].observed_semantic_digest,
	  	  write_effect:$candidate_generation[0].write_effect}},
    fixture_files_unchanged:5,runtime_mode:"source-interpreter",
    generator_binding_support:"UNSUPPORTED",orchestration:"source-feedback",
    continuation:{requested:$continuation[0].continuation.iterations_requested,
      completed:$continuation[0].continuation.iterations_completed,
      feedback_deliveries:$continuation[0].continuation.feedback_deliveries,
      digest:$continuation[0].continuation.digest},
    failed_continuation:{completed:$failed_continuation[0].continuation.iterations_completed,
      observed_iterations:($failed_continuation[0].continuation.executions|length),
      feedback_deliveries:$failed_continuation[0].continuation.feedback_deliveries,
      failure:$failed_continuation[0].continuation.failure},
    external_utility:"UNKNOWN",source_repair_adoption:"EXPLICIT_CALLER",improvement:"UNKNOWN"}' \
  > "$out/observation.json"
 jq -r '"### Executed Gooo budget domain\n- first: \(.first.input) -> \(.first.output); applies=\(.first.apply_calls), deliveries=\(.first.deliveries)\n- next consumes prior output: \(.next.input) -> \(.next.output)\n- real failure: \(.failure.reason), \(.failure.stage)/\(.failure.step)\n- replay/recovery: \(.replay)/\(.recovery); changed-input comparison: \(.changed_input)\n- synthetic corruption: \(.synthetic_corruption); candidate cannot execute\n- source revision: \(.source_revision.state)/\(.source_revision.reason); contract verification: \(.source_revision.contract_verification.state)/\(.source_revision.contract_verification.reason)\n- rejected candidate generated Go reverse observation: \(.source_revision.candidate_generated_go_reverse_observation.semantic_equal), write effect=\(.source_revision.candidate_generated_go_reverse_observation.write_effect)\n- first process: \(.first.wall_ms) ms, peak RSS \(.first.peak_rss_kib) KiB\n- utility/improvement: UNKNOWN; source-repair adoption: BLOCKED_BY_CONTRACT"' \
  "$out/observation.json" > "$out/report.md"
printf '%s\n' '- source revision contract: candidate refuted and adoption remains blocked; see observation.json source_revision.contract_verification' >> "$out/report.md"
