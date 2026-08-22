#!/usr/bin/env bash
set -euo pipefail

before="$(git status --porcelain=v1 --untracked-files=all)"
for pass in 1 2; do
  go run ./cmd/operation-artifact-witness \
    --root "$GITHUB_WORKSPACE" \
    --actionability "$OUTPUT_DIR/actionability/meta-actionability-report.json" \
    --observations "$OUTPUT_DIR/observations.json" \
    --report "$OUTPUT_DIR/report-$pass.json" \
    --check
done

cmp "$OUTPUT_DIR/report-1.json" "$OUTPUT_DIR/report-2.json"
jq -e '.decision == "FIXED_POINT"' "$OUTPUT_DIR/report-1.json"
after="$(git status --porcelain=v1 --untracked-files=all)"
if [[ "$before" != "$after" ]]; then
  printf 'operation artifact witness changed the repository\n' >&2
  diff <(printf '%s\n' "$before") <(printf '%s\n' "$after") || true
  exit 1
fi

