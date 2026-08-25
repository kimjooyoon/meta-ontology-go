#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

paths=(cmd/gooo cmd/language-source-execution-witness internal/sourceexecution internal/meta/languagesourceexecution)
go fix ./cmd/gooo ./cmd/language-source-execution-witness ./internal/sourceexecution ./internal/meta/languagesourceexecution
git diff --exit-code -- "${paths[@]}"
mapfile -t go_files < <(find "${paths[@]}" -name '*.go' -type f | sort)
gofmt -w "${go_files[@]}"
git diff --exit-code -- "${paths[@]}"
go test ./cmd/gooo ./cmd/language-source-execution-witness ./internal/sourceexecution ./internal/meta/languagesourceexecution

mkdir -p source-execution-output
go build -o source-execution-output/gooo ./cmd/gooo
binary=source-execution-output/gooo
source=examples/billing/main.gooo
contract=examples/language-source-execution/contract.json

"$binary" run --json --entry PayOrder "$source" > source-execution-output/positive.json
"$binary" run --json --entry PayOrder "$source" > source-execution-output/replay.json
cmp -s source-execution-output/positive.json source-execution-output/replay.json
if "$binary" run --json --entry Missing "$source" > source-execution-output/unknown-entry.json; then exit 1; fi
if "$binary" run --json --entry Missing examples/language-source-execution/invalid.gooo > source-execution-output/invalid-syntax.json; then exit 1; fi

common=(--head "$HEAD_SHA" --contract "$contract" --replay source-execution-output/replay.json --unknown-entry source-execution-output/unknown-entry.json --invalid-syntax source-execution-output/invalid-syntax.json)
go run ./cmd/language-source-execution-witness "${common[@]}" --positive source-execution-output/positive.json --out source-execution-output/artifact.json
jq -e '.decision=="PASS" and .resolution=="EXACT" and .summary.cases_satisfied==4 and .summary.cases_total==4' source-execution-output/artifact.json
jq -e '.summary.source_executions==1 and .summary.deterministic_replays==1 and .summary.diagnostic_rejections==2 and .summary.execution_events==4' source-execution-output/artifact.json
jq -e '.summary.unknowns==0 and .summary.repository_writes==0 and .summary.mutation_authorities==0' source-execution-output/artifact.json

jq '.decision="UNKNOWN"' source-execution-output/positive.json > source-execution-output/unknown-decision.json
set +e
go run ./cmd/language-source-execution-witness "${common[@]}" --positive source-execution-output/unknown-decision.json --out source-execution-output/unknown-artifact.json
unknown_code=$?
set -e
[[ "$unknown_code" == "1" ]]
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .summary.unknowns>0' source-execution-output/unknown-artifact.json
git diff --exit-code

{
  echo '## Gooo source execution'
  echo
  jq -r '"- decision: \(.decision) / \(.resolution)\n- cases: \(.summary.cases_satisfied)/\(.summary.cases_total)\n- executions: \(.summary.source_executions)\n- deterministic replays: \(.summary.deterministic_replays)\n- diagnostic rejections: \(.summary.diagnostic_rejections)\n- repository writes: \(.summary.repository_writes)\n- receipt: \(.digest)"' source-execution-output/artifact.json
} >> "$GITHUB_STEP_SUMMARY"
