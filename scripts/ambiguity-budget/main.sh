#!/usr/bin/env bash
set -euo pipefail

root="${GITHUB_WORKSPACE:-$(pwd)}"
output="${RUNNER_TEMP:-/tmp}/ambiguity-budget"
mkdir -p "$output"
head_sha="${HEAD_SHA:-$(git -C "$root" rev-parse HEAD)}"
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

snapshot_digest() {
  printf 'sha256:%s' "$(sha256sum "$1" | awk '{print $1}')"
}

make_effects_artifact() {
  local before="$1"
  local after="$2"
  local writes="$3"
  local equal="$4"
  jq -n \
    --arg before "$(snapshot_digest "$before")" \
    --arg after "$(snapshot_digest "$after")" \
    --argjson repository_writes "$writes" \
    --argjson write_set_equal "$equal" \
    '{schema:"gooo/ambiguity-budget-workspace-effects/v1",version:"v1",tracked_and_untracked:true,snapshot_before:$before,snapshot_after:$after,repository_writes:$repository_writes,write_set_equal:$write_set_equal,mutation_authority:"UNKNOWN",mutation_authority_resolution:"NOT_OBSERVED"}'
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

# The first pass is a probe. Its effects artifact is not evidence: the final
# artifact below is rebuilt from the actual tracked+untracked snapshots.
jq -n \
  --arg before "$(snapshot_digest "$output/workspace-before.snapshot")" \
  '{schema:"gooo/ambiguity-budget-workspace-effects/v1",version:"v1",tracked_and_untracked:true,snapshot_before:$before,snapshot_after:$before,repository_writes:0,write_set_equal:true,mutation_authority:"UNKNOWN",mutation_authority_resolution:"NOT_OBSERVED"}' \
  > "$output/workspace-effects-probe.json"
go run ./cmd/ambiguity-budget-witness \
  -head "$head_sha" -contract "$contract" -source "$source" -effects "$output/workspace-effects-probe.json" \
  -output "$output/probe-receipt.json"

workspace_snapshot "$root" "$output/workspace-after-probe.snapshot"
probe_writes=0
probe_equal=true
if ! cmp -s "$output/workspace-before.snapshot" "$output/workspace-after-probe.snapshot"; then
  probe_writes=1
  probe_equal=false
fi
if [[ "$probe_writes" -ne 0 ]]; then
  echo "probe changed the tracked+untracked repository write-set" >&2
  exit 1
fi

make_effects_artifact "$output/workspace-before.snapshot" "$output/workspace-after-probe.snapshot" "$probe_writes" "$probe_equal" \
  > "$output/workspace-effects.json"
jq -e '.schema == "gooo/ambiguity-budget-workspace-effects/v1" and .version == "v1" and .tracked_and_untracked == true and .repository_writes == 0 and .write_set_equal == true and .mutation_authority == "UNKNOWN" and .mutation_authority_resolution == "NOT_OBSERVED"' \
  "$output/workspace-effects.json"

for suffix in first second; do
  go run ./cmd/ambiguity-budget-witness \
    -head "$head_sha" -contract "$contract" -source "$source" -effects "$output/workspace-effects.json" \
    -output "$output/receipt-$suffix.json"
done
cmp "$output/receipt-first.json" "$output/receipt-second.json"

for suffix in first second; do
  go run ./cmd/ambiguity-budget-verifier \
    -contract "$contract" -receipt "$output/receipt-first.json" -source "$source" -effects "$output/workspace-effects.json" \
    -output "$output/judge-$suffix.json"
done
cmp "$output/judge-first.json" "$output/judge-second.json"

jq -e '
  .conformance_decision == "PASS" and .conformance_resolution == "EXACT" and
  .subject_decision == "MIXED" and .subject_resolution == "LOWER_RESOLUTION" and
  .budget_authority == "CONTRACT_POLICY" and .budget_binding == "ambiguity-budget:budget-policy:v1" and
  .summary.denominator.schema == "gooo/ambiguity-budget-denominator/v1" and
  .summary.denominator.version == "v1" and .summary.denominator.cases == 4 and
  .summary.denominator.integer_observations == 12 and .summary.denominator.claims == 4 and
  .summary.denominator.interventions == 2 and .summary.denominator.authority_observations == 1 and
  .summary.numerator.cases_conforming == 4 and .summary.numerator.integer_observations_observed == 11 and
  .summary.numerator.claims_discharged == 2 and .summary.numerator.claims_refuted == 1 and
  .summary.numerator.claims_open == 1 and .summary.numerator.interventions_satisfied == 2 and
  .summary.numerator.authority_observed == 0 and
  .effects.repository_writes == 0 and .effects.write_set_equal == true and
  .effects.mutation_authority == "UNKNOWN" and .effects.mutation_authority_resolution == "NOT_OBSERVED" and
  (.effects.artifact_digest | startswith("sha256:")) and
  (.producer | length) > 0 and (.consumer | length) > 0 and
  (.meta_operation | length) > 0 and (.proof_choice | length) > 0
' "$output/receipt-first.json"

jq -e '
  ([.cases[] | select(.id == "zero-ambiguity" and .counts == {interpretation_candidates:1,unresolved_branches:0,evidence_paths:1} and .class == "ZERO" and .input_state == "KNOWN" and .decision == "PASS" and .resolution == "EXACT" and .claim.from == "OPEN" and .claim.to == "DISCHARGED" and (.claim.proposition | startswith("counts-within-budget(case:zero-ambiguity")))] | length) == 1 and
  ([.cases[] | select(.id == "boundary-ambiguity" and .counts == {interpretation_candidates:2,unresolved_branches:1,evidence_paths:2} and .class == "BOUNDARY" and .input_state == "KNOWN" and .decision == "PASS" and .resolution == "EXACT" and .claim.from == "OPEN" and .claim.to == "DISCHARGED")] | length) == 1 and
  ([.cases[] | select(.id == "over-budget-ambiguity" and .counts == {interpretation_candidates:3,unresolved_branches:2,evidence_paths:3} and .class == "OVER" and .input_state == "KNOWN" and .decision == "FAIL_CLOSED" and .resolution == "LOWER_RESOLUTION" and .reason == "AMBIGUITY_BUDGET_EXCEEDED" and .claim.from == "OPEN" and .claim.to == "REFUTED")] | length) == 1 and
  ([.cases[] | select(.id == "unknown-ambiguity" and .counts == {interpretation_candidates:2,unresolved_branches:0,evidence_paths:2} and .class == "UNKNOWN" and .input_state == "UNKNOWN" and .decision == "UNKNOWN" and .resolution == "LOWER_RESOLUTION" and .coordinate.stage == "AMBIGUITY_OBSERVATION" and .coordinate.step == "unresolved_branches" and .coordinate.reason == "AMBIGUITY_COORDINATE_UNOBSERVED" and .unobserved_dimensions == ["unresolved_branches"] and .claim.from == "OPEN" and .claim.to == "OPEN")] | length) == 1 and
  ([.claims[] | select(.from == "OPEN" and (.proposition | startswith("counts-within-budget(case:")) and (.proposition_digest | startswith("sha256:")) and (.stage | length) > 0 and (.step | length) > 0 and (.reason | length) > 0 and (.evidence_digest | startswith("sha256:")))] | length) == 4 and
  ([.claims[] | select(.case_id == "over-budget-ambiguity" and .to == "REFUTED" and .reason == "AMBIGUITY_BUDGET_EXCEEDED")] | length) == 1 and
  ([.claims[] | select(.case_id == "unknown-ambiguity" and .to == "OPEN" and .stage == "AMBIGUITY_OBSERVATION" and .step == "unresolved_branches" and .reason == "AMBIGUITY_COORDINATE_UNOBSERVED")] | length) == 1 and
  ([.interventions[] | select(.id == "semantic-count-crosses-boundary" and .kind == "SEMANTIC" and .satisfied == true and .counts_before != .counts_after and .semantic_digest_before != .semantic_digest_after and (.elements_before.candidate_ids | length) == 2 and (.elements_after.candidate_ids | length) == 3 and (.elements_before.evidence_path_ids | length) == 2 and (.elements_after.evidence_path_ids | length) == 3 and .resolution_after == "LOWER_RESOLUTION" and .claim_before.from == "OPEN" and .claim_before.to == "DISCHARGED" and .claim_after.from == "OPEN" and .claim_after.to == "REFUTED")] | length) == 1 and
  ([.interventions[] | select(.id == "nonsemantic-comment-only" and .kind == "NONSEMANTIC" and .satisfied == true and .source_digest_before != .source_digest_after and .semantic_digest_before == .semantic_digest_after and .counts_before == .counts_after and .class_before == .class_after and .input_state_before == .input_state_after and .decision_before == .decision_after and .resolution_before == .resolution_after and .claim_before == .claim_after)] | length) == 1 and
  ([.cases[].claim | select(.from != "OPEN")] | length) == 0
' "$output/receipt-first.json"

jq -e '
  .conformance_decision == "PASS" and .conformance_resolution == "EXACT" and
  .subject_decision == "MIXED" and .subject_resolution == "LOWER_RESOLUTION" and
  .denominator.schema == "gooo/ambiguity-budget-denominator/v1" and .denominator.version == "v1" and
  .denominator.cases == 4 and .denominator.integer_observations == 12 and .denominator.claims == 4 and
  .denominator.interventions == 2 and .denominator.authority_observations == 1 and
  .numerator.cases_conforming == 4 and .numerator.integer_observations_observed == 11 and
  .numerator.claims_discharged == 2 and .numerator.claims_refuted == 1 and .numerator.claims_open == 1 and
  .numerator.interventions_satisfied == 2 and .numerator.authority_observed == 0 and
  .forbidden_producer_imports == 0 and .allowed_producer_imports == 0 and
  .repository_writes == 0 and .mutation_authority == "UNKNOWN" and .mutation_authority_resolution == "NOT_OBSERVED" and
  (.effects_artifact_digest | startswith("sha256:"))
' "$output/judge-first.json"

workspace_snapshot "$root" "$output/workspace-after.snapshot"
repository_writes=0
write_set_equal=true
if ! cmp -s "$output/workspace-before.snapshot" "$output/workspace-after.snapshot"; then
  repository_writes=1
  write_set_equal=false
fi
if [[ "$repository_writes" -ne 0 ]]; then
  echo "final tracked+untracked repository write-set changed" >&2
  exit 1
fi
cmp "$output/workspace-before.snapshot" "$output/workspace-after.snapshot"
echo "ambiguity budget: deterministic graph accounting, provenance separation, boundary intervention, independent verification, and zero writes passed"
