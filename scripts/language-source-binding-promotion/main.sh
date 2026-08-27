#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

output="${PROMOTION_OUTPUT:-source-binding-promotion-output}"
source_input="${SOURCE_EXECUTION_INPUT:-source-execution-input}"
oracle_input="${ARTIFACT_ORACLE_INPUT:-artifact-oracle-input}"
policy="examples/self-improvement/main.gooo"
contract="examples/language-source-binding-promotion/contract.json"
mkdir -p "$output/policy/first" "$output/policy/replay"

go fix ./internal/meta/languagesourcebindingpromotion ./cmd/language-source-binding-promotion
test -z "$(git diff -- internal/meta/languagesourcebindingpromotion cmd/language-source-binding-promotion)"
test -z "$(gofmt -l internal/meta/languagesourcebindingpromotion cmd/language-source-binding-promotion)"
go test ./internal/meta/languagesourcebindingpromotion ./cmd/language-source-binding-promotion

dependencies="$(go list -deps ./internal/meta/languagesourcebindingpromotion | \
  rg '/internal/(sourceexecution|meta/languagesourceexecution|meta/languageartifactoracle)$' || true)"
dependency_count="$(printf '%s\n' "$dependencies" | sed '/^$/d' | wc -l | tr -d ' ')"
test "$dependency_count" -eq 0
printf '%s\n' "$dependencies" > "$output/producer-dependencies.txt"
printf '{"schema":"%s","producer_dependencies":%d}\n' \
  'gooo/language-source-binding-promotion-independence/v1' "$dependency_count" > "$output/independence.json"

go run ./cmd/gooo check "$policy"
go run ./cmd/gooo generate "$policy" --out "$output/policy/first"
go run ./cmd/gooo generate "$policy" --out "$output/policy/replay"
first_policy="$(find "$output/policy/first" -type f -name '*.go' -print -quit)"
replay_policy="$(find "$output/policy/replay" -type f -name '*.go' -print -quit)"
test -n "$first_policy" && test -n "$replay_policy"
cmp "$first_policy" "$replay_policy"

reseal() {
  local source="$1" target="$2" filter="$3" temporary body digest
  temporary="$(mktemp)"
  jq "$filter | .digest = \"\"" "$source" > "$temporary"
  body="$(jq -c '.' "$temporary")"
  digest="sha256:$(printf '%s' "$body" | sha256sum | cut -d' ' -f1)"
  jq --arg digest "$digest" '.digest = $digest' "$temporary" > "$target"
  rm -f "$temporary"
}

producer="$source_input/artifact.json"
receipt="$source_input/positive.json"
oracle="$oracle_input/report.json"
reseal "$producer" "$output/unknown-producer.json" '.decision = "UNKNOWN"'
reseal "$oracle" "$output/unknown-oracle.json" '.decision = "UNKNOWN"'
reseal "$oracle" "$output/mismatched-oracle.json" \
  '(.cases[] | select(.id == "genuine-source-bound") | .artifact_digest) = "sha256:0000000000000000000000000000000000000000000000000000000000000000"'

common=(--contract "$contract" --head "$HEAD_SHA" --policy-source "$policy"
  --policy-artifact "$first_policy" --policy-replay "$replay_policy"
  --producer "$producer" --receipt "$receipt" --oracle "$oracle"
  --unknown-producer "$output/unknown-producer.json" --unknown-oracle "$output/unknown-oracle.json"
  --mismatched-oracle "$output/mismatched-oracle.json" --independence "$output/independence.json")
go run ./cmd/language-source-binding-promotion "${common[@]}" --output "$output/report.json"
go run ./cmd/language-source-binding-promotion "${common[@]}" --output "$output/replay.json"
cmp "$output/report.json" "$output/replay.json"
go run ./cmd/language-source-binding-promotion --check "$output/report.json"

jq -e '.decision == "PASS" and .resolution == "EXACT" and
  .summary.cases_satisfied == 5 and .summary.cases_total == 5 and
  .summary.exact_promotions == 1 and .summary.exact_claims == 3 and
  .summary.direct_unknowns == 3 and .summary.dependency_blocked == 3 and
  .summary.link_refutations == 1 and .summary.policy_replays == 1 and
  .summary.producer_dependencies == 0 and .summary.semantic_correctness_claims == 0 and
  .repository_writes == 0 and .mutation_authority == false' "$output/report.json"
jq -e '([.cases[] | select(.status == "SATISFIED")] | length) == 5 and
  ([.cases[] | select((.claims | length) == 3)] | length) == 5 and
  ([.cases[] | select(.id == "exact-promotion") | .claims[] | select(.status == "DISCHARGED")] | length) == 3 and
  ([.cases[].claims[] | select(.unknown_class == "DIRECT_MISSING")] | length) == 3 and
  ([.cases[].claims[] | select(.unknown_class == "DEPENDENCY_BLOCKED")] | length) == 3' "$output/report.json"
sha256sum "$output"/*.json "$output"/policy/*/*.go > "$output/manifest.sha256"
