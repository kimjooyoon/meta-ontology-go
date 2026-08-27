#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

output="${RUNNER_TEMP:-/tmp}/counterexample-first-output-${HEAD_SHA}"
compiler=./cmd/counterexample-first-compiler-witness
judge=./cmd/counterexample-first-judge-witness
intervention=./cmd/counterexample-first-intervention-witness
contract=examples/counterexample-first-compiler/contract.json
source=examples/counterexample-first-compiler/main.gooo
corpus=examples/counterexample-first-compiler/scenarios.json
semantic_intervention=examples/counterexample-first-compiler/semantic-intervention.gooo
comment_intervention=examples/counterexample-first-compiler/comment-only-intervention.gooo
meta_intervention=examples/counterexample-first-compiler/meta-operation-intervention.gooo
paths=(cmd/counterexample-first-compiler-witness cmd/counterexample-first-judge-witness cmd/counterexample-first-intervention-witness internal/meta/counterexamplefirst internal/meta/counterexamplefirstcompiler internal/meta/counterexamplefirstjudge)

mkdir -p "$output"
mapfile -t go_files < <(find "${paths[@]}" -name '*.go' -type f | sort)
gofmt -w "${go_files[@]}"
git diff --exit-code -- "${paths[@]}"
go test ./cmd/counterexample-first-compiler-witness ./cmd/counterexample-first-judge-witness ./cmd/counterexample-first-intervention-witness ./internal/meta/counterexamplefirst ./internal/meta/counterexamplefirstcompiler ./internal/meta/counterexamplefirstjudge

if (command -v rg >/dev/null 2>&1 && rg -n '"(failing|minimal|accepted|expected_decision|expected_resolution|expected_reason|expected_stage|expected_step)"' "$contract" "$corpus") || (! command -v rg >/dev/null 2>&1 && grep -En '"(failing|minimal|accepted|expected_decision|expected_resolution|expected_reason|expected_stage|expected_step)"' "$contract" "$corpus"); then
  printf '%s\n' 'input contract/corpus contains conclusion authority' >&2
  exit 1
fi

go list -deps ./cmd/counterexample-first-judge-witness | sort > "$output/judge-dependencies.txt"
if command -v rg >/dev/null 2>&1; then
  producer_imports=$(rg -c '^github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirstcompiler$' "$output/judge-dependencies.txt" || true)
else
  producer_imports=$(grep -Ec '^github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirstcompiler$' "$output/judge-dependencies.txt" || true)
fi
jq -n --argjson dependencies "${producer_imports:-0}" \
  '{schema:"gooo/counterexample-first-independence/v1",producer_dependencies:$dependencies}' \
  > "$output/independence.json"
[[ "${producer_imports:-0}" == "0" ]]

tracked_status_digest() {
  git status --porcelain --untracked-files=no | sha256sum | awk '{print $1}'
}

before_status=$(tracked_status_digest)
common=(--head "$HEAD_SHA" --contract "$contract" --source "$source" --corpus "$corpus")
go run "$compiler" "${common[@]}" --out "$output/receipts-a.json"
go run "$compiler" "${common[@]}" --out "$output/receipts-b.json"
cmp -s "$output/receipts-a.json" "$output/receipts-b.json"

go run "$intervention" --before "$source" --semantic-after "$semantic_intervention" --comment-after "$comment_intervention" --meta-after "$meta_intervention" --corpus "$corpus" --out "$output/interventions.json"
jq -e '.semantic_intervention.semantic_digest_equal == false and .semantic_intervention.decision_changed == true and .semantic_intervention.counterexample_changed == true and .semantic_intervention.claim_transition_changed == true and .semantic_intervention.all_cases_addressed == true and .comment_only_intervention.semantic_digest_equal == true and .comment_only_intervention.decision_changed == false and .comment_only_intervention.counterexample_changed == false and .comment_only_intervention.claim_transition_changed == false and .comment_only_intervention.all_cases_addressed == true and .meta_operation_intervention.semantic_digest_equal == false and .meta_operation_intervention.decision_changed == true and .meta_operation_intervention.claim_transition_changed == true and .meta_operation_intervention.all_cases_addressed == true and (.meta_operation_intervention.after.cases | all(.[]; .meta_operation_authorized == false))' "$output/interventions.json"

after_status=$(tracked_status_digest)
repository_writes=0
if [[ "$before_status" != "$after_status" ]]; then
  repository_writes=1
fi
jq -n --arg before "$before_status" --arg after "$after_status" --argjson writes "$repository_writes" \
  '{schema:"gooo/counterexample-first-effects/v2",before_status_digest:$before,after_status_digest:$after,repository_writes:$writes,mutation_authority:"DENIED",capability_evidence:"workflow:counterexample-first-compiler.yml;permissions:contents=read"}' \
  > "$output/effects.json"

judge_common=("${common[@]}" --receipts "$output/receipts-a.json" --independence "$output/independence.json" --effects "$output/effects.json")
go run "$judge" "${judge_common[@]}" --out "$output/report-a.json"
jq -e '.decision == "PASS" and .resolution == "EXACT" and .reason == "COUNTEREXAMPLE_FIRST_CONTRACT_OBSERVED" and .fixed_denominator.version == "counterexample-first-denominator/v3" and .fixed_denominator.cases == 5 and .fixed_denominator.unique_claims == 5 and .fixed_denominator.unique_predicates == 5 and .fixed_denominator.indicators == 12 and .fixed_denominator.claim_transitions == 15' "$output/report-a.json"
jq -e '.summary.receipts_reconstructed == 5 and .summary.cases_total == 5 and .summary.corpus_executions == 4 and .summary.discovered_counterexamples == 2 and .summary.shrink_candidate_executions == 3 and .summary.finite_neighborhood_numerator == 2 and .summary.finite_neighborhood_denominator == 2 and .summary.resolution_reruns == 1 and .summary.promotions_after_resolution == 1 and .summary.unknown_coordinates_preserved == 1 and .summary.claim_transitions_preserved == 15 and .summary.receipts_verified == 5 and .summary.unique_claims_observed == 5 and .summary.unique_predicates_observed == 5 and .summary.source_reconstruction_numerator == 5 and .summary.source_reconstruction_denominator == 5 and .summary.producer_import_numerator == 0 and .summary.producer_import_denominator == 1 and .summary.repository_writes == 0 and .summary.mutation_authority == "DENIED" and .summary.capability_evidence != ""' "$output/report-a.json"
jq -e '.indicators | length == 12 and all(.[]; .satisfied == true and .denominator > 0)' "$output/report-a.json"
jq -e '([.receipts[].claim_id] | unique | length) == 5 and ([.receipts[].predicate_id] | unique | length) == 5 and (.receipts | all(.[]; .program_meta_operation.authorized == true)) and ([.receipts[] | . as $receipt | (.claim_transitions | all(.[]; .claim_id == $receipt.claim_id and .proposition_digest == $receipt.proposition_digest and .predicate_id == $receipt.predicate_id and .evidence_digest != ""))] | all(. == true))' "$output/report-a.json"
jq -e '.receipts[] | select(.scenario_id == "resolved-minimal-counterexample") | .decision == "PASS" and .counterexample.predicate.violation_observed == true and .counterexample.finite_neighborhood_irreducible == true and .counterexample.reason == "FINITE_NEIGHBORHOOD_IRREDUCIBLE" and .decision_input.counterexample_id != "" and .decision_input.resolution_id != "" and .resolution_evidence.original_source_digest == .counterexample.source_digest and .resolution_evidence.repair_source_digest == .resolution_evidence.observation.source_digest and .resolution_evidence.repair_delta_digest != "" and .resolution_evidence.same_claim == true and .resolution_evidence.same_predicate == true and (.claim_transitions | map(select(.from == "REFUTED" and .to == "DISCHARGED")) | length == 1)' "$output/report-a.json"
jq -e '.receipts[] | select(.scenario_id == "canonical-control") | .decision == "FAIL_CLOSED" and .reason == "COUNTEREXAMPLE_REQUIRED"' "$output/report-a.json"
jq -e '.receipts[] | select(.scenario_id == "unresolved-counterexample") | .decision == "REFUTED" and .resolution == "LOWER_RESOLUTION" and (.claim_transitions | map(select(.to == "REFUTED")) | length >= 1)' "$output/report-a.json"
jq -e '.receipts[] | select(.scenario_id == "unobserved-input") | .decision == "UNKNOWN" and .resolution == "LOWER_RESOLUTION" and .coordinate.stage == "INPUT_OBSERVATION" and .coordinate.step == "source-acquisition" and .coordinate.reason == "SOURCE_NOT_PROVIDED" and (.claim_transitions | all(.[]; .to == "OPEN" and .reason == "SOURCE_NOT_PROVIDED"))' "$output/report-a.json"

go run "$judge" "${common[@]}" --receipts "$output/receipts-b.json" --independence "$output/independence.json" --effects "$output/effects.json" --out "$output/report-b.json"
cmp -s "$output/report-a.json" "$output/report-b.json"

jq '.[0].reason = "tampered"' "$output/receipts-a.json" > "$output/tampered-receipts.json"
if go run "$judge" "${common[@]}" --receipts "$output/tampered-receipts.json" --independence "$output/independence.json" --effects "$output/effects.json" --out "$output/tampered-report.json"; then
  printf '%s\n' 'independent judge accepted a tampered receipt' >&2
  exit 1
else
  tamper_status=$?
fi
[[ "$tamper_status" == "1" ]]
jq -e '.decision == "FAIL_CLOSED" and .reason == "COUNTEREXAMPLE_RECEIPT_MISMATCH" and .tampered_rejected == 1' "$output/tampered-report.json"

git diff --exit-code

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  jq -r '"### Counterexample-first meta compilation\n- decision: \(.decision) / \(.resolution)\n- corpus executions: \(.summary.corpus_executions)/5\n- discovered counterexamples: \(.summary.discovered_counterexamples)\n- shrink candidate executions: \(.summary.shrink_candidate_executions)\n- finite-neighborhood irreducibility: \(.summary.finite_neighborhood_numerator)/\(.summary.finite_neighborhood_denominator)\n- resolution reruns: \(.summary.resolution_reruns)\n- source reconstruction: \(.summary.source_reconstruction_numerator)/\(.summary.source_reconstruction_denominator)\n- producer import: \(.summary.producer_import_numerator)/\(.summary.producer_import_denominator)\n- claim transitions: \(.summary.claim_transitions_preserved)/15\n- repository writes: \(.summary.repository_writes)\n- mutation authority: \(.summary.mutation_authority)\n- receipt: \(.digest)"' "$output/report-a.json" >> "$GITHUB_STEP_SUMMARY"
fi
