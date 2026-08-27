#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

output="${PROOF_CARRYING_OUTPUT:-proof-carrying-artifact-output}"
source_path="examples/language-proof-carrying-artifact/main.gooo"
recipe="examples/language-proof-carrying-artifact/recipe.json"
contract="examples/language-proof-carrying-artifact/contract.json"
producer_paths=(cmd/language-proof-carrying-artifact internal/meta/languageproofartifact)
verifier_paths=(cmd/language-proof-carrying-artifact-verifier internal/meta/languageproofartifactverifier)
mkdir -p "$output"

go fix ./cmd/language-proof-carrying-artifact ./internal/meta/languageproofartifact \
  ./cmd/language-proof-carrying-artifact-verifier ./internal/meta/languageproofartifactverifier
test -z "$(git diff -- "${producer_paths[@]}" "${verifier_paths[@]}")"
test -z "$(gofmt -l "${producer_paths[@]}" "${verifier_paths[@]}")"
go test ./cmd/language-proof-carrying-artifact ./internal/meta/languageproofartifact \
  ./cmd/language-proof-carrying-artifact-verifier ./internal/meta/languageproofartifactverifier

go run ./cmd/gooo check "$source_path"
go run ./cmd/gooo run --json --entry GenerateProofCarryingArtifact "$source_path" > "$output/operation-receipt.json"
go run ./cmd/language-proof-carrying-artifact \
  -head "$HEAD_SHA" -source-path "$source_path" -source "$source_path" \
  -operation "$output/operation-receipt.json" -recipe "$recipe" -out "$output/artifact.json"

go list -deps ./cmd/language-proof-carrying-artifact-verifier | sort > "$output/verifier-dependencies.txt"
producer_dependencies="$(grep -E '/internal/(meta/languageproofartifact|sourceexecution|syntax|bidir)(/|$)' "$output/verifier-dependencies.txt" | wc -l | tr -d ' ' || true)"
jq -n --argjson dependencies "$producer_dependencies" \
  '{schema:"gooo/language-proof-carrying-artifact-independence/v1",producer_dependencies:$dependencies}' \
  > "$output/independence.json"
test "$producer_dependencies" -eq 0

reseal() {
  local input="$1" target="$2" filter="$3" body digest
  body="$(jq -c "$filter | .digest = \"\"" "$input")"
  digest="sha256:$(printf '%s' "$body" | sha256sum | cut -d' ' -f1)"
  jq -c --arg digest "$digest" '.digest = $digest' <<< "$body" > "$target"
}

reseal "$output/artifact.json" "$output/tampered.json" \
  '(.evidence[] | select(.kind == "SOURCE") | .source_digest) = "sha256:0000000000000000000000000000000000000000000000000000000000000000"'
reseal "$output/artifact.json" "$output/missing-operation.json" \
  'del(.evidence[] | select(.kind == "OPERATION"))'
cp "$output/artifact.json" "$output/byte-only.json"
jq '.steps[0].meta_operation = "accept-source-claim-without-recheck"' "$recipe" > "$output/wrong-recipe.json"

common=(--head "$HEAD_SHA" --contract "$contract" --valid "$output/artifact.json" \
  --tampered "$output/tampered.json" --missing "$output/missing-operation.json" \
  --byte-only "$output/byte-only.json" --wrong-recipe "$output/wrong-recipe.json" \
  --source "$source_path" --operation "$output/operation-receipt.json" --recipe "$recipe" \
  --independence "$output/independence.json")
go run ./cmd/language-proof-carrying-artifact-verifier "${common[@]}" -output "$output/report.json"
go run ./cmd/language-proof-carrying-artifact-verifier "${common[@]}" -output "$output/replay.json"
cmp -s "$output/report.json" "$output/replay.json"
go run ./cmd/language-proof-carrying-artifact-verifier -check "$output/report.json"

jq -e '
  .decision == "PASS" and .resolution == "EXACT" and
  .authority_granted == true and
  .summary.cases_satisfied == 5 and .summary.cases_total == 5 and
  .summary.valid_artifacts == 1 and .summary.evidence_kinds_carried == 3 and
  .summary.exact_evidence_links == 3 and .summary.recipe_matches == 1 and
  .summary.preserved_transitions == 4 and .summary.transition_total == 4 and
  .summary.tampered_rejections == 1 and .summary.missing_evidence_rejections == 1 and
  .summary.byte_only_denials == 1 and .summary.recipe_rejections == 1 and
  .summary.producer_dependencies == 0 and .summary.generated_authority == 0 and
  .summary.semantic_claims == 0 and .summary.repository_writes == 0 and
  .summary.mutation_authorities == 0 and
  ([.indicators[] | select(.satisfied == false)] | length) == 0 and
  ([.claim_transitions[] | select(.from == "CARRIED" and .to == "PRESERVED")] | length) == 3 and
  ([.claim_transitions[] | select(.claim_id == "consumer-authority" and .from == "NOT_GRANTED" and .to == "GRANTED")] | length) == 1
' "$output/report.json"
sha256sum "$output"/*.json > "$output/manifest.sha256"

{
  echo '## Proof-carrying Gooo artifact'
  echo
  jq -r '"- decision: `\(.decision)` / `\(.resolution)`; consumer authority: `\(.authority_granted)`"' "$output/report.json"
  jq -r '"- cases: `\(.summary.cases_satisfied)/\(.summary.cases_total)`; evidence kinds: `\(.summary.evidence_kinds_carried)/3`; exact links: `\(.summary.exact_evidence_links)/3`"' "$output/report.json"
  jq -r '"- preserved transitions: `\(.summary.preserved_transitions)/\(.summary.transition_total)`; recipe matches: `\(.summary.recipe_matches)`"' "$output/report.json"
  jq -r '"- tampered: `\(.summary.tampered_rejections)`; missing: `\(.summary.missing_evidence_rejections)`; byte-only denials: `\(.summary.byte_only_denials)`"' "$output/report.json"
  jq -r '"- producer dependencies: `\(.summary.producer_dependencies)`; generated authority: `\(.summary.generated_authority)`; receipt: `\(.digest)`"' "$output/report.json"
} >> "$GITHUB_STEP_SUMMARY"
