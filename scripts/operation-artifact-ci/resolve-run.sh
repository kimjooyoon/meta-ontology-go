#!/usr/bin/env bash
set -euo pipefail

for _ in $(seq 1 100); do
  runs="$(gh run list --repo "$GITHUB_REPOSITORY" --workflow CI \
    --commit "$HEAD_SHA" --limit 20 \
    --json databaseId,status,conclusion,event,headSha)"
  record="$(jq -r --arg event "$RUN_EVENT" --arg sha "$HEAD_SHA" '
    [.[] | select(.event == $event and .headSha == $sha)]
    | sort_by(.databaseId) | last // empty
    | [.databaseId, .status, (.conclusion // "")] | @tsv
  ' <<<"$runs")"
  if [[ -z "$record" ]]; then
    sleep 6
    continue
  fi
  IFS=$'\t' read -r run_id status conclusion <<<"$record"
  if [[ "$status" != "completed" ]]; then
    sleep 6
    continue
  fi
  if [[ "$conclusion" != "success" ]]; then
    printf 'canonical run %s completed as %s\n' "$run_id" "$conclusion" >&2
    exit 1
  fi
  printf 'run_id=%s\n' "$run_id" >>"$GITHUB_OUTPUT"
  exit 0
done

printf 'exact-head canonical run timed out for %s\n' "$HEAD_SHA" >&2
exit 1

