#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$root"

out="${RUNNER_TEMP:-/tmp}/gooo-ci-plan"
rm -rf "$out"
mkdir -p "$out/bin" "$out/generated-a" "$out/generated-b" "$out/reports" "$out/replays"

go fix ./...
git diff --exit-code
mapfile -t go_files < <(find cmd internal -type f -name '*.go' -print | sort)
gofmt -w "${go_files[@]}"
git diff --exit-code
go test ./internal/metainvocation ./internal/meta/ciplanusecase ./cmd/gooo ./cmd/ci-plan-scorecard
go build -o "$out/bin/gooo" ./cmd/gooo
go build -o "$out/bin/ci-plan-scorecard" ./cmd/ci-plan-scorecard

source_file="examples/ci-plan/main.gooo"
"$out/bin/gooo" check "$source_file"
"$out/bin/gooo" generate "$source_file" --out "$out/generated-a"
"$out/bin/gooo" generate "$source_file" --out "$out/generated-b"

samples="$out/samples.ndjson"
: > "$samples"
for fixture in examples/ci-plan/fixtures/*.json; do
  case_id="$(basename "$fixture" .json)"
  expected_exit=1
  expected_decision="UNKNOWN"
  case "$case_id" in
    pass-*) expected_exit=0; expected_decision="PASS" ;;
    fail-*) expected_decision="FAIL_CLOSED" ;;
  esac

  set +e
  /usr/bin/time -f '%e %M' -o "$out/time.txt" \
    "$out/bin/gooo" invoke --json --entry PlanCI --input "$fixture" "$source_file" \
    > "$out/reports/$case_id.json"
  exit_code=$?
  set -e
  if [[ "$exit_code" -ne "$expected_exit" ]]; then
    echo "$case_id exit=$exit_code want=$expected_exit" >&2
    exit 1
  fi
  jq -e --arg decision "$expected_decision" '.decision == $decision' "$out/reports/$case_id.json" >/dev/null

  set +e
  "$out/bin/gooo" invoke --json --entry PlanCI --input "$fixture" "$source_file" \
    > "$out/replays/$case_id.json"
  replay_exit=$?
  set -e
  if [[ "$replay_exit" -ne "$expected_exit" ]]; then
    echo "$case_id replay exit=$replay_exit want=$expected_exit" >&2
    exit 1
  fi
  cmp "$out/reports/$case_id.json" "$out/replays/$case_id.json"

  read -r seconds peak_rss < "$out/time.txt"
  wall_ms="$(awk -v seconds="$seconds" 'BEGIN { printf "%.0f", seconds * 1000 }')"
  receipt_bytes="$(wc -c < "$out/reports/$case_id.json" | tr -d ' ')"
  jq -n \
    --arg case_id "$case_id" \
    --argjson wall_ms "$wall_ms" \
    --argjson peak_rss_kib "$peak_rss" \
    --argjson receipt_bytes "$receipt_bytes" \
    '{case_id:$case_id,wall_ms:$wall_ms,peak_rss_kib:$peak_rss_kib,receipt_bytes:$receipt_bytes}' \
    >> "$samples"
done

jq -s '{schema:"gooo/ci-plan-resource-profile/v1",samples:.}' "$samples" > "$out/profile.json"

"$out/bin/ci-plan-scorecard" \
  --contract examples/ci-plan/contract.json \
  --source "$source_file" \
  --generated-a "$out/generated-a" \
  --generated-b "$out/generated-b" \
  --reports "$out/reports" \
  --replays "$out/replays" \
  --golden examples/ci-plan/golden \
  --profile "$out/profile.json" \
  --output "$out/scorecard.json"
"$out/bin/ci-plan-scorecard" --check --output "$out/scorecard.json"

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo '## Gooo CI plan use case'
    echo
    echo '| Reader | Indicator | Observed | Target | Status |'
    echo '|---|---|---:|---:|---|'
    jq -r '.indicators[] | "| \(.reader) | `\(.id)` | \(.observed) | \(.comparator) \(.target) | \(.status) |"' "$out/scorecard.json"
    echo
    echo '### Unknown coordinates'
    echo
    jq -r '.cases[] | select(.observed_decision == "UNKNOWN") | .unknowns[] | "- `\(.file)`: `\(.stage)/\(.step)/\(.reason)`"' "$out/scorecard.json"
    echo
    echo '### Interpretation'
    echo
    jq -r '"`\(.decision)` at `\(.resolution)`: `\(.interpretation)`"' "$out/scorecard.json"
  } >> "$GITHUB_STEP_SUMMARY"
fi
