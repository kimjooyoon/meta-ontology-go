#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
source_path="examples/reflective-query-sandbox/main.gooo"
output="$RUNNER_TEMP/reflective-query-sandbox/producer"
mkdir -p "$output"
cd "$root"

before=$(git status --porcelain=v1 --untracked-files=all)
go run ./cmd/gooo check "$source_path"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "$HEAD_SHA" -output "$output/observation.json"
go run ./scripts/reflective-query-sandbox/producer \
	-source "$source_path" -subject-sha "$HEAD_SHA" -output "$output/replay.json"
cmp -s "$output/observation.json" "$output/replay.json"
after=$(git status --porcelain=v1 --untracked-files=all)
test "$before" = "$after"

jq -e --arg sha "$HEAD_SHA" '
  .schema == "gooo/reflective-query-sandbox-observation/v1" and
  .subject_sha == $sha and
  .source == {path:"examples/reflective-query-sandbox/main.gooo", source_digest:.source.source_digest, semantic_digest:.source.semantic_digest, node_count:9, fact_count:8, gooo_lines:11} and
  (.attempts | length) == 5 and
  ([.attempts[] | select(.decision == "PASS" and .resolution == "EXACT")] | length) == 3 and
  ([.attempts[] | select(.decision == "DENIED" and .reason == "READ_ONLY_QUERY_ONLY")] | length) == 1 and
  ([.attempts[] | select(.decision == "UNKNOWN" and .reason == "UNKNOWN_TARGET")] | length) == 1 and
  (.claims | length) == 24 and
  .effects == {repository_writes:0, mutation_authority:false}
' "$output/observation.json" >/dev/null

echo 'reflective query sandbox producer: PASS 5 attempts / 24 transitions'
