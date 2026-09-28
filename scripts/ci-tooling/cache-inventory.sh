#!/usr/bin/env bash
set -euo pipefail

: "${GH_REPO:?GH_REPO is required}"
: "${GH_TOKEN:?GH_TOKEN is required}"
: "${CACHE_INVENTORY_OUTPUT_DIR:?CACHE_INVENTORY_OUTPUT_DIR is required}"
: "${OBSERVE_SOURCE_SHA:?OBSERVE_SOURCE_SHA is required}"
: "${OBSERVE_RUN_ID:?OBSERVE_RUN_ID is required}"
: "${OBSERVE_RUN_ATTEMPT:?OBSERVE_RUN_ATTEMPT is required}"

output_dir="$CACHE_INVENTORY_OUTPUT_DIR"
mkdir -p "$output_dir"
pages_path="$output_dir/api-pages.json"
inventory_path="$output_dir/cache-inventory.json"
text_path="$output_dir/cache-inventory.txt"
tsv_path="$output_dir/cache-inventory.tsv"
receipt_path="$output_dir/receipt.json"
observed_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
start_ms="$(date +%s%3N)"
expected_head="$OBSERVE_SOURCE_SHA"
actual_head="$(git rev-parse HEAD 2>/dev/null || true)"
source_binding_exit=0
if [ "$actual_head" != "$expected_head" ]; then
  source_binding_exit=1
fi
source_binding_reason='accepted'
if [ "$source_binding_exit" -ne 0 ]; then
  source_binding_reason='source_sha_mismatch'
fi
cleanup_protection='PROTECTED'
lookup_budget_ms=120000

gh_version_start_ms="$(date +%s%3N)"
set +e
timeout 30s gh version >"$output_dir/gh-version.txt" 2>"$output_dir/gh-version-stderr.log"
gh_version_exit=$?
set -e
gh_version_end_ms="$(date +%s%3N)"
gh_version_wall_ms=$((gh_version_end_ms - gh_version_start_ms))

api_start_ms="$(date +%s%3N)"
set +e
timeout 180s gh api --paginate --slurp \
  -H 'Accept: application/vnd.github+json' \
  "repos/$GH_REPO/actions/caches?per_page=100" >"$pages_path" 2>"$output_dir/pagination-stderr.log"
api_exit=$?
set -e
api_end_ms="$(date +%s%3N)"
api_wall_ms=$((api_end_ms - api_start_ms))

if [ "$api_exit" -ne 0 ]; then
  api_verdict='API_FAILED'
  if [ "$source_binding_exit" -ne 0 ]; then
    api_verdict='SOURCE_BINDING_FAILED'
  fi
  printf 'schema=ci-tooling/cache-inventory/v1\npagination=FAILED\napi_exit=%s\nsource_binding_exit=%s\nverdict=%s\nplan=read_only\nplanned_deletions=0\n' "$api_exit" "$source_binding_exit" "$api_verdict" | tee "$text_path"
  jq -n \
    --arg schema 'ci-tooling/cache-inventory/v1' \
    --arg source_sha "$actual_head" \
    --arg expected_source_sha "$expected_head" \
    --arg source_binding_reason "$source_binding_reason" \
    --arg run_id "$OBSERVE_RUN_ID" \
    --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
    --arg observed_at "$observed_at" \
    --arg scope "${OBSERVE_SCOPE:-repository Actions caches; no deletions}" \
    --arg wall_scope 'pagination start through API failure; checkout and receipt upload excluded' \
    --argjson wall_ms "$(( $(date +%s%3N) - start_ms ))" \
    --argjson source_binding_exit "$source_binding_exit" \
    --argjson original_exit "$api_exit" \
    --arg verdict "$api_verdict" \
    '{schema: $schema, source_sha: $source_sha, expected_source_sha: $expected_source_sha, source_binding_exit: $source_binding_exit, source_binding_reason: $source_binding_reason, run_id: $run_id, run_attempt: $run_attempt, observed_at: $observed_at, wall_ms: $wall_ms, wall_scope: $wall_scope, original_exit: $original_exit, scope: $scope, pagination: "FAILED", verdict: $verdict, plan: {read_only: true, planned_deletions: 0}, caches: []}' \
    > "$receipt_path"
  printf '[]\n' > "$inventory_path"
  exit 1
fi

set +e
jq -e 'length > 0 and all(.[]; type == "object" and (.actions_caches | type == "array") and (.total_count | type == "number"))' \
  "$pages_path" >/dev/null 2>"$output_dir/page-validation-stderr.log"
page_schema_exit=$?
set -e

if [ "$page_schema_exit" -ne 0 ]; then
  page_verdict='PAGE_SCHEMA_FAILED'
  if [ "$source_binding_exit" -ne 0 ]; then
    page_verdict='SOURCE_BINDING_FAILED'
  fi
  printf 'schema=ci-tooling/cache-inventory/v1\npagination=FAILED\napi_exit=0\npage_schema_exit=%s\nsource_binding_exit=%s\nverdict=%s\nplan=read_only\nplanned_deletions=0\n' "$page_schema_exit" "$source_binding_exit" "$page_verdict" | tee "$text_path"
  jq -n \
    --arg schema 'ci-tooling/cache-inventory/v1' \
    --arg source_sha "$actual_head" \
    --arg expected_source_sha "$expected_head" \
    --arg source_binding_reason "$source_binding_reason" \
    --arg run_id "$OBSERVE_RUN_ID" \
    --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
    --arg observed_at "$observed_at" \
    --arg scope "${OBSERVE_SCOPE:-repository Actions caches; no deletions}" \
    --arg wall_scope 'pagination start through page schema validation; checkout and receipt upload excluded' \
    --arg verdict "$page_verdict" \
    --argjson wall_ms "$(( $(date +%s%3N) - start_ms ))" \
    --argjson source_binding_exit "$source_binding_exit" \
    --argjson original_exit "$page_schema_exit" \
    '{schema: $schema, source_sha: $source_sha, expected_source_sha: $expected_source_sha, source_binding_exit: $source_binding_exit, source_binding_reason: $source_binding_reason, run_id: $run_id, run_attempt: $run_attempt, observed_at: $observed_at, wall_ms: $wall_ms, wall_scope: $wall_scope, original_exit: $original_exit, scope: $scope, pagination: "FAILED", page_schema_exit: $original_exit, verdict: $verdict, plan: {read_only: true, planned_deletions: 0}, caches: []}' \
    > "$receipt_path"
  printf '[]\n' > "$inventory_path"
  exit 1
fi

jq '[.[] | {total_count, page_count: (.actions_caches | length)}]' "$pages_path" >"$output_dir/api-metadata.json"

set +e
jq '[.[] | .actions_caches[] | {id, ref, key, size_in_bytes, created_at, last_accessed_at}] | sort_by([(.ref // ""), (.key // ""), (.id | tostring)])' \
  "$pages_path" >"$inventory_path"
parse_exit=$?
set -e

if [ "$parse_exit" -ne 0 ]; then
  parse_verdict='PARSE_FAILED'
  if [ "$source_binding_exit" -ne 0 ]; then
    parse_verdict='SOURCE_BINDING_FAILED'
  fi
  printf 'schema=ci-tooling/cache-inventory/v1\npagination=FAILED\napi_exit=0\nparse_exit=%s\nsource_binding_exit=%s\nverdict=%s\nplan=read_only\nplanned_deletions=0\n' "$parse_exit" "$source_binding_exit" "$parse_verdict" | tee "$text_path"
  jq -n \
    --arg schema 'ci-tooling/cache-inventory/v1' \
    --arg source_sha "$actual_head" \
    --arg expected_source_sha "$expected_head" \
    --arg source_binding_reason "$source_binding_reason" \
    --arg run_id "$OBSERVE_RUN_ID" \
    --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
    --arg observed_at "$observed_at" \
    --arg scope "${OBSERVE_SCOPE:-repository Actions caches; no deletions}" \
    --arg wall_scope 'pagination start through cache parsing failure; checkout and receipt upload excluded' \
    --argjson wall_ms "$(( $(date +%s%3N) - start_ms ))" \
    --argjson source_binding_exit "$source_binding_exit" \
    --argjson original_exit "$parse_exit" \
    --arg verdict "$parse_verdict" \
    '{schema: $schema, source_sha: $source_sha, expected_source_sha: $expected_source_sha, source_binding_exit: $source_binding_exit, source_binding_reason: $source_binding_reason, run_id: $run_id, run_attempt: $run_attempt, observed_at: $observed_at, wall_ms: $wall_ms, wall_scope: $wall_scope, original_exit: $original_exit, scope: $scope, pagination: "FAILED", verdict: $verdict, plan: {read_only: true, planned_deletions: 0}, caches: []}' \
    > "$receipt_path"
  printf '[]\n' > "$inventory_path"
  exit 1
fi

default_path="$output_dir/repository.json"
set +e
timeout 30s gh api -H 'Accept: application/vnd.github+json' "repos/$GH_REPO" >"$default_path" 2>"$output_dir/repository-stderr.log"
repository_exit=$?
set -e
default_branch='UNKNOWN'
if [ "$repository_exit" -eq 0 ]; then
  default_candidate="$(jq -r '.default_branch // empty' "$default_path" 2>/dev/null || true)"
  if [ -n "$default_candidate" ]; then
    default_branch="$default_candidate"
  fi
fi

pull_states='{}'
lookup_start_ms="$(date +%s%3N)"
lookup_deadline_ms=$((lookup_start_ms + lookup_budget_ms))
pull_lookup_status='COMPLETED'
while IFS= read -r ref; do
  if [[ "$ref" =~ ^refs/pull/([0-9]+)/merge$ ]]; then
    number="${BASH_REMATCH[1]}"
    lookup_path="$output_dir/pull-$number.json"
    state='UNKNOWN'
    now_ms="$(date +%s%3N)"
    if [ "$now_ms" -ge "$lookup_deadline_ms" ]; then
      state='UNKNOWN_BUDGET'
      pull_lookup_status='BUDGET_EXHAUSTED'
    else
      remaining_ms=$((lookup_deadline_ms - now_ms))
      lookup_timeout_seconds=$(( (remaining_ms + 999) / 1000 ))
      if [ "$lookup_timeout_seconds" -gt 30 ]; then
        lookup_timeout_seconds=30
      elif [ "$lookup_timeout_seconds" -lt 1 ]; then
        lookup_timeout_seconds=1
      fi
      set +e
      timeout "${lookup_timeout_seconds}s" gh api -H 'Accept: application/vnd.github+json' "repos/$GH_REPO/pulls/$number" >"$lookup_path" 2>"$output_dir/pull-$number-stderr.log"
      lookup_exit=$?
      set -e
      if [ "$lookup_exit" -eq 124 ]; then
        state='UNKNOWN_BUDGET'
        pull_lookup_status='BUDGET_EXHAUSTED'
      elif [ "$lookup_exit" -eq 0 ]; then
        state_candidate="$(jq -r '.state // empty' "$lookup_path" 2>/dev/null || true)"
        if [ "$state_candidate" = 'open' ] || [ "$state_candidate" = 'closed' ]; then
          state="$state_candidate"
        else
          pull_lookup_status='COMPLETED_WITH_UNKNOWN'
        fi
      else
        pull_lookup_status='COMPLETED_WITH_UNKNOWN'
      fi
    fi
    pull_states="$(jq --arg number "$number" --arg state "$state" '. + {($number): $state}' <<< "$pull_states")"
  fi
done < <(jq -r '.[].ref // empty' "$inventory_path" | sort -u)

pull_lookup_end_ms="$(date +%s%3N)"
pull_lookup_wall_ms=$((pull_lookup_end_ms - lookup_start_ms))
enriched_path="$output_dir/cache-inventory-enriched.json"
jq \
  --argjson pull_states "$pull_states" \
  --arg cleanup_protection "$cleanup_protection" \
  --arg observed_at "$observed_at" \
  'map(. + ((.ref // "") as $ref |
    if ($ref | test("^refs/pull/[0-9]+/merge$")) then
      ($ref | capture("^refs/pull/(?<number>[0-9]+)/merge$").number) as $number |
      ($pull_states[$number] // "UNKNOWN") as $state |
      {pull_request_number: ($number | tonumber), pull_request_state: $state, protection_status: (if $state == "open" or $state == "closed" then "STATE_CHECKED" elif $state == "UNKNOWN_BUDGET" then "UNKNOWN_BUDGET" else "UNKNOWN" end), eligible_candidate: ($state == "closed"), candidate_reason: (if $state == "closed" then "observed_closed_pull_request" elif $state == "open" then "open_pull_request_not_eligible" elif $state == "UNKNOWN_BUDGET" then "pull_request_lookup_budget_exhausted" else "pull_request_lookup_unknown" end), observed_at: $observed_at}
    else
      {pull_request_state: "NOT_APPLICABLE", protection_status: $cleanup_protection, eligible_candidate: false, candidate_reason: "non_pull_request_ref", observed_at: $observed_at}
    end))' "$inventory_path" > "$enriched_path"

mv "$enriched_path" "$inventory_path"

jq -r '.[] | [.id, (.ref // "unknown"), (.key // "unknown"), (.size_in_bytes // "UNKNOWN"), (.created_at // "unknown"), (.last_accessed_at // "unknown")] | @tsv' \
  "$inventory_path" >"$tsv_path"

cache_count="$(jq 'length' "$inventory_path")"
api_pages="$(jq 'length' "$output_dir/api-metadata.json")"
reported_total_count="$(jq 'if length > 0 and ([.[].total_count] | unique | length == 1) then .[0].total_count else null end' "$output_dir/api-metadata.json")"
{
  printf 'schema=ci-tooling/cache-inventory/v1\n'
  printf 'repository=%s\n' "$GH_REPO"
  printf 'pagination=full\n'
  printf 'cache_count=%s\n' "$cache_count"
  printf 'plan=read_only\n'
  printf 'planned_deletions=0\n'
  printf 'cleanup_protection_dev=%s\n' "$cleanup_protection"
  printf 'cleanup_protection_main=%s\n' "$cleanup_protection"
  printf 'default_branch=%s\n' "$default_branch"
  printf 'default_branch_cleanup_protection=%s\n' "$cleanup_protection"
  printf 'gh_version_exit=%s\n' "$gh_version_exit"
  printf 'gh_version_wall_ms=%s\n' "$gh_version_wall_ms"
  printf 'pull_lookup_budget_ms=%s\n' "$lookup_budget_ms"
  printf 'pull_lookup_wall_ms=%s\n' "$pull_lookup_wall_ms"
  printf 'pull_lookup_status=%s\n' "$pull_lookup_status"
  jq -r '.[] | "cache id=\(.id) ref=\(.ref // "UNKNOWN") key=\(.key // "UNKNOWN") bytes=\(.size_in_bytes // "UNKNOWN") created_at=\(.created_at // "UNKNOWN") last_accessed_at=\(.last_accessed_at // "UNKNOWN") candidate_reason=\(.candidate_reason) observed_at=\(.observed_at)"' "$inventory_path"
} | tee "$text_path"

cache_verdict='PASS'
if [ "$source_binding_exit" -ne 0 ]; then
  cache_verdict='SOURCE_BINDING_FAILED'
elif [ "$gh_version_exit" -ne 0 ]; then
  cache_verdict='FAIL_TOOL_METADATA'
fi

jq -n \
  --arg schema 'ci-tooling/cache-inventory/v1' \
  --arg repository "$GH_REPO" \
  --arg source_sha "$actual_head" \
  --arg expected_source_sha "$expected_head" \
  --arg source_binding_reason "$source_binding_reason" \
  --arg run_id "$OBSERVE_RUN_ID" \
  --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
  --arg observed_at "$observed_at" \
  --arg scope "${OBSERVE_SCOPE:-repository Actions caches; no deletions}" \
  --arg default_branch "$default_branch" \
  --arg cleanup_protection "$cleanup_protection" \
  --arg pull_lookup_status "$pull_lookup_status" \
  --arg verdict "$cache_verdict" \
  --arg wall_scope 'pagination start through receipt generation; checkout and receipt upload excluded' \
  --argjson caches "$(jq -c . "$inventory_path")" \
  --argjson api_pages "$api_pages" \
  --argjson reported_total_count "$reported_total_count" \
  --argjson pull_lookup_budget_ms "$lookup_budget_ms" \
  --argjson pull_lookup_wall_ms "$pull_lookup_wall_ms" \
  --argjson api_wall_ms "$api_wall_ms" \
  --argjson gh_version_exit "$gh_version_exit" \
  --argjson gh_version_wall_ms "$gh_version_wall_ms" \
  --argjson wall_ms "$(( $(date +%s%3N) - start_ms ))" \
  --argjson source_binding_exit "$source_binding_exit" \
  --argjson original_exit 0 \
  '{schema: $schema, repository: $repository, source_sha: $source_sha, expected_source_sha: $expected_source_sha, source_binding_exit: $source_binding_exit, source_binding_reason: $source_binding_reason, run_id: $run_id, run_attempt: $run_attempt, observed_at: $observed_at, wall_ms: $wall_ms, wall_scope: $wall_scope, api_wall_ms: $api_wall_ms, original_exit: $original_exit, gh_version_exit: $gh_version_exit, gh_version_wall_ms: $gh_version_wall_ms, scope: $scope, pagination: "full", api: {pages: $api_pages, reported_total_count: $reported_total_count, observed_count: ($caches | length)}, plan: {read_only: true, planned_deletions: 0, mutations_attempted: 0}, cache_cleanup_protection: {dev: $cleanup_protection, main: $cleanup_protection, default_branch: $cleanup_protection, non_pull_request_refs: $cleanup_protection, pull_request_refs: "STATE_CHECKED"}, pull_lookup: {budget_ms: $pull_lookup_budget_ms, wall_ms: $pull_lookup_wall_ms, status: $pull_lookup_status}, verdict: $verdict, caches: $caches}' \
  >"$receipt_path"

if [ "$cache_verdict" != PASS ]; then
  exit 1
fi
