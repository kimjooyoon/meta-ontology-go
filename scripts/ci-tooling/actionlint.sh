#!/usr/bin/env bash
set -euo pipefail

: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
: "${OBSERVE_SOURCE_SHA:?OBSERVE_SOURCE_SHA is required}"
: "${OBSERVE_RUN_ID:?OBSERVE_RUN_ID is required}"
: "${OBSERVE_RUN_ATTEMPT:?OBSERVE_RUN_ATTEMPT is required}"

output_dir="${ACTIONLINT_OUTPUT_DIR:-$RUNNER_TEMP/workflow-lint}"
mkdir -p "$output_dir"

log_path="$output_dir/actionlint.log"
receipt_path="$output_dir/receipt.json"
expected_head="$OBSERVE_SOURCE_SHA"
actual_head="$(git rev-parse HEAD 2>/dev/null || true)"
source_binding_exit=0
if [ "$actual_head" != "$expected_head" ]; then
  source_binding_exit=1
fi
install_start_ms="$(date +%s%3N)"
set +e
timeout 120s go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
install_exit=$?
set -e
install_end_ms="$(date +%s%3N)"
install_wall_ms=$((install_end_ms - install_start_ms))

actionlint="$(go env GOPATH)/bin/actionlint"

if [ "$install_exit" -eq 0 ]; then
  "$actionlint" -version | tee "$output_dir/actionlint-version.txt"
  run_start_ms="$(date +%s%3N)"
  set +e
  timeout 180s "$actionlint" -shellcheck= -pyflakes= 2>&1 | tee "$log_path"
  pipeline_status=("${PIPESTATUS[@]}")
  lint_exit="${pipeline_status[0]}"
  lint_pipeline_exit="${pipeline_status[1]}"
  set -e
  run_end_ms="$(date +%s%3N)"
else
  lint_pipeline_exit=1
  lint_exit=125
  run_start_ms="$install_end_ms"
  run_end_ms="$install_end_ms"
  : > "$log_path"
fi

wall_ms=$((run_end_ms - run_start_ms))
observed_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
jq -n \
  --arg schema 'ci-tooling/workflow-lint/v1' \
  --arg source_sha "$actual_head" \
  --arg expected_source_sha "$expected_head" \
  --arg run_id "$OBSERVE_RUN_ID" \
  --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
  --arg tool 'github.com/rhysd/actionlint/cmd/actionlint@v1.7.12' \
  --arg observed_at "$observed_at" \
  --arg scope "${OBSERVE_SCOPE:-.github/workflows/**,scripts/ci-tooling/actionlint.sh}" \
  --arg wall_scope 'actionlint command only; checkout and installation measured separately' \
  --argjson wall_ms "$wall_ms" \
  --argjson install_wall_ms "$install_wall_ms" \
  --argjson install_exit "$install_exit" \
  --argjson original_exit "$lint_exit" \
  --argjson capture_exit "$lint_pipeline_exit" \
  --argjson source_binding_exit "$source_binding_exit" \
  '{schema: $schema, source_sha: $source_sha, expected_source_sha: $expected_source_sha, source_binding_exit: $source_binding_exit, run_id: $run_id, run_attempt: $run_attempt, tool: $tool, observed_at: $observed_at, wall_ms: $wall_ms, wall_scope: $wall_scope, install_wall_ms: $install_wall_ms, install_exit: $install_exit, original_exit: $original_exit, capture_exit: $capture_exit, scope: $scope}' \
  > "$receipt_path"

if [ "$source_binding_exit" -ne 0 ] || [ "$install_exit" -ne 0 ] || [ "$lint_pipeline_exit" -ne 0 ] || [ "$lint_exit" -ne 0 ]; then
  exit 1
fi
