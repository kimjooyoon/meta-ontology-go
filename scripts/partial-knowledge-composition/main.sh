#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"

source_path="examples/partial-knowledge-composition/main.gooo"
source_file="$GITHUB_WORKSPACE/$source_path"
work="$RUNNER_TEMP/partial-knowledge-composition"
mkdir -p "$work/base" "$work/semantic" "$work/comment-only" "$work/snapshots"

snapshot_repository() {
  local prefix="$1"
  git ls-files --cached | sort > "$work/snapshots/${prefix}-tracked.txt"
  git ls-files --others --exclude-standard | sort > "$work/snapshots/${prefix}-untracked.txt"
  git status --porcelain=v1 --untracked-files=all | sort > "$work/snapshots/${prefix}-status.txt"
}

snapshot_repository before

go_files=(
  cmd/partial-knowledge-composition-observer/main.go
  cmd/partial-knowledge-composition-witness/main.go
  cmd/partial-knowledge-composition-verifier/main.go
  internal/meta/partialknowledgecomposition
  internal/meta/partialknowledgecomposition/provider
  internal/meta/partialknowledgecomposition/verify
)
test -z "$(gofmt -l "${go_files[@]}")"
go test ./cmd/partial-knowledge-composition-observer ./cmd/partial-knowledge-composition-witness ./cmd/partial-knowledge-composition-verifier ./internal/meta/partialknowledgecomposition ./internal/meta/partialknowledgecomposition/provider ./internal/meta/partialknowledgecomposition/verify

producer_package="github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition"
producer_imports_count="$(go list -deps ./internal/meta/partialknowledgecomposition/verify | awk -v package="$producer_package" '$0 == package { count++ } END { print count + 0 }')"
if test "$producer_imports_count" -eq 0; then
  imports_producer_package=1
else
  imports_producer_package=0
fi
imports_producer_package_total=1
test "$imports_producer_package" -eq 1

snapshot_repository after

snapshot_args=(
  --before-tracked "$work/snapshots/before-tracked.txt"
  --before-untracked "$work/snapshots/before-untracked.txt"
  --before-status "$work/snapshots/before-status.txt"
  --after-tracked "$work/snapshots/after-tracked.txt"
  --after-untracked "$work/snapshots/after-untracked.txt"
  --after-status "$work/snapshots/after-status.txt"
)

run_case() {
  local mode="$1" input_source="$2" output_dir="$3"
  mkdir -p "$output_dir"
  go run ./cmd/partial-knowledge-composition-observer \
    --head-sha "$HEAD_SHA" --source-file "$input_source" --source-path "$source_path" \
    "${snapshot_args[@]}" --output "$output_dir/raw-evidence.json"
  go run ./cmd/partial-knowledge-composition-witness \
    --head-sha "$HEAD_SHA" --source-file "$input_source" --source-path "$source_path" \
    --evidence "$output_dir/raw-evidence.json" --intervention "$mode" --output "$output_dir/receipt.json"
  go run ./cmd/partial-knowledge-composition-verifier \
    --head-sha "$HEAD_SHA" --source-file "$input_source" --source-path "$source_path" \
    --evidence "$output_dir/raw-evidence.json" --intervention "$mode" \
    --receipt "$output_dir/receipt.json" --output "$output_dir/verification.json"
  jq -e '
    .status == "VERIFIED" and .decision == "CALCULUS_PROVEN" and .resolution == "CALCULUS" and
    .subject_resolution == "PARTIAL_KNOWLEDGE" and .evidence_coverage == "COMPLETE" and
    .independent_evaluator == true and .repository_writes == 0 and .promotion_authorized == false and
    .authority_state == "UNKNOWN" and .authority_resolution == "LOWER_RESOLUTION"
  ' "$output_dir/verification.json"
}

run_case none "$source_file" "$work/base"

{
  printf '%s\n' '// nonsemantic comment intervention'
  cat "$source_file"
} > "$work/comment-only/main.gooo"

sed '/case=direct-unknown/ s/left.observation_recipe=missing/left.observation_recipe=exact/' \
  "$source_file" > "$work/semantic/main.gooo"

run_case semantic "$work/semantic/main.gooo" "$work/semantic"
run_case comment-only "$work/comment-only/main.gooo" "$work/comment-only"

jq -e '
  .source_cases == 5 and .source_cases_total == 5 and
  .exact_cases == 1 and .direct_unknown_cases == 1 and
  .dependency_blocked_cases == 1 and .invariant_only_cases == 1 and
  .mixed_unresolved_cases == 1 and .top_success_cases == 1 and
  .open_claims == 4 and .claim_transition_total == 5 and
  .distinct_predicate_count == 5 and .predicate_denominator == 5 and
  .authority_state == "UNKNOWN" and .authority_resolution == "LOWER_RESOLUTION"
' "$work/base/receipt.json"

base_source_digest="$(jq -r '.source_digest' "$work/base/receipt.json")"
semantic_source_digest="$(jq -r '.source_digest' "$work/semantic/receipt.json")"
comment_source_digest="$(jq -r '.source_digest' "$work/comment-only/receipt.json")"
base_semantic_ir_digest="$(jq -r '.semantic_ir_digest' "$work/base/receipt.json")"
semantic_semantic_ir_digest="$(jq -r '.semantic_ir_digest' "$work/semantic/receipt.json")"
comment_semantic_ir_digest="$(jq -r '.semantic_ir_digest' "$work/comment-only/receipt.json")"
base_raw_evidence_digest="$(jq -r '.raw_evidence_digest' "$work/base/receipt.json")"
semantic_raw_evidence_digest="$(jq -r '.raw_evidence_digest' "$work/semantic/receipt.json")"
comment_raw_evidence_digest="$(jq -r '.raw_evidence_digest' "$work/comment-only/receipt.json")"
base_projection="$(jq -r '.semantic_projection_digest' "$work/base/receipt.json")"
semantic_projection="$(jq -r '.semantic_projection_digest' "$work/semantic/receipt.json")"
comment_projection="$(jq -r '.semantic_projection_digest' "$work/comment-only/receipt.json")"

test "$base_source_digest" != "$semantic_source_digest"
test "$base_semantic_ir_digest" != "$semantic_semantic_ir_digest"
test "$base_raw_evidence_digest" != "$semantic_raw_evidence_digest"
test "$base_projection" != "$semantic_projection"
jq -e '
  ([.cases[] | select(.id == "direct-unknown") | {decision, resolution}] == [{decision:"PASS", resolution:"EXACT"}]) and
  ([.claims[] | select(.proposition | contains("case=direct-unknown")) | {from, to}] == [{from:"OPEN", to:"DISCHARGED"}])
' "$work/semantic/receipt.json"
semantic_causality=1

test "$base_source_digest" != "$comment_source_digest"
test "$base_raw_evidence_digest" != "$comment_raw_evidence_digest"
test "$base_semantic_ir_digest" = "$comment_semantic_ir_digest"
test "$base_projection" = "$comment_projection"
base_comment_outcome="$(jq -c '[.cases[] | {id, decision, resolution}] + [.claims[] | {from, to, predicate, proposition_digest}]' "$work/base/receipt.json")"
comment_comment_outcome="$(jq -c '[.cases[] | {id, decision, resolution}] + [.claims[] | {from, to, predicate, proposition_digest}]' "$work/comment-only/receipt.json")"
test "$base_comment_outcome" = "$comment_comment_outcome"
nonsemantic_preservation=1

open_claims_preserved="$(jq '[.claims[] | select(.from == "OPEN" and .to == "OPEN")] | length' "$work/base/receipt.json")"
test "$open_claims_preserved" -eq 4
open_claims_preserved_total=4
source_cases="$(jq -r '.source_cases' "$work/base/receipt.json")"
source_cases_total="$(jq -r '.source_cases_total' "$work/base/receipt.json")"
distinct_predicates="$(jq -r '.summary.distinct_predicate_count' "$work/base/receipt.json")"
predicate_denominator="$(jq -r '.summary.predicate_denominator' "$work/base/receipt.json")"
repository_writes="$(jq -r '.repository_writes' "$work/base/receipt.json")"
promotion_authorized="$(jq -r '.promotion_authorized' "$work/base/receipt.json")"
authority_state="$(jq -r '.authority_state' "$work/base/receipt.json")"
authority_resolution="$(jq -r '.authority_resolution' "$work/base/receipt.json")"
test "$repository_writes" -eq 0
test "$authority_state" = UNKNOWN
test "$authority_resolution" = LOWER_RESOLUTION

semantic_direct_base="$(jq -c '[.cases[] | select(.id == "direct-unknown") | {decision, resolution}] + [.claims[] | select(.proposition | contains("case=direct-unknown")) | {from, to}]' "$work/base/receipt.json")"
semantic_direct_variant="$(jq -c '[.cases[] | select(.id == "direct-unknown") | {decision, resolution}] + [.claims[] | select(.proposition | contains("case=direct-unknown")) | {from, to}]' "$work/semantic/receipt.json")"
comment_direct_variant="$(jq -c '[.cases[] | select(.id == "direct-unknown") | {decision, resolution}] + [.claims[] | select(.proposition | contains("case=direct-unknown")) | {from, to}]' "$work/comment-only/receipt.json")"

jq -n \
  --arg schema 'gooo/meta/partial-knowledge-composition-ci/v2' \
  --arg head "$HEAD_SHA" \
  --arg base_receipt "$(jq -r '.digest' "$work/base/receipt.json")" \
  --arg semantic_receipt "$(jq -r '.digest' "$work/semantic/receipt.json")" \
  --arg comment_receipt "$(jq -r '.digest' "$work/comment-only/receipt.json")" \
  --arg base_source_digest "$base_source_digest" \
  --arg semantic_source_digest "$semantic_source_digest" \
  --arg comment_source_digest "$comment_source_digest" \
  --arg base_semantic_ir_digest "$base_semantic_ir_digest" \
  --arg semantic_semantic_ir_digest "$semantic_semantic_ir_digest" \
  --arg comment_semantic_ir_digest "$comment_semantic_ir_digest" \
  --arg base_raw_evidence_digest "$base_raw_evidence_digest" \
  --arg semantic_raw_evidence_digest "$semantic_raw_evidence_digest" \
  --arg comment_raw_evidence_digest "$comment_raw_evidence_digest" \
  --arg base_projection "$base_projection" \
  --arg semantic_projection "$semantic_projection" \
  --arg comment_projection "$comment_projection" \
  --arg semantic_direct_base "$semantic_direct_base" \
  --arg semantic_direct_variant "$semantic_direct_variant" \
  --arg comment_direct_variant "$comment_direct_variant" \
  --argjson imports_producer_package "$imports_producer_package" \
  --argjson imports_producer_package_count "$producer_imports_count" \
  --argjson imports_producer_package_total "$imports_producer_package_total" \
  --argjson source_cases "$source_cases" \
  --argjson source_cases_total "$source_cases_total" \
  --argjson distinct_predicates "$distinct_predicates" \
  --argjson predicate_denominator "$predicate_denominator" \
  --argjson semantic_causality "$semantic_causality" \
  --argjson nonsemantic_preservation "$nonsemantic_preservation" \
  --argjson open_claims_preserved "$open_claims_preserved" \
  --argjson open_claims_preserved_total "$open_claims_preserved_total" \
  --argjson repository_writes "$repository_writes" \
  --argjson promotion_authorized false \
  '{
    schema:$schema, subject_sha:$head,
    calculus:{decision:"CALCULUS_PROVEN", resolution:"CALCULUS"},
    subject:{resolution:"PARTIAL_KNOWLEDGE", source_cases:$source_cases, source_cases_total:$source_cases_total, distinct_predicates:$distinct_predicates, predicate_denominator:$predicate_denominator},
    evidence:{coverage:"COMPLETE", repository_writes:$repository_writes},
    authority:{state:"UNKNOWN", resolution:"LOWER_RESOLUTION", promotion_authorized:$promotion_authorized},
    independent_consumer:{imports_producer_package:$imports_producer_package, imports_producer_package_count:$imports_producer_package_count, imports_producer_package_total:$imports_producer_package_total},
    semantic_causality:{observed:$semantic_causality, total:1, base_source_digest:$base_source_digest, semantic_source_digest:$semantic_source_digest, base_semantic_ir_digest:$base_semantic_ir_digest, semantic_semantic_ir_digest:$semantic_semantic_ir_digest, base_raw_evidence_digest:$base_raw_evidence_digest, semantic_raw_evidence_digest:$semantic_raw_evidence_digest, base_projection:$base_projection, semantic_projection:$semantic_projection, base_direct:$semantic_direct_base, semantic_direct:$semantic_direct_variant},
    nonsemantic_preservation:{observed:$nonsemantic_preservation, total:1, base_source_digest:$base_source_digest, comment_source_digest:$comment_source_digest, base_semantic_ir_digest:$base_semantic_ir_digest, comment_semantic_ir_digest:$comment_semantic_ir_digest, base_raw_evidence_digest:$base_raw_evidence_digest, comment_raw_evidence_digest:$comment_raw_evidence_digest, base_projection:$base_projection, comment_projection:$comment_projection, base_direct:$semantic_direct_base, comment_direct:$comment_direct_variant},
    open_claims_preserved:$open_claims_preserved, open_claims_preserved_total:$open_claims_preserved_total,
    receipts:{base:$base_receipt, semantic_intervention:$semantic_receipt, comment_only:$comment_receipt}
  }' > "$work/report.json"

cat > "$work/summary.md" <<EOF
# Partial knowledge composition — exact-head Action evidence

- subject: $HEAD_SHA
- calculus conformance: CALCULUS_PROVEN / CALCULUS
- subject resolution: PARTIAL_KNOWLEDGE
- source-derived cases: $source_cases/$source_cases_total
- distinct predicates: $distinct_predicates/$predicate_denominator
- independent consumer import predicate: $imports_producer_package/$imports_producer_package_total (actual imports: $producer_imports_count)
- semantic intervention causality: $semantic_causality/1
- nonsemantic comment preservation: $nonsemantic_preservation/1
- open claims preserved: $open_claims_preserved/$open_claims_preserved_total
- repository writes observed: $repository_writes; promotion authority: $authority_state / $authority_resolution; promotion authorized: $promotion_authorized

## Subject outcomes

| Case | Decision | Resolution | Claim transition |
|---|---|---|---|
| exact-pair | PASS | EXACT | OPEN -> DISCHARGED |
| direct-unknown | UNKNOWN | LOWER_RESOLUTION | OPEN -> OPEN |
| dependency-blocked | UNKNOWN | LOWER_RESOLUTION | OPEN -> OPEN |
| invariant-preservation | HOLD | INVARIANT_ONLY | OPEN -> OPEN |
| mixed-unknown-and-blocked | UNKNOWN | LOWER_RESOLUTION | OPEN -> OPEN |

## A/B digest and transition evidence

| Variant | Raw source digest | Semantic IR digest | Raw evidence digest | Semantic projection |
|---|---|---|---|---|
| base | $base_source_digest | $base_semantic_ir_digest | $base_raw_evidence_digest | $base_projection |
| semantic source | $semantic_source_digest | $semantic_semantic_ir_digest | $semantic_raw_evidence_digest | $semantic_projection |
| comment-only source | $comment_source_digest | $comment_semantic_ir_digest | $comment_raw_evidence_digest | $comment_projection |

Semantic direct case: $semantic_direct_base -> $semantic_direct_variant.
Comment-only direct case: $semantic_direct_base -> $comment_direct_variant.
Raw provenance changes are kept out of the semantic projection; they remain visible in the raw source/evidence digests.
EOF

cp "$source_file" "$work/source.gooo"
find "$work" -type f \( -name '*.json' -o -name '*.gooo' -o -name '*.txt' \) -print0 | sort -z | while IFS= read -r -d '' file; do sha256sum "$file"; done > "$work/manifest.sha256"
cat "$work/summary.md" >> "$GITHUB_STEP_SUMMARY"
