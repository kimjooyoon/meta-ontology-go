#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

conformance_stage="${GOOO_CONFORMANCE_STAGE:-0}"
case "$conformance_stage" in
  0)
    echo "semantic conformance stage 0: Go verifier authoritative"
    ;;
  1|2|3)
    echo "semantic conformance stage $conformance_stage is not enabled; see .github/conformance-plan.md" >&2
    exit 1
    ;;
  *)
    echo "invalid GOOO_CONFORMANCE_STAGE: $conformance_stage" >&2
    exit 1
    ;;
esac

# The semantic verifier consumes the same deterministic projected workspace as
# repository-projection. This keeps its source policy observation on the CI
# materialization that closes the selected density residuals, rather than on
# the unprojected checkout. Missing projection inputs are fail-closed.
projection_work="${RUNNER_TEMP:-${TMPDIR:-/tmp}}/gooo-semantic-projection"
storage_root="${GOOO_STORAGE_ROOT:-}"
runner_temp="${RUNNER_TEMP:-${TMPDIR:-/tmp}}"
job_id="${GITHUB_JOB:-semantic}"
projection_evidence="$runner_temp/repository-materialize-$job_id/projector/evidence.json"
if [[ -z "$storage_root" ]]; then
  echo "semantic conformance: exact materialized storage root is required" >&2
  exit 1
fi
if [[ ! -s "$projection_evidence" ]]; then
  echo "semantic conformance: exact materialization projection evidence is unavailable" >&2
  exit 1
fi
head_sha="$(git rev-parse HEAD)"
export METRICS_COMMIT_SHA="$head_sha"
mkdir -p "$projection_work"
go run ./scripts/line-metrics \
  -root "$repo_root" \
  -storage-root "$repo_root" \
  -json > "$projection_work/split-source-metrics-raw.json"
go run ./bootstrap/meta-binding-witness \
  -root "$repo_root" \
  -metrics "$projection_work/split-source-metrics-raw.json" \
  -bound-metrics "$projection_work/split-source-metrics.json" \
  -report "$projection_work/split-binding-report.json" \
  -check
go run ./bootstrap/repository-cutover \
  --root "$repo_root" \
  --physical-root "$storage_root" \
  --authority-manifest "$storage_root/projection/catalog/manifest.json" \
  --expected-sha "$head_sha" \
  --evidence "$projection_work/cutover-evidence.json"
go run ./bootstrap/logical-split-planner \
  --root "$repo_root" \
  --evidence "$projection_evidence" \
  --expected-sha "$head_sha" \
  --output "$projection_work/split-plan.json"
go run ./bootstrap/line-density-rewriter \
  --root "$repo_root" \
  --plan "$projection_work/split-plan.json" \
  --expected-sha "$head_sha" \
  --output "$projection_work/density-report.json"

EXTRACTION_PARTIAL=0
refuted_failures="$projection_work/refuted-failures.json"
partial_subjects="$projection_work/partial-subjects.json"
printf '{}\n' > "$projection_work/projected-compile.json"
printf '[]\n' > "$refuted_failures"
printf '[]\n' > "$partial_subjects"
jq '[.counterexamples[]? | select(.decision == "REFUTED")]' \
  "$projection_work/split-plan.json" > "$projection_work/planner-refuted.json"
jq -s 'add | unique_by(tojson) | sort_by(tojson)' \
  "$refuted_failures" "$projection_work/planner-refuted.json" \
  > "$projection_work/planner-refuted-merged.json"
mv "$projection_work/planner-refuted-merged.json" "$refuted_failures"

record_extraction_report() {
  local report="$1"
  if [[ ! -s "$report" ]]; then
    return
  fi
  jq '[.failures[]? | select(.decision == "REFUTED")]' "$report" > "$projection_work/refuted-current.json"
  jq -s 'add | unique_by(tojson) | sort_by(tojson)' \
    "$refuted_failures" "$projection_work/refuted-current.json" \
    > "$projection_work/refuted-merged.json"
  mv "$projection_work/refuted-merged.json" "$refuted_failures"
  jq '(.subjects // [])' "$report" > "$projection_work/subjects-current.json"
  jq -s 'add | unique_by(tojson) | sort_by([(.logical // ""), tojson])' \
    "$partial_subjects" "$projection_work/subjects-current.json" \
    > "$projection_work/subjects-merged.json"
  mv "$projection_work/subjects-merged.json" "$partial_subjects"
}

write_extraction_result() {
  local report="$1"
  local output="$2"
  local fallback_reason="$3"
  local projected_metrics="$4"
  if [[ -s "$report" ]]; then
    jq --slurpfile historical_refuted "$refuted_failures" \
      --slurpfile historical_subjects "$partial_subjects" \
      --slurpfile raw_metrics "$projection_work/split-source-metrics-raw.json" \
      --slurpfile projected_metrics "$projected_metrics" \
      --slurpfile compile_observation "$projection_work/projected-compile.json" \
      --arg source_sha "$head_sha" --arg reason "$fallback_reason" '
      def stable_unique: unique_by(tojson) | sort_by(tojson);
      def metric_blockers($report; $metric):
        [$report.meta.indicators[]? |
          select(.metric_id == $metric and .blocking == true and .satisfied == false and .applicability != "NOT_APPLICABLE")] | length;
      (($historical_refuted[0] // []) + [.failures[]? | select(.decision == "REFUTED")] | stable_unique) as $refuted |
      ($refuted | length) as $refuted_count |
      (($historical_refuted[0] // []) + (.failures // []) | stable_unique) as $counterexamples |
      (($historical_subjects[0] // []) + (.subjects // []) |
        unique_by(tojson) | sort_by([(.logical // ""), tojson])) as $subjects |
      (metric_blockers($raw_metrics[0]; "gooo.metric.source.go-file-lines.v1")) as $raw_go_file_line_blockers |
      (metric_blockers($raw_metrics[0]; "gooo.metric.source.function-lines.v1")) as $raw_function_line_blockers |
      (metric_blockers($projected_metrics[0]; "gooo.metric.source.go-file-lines.v1")) as $projected_go_file_line_blockers |
      (metric_blockers($projected_metrics[0]; "gooo.metric.source.function-lines.v1")) as $projected_function_line_blockers |
      {
        schema: "gooo.semantic-conformance-extraction-result/v1",
        source_sha: $source_sha,
        decision: (if $refuted_count > 0 then "REFUTED" else "UNKNOWN" end),
        stage: "semantic-conformance",
        step: "extractor-observe",
        reason: (if $refuted_count > 0 then "KNOWN_CONTRADICTION_PERSISTED" else $reason end),
        unknown_class: (if $refuted_count > 0 then "KNOWN_CONTRADICTION" else "DIRECT_MISSING" end),
        next_operation: "restore-decomposition-evidence",
        blocked_by: ([$counterexamples[]?.blocked_by[]?] | unique | sort),
        blocker_ids: ([
          $refuted[] |
          .logical as $logical |
          (.blocker_id // "") as $blocker_id |
          if $blocker_id != "" then $blocker_id else
            ((.diagnostics // []) | map(select(startswith("declaration="))) | .[0] // "declaration=unknown") as $declaration |
            ($declaration | sub("^declaration="; "")) as $identity |
            ($logical + "#" + $identity)
          end
        ] | unique | sort),
        counterexamples: $counterexamples,
        partial_subjects: $subjects,
        partial_writes: {
          changed: ([$subjects[]?.changed_files[]?] | unique | sort),
          created: ([$subjects[]?.created_files[]?] | unique | sort)
        },
        raw_projected_progress: {
          raw_go_file_line_blockers: $raw_go_file_line_blockers,
          projected_go_file_line_blockers: $projected_go_file_line_blockers,
          raw_function_line_blockers: $raw_function_line_blockers,
          projected_function_line_blockers: $projected_function_line_blockers
        },
        projected_compile: (if (($compile_observation[0].schema // "") != "") then $compile_observation[0] else null end),
        extraction: {
          observed: ([.indicators[]? | select(.id == "extraction.observed") | .value] | first // 0),
          applied: ([.indicators[]? | select(.id == "extraction.applied") | .value] | first // 0),
          created: ([.indicators[]? | select(.id == "extraction.created") | .value] | first // 0),
          unhandled: ([.indicators[]? | select(.id == "extraction.unhandled") | .value] | first // 0)
        }
      }' "$report" > "$output"
    return
  fi
  jq -n --slurpfile historical_refuted "$refuted_failures" \
    --slurpfile historical_subjects "$partial_subjects" \
    --slurpfile raw_metrics "$projection_work/split-source-metrics-raw.json" \
    --slurpfile projected_metrics "$projected_metrics" \
    --slurpfile compile_observation "$projection_work/projected-compile.json" \
    --arg source_sha "$head_sha" --arg reason "$fallback_reason" \
    'def metric_blockers($report; $metric):
       [$report.meta.indicators[]? |
         select(.metric_id == $metric and .blocking == true and .satisfied == false and .applicability != "NOT_APPLICABLE")] | length;
     ([$historical_refuted[0][]?] | unique_by(tojson) | sort_by(tojson)) as $refuted |
     ([$historical_subjects[0][]?] | unique_by(tojson) | sort_by([(.logical // ""), tojson])) as $subjects |
     (metric_blockers($raw_metrics[0]; "gooo.metric.source.go-file-lines.v1")) as $raw_go_file_line_blockers |
     (metric_blockers($raw_metrics[0]; "gooo.metric.source.function-lines.v1")) as $raw_function_line_blockers |
     (metric_blockers($projected_metrics[0]; "gooo.metric.source.go-file-lines.v1")) as $projected_go_file_line_blockers |
     (metric_blockers($projected_metrics[0]; "gooo.metric.source.function-lines.v1")) as $projected_function_line_blockers |
     {schema:"gooo.semantic-conformance-extraction-result/v1", source_sha:$source_sha,
      decision:(if ($refuted | length) > 0 then "REFUTED" else "UNKNOWN" end),
      stage:"semantic-conformance", step:"extractor-observe",
      reason:(if ($refuted | length) > 0 then "KNOWN_CONTRADICTION_PERSISTED" else $reason end),
      unknown_class:(if ($refuted | length) > 0 then "KNOWN_CONTRADICTION" else "DIRECT_MISSING" end),
      next_operation:"restore-decomposition-evidence",
      blocked_by:([$refuted[]?.blocked_by[]?] | unique | sort),
      blocker_ids:([$refuted[]?.blocker_id // empty] | unique | sort),
      counterexamples:$refuted, partial_subjects:$subjects,
      partial_writes:{changed:([$subjects[]?.changed_files[]?] | unique | sort),
        created:([$subjects[]?.created_files[]?] | unique | sort)},
      raw_projected_progress:{raw_go_file_line_blockers:$raw_go_file_line_blockers,
        projected_go_file_line_blockers:$projected_go_file_line_blockers,
        raw_function_line_blockers:$raw_function_line_blockers,
        projected_function_line_blockers:$projected_function_line_blockers},
      projected_compile:(if (($compile_observation[0].schema // "") != "") then $compile_observation[0] else null end),
      extraction:{observed:0, applied:0, created:0, unhandled:null}}' > "$output"
}

run_projected_compile() {
  local stdout="$projection_work/projected-compile.stdout"
  local stderr="$projection_work/projected-compile.stderr"
  set +e
  go test ./... > "$stdout" 2> "$stderr"
  local status=$?
  set -e
  local stdout_bytes
  local stderr_bytes
  local stdout_digest
  local stderr_digest
  stdout_bytes="$(wc -c < "$stdout" | tr -d ' ')"
  stderr_bytes="$(wc -c < "$stderr" | tr -d ' ')"
  stdout_digest="sha256:$(sha256sum "$stdout" | cut -d' ' -f1)"
  stderr_digest="sha256:$(sha256sum "$stderr" | cut -d' ' -f1)"
  jq -n \
    --arg cwd "$repo_root" \
    --arg stdout_digest "$stdout_digest" \
    --arg stderr_digest "$stderr_digest" \
    --argjson exit_code "$status" \
    --argjson stdout_bytes "$stdout_bytes" \
    --argjson stderr_bytes "$stderr_bytes" \
    '{schema:"gooo.projected-compile-observation/v1", command:["go","test","./..."], cwd:$cwd,
      exit_code:$exit_code, stdout_bytes:$stdout_bytes, stdout_digest:$stdout_digest,
      stderr_bytes:$stderr_bytes, stderr_digest:$stderr_digest, observed:($exit_code == 0)}' \
    > "$projection_work/projected-compile.json"
  return "$status"
}

run_extraction_pass() {
  local output="$1"
  local fixed_point="$2"
  local plan_path="$3"
  local density_path="$4"
  EXTRACTION_PARTIAL=0
  set +e
  if [[ "$fixed_point" == true ]]; then
    go run ./bootstrap/function-extractor \
      --root "$repo_root" \
      --plan "$plan_path" \
      --density-report "$density_path" \
      --expected-sha "$head_sha" \
      --output "$output" \
      --fixed-point
  else
    go run ./bootstrap/function-extractor \
      --root "$repo_root" \
      --plan "$plan_path" \
      --density-report "$density_path" \
      --expected-sha "$head_sha" \
      --output "$output"
  fi
  local status=$?
  set -e
  if [[ "$status" -eq 0 ]]; then
    return 0
  fi
  if [[ ! -s "$output" ]]; then
    write_extraction_result "$output" "$projection_work/extraction-counterexample.json" "EXTRACTION_REPORT_UNAVAILABLE" "$projection_work/split-source-metrics-raw.json"
    cat "$projection_work/extraction-counterexample.json"
    return "$status"
  fi
  record_extraction_report "$output"
  cat "$output"
  if jq -e '.indicators | any(.[]; .id == "extraction.applied" and .value > 0)' "$output" >/dev/null; then
    EXTRACTION_PARTIAL=1
    return 0
  fi
  write_extraction_result "$output" "$projection_work/extraction-counterexample.json" "EXTRACTION_APPLIED_ZERO" "$projection_work/split-source-metrics-raw.json"
  cat "$projection_work/extraction-counterexample.json"
  return "$status"
}

set +e
run_extraction_pass "$projection_work/extraction-report.json" false \
  "$projection_work/split-plan.json" "$projection_work/density-report.json"
extraction_status=$?
set -e
if [[ "$extraction_status" -ne 0 ]]; then
  exit "$extraction_status"
fi
go run ./scripts/source-splitter \
  -root "$repo_root" \
  -metrics "$projection_work/split-source-metrics.json" \
  -sha "$head_sha" \
  -plan "$projection_work/split-plan.json" \
  -output "$projection_work/split-report.json"

last_extraction="$projection_work/extraction-report.json"
post_split_metrics="$projection_work/split-source-metrics-projected.json"
go run ./scripts/line-metrics \
  -root "$repo_root" \
  -storage-root "$storage_root" \
  -json > "$post_split_metrics"
set +e
run_projected_compile
compile_status=$?
set -e
if [[ "$compile_status" -ne 0 ]]; then
  write_extraction_result "$last_extraction" "$projection_work/projected-compile-failure.json" "PROJECTED_COMPILE_FAILED" "$post_split_metrics"
  cat "$projection_work/projected-compile-failure.json"
  exit "$compile_status"
fi

# Re-observe the active materialized workspace after each projection pass.
# The fixed point is a bounded metaprogram operation: generated helpers are
# ordinary source subjects, and a non-decreasing residual is an explicit
# unknown rather than an authorization to repeat names forever.
previous_blocking=-1
fixed_point_closed=false
last_metrics="$projection_work/split-source-metrics-raw.json"
for iteration in 1 2 3 4 5 6 7 8; do
  pass_work="$projection_work/fixed-point-$iteration"
  mkdir -p "$pass_work"
  metrics="$pass_work/source-metrics.json"
  plan="$pass_work/split-plan.json"
  density="$pass_work/density-report.json"
  extraction="$pass_work/extraction-report.json"
  split="$pass_work/split-report.json"
  go run ./scripts/line-metrics \
    -root "$repo_root" \
    -storage-root "$storage_root" \
    -json > "$metrics"
  go_file_line_blockers="$(jq -r '[.meta.indicators[]? | select(.metric_id == "gooo.metric.source.go-file-lines.v1" and .blocking == true and .satisfied == false and .applicability != "NOT_APPLICABLE")] | length' "$metrics")"
  function_line_blockers="$(jq -r '[.meta.indicators[]? | select(.metric_id == "gooo.metric.source.function-lines.v1" and .blocking == true and .satisfied == false and .applicability != "NOT_APPLICABLE")] | length' "$metrics")"
  blocking=$((go_file_line_blockers + function_line_blockers))
  last_metrics="$metrics"
  if [[ "$blocking" -eq 0 ]]; then
    fixed_point_closed=true
    break
  fi
  if [[ "$previous_blocking" -ge 0 && "$blocking" -ge "$previous_blocking" ]]; then
    write_extraction_result "$last_extraction" "$projection_work/fixed-point-unknown.json" "NO_PROGRESS_FIXED_POINT" "$last_metrics"
    jq --arg blocking "$blocking" --arg iteration "$iteration" \
      '. + {diagnostics:{blocking_residual:($blocking|tonumber), iteration:($iteration|tonumber)}}' \
      "$projection_work/fixed-point-unknown.json" \
      > "$projection_work/fixed-point-unknown.sealed.json"
    mv "$projection_work/fixed-point-unknown.sealed.json" "$projection_work/fixed-point-unknown.json"
    cat "$projection_work/fixed-point-unknown.json"
    exit 1
  fi
  previous_blocking="$blocking"
  go run ./bootstrap/meta-binding-witness \
    -root "$repo_root" \
    -metrics "$metrics" \
    -bound-metrics "$pass_work/bound-metrics.json" \
    -report "$pass_work/binding-report.json" \
    -check
  go run ./bootstrap/logical-split-planner \
    --root "$repo_root" \
    --metrics "$metrics" \
    --expected-sha "$head_sha" \
    --output "$plan"
  go run ./bootstrap/line-density-rewriter \
    --root "$repo_root" \
    --plan "$plan" \
    --expected-sha "$head_sha" \
    --output "$density"
  set +e
  run_extraction_pass "$extraction" true "$plan" "$density"
  extraction_status=$?
  set -e
  if [[ "$extraction_status" -ne 0 ]]; then
    exit "$extraction_status"
  fi
  last_extraction="$extraction"
  if [[ "$EXTRACTION_PARTIAL" -eq 1 ]]; then
    continue
  fi
  go run ./scripts/source-splitter \
    -root "$repo_root" \
    -metrics "$metrics" \
    -sha "$head_sha" \
    -plan "$plan" \
    -output "$split"
done

if [[ "$fixed_point_closed" != true ]]; then
  write_extraction_result "$last_extraction" "$projection_work/fixed-point-unknown.json" "FIXED_POINT_ITERATION_BOUND" "$last_metrics"
  jq --arg iteration "8" \
    '. + {diagnostics:{iteration_limit:($iteration|tonumber)}}' \
    "$projection_work/fixed-point-unknown.json" \
    > "$projection_work/fixed-point-unknown.sealed.json"
  mv "$projection_work/fixed-point-unknown.sealed.json" "$projection_work/fixed-point-unknown.json"
  cat "$projection_work/fixed-point-unknown.json"
  exit 1
fi

# Stage 0 keeps the deterministic Go verifier as the baseline while gooo is
# bootstrapped. The policy job adds the full changed-path scope check.
go run ./scripts/verify --root "$repo_root" --storage-root "$storage_root"

if [[ ! -f cmd/gooo/main.go ]]; then
  echo "semantic conformance deferred: cmd/gooo is not present in this baseline"
  exit 0
fi

if ! grep -Fq 'case "check":' cmd/gooo/main.go; then
  echo "semantic conformance deferred: cmd/gooo check is not implemented in this baseline"
  exit 0
fi

go run ./cmd/gooo check examples/billing/main.gooo
go test -tags semantic_conformance ./internal/verify -run 'TestSemantic|TestGenerated' -count=1
./scripts/generated-freshness.sh

repro_work="${RUNNER_TEMP:-${TMPDIR:-/tmp}}/gooo-reproducibility-semantics"
mkdir -p "$repro_work"
head_sha="$(git rev-parse HEAD)"
go run ./cmd/gooo check examples/reproducibility-semantics/main.gooo
go run ./scripts/reproducibility-semantics \
  -mode produce \
  -source examples/reproducibility-semantics/main.gooo \
  -head-sha "$head_sha" \
  -output "$repro_work/receipt.json"
go run ./scripts/reproducibility-semantics \
  -mode produce \
  -source examples/reproducibility-semantics/main.gooo \
  -head-sha "$head_sha" \
  -output "$repro_work/receipt-replay.json"
cmp -s "$repro_work/receipt.json" "$repro_work/receipt-replay.json"
go run ./scripts/reproducibility-semantics \
  -mode judge \
  -source examples/reproducibility-semantics/main.gooo \
  -head-sha "$head_sha" \
  -receipt "$repro_work/receipt.json" \
  -output "$repro_work/judgment.json" \
  -check
go run ./scripts/reproducibility-semantics \
  -mode judge \
  -source examples/reproducibility-semantics/main.gooo \
  -head-sha "$head_sha" \
  -receipt "$repro_work/receipt.json" \
  -output "$repro_work/judgment-replay.json" \
  -check
cmp -s "$repro_work/judgment.json" "$repro_work/judgment-replay.json"
jq -e '.decision == "DISCHARGED" and .conformance_decision == "DISCHARGED" and .conformance_resolution == "EXACT" and .subject_decision == "OPEN" and .subject_resolution == "LOWER_RESOLUTION" and .summary.case_matrix.numerator == 4 and .summary.case_matrix.denominator == 4 and .summary.byte_claim.numerator == 2 and .summary.meaning_claim.numerator == 2 and .summary.joint_claim.numerator == 1 and .summary.counterexamples.numerator == 2 and .summary.open_cases.numerator == 1 and .summary.source_digest_binding.numerator == 4 and .summary.semantic_causality.numerator == 4' "$repro_work/judgment.json"

consumer_sources=(internal/meta/reproducibilitysemanticsconsumer/*.go)
if rg -n 'Produce\(|internal/meta/reproducibilitysemantics("|/)' "${consumer_sources[@]}" >/dev/null; then
  echo "production consumer producer-dependency guardrail failed" >&2
  exit 1
fi
echo "production consumer producer imports: 0/0"

sed 's/meaning.observed=meaning\/render-approved\/v1/meaning.observed=meaning\/charge-and-ledger\/v1/' \
  examples/reproducibility-semantics/main.gooo > "$repro_work/semantic-intervention.gooo"
sed '1i // presentation-only comment' examples/reproducibility-semantics/main.gooo > "$repro_work/presentation-intervention.gooo"
go run ./scripts/reproducibility-semantics \
  -mode intervention \
  -source examples/reproducibility-semantics/main.gooo \
  -semantic-source "$repro_work/semantic-intervention.gooo" \
  -presentation-source "$repro_work/presentation-intervention.gooo" \
  -head-sha "$head_sha" \
  -output "$repro_work/intervention.json"
jq -e '(.schema == "gooo/reproducibility-semantics-intervention/v1") and (.denominator == 2) and ((.cases | length) == 2) and (.decision == "DISCHARGED") and (.resolution == "EXACT") and (.authority.repository_writes == 0) and (.authority.mutation_authorized == false) and (.authority.promotion_authorized == false) and (.cases[0].id == "semantic-source-change") and (.cases[0].kind == "SEMANTIC_SOURCE_CHANGE") and (.cases[0].source_digest_before != .cases[0].source_digest_after) and (.cases[0].semantic_digest_before != .cases[0].semantic_digest_after) and (.cases[0].meaning_before != .cases[0].meaning_after) and (.cases[0].joint_before != .cases[0].joint_after) and (.cases[0].transitions_before != .cases[0].transitions_after) and (.cases[1].id == "presentation-only-source-change") and (.cases[1].kind == "PRESENTATION_ONLY_SOURCE_CHANGE") and (.cases[1].source_digest_before != .cases[1].source_digest_after) and (.cases[1].semantic_digest_before == .cases[1].semantic_digest_after) and (.cases[1].meaning_before == .cases[1].meaning_after) and (.cases[1].joint_before == .cases[1].joint_after) and (.cases[1].transitions_before == .cases[1].transitions_after) and (has("score") | not) and (has("aggregate_score") | not)' "$repro_work/intervention.json"
echo "semantic causality intervention artifact: fixed-denominator=2/2; semantic and presentation cases separate"
