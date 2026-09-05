#!/usr/bin/env bash
set -euo pipefail

: "${RUNNER_TEMP:?RUNNER_TEMP is required}"

output_dir="${ACTIONLINT_OUTPUT_DIR:-$RUNNER_TEMP/workflow-lint}"
mkdir -p "$output_dir"

go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
actionlint="$(go env GOPATH)/bin/actionlint"

"$actionlint" -version | tee "$output_dir/actionlint-version.txt"
timeout 240s "$actionlint" -shellcheck= -pyflakes= 2>&1 | tee "$output_dir/actionlint.log"
