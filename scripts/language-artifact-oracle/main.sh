#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
source_dir="${SOURCE_EXECUTION_DIR:-source-execution-output}"
output=artifact-oracle-output
paths=(cmd/language-artifact-oracle internal/meta/languageartifactoracle)

go fix ./cmd/language-artifact-oracle ./internal/meta/languageartifactoracle
git diff --exit-code -- "${paths[@]}"
mapfile -t go_files < <(find "${paths[@]}" -name '*.go' -type f | sort)
gofmt -w "${go_files[@]}"
git diff --exit-code -- "${paths[@]}"
go test ./cmd/language-artifact-oracle ./internal/meta/languageartifactoracle

mkdir -p "$output"
go list -deps ./cmd/language-artifact-oracle | sort > "$output/oracle-dependencies.txt"
producer_dependencies=$(grep -Ec '^github.com/kimjooyoon/meta-ontology-go/internal/(sourceexecution|syntax|bidir|meta/languagesourceexecution)$' "$output/oracle-dependencies.txt" || true)
jq -n --argjson dependencies "$producer_dependencies" \
  '{schema:"gooo/language-artifact-oracle-independence/v1",producer_dependencies:$dependencies}' \
  > "$output/independence.json"
[[ "$producer_dependencies" == "0" ]]

forge_body=$(jq -c '.entry.output.id="forged://entity/payment" | .events[3].subject="forged://entity/payment" | .digest=""' "$source_dir/positive.json")
forge_digest=$(printf '%s' "$forge_body" | sha256sum | awk '{print $1}')
printf '%s' "$forge_body" | jq -c --arg digest "sha256:$forge_digest" '.digest=$digest' > "$output/forged.json"
unknown_body=$(jq -c '.decision="UNKNOWN" | .digest=""' "$source_dir/positive.json")
unknown_digest=$(printf '%s' "$unknown_body" | sha256sum | awk '{print $1}')
printf '%s' "$unknown_body" | jq -c --arg digest "sha256:$unknown_digest" '.digest=$digest' > "$output/unknown-decision.json"

legacy_common=(--head "$HEAD_SHA" --contract examples/language-source-execution/contract.json \
  --positive "$output/forged.json" --replay "$output/forged.json" \
  --unknown-entry "$source_dir/unknown-entry.json" --invalid-syntax "$source_dir/invalid-syntax.json")
go run ./cmd/language-source-execution-witness "${legacy_common[@]}" --out "$output/legacy-forgery-accepted.json"
jq -e '.decision=="PASS" and .summary.cases_satisfied==4 and .summary.cases_total==4' "$output/legacy-forgery-accepted.json"

common=(--head "$HEAD_SHA" --contract examples/language-artifact-oracle/contract.json \
  --source examples/billing/main.gooo --unsupported-source examples/language-artifact-oracle/unsupported.txt \
  --entry PayOrder --genuine "$source_dir/positive.json" --forged "$output/forged.json" \
  --unknown-decision "$output/unknown-decision.json" --legacy-acceptance "$output/legacy-forgery-accepted.json" \
  --independence "$output/independence.json")
go run ./cmd/language-artifact-oracle "${common[@]}" --out "$output/report.json"
go run ./cmd/language-artifact-oracle "${common[@]}" --out "$output/replay.json"
cmp -s "$output/report.json" "$output/replay.json"
jq -e '.decision=="PASS" and .resolution=="EXACT" and .summary.cases_satisfied==4 and .summary.cases_total==4' "$output/report.json"
jq -e '.summary.exact_source_bindings==1 and .summary.resealed_forgeries_rejected==1 and .summary.unknown_decisions_rejected==1' "$output/report.json"
jq -e '.summary.lower_resolutions==1 and .summary.legacy_validator_counterexamples==1 and .summary.producer_dependencies==0' "$output/report.json"
jq -e '.summary.unknown_checks==9 and .summary.semantic_correctness_claims==0 and ([.indicators[]|select(.satisfied==false)]|length)==0' "$output/report.json"
git diff --exit-code

{
  echo '## Independent Gooo artifact oracle'
  echo
  jq -r '"- decision: \(.decision) / \(.resolution)\n- cases: \(.summary.cases_satisfied)/\(.summary.cases_total)\n- exact source bindings: \(.summary.exact_source_bindings)\n- resealed forgeries rejected: \(.summary.resealed_forgeries_rejected)\n- legacy counterexamples captured: \(.summary.legacy_validator_counterexamples)\n- producer dependencies: \(.summary.producer_dependencies)\n- unknown checks: \(.summary.unknown_checks) at ORACLE_PARSE/project-source\n- semantic correctness claims: \(.summary.semantic_correctness_claims)\n- receipt: \(.digest)"' "$output/report.json"
} >> "$GITHUB_STEP_SUMMARY"
