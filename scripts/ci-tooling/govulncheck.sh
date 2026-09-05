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
expected_head="$OBSERVE_SOURCE_SHA"
actual_head="$(git rev-parse HEAD 2>/dev/null || true)"
source_binding_exit=0
if [ "$actual_head" != "$expected_head" ]; then
  source_binding_exit=1
fi

install_start_ms="$(date +%s%3N)"
set +e
timeout 120s go install golang.org/x/vuln/cmd/govulncheck@v1.7.0
install_exit=$?
set -e
install_end_ms="$(date +%s%3N)"
install_wall_ms=$((install_end_ms - install_start_ms))
govulncheck="$(go env GOPATH)/bin/govulncheck"

source_binding_reason='accepted'
if [ "$source_binding_exit" -ne 0 ]; then
  source_binding_reason='source_sha_mismatch'
fi

if [ "$install_exit" -ne 0 ]; then
  observed_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  install_verdict='INSTALL_FAILED'
  if [ "$source_binding_exit" -ne 0 ]; then
    install_verdict='SOURCE_BINDING_FAILED'
  fi
  jq -n \
    --arg schema 'ci-tooling/govulncheck/v1' \
    --arg source_sha "$actual_head" \
    --arg expected_source_sha "$expected_head" \
    --arg run_id "$OBSERVE_RUN_ID" \
    --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
    --arg tool 'golang.org/x/vuln/cmd/govulncheck@v1.7.0' \
    --arg tool_commit '617f44b718537dccdea1915395650e0529e3b72e' \
    --arg observed_at "$observed_at" \
    --arg scope "${OBSERVE_SCOPE:-./...}" \
    --arg source_binding_reason "$source_binding_reason" \
    --arg wall_scope 'installation start through installation end; checkout and receipt upload measured separately' \
    --argjson wall_ms "$install_wall_ms" \
    --argjson source_binding_exit "$source_binding_exit" \
    --argjson install_exit "$install_exit" \
    --argjson original_exit "$install_exit" \
    --arg verdict "$install_verdict" \
    '{schema: $schema, source_sha: $source_sha, expected_source_sha: $expected_source_sha, source_binding_exit: $source_binding_exit, source_binding_reason: $source_binding_reason, run_id: $run_id, run_attempt: $run_attempt, tool: $tool, tool_commit: $tool_commit, observed_at: $observed_at, wall_ms: $wall_ms, wall_scope: $wall_scope, install_exit: $install_exit, original_exit: $original_exit, verdict: $verdict, scope: $scope, govuln_db: "UNKNOWN"}' \
    > "$receipt_path"
  printf 'verdict=%s\ninstall_exit=%s\nsource_binding_exit=%s\n' "$install_verdict" "$install_exit" "$source_binding_exit" | tee "$text_path"
  exit 1
fi

tool_start_ms="$(date +%s%3N)"
set +e
timeout 30s go version -m "$govulncheck" >"$output_dir/toolchain.txt" 2>"$output_dir/toolchain-stderr.log"
tool_version_exit=$?
set -e
tool_end_ms="$(date +%s%3N)"
tool_metadata_wall_ms=$((tool_end_ms - tool_start_ms))

scan_start_ms="$(date +%s%3N)"
set +e
timeout 480s "$govulncheck" -format=json ./... >"$json_path" 2>"$json_stderr_path"
json_exit=$?
set -e
scan_end_ms="$(date +%s%3N)"

json_schema_filter='length > 0 and all(.[]; type == "object") and ([.[] | select(has("config"))] | length == 1) and ([.[] | select(has("config")) | .config] | all(type == "object" and (.protocol_version | type == "string") and (.protocol_version | length > 0)))'
set +e
jq -s -e "$json_schema_filter" "$json_path" >/dev/null 2>"$output_dir/json-validation-stderr.log"
raw_parse_exit=$?
set -e
raw_rejection_reason='accepted'
if [ ! -s "$json_path" ]; then
  raw_rejection_reason='empty_scan_output'
elif [ "$raw_parse_exit" -ne 0 ]; then
  raw_rejection_reason='invalid_ndjson_or_missing_single_config'
fi

fixture_path="$output_dir/fixture.json"
fixture_log="$output_dir/fixture-validation.log"
fixture_failures=0
: > "$fixture_log"
validate_rejected_fixture() {
  local fixture_name="$1"
  local fixture_content="$2"
  printf '%s' "$fixture_content" > "$fixture_path"
  if jq -s -e "$json_schema_filter" "$fixture_path" >/dev/null 2>>"$fixture_log"; then
    printf '%s=UNEXPECTED_ACCEPT\n' "$fixture_name" >>"$fixture_log"
    fixture_failures=$((fixture_failures + 1))
  else
    printf '%s=REJECTED\n' "$fixture_name" >>"$fixture_log"
  fi
}
validate_rejected_fixture invalid_json '{'
validate_rejected_fixture empty ''
validate_rejected_fixture whitespace '   '
validate_rejected_fixture empty_object '{}'
rm -f "$fixture_path"

govuln_db='UNKNOWN'
if [ "$raw_parse_exit" -eq 0 ] && db_candidate="$(jq -s -r 'map(select(.config? != null) | .config.db // empty) | first // empty' "$json_path" 2>/dev/null)" && [ -n "$db_candidate" ]; then
  govuln_db="$db_candidate"
fi

text_start_ms="$(date +%s%3N)"
set +e
timeout 30s "$govulncheck" -mode=convert -format=text <"$json_path" >"$text_path" 2>"$output_dir/govulncheck-text-stderr.log"
text_exit=$?
set -e
text_end_ms="$(date +%s%3N)"

sarif_start_ms="$(date +%s%3N)"
set +e
timeout 30s "$govulncheck" -mode=convert -format=sarif <"$json_path" >"$sarif_path" 2>"$sarif_stderr_path"
sarif_exit=$?
set -e
sarif_end_ms="$(date +%s%3N)"

scan_wall_ms=$((scan_end_ms - scan_start_ms))
text_wall_ms=$((text_end_ms - text_start_ms))
sarif_wall_ms=$((sarif_end_ms - sarif_start_ms))
total_wall_ms=$((sarif_end_ms - scan_start_ms))
wall_scope='scan start through SARIF conversion end; installation, tool metadata, checkout, and receipt upload measured separately'

if [ "$sarif_exit" -ne 0 ] || [ ! -s "$sarif_path" ]; then
  rm -f "$sarif_path"
fi

if [ "$source_binding_exit" -ne 0 ]; then
  verdict=FAIL_SOURCE_BINDING
elif [ "$install_exit" -ne 0 ]; then
  verdict=FAIL_INSTALL
elif [ "$tool_version_exit" -ne 0 ]; then
  verdict=FAIL_TOOL_METADATA
elif [ "$json_exit" -ne 0 ]; then
  verdict=FAIL_SCAN
elif [ "$raw_parse_exit" -ne 0 ]; then
  verdict=FAIL_SCAN_OUTPUT
elif [ "$fixture_failures" -ne 0 ]; then
  verdict=FAIL_FIXTURE_CONTRACT
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
  --arg source_sha "$actual_head" \
  --arg expected_source_sha "$expected_head" \
  --arg source_binding_reason "$source_binding_reason" \
  --arg run_id "$OBSERVE_RUN_ID" \
  --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
  --arg tool 'golang.org/x/vuln/cmd/govulncheck@v1.7.0' \
  --arg tool_commit '617f44b718537dccdea1915395650e0529e3b72e' \
  --arg observed_at "$observed_at" \
  --arg scope "${OBSERVE_SCOPE:-./...}" \
  --arg govuln_db "$govuln_db" \
  --arg raw_rejection_reason "$raw_rejection_reason" \
  --arg wall_scope "$wall_scope" \
  --arg verdict "$verdict" \
  --argjson wall_ms "$total_wall_ms" \
  --argjson install_exit "$install_exit" \
  --argjson original_exit "$json_exit" \
  --argjson text_exit "$text_exit" \
  --argjson sarif_exit "$sarif_exit" \
  --argjson source_binding_exit "$source_binding_exit" \
  --argjson tool_version_exit "$tool_version_exit" \
  --argjson raw_parse_exit "$raw_parse_exit" \
  --argjson fixture_failures "$fixture_failures" \
  --argjson install_wall_ms "$install_wall_ms" \
  --argjson tool_metadata_wall_ms "$tool_metadata_wall_ms" \
  --argjson scan_wall_ms "$scan_wall_ms" \
  --argjson text_wall_ms "$text_wall_ms" \
  --argjson sarif_wall_ms "$sarif_wall_ms" \
  '{schema: $schema, source_sha: $source_sha, expected_source_sha: $expected_source_sha, source_binding_exit: $source_binding_exit, source_binding_reason: $source_binding_reason, run_id: $run_id, run_attempt: $run_attempt, tool: $tool, tool_commit: $tool_commit, observed_at: $observed_at, wall_ms: $wall_ms, wall_scope: $wall_scope, install_wall_ms: $install_wall_ms, tool_metadata_wall_ms: $tool_metadata_wall_ms, install_exit: $install_exit, tool_version_exit: $tool_version_exit, original_exit: $original_exit, raw_parse_exit: $raw_parse_exit, raw_rejection_reason: $raw_rejection_reason, fixture_failures: $fixture_failures, text_conversion_exit: $text_exit, sarif_conversion_exit: $sarif_exit, verdict: $verdict, scope: $scope, govuln_db: $govuln_db, json_scan: {wall_ms: $scan_wall_ms, original_exit: $original_exit, raw_parse_exit: $raw_parse_exit, rejection_reason: $raw_rejection_reason}, text_conversion: {wall_ms: $text_wall_ms, original_exit: $text_exit}, sarif_conversion: {wall_ms: $sarif_wall_ms, original_exit: $sarif_exit}}' \
  > "$receipt_path"

{
  printf 'schema=ci-tooling/govulncheck/v1\n'
  printf 'tool=govulncheck\n'
  printf 'tool_version=v1.7.0\n'
  printf 'tool_commit=617f44b718537dccdea1915395650e0529e3b72e\n'
  printf 'govuln_db=%s\n' "$govuln_db"
  printf 'scan_exit=%s\n' "$json_exit"
  printf 'raw_parse_exit=%s\n' "$raw_parse_exit"
  printf 'raw_rejection_reason=%s\n' "$raw_rejection_reason"
  printf 'text_conversion_exit=%s\n' "$text_exit"
  printf 'sarif_conversion_exit=%s\n' "$sarif_exit"
  printf 'fixture_failures=%s\n' "$fixture_failures"
  printf 'source_binding_exit=%s\n' "$source_binding_exit"
  printf 'verdict=%s\n' "$verdict"
} | tee -a "$text_path"

if [ "$verdict" != PASS ]; then
  exit 1
fi
