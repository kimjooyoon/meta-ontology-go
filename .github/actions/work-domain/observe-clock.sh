#!/usr/bin/env bash
set -euo pipefail

# The composite owns one 30-second acquisition deadline. Only the native clock
# classifier can request another read; it never grants a new work allowance.
tool="$1"
origins="$2"
job_name="$3"
observations="$(mktemp -d "$origins/clock-observations.XXXXXXXX")"

capture() {
  local output="$1"
  shift
  "$@" > "$output" 2> "$output.stderr" || {
    local code=$?
    cat "$output.stderr" >&2
    return "$code"
  }
}

for observation in {1..16}; do
  frame="$observations/$(printf '%02d' "$observation")"
  mkdir "$frame"
  capture "$frame/workflow-run.json" gh api \
    "repos/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID"
  capture "$frame/attempt-job-pages.json" gh api --paginate --slurp \
    "repos/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID/attempts/$GITHUB_RUN_ATTEMPT/jobs?per_page=100"
  # Retain the client's paginated output separately from this derived merge.
  capture "$frame/attempt-jobs.json" jq -e '{total_count:.[0].total_count,jobs:[.[].jobs[]]}' \
    "$frame/attempt-job-pages.json"
  if capture "$frame/runner-clock.json" "$tool" runner-clock \
    --run-json "$frame/workflow-run.json" --jobs-json "$frame/attempt-jobs.json" \
    --head "$EXPECTED_SOURCE_HEAD" --run-id "$GITHUB_RUN_ID" --attempt "$GITHUB_RUN_ATTEMPT" \
    --budget-ms "$GOOO_CI_BUDGET_MS" --job "$job_name" --producer 'Work domain preparation'; then
    code=0
  else
    code=$?
  fi
  printf '%s\n' "$code" > "$frame/runner-clock.exit"
  if test "$code" -eq 0; then
    for name in workflow-run.json attempt-job-pages.json attempt-jobs.json runner-clock.json; do
      cp "$frame/$name" "$origins/$name"
    done
    printf '%s\n' "$frame" > "$origins/selected-clock-observation.txt"
    exit 0
  fi
  if test "$code" -ne 75; then exit "$code"; fi
  if test "$observation" -lt 16; then sleep 1; fi
done
printf 'Current job was not observed in progress within 16 snapshots; no clock admitted.\n' >&2
exit 75
