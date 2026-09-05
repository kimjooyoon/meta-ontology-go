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
install_start_ms="$(date +%s%3N)"
set +e
timeout 120s go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
install_exit=$?
set -e
install_end_ms="$(date +%s%3N)"

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
  --arg source_sha "$OBSERVE_SOURCE_SHA" \
  --arg run_id "$OBSERVE_RUN_ID" \
  --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
  --arg tool 'github.com/rhysd/actionlint/cmd/actionlint@v1.7.12' \
  --arg observed_at "$observed_at" \
  --arg scope "${OBSERVE_SCOPE:-.github/workflows/**,scripts/ci-tooling/actionlint.sh}" \
  --argjson wall_ms "$wall_ms" \
  --argjson install_exit "$install_exit" \
  --argjson original_exit "$lint_exit" \
  --argjson capture_exit "$lint_pipeline_exit" \
  '{schema: $schema, source_sha: $source_sha, run_id: $run_id, run_attempt: $run_attempt, tool: $tool, observed_at: $observed_at, wall_ms: $wall_ms, install_exit: $install_exit, original_exit: $original_exit, capture_exit: $capture_exit, scope: $scope}' \
  > "$receipt_path"

if [ "$install_exit" -ne 0 ] || [ "$lint_pipeline_exit" -ne 0 ] || [ "$lint_exit" -ne 0 ]; then
  exit 1
fi
