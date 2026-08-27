#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"

source_path="examples/partial-knowledge-composition/main.gooo"
source_file="$GITHUB_WORKSPACE/$source_path"
work="$RUNNER_TEMP/partial-knowledge-composition"
mkdir -p "$work/base" "$work/semantic" "$work/comment-only"
git status --porcelain=v1 > "$work/repository-before.txt"

go_files=(
  cmd/partial-knowledge-composition-witness/main.go
  cmd/partial-knowledge-composition-verifier/main.go
  internal/meta/partialknowledgecomposition
  internal/meta/partialknowledgecomposition/verify
)
test -z "$(gofmt -l "${go_files[@]}")"
go test ./cmd/partial-knowledge-composition-witness ./cmd/partial-knowledge-composition-verifier ./internal/meta/partialknowledgecomposition ./internal/meta/partialknowledgecomposition/verify

producer_package="github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition"
producer_imports="$(go list -deps ./internal/meta/partialknowledgecomposition/verify | awk -v package="$producer_package" '$0 == package { count++ } END { print count + 0 }')"
test "$producer_imports" -eq 0
producer_imports_total=0

run_case() {
  local mode="$1" input_source="$2" output_dir="$3"
  go run ./cmd/partial-knowledge-composition-witness \
    --head-sha "$HEAD_SHA" --source-file "$input_source" --source-path "$source_path" \
    --intervention "$mode" --output "$output_dir/receipt.json"
  go run ./cmd/partial-knowledge-composition-verifier \
    --head-sha "$HEAD_SHA" --source-file "$input_source" --source-path "$source_path" \
    --intervention "$mode" --receipt "$output_dir/receipt.json" --output "$output_dir/verification.json"
  jq -e '.status == "VERIFIED" and .decision == "CALCULUS_PROVEN" and .resolution == "CALCULUS" and .independent_evaluator == true and .repository_writes == 0 and .promotion_authorized == false' "$output_dir/verification.json"
}

run_case none "$source_file" "$work/base"
{
  printf '%s\n' '// nonsemantic comment intervention'
  cat "$source_file"
} > "$work/comment-only/main.gooo"
run_case comment-only "$work/comment-only/main.gooo" "$work/comment-only"
run_case semantic "$source_file" "$work/semantic"

jq -e '
  .source_cases == 5 and .source_cases_total == 5 and
  .exact_cases == 1 and .direct_unknown_cases == 1 and
  .dependency_blocked_cases == 1 and .invariant_only_cases == 1 and
  .mixed_unresolved_cases == 1 and .top_success_cases == 1 and
  .open_claims == 4 and .claim_transition_total == 5
' "$work/base/verification.json"

base_projection="$(jq -r '.semantic_projection_digest' "$work/base/receipt.json")"
semantic_projection="$(jq -r '.semantic_projection_digest' "$work/semantic/receipt.json")"
test "$base_projection" != "$semantic_projection"
jq -e '
  ([.cases[] | select(.id == "direct-unknown") | .result.state] == ["EXACT"]) and
  ([.claims[] | select(.claim_id == "composition/direct-unknown") | .from, .to] == ["OPEN", "DISCHARGED"])
' "$work/semantic/receipt.json"
semantic_causality=1

test "$(jq -r '.semantic_ir_digest' "$work/base/receipt.json")" = "$(jq -r '.semantic_ir_digest' "$work/comment-only/receipt.json")"
test "$base_projection" = "$(jq -r '.semantic_projection_digest' "$work/comment-only/receipt.json")"
test "$(jq -r '.source_digest' "$work/base/receipt.json")" != "$(jq -r '.source_digest' "$work/comment-only/receipt.json")"
nonsemantic_preservation=1

open_claims_preserved="$(jq '[.claims[] | select(.from == "OPEN" and .to == "OPEN")] | length' "$work/base/receipt.json")"
test "$open_claims_preserved" -eq 4
open_claims_preserved_total=4
source_cases="$(jq -r '.source_cases' "$work/base/receipt.json")"
source_cases_total="$(jq -r '.source_cases_total' "$work/base/receipt.json")"
repository_writes="$(jq -r '.repository_writes' "$work/base/receipt.json")"
promotion_authorized="$(jq -r '.promotion_authorized' "$work/base/receipt.json")"
git status --porcelain=v1 > "$work/repository-after.txt"
if cmp -s "$work/repository-before.txt" "$work/repository-after.txt"; then
  repository_writes_observed=0
else
  repository_writes_observed=1
fi
test "$repository_writes" -eq 0
test "$repository_writes_observed" -eq 0

jq -n \
  --arg schema 'gooo/meta/partial-knowledge-composition-ci/v1' \
  --arg head "$HEAD_SHA" \
  --arg base_receipt "$(jq -r '.digest' "$work/base/receipt.json")" \
  --arg semantic_receipt "$(jq -r '.digest' "$work/semantic/receipt.json")" \
  --arg comment_receipt "$(jq -r '.digest' "$work/comment-only/receipt.json")" \
  --arg semantic_projection "$base_projection" \
  --arg semantic_intervention_projection "$semantic_projection" \
  --argjson producer_imports "$producer_imports" \
  --argjson producer_imports_total "$producer_imports_total" \
  --argjson source_cases "$source_cases" \
  --argjson source_cases_total "$source_cases_total" \
  --argjson semantic_causality "$semantic_causality" \
  --argjson semantic_causality_total 1 \
  --argjson nonsemantic_preservation "$nonsemantic_preservation" \
  --argjson nonsemantic_preservation_total 1 \
  --argjson open_claims_preserved "$open_claims_preserved" \
  --argjson open_claims_preserved_total "$open_claims_preserved_total" \
  --argjson repository_writes "$repository_writes_observed" \
  --argjson receipt_repository_writes "$repository_writes" \
  --argjson promotion_authorized false \
  '{schema:$schema,subject_sha:$head,decision:"CALCULUS_PROVEN",resolution:"CALCULUS",producer_imports:$producer_imports,producer_imports_total:$producer_imports_total,source_cases:$source_cases,source_cases_total:$source_cases_total,semantic_causality:$semantic_causality,semantic_causality_total:$semantic_causality_total,nonsemantic_preservation:$nonsemantic_preservation,nonsemantic_preservation_total:$nonsemantic_preservation_total,open_claims_preserved:$open_claims_preserved,open_claims_preserved_total:$open_claims_preserved_total,repository_writes:$repository_writes,receipt_repository_writes:$receipt_repository_writes,promotion_authorized:$promotion_authorized,receipts:{base:$base_receipt,semantic_intervention:$semantic_receipt,comment_only:$comment_receipt},semantic_projection_digest:$semantic_projection,semantic_intervention_projection_digest:$semantic_intervention_projection}' \
  > "$work/report.json"

cat > "$work/summary.md" <<EOF
# Partial knowledge composition — exact-head Action evidence

- subject: `$HEAD_SHA`
- decision: `CALCULUS_PROVEN` / resolution: `CALCULUS`
- source-derived cases: `$source_cases/$source_cases_total`
- producer imports in independent verifier: `$producer_imports/$producer_imports_total`
- semantic intervention causality: `$semantic_causality/1`
- nonsemantic comment preservation: `$nonsemantic_preservation/1`
- open claims preserved: `$open_claims_preserved/$open_claims_preserved_total`
- repository writes observed: `$repository_writes_observed`; receipt writes: `$repository_writes`; promotion authorized: `$promotion_authorized`

## Subject outcomes

| Case | Resolution | Claim transition |
|---|---|---|
| exact-pair | EXACT | OPEN → DISCHARGED |
| direct-unknown | LOWER_RESOLUTION | OPEN → OPEN |
| dependency-blocked | LOWER_RESOLUTION | OPEN → OPEN |
| invariant-preservation | INVARIANT_ONLY | OPEN → OPEN |
| mixed-unknown-and-blocked | LOWER_RESOLUTION | OPEN → OPEN |

The receipt-level calculus proof is separate from subject-case promotion.
The semantic intervention makes the direct observation available and changes
its result; the comment-only intervention changes only the source digest.
EOF

cp "$source_file" "$work/source.gooo"
sha256sum "$work"/*.json "$work"/*/*.json "$work"/*.gooo > "$work/manifest.sha256"
cat "$work/summary.md" >> "$GITHUB_STEP_SUMMARY"
