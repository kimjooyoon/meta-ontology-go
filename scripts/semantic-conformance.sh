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
jq -e '.decision == "DISCHARGED" and .conformance_decision == "DISCHARGED" and .conformance_resolution == "EXACT" and .subject_decision == "OPEN" and .subject_resolution == "LOWER_RESOLUTION" and .summary.case_matrix.numerator == 4 and .summary.case_matrix.denominator == 4 and .summary.byte_claim.numerator == 2 and .summary.meaning_claim.numerator == 2 and .summary.joint_claim.numerator == 1 and .summary.counterexamples.numerator == 2 and .summary.open_cases.numerator == 1 and .summary.source_digest_binding.numerator == 4 and .summary.semantic_causality.numerator == 4' "$repro_work/judgment.json"

consumer_sources=(internal/meta/reproducibilitysemanticsconsumer/*.go)
if rg -n 'Produce\(|internal/meta/reproducibilitysemantics("|/)' "${consumer_sources[@]}" >/dev/null; then
  echo "production consumer producer-dependency guardrail failed" >&2
  exit 1
fi
echo "production consumer producer imports: 0/0"

sed 's/meaning.observed=meaning\/render-approved\/v1/meaning.observed=meaning\/charge-and-ledger\/v1/' \
  examples/reproducibility-semantics/main.gooo > "$repro_work/semantic-intervention.gooo"
sed '1i // presentation-only comment' examples/reproducibility-semantics/main.gooo > "$repro_work/presentation-intervention.gooo"
go run ./scripts/reproducibility-semantics \
  -mode intervention \
  -source examples/reproducibility-semantics/main.gooo \
  -semantic-source "$repro_work/semantic-intervention.gooo" \
  -presentation-source "$repro_work/presentation-intervention.gooo" \
  -head-sha "$head_sha" \
  -output "$repro_work/intervention.json"
jq -e '(.schema == "gooo/reproducibility-semantics-intervention/v1") and (.denominator == 2) and ((.cases | length) == 2) and (.decision == "DISCHARGED") and (.resolution == "EXACT") and (.authority.repository_writes == 0) and (.authority.mutation_authorized == false) and (.authority.promotion_authorized == false) and (.cases[0].id == "semantic-source-change") and (.cases[0].kind == "SEMANTIC_SOURCE_CHANGE") and (.cases[0].source_digest_before != .cases[0].source_digest_after) and (.cases[0].semantic_digest_before != .cases[0].semantic_digest_after) and (.cases[0].meaning_before != .cases[0].meaning_after) and (.cases[0].joint_before != .cases[0].joint_after) and (.cases[0].transitions_before != .cases[0].transitions_after) and (.cases[1].id == "presentation-only-source-change") and (.cases[1].kind == "PRESENTATION_ONLY_SOURCE_CHANGE") and (.cases[1].source_digest_before != .cases[1].source_digest_after) and (.cases[1].semantic_digest_before == .cases[1].semantic_digest_after) and (.cases[1].meaning_before == .cases[1].meaning_after) and (.cases[1].joint_before == .cases[1].joint_after) and (.cases[1].transitions_before == .cases[1].transitions_after) and (has("score") | not) and (has("aggregate_score") | not)' "$repro_work/intervention.json"
echo "semantic causality intervention artifact: fixed-denominator=2/2; semantic and presentation cases separate"
