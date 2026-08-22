#!/usr/bin/env bash
set -euo pipefail

source_dir="${1:?source metrics directory required}"
output_dir="$source_dir/operation-artifact"
before="$(git status --porcelain=v1 --untracked-files=all)"
for pass in 1 2; do
  go run ./cmd/operation-artifact-witness \
    --root "$GITHUB_WORKSPACE" \
    --actionability "$source_dir/meta-actionability-report.json" \
    --observations "$output_dir/observations.json" \
    --report "$output_dir/report-$pass.json" \
    --check
done

cmp "$output_dir/report-1.json" "$output_dir/report-2.json"
jq -e '.decision == "FIXED_POINT"' "$output_dir/report-1.json"
after="$(git status --porcelain=v1 --untracked-files=all)"
if [[ "$before" != "$after" ]]; then
  printf 'operation artifact witness changed the repository\n' >&2
  diff <(printf '%s\n' "$before") <(printf '%s\n' "$after") || true
  exit 1
fi
