#!/usr/bin/env bash
set -euo pipefail

: "${GOVULNCHECK_OUTPUT_DIR:?GOVULNCHECK_OUTPUT_DIR is required}"

output_dir="$GOVULNCHECK_OUTPUT_DIR"
mkdir -p "$output_dir"
json_path="$output_dir/govulncheck.json"
text_path="$output_dir/govulncheck.txt"
sarif_path="$output_dir/govulncheck.sarif"
json_stderr_path="$output_dir/govulncheck-stderr.log"
sarif_stderr_path="$output_dir/govulncheck-sarif-stderr.log"

go install golang.org/x/vuln/cmd/govulncheck@v1.7.0
govulncheck="$(go env GOPATH)/bin/govulncheck"
go version -m "$govulncheck" | tee "$output_dir/toolchain.txt"

set +e
"$govulncheck" -format=json ./... >"$json_path" 2>"$json_stderr_path"
json_exit=$?
set -e

json_parse_exit=0
if ! jq -e -s 'type == "array"' "$json_path" >/dev/null; then
  json_parse_exit=$?
fi

finding_count=0
if [ "$json_parse_exit" -eq 0 ]; then
  finding_count="$(jq -s '[.[] | select(.finding? != null)] | length' "$json_path")"
fi

if [ "$json_parse_exit" -ne 0 ]; then
  verdict=FAIL_SCAN_OUTPUT
elif [ "$json_exit" -ne 0 ]; then
  verdict=FAIL_SCAN
elif [ "$finding_count" -gt 0 ]; then
  verdict=FAIL_FINDINGS
else
  verdict=PASS
fi

{
  printf 'schema=ci-tooling/govulncheck/v1\n'
  printf 'tool=govulncheck\n'
  printf 'tool_version=v1.7.0\n'
  printf 'tool_commit=617f44b718537dccdea1915395650e0529e3b72e\n'
  printf 'scan_exit=%s\n' "$json_exit"
  printf 'json_parse_exit=%s\n' "$json_parse_exit"
  printf 'finding_count=%s\n' "$finding_count"
  printf 'verdict=%s\n' "$verdict"
} | tee "$text_path"

set +e
"$govulncheck" -format=sarif ./... >"$sarif_path" 2>"$sarif_stderr_path"
sarif_exit=$?
set -e
printf 'sarif_exit=%s\n' "$sarif_exit" | tee -a "$text_path"

if [ "$json_parse_exit" -ne 0 ] || [ "$json_exit" -ne 0 ] || [ "$finding_count" -gt 0 ] || [ "$sarif_exit" -ne 0 ]; then
  exit 1
fi
