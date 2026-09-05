#!/usr/bin/env bash
set -euo pipefail

: "${GOVULNCHECK_OUTPUT_DIR:?GOVULNCHECK_OUTPUT_DIR is required}"
: "${OBSERVE_SOURCE_SHA:?OBSERVE_SOURCE_SHA is required}"
: "${OBSERVE_RUN_ID:?OBSERVE_RUN_ID is required}"
: "${OBSERVE_RUN_ATTEMPT:?OBSERVE_RUN_ATTEMPT is required}"

output_dir="$GOVULNCHECK_OUTPUT_DIR"
mkdir -p "$output_dir"
json_path="$output_dir/govulncheck.json"
text_path="$output_dir/govulncheck.txt"
sarif_path="$output_dir/govulncheck.sarif"
json_stderr_path="$output_dir/govulncheck-stderr.log"
sarif_stderr_path="$output_dir/govulncheck-sarif-stderr.log"
receipt_path="$output_dir/receipt.json"

install_start_ms="$(date +%s%3N)"
set +e
timeout 120s go install golang.org/x/vuln/cmd/govulncheck@v1.7.0
install_exit=$?
set -e
install_end_ms="$(date +%s%3N)"
govulncheck="$(go env GOPATH)/bin/govulncheck"

if [ "$install_exit" -ne 0 ]; then
  observed_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  jq -n \
    --arg schema 'ci-tooling/govulncheck/v1' \
    --arg source_sha "$OBSERVE_SOURCE_SHA" \
    --arg run_id "$OBSERVE_RUN_ID" \
    --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
    --arg tool 'golang.org/x/vuln/cmd/govulncheck@v1.7.0' \
    --arg tool_commit '617f44b718537dccdea1915395650e0529e3b72e' \
    --arg observed_at "$observed_at" \
    --arg scope "${OBSERVE_SCOPE:-./...}" \
    --argjson wall_ms "$((install_end_ms - install_start_ms))" \
    --argjson install_exit "$install_exit" \
    --argjson original_exit "$install_exit" \
    --arg verdict INSTALL_FAILED \
    '{schema: $schema, source_sha: $source_sha, run_id: $run_id, run_attempt: $run_attempt, tool: $tool, tool_commit: $tool_commit, observed_at: $observed_at, wall_ms: $wall_ms, install_exit: $install_exit, original_exit: $original_exit, verdict: $verdict, scope: $scope, govuln_db: "UNKNOWN"}' \
    > "$receipt_path"
  printf 'verdict=INSTALL_FAILED\ninstall_exit=%s\n' "$install_exit" | tee "$text_path"
  exit 1
fi

timeout 30s go version -m "$govulncheck" | tee "$output_dir/toolchain.txt"

scan_start_ms="$(date +%s%3N)"
set +e
timeout 480s "$govulncheck" -format=json ./... >"$json_path" 2>"$json_stderr_path"
json_exit=$?
set -e
scan_end_ms="$(date +%s%3N)"

govuln_db='UNKNOWN'
if db_candidate="$(jq -s -r 'map(select(.config? != null) | .config.db // empty) | first // empty' "$json_path" 2>/dev/null)" && [ -n "$db_candidate" ]; then
  govuln_db="$db_candidate"
fi

text_start_ms="$(date +%s%3N)"
set +e
timeout 120s "$govulncheck" -mode=convert -format=text <"$json_path" >"$text_path" 2>"$output_dir/govulncheck-text-stderr.log"
text_exit=$?
set -e
text_end_ms="$(date +%s%3N)"

sarif_start_ms="$(date +%s%3N)"
set +e
timeout 120s "$govulncheck" -mode=convert -format=sarif <"$json_path" >"$sarif_path" 2>"$sarif_stderr_path"
sarif_exit=$?
set -e
sarif_end_ms="$(date +%s%3N)"

scan_wall_ms=$((scan_end_ms - scan_start_ms))
text_wall_ms=$((text_end_ms - text_start_ms))
sarif_wall_ms=$((sarif_end_ms - sarif_start_ms))
total_wall_ms=$((sarif_end_ms - scan_start_ms))

if [ "$sarif_exit" -ne 0 ] || [ ! -s "$sarif_path" ]; then
  rm -f "$sarif_path"
fi

if [ "$json_exit" -ne 0 ]; then
  verdict=FAIL_SCAN
elif [ ! -s "$json_path" ]; then
  verdict=FAIL_SCAN_OUTPUT_EMPTY
elif [ "$text_exit" -eq 3 ]; then
  verdict=FAIL_FINDINGS
elif [ "$text_exit" -ne 0 ]; then
  verdict=FAIL_TEXT_CONVERSION
elif [ "$sarif_exit" -ne 0 ] || [ ! -s "$output_dir/govulncheck.sarif" ]; then
  verdict=FAIL_SARIF_CONVERSION
else
  verdict=PASS
fi

observed_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
jq -n \
  --arg schema 'ci-tooling/govulncheck/v1' \
  --arg source_sha "$OBSERVE_SOURCE_SHA" \
  --arg run_id "$OBSERVE_RUN_ID" \
  --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
  --arg tool 'golang.org/x/vuln/cmd/govulncheck@v1.7.0' \
  --arg tool_commit '617f44b718537dccdea1915395650e0529e3b72e' \
  --arg observed_at "$observed_at" \
  --arg scope "${OBSERVE_SCOPE:-./...}" \
  --arg govuln_db "$govuln_db" \
  --arg verdict "$verdict" \
  --argjson wall_ms "$total_wall_ms" \
  --argjson install_exit "$install_exit" \
  --argjson original_exit "$json_exit" \
  --argjson text_exit "$text_exit" \
  --argjson sarif_exit "$sarif_exit" \
  --argjson scan_wall_ms "$scan_wall_ms" \
  --argjson text_wall_ms "$text_wall_ms" \
  --argjson sarif_wall_ms "$sarif_wall_ms" \
  '{schema: $schema, source_sha: $source_sha, run_id: $run_id, run_attempt: $run_attempt, tool: $tool, tool_commit: $tool_commit, observed_at: $observed_at, wall_ms: $wall_ms, install_exit: $install_exit, original_exit: $original_exit, text_conversion_exit: $text_exit, sarif_conversion_exit: $sarif_exit, verdict: $verdict, scope: $scope, govuln_db: $govuln_db, json_scan: {wall_ms: $scan_wall_ms, original_exit: $original_exit}, text_conversion: {wall_ms: $text_wall_ms, original_exit: $text_exit}, sarif_conversion: {wall_ms: $sarif_wall_ms, original_exit: $sarif_exit}}' \
  > "$receipt_path"

{
  printf 'schema=ci-tooling/govulncheck/v1\n'
  printf 'tool=govulncheck\n'
  printf 'tool_version=v1.7.0\n'
  printf 'tool_commit=617f44b718537dccdea1915395650e0529e3b72e\n'
  printf 'govuln_db=%s\n' "$govuln_db"
  printf 'scan_exit=%s\n' "$json_exit"
  printf 'text_conversion_exit=%s\n' "$text_exit"
  printf 'sarif_conversion_exit=%s\n' "$sarif_exit"
  printf 'verdict=%s\n' "$verdict"
} | tee -a "$text_path"

if [ "$verdict" != PASS ]; then
  exit 1
fi
