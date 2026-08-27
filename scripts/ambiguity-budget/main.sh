#!/usr/bin/env bash
set -euo pipefail

root="${GITHUB_WORKSPACE:-$(pwd)}"
output="${RUNNER_TEMP:-/tmp}/ambiguity-budget"
mkdir -p "$output"
head_sha="${HEAD_SHA:-$(git rev-parse HEAD)}"
contract="$root/examples/ambiguity-budget/contract.json"
source="$root/examples/ambiguity-budget/main.gooo"

workspace_snapshot() {
  local repo="$1"
  local target="$2"
  {
    git -C "$repo" status --porcelain=v1 --untracked-files=all
    while IFS= read -r -d '' relative; do
      if [[ -f "$repo/$relative" ]]; then
        printf 'tracked\t%s\t%s\n' "$relative" "$(sha256sum "$repo/$relative" | awk '{print $1}')"
      else
        printf 'tracked-missing\t%s\n' "$relative"
      fi
    done < <(git -C "$repo" ls-files -z)
    while IFS= read -r -d '' relative; do
      printf 'untracked\t%s\t%s\n' "$relative" "$(sha256sum "$repo/$relative" | awk '{print $1}')"
    done < <(git -C "$repo" ls-files --others --exclude-standard -z)
  } | LC_ALL=C sort > "$target"
}

workspace_snapshot "$root" "$output/workspace-before.snapshot"

forbidden_import='github.com/kimjooyoon/meta-ontology-go/internal/meta/ambiguitybudget"'
forbidden_imports="$(git -C "$root" grep -n -F "$forbidden_import" -- internal/meta/ambiguitybudgetjudge cmd/ambiguity-budget-verifier || true)"
if [[ -n "$forbidden_imports" ]]; then
  echo "forbidden producer imports detected:" >&2
  echo "$forbidden_imports" >&2
  exit 1
fi
producer_package='github.com/kimjooyoon/meta-ontology-go/internal/meta/ambiguitybudget'
allowed_producer_imports=0
while IFS= read -r package; do
  if go list -deps "$package" | grep -Fxq "$producer_package"; then
    allowed_producer_imports=$((allowed_producer_imports + 1))
  fi
done <<'EOF'
./internal/meta/ambiguitybudgetjudge
./cmd/ambiguity-budget-verifier
EOF
if [[ "$allowed_producer_imports" -ne 0 ]]; then
  echo "producer dependency detected in independent consumer graph" >&2
  exit 1
fi
for suffix in first second; do
  go run ./cmd/ambiguity-budget-witness \
    -head "$head_sha" -contract "$contract" -source "$source" \
    -output "$output/receipt-$suffix.json"
done
cmp "$output/receipt-first.json" "$output/receipt-second.json"

for suffix in first second; do
  go run ./cmd/ambiguity-budget-verifier \
    -contract "$contract" -receipt "$output/receipt-first.json" -source "$source" \
    -output "$output/judge-$suffix.json"
done
cmp "$output/judge-first.json" "$output/judge-second.json"

jq -e '
  .conformance_decision == "PASS" and .conformance_resolution == "EXACT" and
  .subject_decision == "MIXED" and .subject_resolution == "LOWER_RESOLUTION" and
  .summary.cases_total == 4 and .summary.known_cases == 3 and
  .summary.fixed_denominator == 2 and .summary.integer_dimensions == 3 and
  .summary.interventions_total == 2 and .summary.open_claims == 1 and
  .summary.zero_ambiguity_cases == 1 and
  .summary.boundary_cases == 1 and .summary.over_budget_cases == 1 and
  .summary.unknown_cases == 1 and .summary.lower_resolution_cases == 2 and
  .effects.repository_writes == 0 and .effects.mutation_authority == false and
  (.producer | length) > 0 and (.consumer | length) > 0 and
  (.meta_operation | length) > 0 and (.proof_choice | length) > 0
' "$output/receipt-first.json"
jq -e '
  ([.cases[] | select(.id == "zero-ambiguity" and .class == "ZERO" and .input_state == "KNOWN" and .decision == "PASS" and .resolution == "EXACT" and .claim.from == "OPEN" and .claim.to == "DISCHARGED")] | length) == 1 and
  ([.cases[] | select(.id == "boundary-ambiguity" and .class == "BOUNDARY" and .input_state == "KNOWN" and .decision == "PASS" and .resolution == "EXACT" and .claim.from == "OPEN" and .claim.to == "DISCHARGED")] | length) == 1 and
  ([.cases[] | select(.id == "over-budget-ambiguity" and .class == "OVER" and .input_state == "KNOWN" and .decision == "FAIL_CLOSED" and .resolution == "LOWER_RESOLUTION" and .reason == "AMBIGUITY_BUDGET_EXCEEDED" and .claim.from == "OPEN" and .claim.to == "REFUTED")] | length) == 1 and
  ([.cases[] | select(.id == "unknown-ambiguity" and .class == "UNKNOWN" and .input_state == "UNKNOWN" and .decision == "UNKNOWN" and .resolution == "LOWER_RESOLUTION" and .coordinate.stage == "AMBIGUITY_OBSERVATION" and .coordinate.step == "unresolved_branches" and .coordinate.reason == "AMBIGUITY_COORDINATE_UNOBSERVED" and .unobserved_dimensions == ["unresolved_branches"] and .claim.from == "OPEN" and .claim.to == "OPEN")] | length) == 1 and
  ([.claims[] | select(.from == "OPEN" and .case_id == "zero-ambiguity" and .to == "DISCHARGED" and (.stage | length) > 0 and (.step | length) > 0 and (.reason | length) > 0 and (.evidence_digest | length) > 0)] | length) == 1 and
  ([.claims[] | select(.from == "OPEN" and .case_id == "boundary-ambiguity" and .to == "DISCHARGED" and (.stage | length) > 0 and (.step | length) > 0 and (.reason | length) > 0 and (.evidence_digest | length) > 0)] | length) == 1 and
  ([.claims[] | select(.from == "OPEN" and .case_id == "over-budget-ambiguity" and .to == "REFUTED" and (.stage | length) > 0 and (.step | length) > 0 and (.reason | length) > 0 and (.evidence_digest | length) > 0)] | length) == 1 and
  ([.claims[] | select(.from == "OPEN" and .case_id == "unknown-ambiguity" and .to == "OPEN" and (.stage | length) > 0 and (.step | length) > 0 and (.reason | length) > 0 and (.evidence_digest | length) > 0)] | length) == 1 and
  ([.claims[] | select((.from == "OPEN") and (.stage | length) > 0 and (.step | length) > 0 and (.reason | length) > 0 and (.evidence_digest | length) > 0)] | length) == 4 and
  ([.cases[] | select(.id == "unknown-ambiguity" and .unobserved_dimensions == ["unresolved_branches"] and .coordinate.stage == "AMBIGUITY_OBSERVATION" and .coordinate.step == "unresolved_branches" and .coordinate.reason == "AMBIGUITY_COORDINATE_UNOBSERVED")] | length) == 1 and
  ([.interventions[] | select(.id == "semantic-count-crosses-boundary" and .kind == "SEMANTIC" and .satisfied == true and .counts_before != .counts_after and .semantic_digest_before != .semantic_digest_after and .resolution_after == "LOWER_RESOLUTION" and .claim_before.from == "OPEN" and .claim_before.to == "DISCHARGED" and .claim_after.from == "OPEN" and .claim_after.to == "REFUTED")] | length) == 1 and
  ([.interventions[] | select(.id == "nonsemantic-comment-only" and .kind == "NONSEMANTIC" and .satisfied == true and .source_digest_before != .source_digest_after and .semantic_digest_before == .semantic_digest_after and .counts_before == .counts_after and .class_before == .class_after and .input_state_before == .input_state_after and .decision_before == .decision_after and .resolution_before == .resolution_after and .claim_before == .claim_after)] | length) == 1
' "$output/receipt-first.json"
jq -e '
  .conformance_decision == "PASS" and .conformance_resolution == "EXACT" and
  .subject_decision == "MIXED" and .subject_resolution == "LOWER_RESOLUTION" and
  .fixed_denominator == 2 and .cases_total == 4 and .interventions_total == 2 and
  .forbidden_producer_imports == 0 and .allowed_producer_imports == 0 and
  .repository_writes == 0 and .mutation_authority == false
' "$output/judge-first.json"

workspace_snapshot "$root" "$output/workspace-after.snapshot"
repository_writes=0
write_set_equal=true
if ! cmp -s "$output/workspace-before.snapshot" "$output/workspace-after.snapshot"; then
  repository_writes=1
  write_set_equal=false
fi
before_snapshot_digest="sha256:$(sha256sum "$output/workspace-before.snapshot" | awk '{print $1}')"
after_snapshot_digest="sha256:$(sha256sum "$output/workspace-after.snapshot" | awk '{print $1}')"
jq -n \
  --arg before "$before_snapshot_digest" \
  --arg after "$after_snapshot_digest" \
  --argjson repository_writes "$repository_writes" \
  --argjson write_set_equal "$write_set_equal" \
  '{schema:"gooo/ambiguity-budget-workspace-effects/v1", tracked_and_untracked:true, snapshot_before:$before, snapshot_after:$after, repository_writes:$repository_writes, write_set_equal:$write_set_equal}' \
  > "$output/workspace-effects.json"
if [[ "$repository_writes" -ne 0 ]]; then
  echo "repository write-set changed between tracked+untracked snapshots" >&2
  exit 1
fi
test "$allowed_producer_imports" -eq 0
jq -e '.tracked_and_untracked == true and .repository_writes == 0 and .write_set_equal == true' "$output/workspace-effects.json"

cp "$output/receipt-first.json" "$output/receipt.json"
cp "$output/judge-first.json" "$output/judge.json"
echo "ambiguity budget: deterministic replay, boundary descent, and independent verification passed"
