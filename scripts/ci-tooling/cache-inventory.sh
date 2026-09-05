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

set +e
timeout 180s gh api --paginate --slurp \
  -H 'Accept: application/vnd.github+json' \
  "repos/$GH_REPO/actions/caches?per_page=100" >"$pages_path" 2>"$output_dir/pagination-stderr.log"
api_exit=$?
set -e

if [ "$api_exit" -ne 0 ]; then
  printf 'schema=ci-tooling/cache-inventory/v1\npagination=FAILED\napi_exit=%s\nplan=read_only\nplanned_deletions=0\n' "$api_exit" | tee "$text_path"
  jq -n \
    --arg schema 'ci-tooling/cache-inventory/v1' \
    --arg source_sha "$OBSERVE_SOURCE_SHA" \
    --arg run_id "$OBSERVE_RUN_ID" \
    --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
    --arg observed_at "$observed_at" \
    --arg scope "${OBSERVE_SCOPE:-repository Actions caches; no deletions}" \
    --argjson wall_ms "$(( $(date +%s%3N) - start_ms ))" \
    --argjson original_exit "$api_exit" \
    '{schema: $schema, source_sha: $source_sha, run_id: $run_id, run_attempt: $run_attempt, observed_at: $observed_at, wall_ms: $wall_ms, original_exit: $original_exit, scope: $scope, pagination: "FAILED", plan: {read_only: true, planned_deletions: 0}, caches: []}' \
    > "$receipt_path"
  printf '[]\n' > "$inventory_path"
  exit 1
fi

set +e
jq '[.[] | .actions_caches[]? | {id, ref, key, size_in_bytes, created_at, last_accessed_at}] | sort_by([(.ref // ""), (.key // ""), (.id | tostring)])' \
  "$pages_path" >"$inventory_path"
parse_exit=$?
set -e

if [ "$parse_exit" -ne 0 ]; then
  printf 'schema=ci-tooling/cache-inventory/v1\npagination=FAILED\napi_exit=0\nparse_exit=%s\nplan=read_only\nplanned_deletions=0\n' "$parse_exit" | tee "$text_path"
  jq -n \
    --arg schema 'ci-tooling/cache-inventory/v1' \
    --arg source_sha "$OBSERVE_SOURCE_SHA" \
    --arg run_id "$OBSERVE_RUN_ID" \
    --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
    --arg observed_at "$observed_at" \
    --arg scope "${OBSERVE_SCOPE:-repository Actions caches; no deletions}" \
    --argjson wall_ms "$(( $(date +%s%3N) - start_ms ))" \
    --argjson original_exit "$parse_exit" \
    '{schema: $schema, source_sha: $source_sha, run_id: $run_id, run_attempt: $run_attempt, observed_at: $observed_at, wall_ms: $wall_ms, original_exit: $original_exit, scope: $scope, pagination: "FAILED", plan: {read_only: true, planned_deletions: 0}, caches: []}' \
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

branch_protection() {
  local branch="$1"
  local branch_path="$output_dir/protection-$branch.json"
  local branch_exit
  set +e
  timeout 30s gh api -H 'Accept: application/vnd.github+json' "repos/$GH_REPO/branches/$branch/protection" >"$branch_path" 2>"$output_dir/protection-$branch-stderr.log"
  branch_exit=$?
  set -e
  if [ "$branch_exit" -eq 0 ]; then
    printf 'PROTECTED'
  else
    printf 'UNKNOWN'
  fi
}

dev_protection="$(branch_protection dev)"
main_protection="$(branch_protection main)"
default_protection='UNKNOWN'
if [ "$default_branch" != 'UNKNOWN' ] && [ "$default_branch" != 'dev' ] && [ "$default_branch" != 'main' ]; then
  default_protection="$(branch_protection "$default_branch")"
elif [ "$default_branch" = 'dev' ]; then
  default_protection="$dev_protection"
elif [ "$default_branch" = 'main' ]; then
  default_protection="$main_protection"
fi

pull_states='{}'
while IFS= read -r ref; do
  if [[ "$ref" =~ ^refs/pull/([0-9]+)/merge$ ]]; then
    number="${BASH_REMATCH[1]}"
    lookup_path="$output_dir/pull-$number.json"
    set +e
    timeout 30s gh api -H 'Accept: application/vnd.github+json' "repos/$GH_REPO/pulls/$number" >"$lookup_path" 2>"$output_dir/pull-$number-stderr.log"
    lookup_exit=$?
    set -e
    state='UNKNOWN'
    if [ "$lookup_exit" -eq 0 ]; then
      state_candidate="$(jq -r '.state // empty' "$lookup_path" 2>/dev/null || true)"
      if [ "$state_candidate" = 'open' ] || [ "$state_candidate" = 'closed' ]; then
        state="$state_candidate"
      fi
    fi
    pull_states="$(jq --arg number "$number" --arg state "$state" '. + {($number): $state}' <<< "$pull_states")"
  fi
done < <(jq -r '.[].ref // empty' "$inventory_path" | sort -u)

protection_json="$(jq -n --arg dev "$dev_protection" --arg main "$main_protection" --arg default "$default_protection" '{dev: $dev, main: $main, default: $default}')"
enriched_path="$output_dir/cache-inventory-enriched.json"
jq \
  --argjson pull_states "$pull_states" \
  --argjson protection "$protection_json" \
  --arg observed_at "$observed_at" \
  'map(. + ((.ref // "") as $ref |
    if ($ref | test("^refs/pull/[0-9]+/merge$")) then
      ($ref | capture("^refs/pull/(?<number>[0-9]+)/merge$").number) as $number |
      ($pull_states[$number] // "UNKNOWN") as $state |
      {pull_request_number: ($number | tonumber), pull_request_state: $state, protection_status: "NOT_APPLICABLE", eligible_candidate: ($state == "closed"), candidate_reason: (if $state == "closed" then "observed_closed_pull_request" elif $state == "open" then "open_pull_request_not_eligible" else "pull_request_lookup_unknown" end), observed_at: $observed_at}
    elif $ref == "refs/heads/dev" then
      {pull_request_state: "NOT_APPLICABLE", protection_status: $protection.dev, eligible_candidate: false, candidate_reason: "non_pull_request_ref", observed_at: $observed_at}
    elif $ref == "refs/heads/main" then
      {pull_request_state: "NOT_APPLICABLE", protection_status: $protection.main, eligible_candidate: false, candidate_reason: "non_pull_request_ref", observed_at: $observed_at}
    else
      {pull_request_state: "NOT_APPLICABLE", protection_status: "UNKNOWN", eligible_candidate: false, candidate_reason: "non_pull_request_ref", observed_at: $observed_at}
    end))' "$inventory_path" > "$enriched_path"

mv "$enriched_path" "$inventory_path"

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
  printf 'protected_dev=%s\n' "$dev_protection"
  printf 'protected_main=%s\n' "$main_protection"
  printf 'default_branch=%s\n' "$default_branch"
  printf 'default_branch_protection=%s\n' "$default_protection"
  jq -r '.[] | "cache id=\(.id) ref=\(.ref // "UNKNOWN") key=\(.key // "UNKNOWN") bytes=\(.size_in_bytes // "UNKNOWN") created_at=\(.created_at // "UNKNOWN") last_accessed_at=\(.last_accessed_at // "UNKNOWN") candidate_reason=\(.candidate_reason) observed_at=\(.observed_at)"' "$inventory_path"
} | tee "$text_path"

jq -n \
  --arg schema 'ci-tooling/cache-inventory/v1' \
  --arg repository "$GH_REPO" \
  --arg source_sha "$OBSERVE_SOURCE_SHA" \
  --arg run_id "$OBSERVE_RUN_ID" \
  --arg run_attempt "$OBSERVE_RUN_ATTEMPT" \
  --arg observed_at "$observed_at" \
  --arg scope "${OBSERVE_SCOPE:-repository Actions caches; no deletions}" \
  --arg default_branch "$default_branch" \
  --arg default_protection "$default_protection" \
  --arg dev_protection "$dev_protection" \
  --arg main_protection "$main_protection" \
  --argjson caches "$(jq -c . "$inventory_path")" \
  --argjson wall_ms "$(( $(date +%s%3N) - start_ms ))" \
  --argjson original_exit 0 \
  '{schema: $schema, repository: $repository, source_sha: $source_sha, run_id: $run_id, run_attempt: $run_attempt, observed_at: $observed_at, wall_ms: $wall_ms, original_exit: $original_exit, scope: $scope, pagination: "full", plan: {read_only: true, planned_deletions: 0, mutations_attempted: 0}, protection: {dev: $dev_protection, main: $main_protection, default_branch: $default_branch, default_branch_protection: $default_protection}, caches: $caches}' \
  >"$receipt_path"
