#!/usr/bin/env bash
set -euo pipefail

: "${GH_REPO:?GH_REPO is required}"
: "${GH_TOKEN:?GH_TOKEN is required}"
: "${CACHE_INVENTORY_OUTPUT_DIR:?CACHE_INVENTORY_OUTPUT_DIR is required}"

output_dir="$CACHE_INVENTORY_OUTPUT_DIR"
mkdir -p "$output_dir"
pages_path="$output_dir/api-pages.json"
inventory_path="$output_dir/cache-inventory.json"
text_path="$output_dir/cache-inventory.txt"
tsv_path="$output_dir/cache-inventory.tsv"
receipt_path="$output_dir/receipt.json"

gh api --paginate --slurp \
  -H 'Accept: application/vnd.github+json' \
  "repos/$GH_REPO/actions/caches?per_page=100" >"$pages_path"

jq '[.[] | .actions_caches[]? | {id, ref, key, size_in_bytes, created_at, last_accessed_at}] | sort_by([(.ref // ""), (.key // ""), (.id | tostring)])' \
  "$pages_path" >"$inventory_path"

jq -r '.[] | [.id, (.ref // "unknown"), (.key // "unknown"), (.size_in_bytes // 0), (.created_at // "unknown"), (.last_accessed_at // "unknown")] | @tsv' \
  "$inventory_path" >"$tsv_path"

cache_count="$(jq 'length' "$inventory_path")"
{
  printf 'schema=ci-tooling/cache-inventory/v1\n'
  printf 'repository=%s\n' "$GH_REPO"
  printf 'pagination=full\n'
  printf 'cache_count=%s\n' "$cache_count"
  printf 'plan=read_only\n'
  printf 'planned_deletions=0\n'
  printf 'protected_dev=UNKNOWN\n'
  printf 'protected_main=UNKNOWN\n'
  printf 'default_branch=UNKNOWN\n'
  printf 'open_pull_requests=UNKNOWN\n'
  printf 'lookup_status=UNKNOWN\n'
  jq -r '.[] | "cache id=\(.id) ref=\(.ref // "unknown") key=\(.key // "unknown") bytes=\(.size_in_bytes // 0) created_at=\(.created_at // "unknown") last_accessed_at=\(.last_accessed_at // "unknown")"' "$inventory_path"
} | tee "$text_path"

jq -n \
  --arg schema 'ci-tooling/cache-inventory/v1' \
  --arg repository "$GH_REPO" \
  --argjson caches "$(cat "$inventory_path")" \
  '{schema: $schema, repository: $repository, pagination: "full", plan: {read_only: true, planned_deletions: 0}, protection: {dev: "UNKNOWN", main: "UNKNOWN", default_branch: "UNKNOWN", open_pull_requests: "UNKNOWN", lookup_status: "UNKNOWN"}, caches: $caches}' \
  >"$receipt_path"
