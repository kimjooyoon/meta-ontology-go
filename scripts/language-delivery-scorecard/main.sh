#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_RUN_ID:?GITHUB_RUN_ID is required}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
: "${GH_TOKEN:?GH_TOKEN is required}"

go fix ./cmd/language-delivery-scorecard ./internal/meta/languagedelivery
git diff --exit-code -- cmd/language-delivery-scorecard internal/meta/languagedelivery
mapfile -t go_files < <(find cmd/language-delivery-scorecard internal/meta/languagedelivery -name '*.go' -type f | sort)
gofmt -w "${go_files[@]}"
git diff --exit-code -- cmd/language-delivery-scorecard internal/meta/languagedelivery
go test ./cmd/language-delivery-scorecard ./internal/meta/languagedelivery

mkdir -p delivery-evidence delivery-output
artifacts=$(gh api "repos/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID/artifacts?per_page=100")
specs=(
  'USER_JOURNEY|user-journey-scorecard|first/report.json|user_journey'
  'TOOLCHAIN_CONFORMANCE|toolchain-conformance|artifact.json|conformance'
  'TOOLCHAIN_LSP|toolchain-lsp|artifact.json|lsp'
  'CROSS_PLATFORM_RELEASE|toolchain-cross-platform-release|artifact.json|release'
  'LANGUAGE_SOURCE_EXECUTION|language-source-execution|artifact.json|execution'
  'LANGUAGE_PROFILE|language-source-execution|profile/report.json|profile'
  'LANGUAGE_DEBUG|language-source-execution|debug/report.json|debug'
  'LANGUAGE_READINESS|language-readiness-artifact|artifact.json|readiness'
)
: > delivery-evidence/entries.jsonl

for spec in "${specs[@]}"; do
  IFS='|' read -r source prefix relative key <<<"$spec"
  name="$prefix-$HEAD_SHA"
  matches=$(jq --arg name "$name" '[.artifacts[] | select(.name == $name)] | length' <<<"$artifacts")
  [[ "$matches" == "1" ]]
  artifact=$(jq -c --arg name "$name" '.artifacts[] | select(.name == $name)' <<<"$artifacts")
  id=$(jq -r '.id' <<<"$artifact")
  archive=$(jq -r '.digest // ""' <<<"$artifact")
  [[ "$id" -gt 0 && "$archive" == sha256:* ]]
  mkdir -p "delivery-evidence/$key"
  gh api "repos/$GITHUB_REPOSITORY/actions/artifacts/$id/zip" > "delivery-evidence/$key.zip"
  unzip -q "delivery-evidence/$key.zip" -d "delivery-evidence/$key"
  report="delivery-evidence/$key/$relative"
  [[ -f "$report" ]]
  digest="sha256:$(sha256sum "$report" | cut -d' ' -f1)"
  jq -nc --arg source "$source" --argjson id "$id" --arg name "$name" --arg archive "$archive" --arg digest "$digest" '{source:$source,artifact_id:$id,artifact_name:$name,archive_digest:$archive,report_digest:$digest}' >> delivery-evidence/entries.jsonl
  printf -v "$key" '%s' "$report"
done

jq -s --arg sha "$HEAD_SHA" --argjson run "$GITHUB_RUN_ID" '{schema:"gooo/language-delivery-source-manifest/v3",subject_sha:$sha,workflow_run_id:$run,workflow_name:"Transformation effect ledger",workflow_decision:"upstream_jobs_success",artifacts:.,repository_writes:0}' delivery-evidence/entries.jsonl > delivery-evidence/manifest.json

run_scorecard() {
  local contract=$1 manifest=$2 user=$3 output=$4
  go run ./cmd/language-delivery-scorecard \
    --expected-head "$HEAD_SHA" \
    --contract "$contract" \
    --manifest "$manifest" \
    --user-journey "$user" \
    --conformance "$conformance" \
    --lsp "$lsp" \
    --release "$release" \
    --execution "$execution" \
    --profile "$profile" \
    --debug "$debug" \
    --readiness "$readiness" \
    --out "$output"
}

set +e
run_scorecard examples/language-delivery-scorecard/contract.json delivery-evidence/manifest.json "$user_journey" delivery-output/report.json
scorecard_code=$?
set -e
if [[ "$scorecard_code" != "0" ]]; then
  jq -r '.sources[] | select(.state != "PASS") | "language delivery source: \(.source) state=\(.state) reason=\(.reason) decision=\(.decision) resolution=\(.resolution)"' delivery-output/report.json
  jq -r '.obligations[] | select(.status == "UNKNOWN" or .status == "NOT_SATISFIED") | "language delivery obligation: \(.id) status=\(.status) reason=\(.reason) observed=\(.observed) expected=\(.expected)"' delivery-output/report.json
  exit "$scorecard_code"
fi
jq -e '.decision == "INCOMPLETE" and .resolution == "EXACT"' delivery-output/report.json
jq -e '.summary.coordinates == {satisfied:34,not_implemented:2,not_satisfied:0,unknown:0,total:36,basis_points:9444}' delivery-output/report.json
jq -e '[.views[] | [.audience,.coordinates.satisfied,.coordinates.total]] == [["USER",10,12],["TOOL_AUTHOR",22,24],["GOVERNOR",34,36]]' delivery-output/report.json
jq -e '.summary.internal_readiness.satisfied == 24 and .summary.internal_readiness.total == 24' delivery-output/report.json
jq -e '.summary.meta_bindings == 36 and .summary.source_receipts == 8 and .summary.source_receipts_total == 8 and .summary.effects.repository_writes == 0 and .summary.effects.mutation_authority == false' delivery-output/report.json

jq '.decision="UNKNOWN"' "$user_journey" > delivery-output/unknown-user.json
unknown_digest="sha256:$(sha256sum delivery-output/unknown-user.json | cut -d' ' -f1)"
jq --arg digest "$unknown_digest" '(.artifacts[] | select(.source=="USER_JOURNEY") | .report_digest)=$digest' delivery-evidence/manifest.json > delivery-output/unknown-manifest.json
set +e
run_scorecard examples/language-delivery-scorecard/contract.json delivery-output/unknown-manifest.json delivery-output/unknown-user.json delivery-output/unknown-report.json
unknown_code=$?
set -e
[[ "$unknown_code" == "1" ]]
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .summary.coordinates.unknown==7' delivery-output/unknown-report.json

jq 'del(.obligations[-1])' examples/language-delivery-scorecard/contract.json > delivery-output/drift-contract.json
if run_scorecard delivery-output/drift-contract.json delivery-evidence/manifest.json "$user_journey" delivery-output/drift-report.json; then
  exit 1
fi
jq '(.obligations[] | select(.id=="USER-RUN-SOURCE") | .evidence)={source:"USER_JOURNEY",kind:"JOURNEY",id:"version-text",target:1}' examples/language-delivery-scorecard/contract.json > delivery-output/self-minted-contract.json
if run_scorecard delivery-output/self-minted-contract.json delivery-evidence/manifest.json "$user_journey" delivery-output/self-minted-report.json; then
  exit 1
fi

{
  echo '## Language delivery scorecard'
  echo
  echo '| Reader | Satisfied | Fixed total | Basis points |'
  echo '|---|---:|---:|---:|'
  jq -r '.views[] | "| \(.audience) | \(.coordinates.satisfied) | \(.coordinates.total) | \(.coordinates.basis_points) |"' delivery-output/report.json
  echo
  jq -r '"Known gaps: **\(.summary.coordinates.not_implemented)**  Unknown evidence: **\(.summary.coordinates.unknown)**  Repository writes: **\(.summary.effects.repository_writes)**"' delivery-output/report.json
  echo
  jq -r '"Internal readiness is separately labeled: **\(.summary.internal_readiness.satisfied)/\(.summary.internal_readiness.total)** (`\(.summary.internal_readiness.claim)`)."' delivery-output/report.json
} >> "$GITHUB_STEP_SUMMARY"
