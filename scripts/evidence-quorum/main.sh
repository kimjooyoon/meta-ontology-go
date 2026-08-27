#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

output="${QUORUM_OUTPUT:-evidence-quorum-output}"
source="examples/billing/main.gooo"
contract="examples/evidence-quorum/contract.json"
mkdir -p "$output/cases"

test -z "$(gofmt -l internal/meta/evidencequorum cmd/evidence-quorum-witness)"
go test ./internal/meta/evidencequorum ./cmd/evidence-quorum-witness

forbidden="$(go list -deps ./internal/meta/evidencequorum | grep -E '/internal/(sourceexecution|meta/language(sourceexecution|artifactoracle|sourcebindingpromotion))$' || true)"
test -z "$forbidden"
printf '%s\n' "$forbidden" > "$output/evaluator-forbidden-dependencies.txt"

emit() {
  local directory="$1" role="$2" group="$3" evidence="$4" value="$5" confidence="$6"
  mkdir -p "$output/cases/$directory"
  go run ./cmd/evidence-quorum-witness --mode emit --contract "$contract" --source "$source" \
    --head "$HEAD_SHA" --source-path "$source" --role "$role" --origin-group "$group" \
    --evidence-id "$evidence" --value "$value" --confidence-bps "$confidence" \
    --out "$output/cases/$directory/$evidence.json"
}

emit sufficient-independent producer producer-run producer-1 SUPPORTS 9100
emit sufficient-independent consumer consumer-check consumer-1 SUPPORTS 8800
emit sufficient-independent meta-operation quorum-meta meta-1 SUPPORTS 7600

emit same-origin-replica producer producer-run producer-1 SUPPORTS 10000
emit same-origin-replica producer producer-run producer-replica SUPPORTS 10000
emit same-origin-replica consumer consumer-check consumer-1 SUPPORTS 10000

emit conflicting-independent producer producer-run producer-1 SUPPORTS 9900
emit conflicting-independent consumer consumer-check consumer-1 SUPPORTS 9900
emit conflicting-independent meta-operation quorum-meta meta-1 SUPPORTS 9900
emit conflicting-independent consumer contradictory-check contradictory-1 CONTRADICTS 100

emit insufficient-independent producer producer-run producer-1 SUPPORTS 10000
emit insufficient-independent consumer consumer-check consumer-1 SUPPORTS 10000

receipts_for() {
  local directory="$1"
  find "$output/cases/$directory" -type f -name '*.json' | sort | paste -sd, -
}

spec="$(receipts_for sufficient-independent);$(receipts_for same-origin-replica);$(receipts_for conflicting-independent);$(receipts_for insufficient-independent)"
go run ./cmd/evidence-quorum-witness --mode evaluate --contract "$contract" --source "$source" \
  --head "$HEAD_SHA" --source-path "$source" --receipts "$spec" --out "$output/report.json"
go run ./cmd/evidence-quorum-witness --check "$output/report.json"

jq -e '.decision == "PASS" and .resolution == "EXACT" and
  .summary.cases_satisfied == 4 and .summary.cases_total == 4 and
  .summary.raw_evidence_total == 12 and .summary.independent_groups_total == 11 and
  .summary.duplicate_evidence_total == 1 and .summary.conflict_cases == 1 and
  .summary.quorum_satisfied_cases == 1 and .summary.lower_resolution_cases == 2 and
  .summary.minimum_independent_groups == 3 and .summary.confidence_aggregated == false and
  .summary.repository_writes == 0 and .summary.mutation_authority == false' "$output/report.json"
jq -e '([.cases[] | select(.status == "SATISFIED")] | length) == 4 and
  ([.cases[] | select(.id == "same-origin-replica") | .independent_groups] | .[0]) == 2 and
  ([.cases[] | select(.id == "same-origin-replica") | .duplicate_evidence] | .[0]) == 1 and
  ([.cases[] | select(.id == "conflicting-independent") | .observed_reason] | .[0]) == "QUORUM_CONFLICT" and
  ([.cases[] | select(.id == "insufficient-independent") | .observed_resolution] | .[0]) == "LOWER_RESOLUTION" and
  ([.cases[].claims[] | .transitions[]] | length) == 4' "$output/report.json"
sha256sum "$output"/report.json "$output"/cases/*/*.json > "$output/manifest.sha256"
