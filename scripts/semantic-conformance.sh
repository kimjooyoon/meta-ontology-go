#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

conformance_stage="${GOOO_CONFORMANCE_STAGE:-0}"
case "$conformance_stage" in
  0)
    echo "semantic conformance stage 0: Go verifier authoritative"
    ;;
  1|2|3)
    echo "semantic conformance stage $conformance_stage is not enabled; see .github/conformance-plan.md" >&2
    exit 1
    ;;
  *)
    echo "invalid GOOO_CONFORMANCE_STAGE: $conformance_stage" >&2
    exit 1
    ;;
esac

# Stage 0 keeps the deterministic Go verifier as the baseline while gooo is
# bootstrapped. The policy job adds the full changed-path scope check.
go run ./scripts/verify

if [[ ! -f cmd/gooo/main.go ]]; then
  echo "semantic conformance deferred: cmd/gooo is not present in this baseline"
  exit 0
fi

if ! grep -Fq 'case "check":' cmd/gooo/main.go; then
  echo "semantic conformance deferred: cmd/gooo check is not implemented in this baseline"
  exit 0
fi

go run ./cmd/gooo check examples/billing/main.gooo
go test -tags semantic_conformance ./internal/verify -run 'TestSemantic|TestGenerated' -count=1
./scripts/generated-freshness.sh

repro_work="${RUNNER_TEMP:-${TMPDIR:-/tmp}}/gooo-reproducibility-semantics"
mkdir -p "$repro_work"
head_sha="$(git rev-parse HEAD)"
go run ./cmd/gooo check examples/reproducibility-semantics/main.gooo
go run ./scripts/reproducibility-semantics \
  -mode produce \
  -source examples/reproducibility-semantics/main.gooo \
  -head-sha "$head_sha" \
  -output "$repro_work/receipt.json"
go run ./scripts/reproducibility-semantics \
  -mode produce \
  -source examples/reproducibility-semantics/main.gooo \
  -head-sha "$head_sha" \
  -output "$repro_work/receipt-replay.json"
cmp -s "$repro_work/receipt.json" "$repro_work/receipt-replay.json"
go run ./scripts/reproducibility-semantics \
  -mode judge \
  -source examples/reproducibility-semantics/main.gooo \
  -head-sha "$head_sha" \
  -receipt "$repro_work/receipt.json" \
  -output "$repro_work/judgment.json" \
  -check
go run ./scripts/reproducibility-semantics \
  -mode judge \
  -source examples/reproducibility-semantics/main.gooo \
  -head-sha "$head_sha" \
  -receipt "$repro_work/receipt.json" \
  -output "$repro_work/judgment-replay.json" \
  -check
cmp -s "$repro_work/judgment.json" "$repro_work/judgment-replay.json"
jq -e '.decision == "DISCHARGED" and .summary.case_matrix.numerator == 4 and .summary.case_matrix.denominator == 4 and .summary.byte_claim.numerator == 2 and .summary.meaning_claim.numerator == 2 and .summary.joint_claim.numerator == 1 and .summary.counterexamples.numerator == 2 and .summary.open_cases.numerator == 1 and .summary.source_digest_binding.numerator == 4 and .summary.semantic_causality.numerator == 4' "$repro_work/judgment.json"

judge_sources=(internal/meta/reproducibilitysemantics/judge.go
  internal/meta/reproducibilitysemantics/judge_source.go
  internal/meta/reproducibilitysemantics/judge_proofs.go
  internal/meta/reproducibilitysemantics/judge_provenance.go
  internal/meta/reproducibilitysemantics/judge_receipt.go
  internal/meta/reproducibilitysemantics/judge_summary.go
  internal/meta/reproducibilitysemantics/judge_validate.go)
if rg -n 'Produce\(|internal/meta/reproducibilitysemantics' "${judge_sources[@]}" >/dev/null; then
  echo "independent judge dependency guardrail failed" >&2
  exit 1
fi
echo "independent judge dependency guardrail: 1/1"

sed 's/meaning.observed=meaning\/render-approved\/v1/meaning.observed=meaning\/charge-and-ledger\/v1/' \
  examples/reproducibility-semantics/main.gooo > "$repro_work/semantic-intervention.gooo"
sed '1i // presentation-only comment' examples/reproducibility-semantics/main.gooo > "$repro_work/presentation-intervention.gooo"
for variant in semantic-intervention presentation-intervention; do
  go run ./scripts/reproducibility-semantics \
    -mode produce \
    -source "$repro_work/$variant.gooo" \
    -head-sha "$head_sha" \
    -output "$repro_work/$variant-receipt.json"
  go run ./scripts/reproducibility-semantics \
    -mode judge \
    -source "$repro_work/$variant.gooo" \
    -head-sha "$head_sha" \
    -receipt "$repro_work/$variant-receipt.json" \
    -output "$repro_work/$variant-judgment.json" \
    -check
done
base_meaning="$(jq -c '.summary.meaning_claim' "$repro_work/judgment.json")"
semantic_meaning="$(jq -c '.summary.meaning_claim' "$repro_work/semantic-intervention-judgment.json")"
base_joint="$(jq -c '.summary.joint_claim' "$repro_work/judgment.json")"
semantic_joint="$(jq -c '.summary.joint_claim' "$repro_work/semantic-intervention-judgment.json")"
[[ "$base_meaning" != "$semantic_meaning" && "$base_joint" != "$semantic_joint" ]]
[[ "$(jq -r .source_digest "$repro_work/judgment.json")" != "$(jq -r .source_digest "$repro_work/semantic-intervention-judgment.json")" ]]
[[ "$(jq -r .semantic_digest "$repro_work/judgment.json")" != "$(jq -r .semantic_digest "$repro_work/semantic-intervention-judgment.json")" ]]
[[ "$(jq -r .source_digest "$repro_work/judgment.json")" != "$(jq -r .source_digest "$repro_work/presentation-intervention-judgment.json")" ]]
[[ "$(jq -r .semantic_digest "$repro_work/judgment.json")" == "$(jq -r .semantic_digest "$repro_work/presentation-intervention-judgment.json")" ]]
[[ "$(jq -c .summary "$repro_work/judgment.json")" == "$(jq -c .summary "$repro_work/presentation-intervention-judgment.json")" ]]
echo "semantic causality intervention contract: 2/2"
