#!/usr/bin/env bash

set -euo pipefail

out=${1:?output directory is required}
expected_sha=${2:?expected checkout SHA is required}
observed_checkout_sha=${3:?observed checkout SHA is required}
expectation_digest=${4:?expectation artifact digest is required}
first="$out/first.json"
expectation="$out/claim-transition-expectations.json"
evolution="$out/claim-transition-expectation-evolution.json"
gates="$out/contract-gates.json"

printf '%s\n' '{"gates":[]}' > "$gates"
failed=0

record_gate() {
  local gate_id="$1"
  local stage="$2"
  local step="$3"
  local expected="$4"
  local observed="$5"
  local reason="$6"
  local passed="$7"
  local next="$gates.next"
  jq --arg id "$gate_id" --arg stage "$stage" --arg step "$step" \
    --arg expected "$expected" --arg observed "$observed" --arg reason "$reason" \
    --argjson passed "$passed" \
    '.gates += [{gate_id:$id,stage:$stage,step:$step,expected:$expected,observed:$observed,reason:(if $passed then "GATE_SATISFIED" else $reason end),decision:(if $passed then "PASS" else "FAIL_CLOSED" end),resolution:(if $passed then "EXACT" else "LOWER_RESOLUTION" end),passed:$passed}]' \
    "$gates" > "$next"
  mv "$next" "$gates"
  if [[ "$passed" != true ]]; then
    failed=$((failed + 1))
  fi
}

run_gate() {
  local gate_id="$1"
  local stage="$2"
  local step="$3"
  local expected="$4"
  local observed="$5"
  local reason="$6"
  local filter="$7"
  local passed=true
  if ! jq -e --arg output_path "$first" --arg expected_sha "$expected_sha" \
    --arg observed_checkout_sha "$observed_checkout_sha" --arg expectation_digest "$expectation_digest" \
    --slurpfile expectation "$expectation" --slurpfile diff "$out/claim-identity-diff.json" --slurpfile evolution "$evolution" "$filter" "$first" >/dev/null; then
    passed=false
  fi
  record_gate "$gate_id" "$stage" "$step" "$expected" "$observed" "$reason" "$passed"
}

expected_case_ids='["equivalent","semantic-change","value-program-change","indeterminate","ambiguous-match"]'
expected_claim_counts='[{"case_id":"equivalent","total":7},{"case_id":"semantic-change","total":7},{"case_id":"value-program-change","total":7},{"case_id":"indeterminate","total":3},{"case_id":"ambiguous-match","total":7}]'
expected_status='{"open":10,"discharged":18,"refuted":3}'

observed_suite=$(jq -c '{schema,decision,resolution,contract_reproduction,subject_semantic_equivalence,semantic_equivalence_claim}' "$first")
observed_subject=$(jq -c '{subject_sha,observed_checkout_sha}' "$first")
observed_case_ids=$(jq -c '{expected:.case_contract_expected_ids,observed:.case_contract_observed_ids,observed_recipe_ids:.case_contract_observed_recipe_ids}' "$first")
observed_claim_counts=$(jq -c '[.cases[] | {case_id:.report.case_id,total:.report.receipt.total_claims}]' "$first")
observed_status=$(jq -c '{open:.summary.open_claims,discharged:.summary.discharged_claims,refuted:.summary.refuted_claims}' "$first")
jq '[.cases[].report.claim_identity |
  (.expected_claim_ids | unique) as $expected_set |
  (.observed_claim_ids | unique) as $observed_set |
  {case_id,
   expected_claim_ids,
   observed_claim_ids,
   expected_unique: ((.expected_claim_ids | length) == ($expected_set | length)),
   observed_unique: ((.observed_claim_ids | length) == ($observed_set | length)),
   expected_only: ($expected_set - $observed_set),
   observed_only: ($observed_set - $expected_set),
   same_set_different_order: ($expected_set == $observed_set and .expected_claim_ids != .observed_claim_ids)}
]' "$first" > "$out/claim-identity-diff.json"
observed_identity=$(jq -c '[.[] | {case_id,expected_count:(.expected_claim_ids|length),observed_count:(.observed_claim_ids|length),expected_only_count:(.expected_only|length),observed_only_count:(.observed_only|length),expected_unique,observed_unique,same_set_different_order}]' "$out/claim-identity-diff.json")
observed_identity_pass=$(jq -r --arg digest "$expectation_digest" --slurpfile diff "$out/claim-identity-diff.json" '[.cases[].report.claim_identity as $identity | $diff[0][] | select(.case_id == $identity.case_id and .expected_unique and .observed_unique and (.expected_only | length) == 0 and (.observed_only | length) == 0) | select($identity.decision == "PASS" and $identity.resolution == "EXACT" and $identity.reason == "FIXED_CLAIM_IDENTITY_EXACT" and $identity.expectation_artifact_digest == $digest and $identity.expected_transition_identity_digest == $identity.observed_transition_identity_digest and $identity.coverage_bps == 10000)] | length' "$first")
observed_inventory_pass=$(jq -r '[.[] | select(.expected_unique and .observed_unique and (.expected_only | length) == 0 and (.observed_only | length) == 0)] | length' "$out/claim-identity-diff.json")
observed_transition_pass=$(jq -r '[.cases[].report.claim_identity | select(.expected_transition_identity_digest == .observed_transition_identity_digest)] | length' "$first")
observed_artifact_pass=$(jq -r '[.cases[].report.claim_identity | select(.expectation_artifact_digest == $digest)] | length' --arg digest "$expectation_digest" "$first")
observed_case_row_pass=$(jq -r --slurpfile expected "$expectation" '($expected[0].cases | map({id,case_row_digest}) | sort_by(.id)) as $fixed | ([.cases[] | {id:.report.case_id,case_row_digest:.report.claim_identity.expectation_case_row_digest}] | sort_by(.id)) as $observed | if $fixed == $observed then 5 else 0 end' "$first")
observed_claim_identity_record_pass=$(jq -r '[.cases[].report | select(.claim_identity.case_id == .case_id and .claim_identity.denominator_id == "gooo://semantic-delta-receipt-denominator/v2" and (.claim_identity.expected_claim_ids | length) > 0 and (.claim_identity.observed_claim_ids | length) == (.claim_identity.expected_claim_ids | length) and (.claim_identity.expected_claims | length) == (.claim_identity.observed_claims | length))] | length' "$first")
observed_consumer=$(jq -r '[.cases[].report.independent_verdict | select(.consumer == "semanticdeltareceiptconsumer.AdjudicateFiles")] | length' "$first")
observed_effects=$(jq -r --arg expected "$expected_sha" --arg observed "$observed_checkout_sha" '[.cases[].report.receipt | select(.expected_subject_sha == $expected and .observed_checkout_sha == $observed and .effects.status == "NET_REPOSITORY_STATE_UNCHANGED" and .effects.changed_path_or_content_count == 0 and .effects.output_location == "OUTSIDE_REPOSITORY")] | length' "$first")

run_gate "suite-contract-decision" "suite-contract" "reconstruct-fixed-contract" \
  'schema=conformance/v2, decision=FIXED_POINT, resolution=EXACT' "$observed_suite" "SUITE_CONTRACT_NOT_REPRODUCED" \
  '.schema == "gooo/semantic-delta-receipt-conformance/v2" and .decision == "FIXED_POINT" and .resolution == "EXACT" and .contract_reproduction == "FIXED_FIVE_CASE_CONTRACT_REPRODUCED" and .subject_semantic_equivalence == "NOT_ASSERTED" and .semantic_equivalence_claim == "NOT_ASSERTED" and .output_path == $output_path'

run_gate "subject-binding" "bind-subject" "observe-checkout-sha" \
  "expected=$expected_sha, observed=$observed_checkout_sha" "$observed_subject" "SUBJECT_CHECKOUT_BINDING_MISMATCH" \
  '.subject_sha == $expected_sha and .observed_checkout_sha == $observed_checkout_sha'

run_gate "source-and-meta-contract" "suite-contract" "validate-source-addresses" \
  'source_paths=11, denominator=v2, meta=main.gooo' \
  "paths=$(jq -r '.source_paths | length' "$first"), denominator=$(jq -r '.denominator_version' "$first"), meta=$(jq -r '.meta_source_path' "$first")" "SOURCE_META_CONTRACT_MISMATCH" \
  '(.source_paths | length) == 11 and .denominator_version == "v2" and .meta_source_path == "examples/semantic-delta-receipt/main.gooo"'

run_gate "projection-coverage" "semantic-projection" "validate-modeled-components" \
  'modeled=5, total=5, coverage_bps=10000' \
  "modeled=$(jq -r '.modeled_semantic_components' "$first"), total=$(jq -r '.total_semantic_components' "$first"), coverage_bps=$(jq -r '.declared_projection_component_kind_coverage_bps' "$first")" "DECLARED_PROJECTION_COMPONENT_COVERAGE_MISMATCH" \
  '.modeled_semantic_components == 5 and .total_semantic_components == 5 and .declared_projection_component_kind_coverage_bps == 10000'

run_gate "case-inventory" "suite-contract" "validate-case-inventory" \
  "expected=$expected_case_ids" "$observed_case_ids" "FIXED_CASE_ID_INVENTORY_MISMATCH" \
  '.case_contract_denominator_id == "gooo://semantic-delta-receipt-denominator/v2" and .case_contract_fixed_total == 5 and .case_contract_expected_ids == ["equivalent","semantic-change","value-program-change","indeterminate","ambiguous-match"] and ((.case_contract_expected_ids | sort) == (.case_contract_observed_ids | sort)) and ((.case_contract_expected_ids | sort) == (.case_contract_observed_recipe_ids | sort)) and .stage == "suite-contract" and .step == "validate-case-inventory" and .case_contract_stage == "suite-contract" and .case_contract_step == "validate-case-inventory" and .case_contract_reason == "FIXED_CASE_ID_INVENTORY_EXACT"'

run_gate "claim-count-contract" "claim-ledger" "validate-fixed-counts" \
  "version=v1, total=31, cases=$expected_claim_counts" \
  "version=$(jq -r '.claim_count_contract_version' "$first"), total=$(jq -r '.claim_count_expected_total' "$first"), cases=$observed_claim_counts" "FIXED_CLAIM_COUNT_CONTRACT_MISMATCH" \
  '.claim_count_contract_version == "v1" and .claim_count_expected_total == 31 and .claim_count_expected_by_case == [{"case_id":"equivalent","total":7},{"case_id":"semantic-change","total":7},{"case_id":"value-program-change","total":7},{"case_id":"indeterminate","total":3},{"case_id":"ambiguous-match","total":7}]'

run_gate "structural-observations" "semantic-classification" "count-structural-observations" \
  'structural_observations=4' "structural_observations=$(jq -r '.summary.structural_observations' "$first")" "STRUCTURAL_OBSERVATION_COUNT_MISMATCH" \
  '.summary.structural_observations == 4'

run_gate "claim-total" "claim-ledger" "count-fixed-claims" \
  'total_claims=31, explained=31' \
  "total_claims=$(jq -r '.summary.total_claims' "$first"), explained=$(jq -r '.summary.claims_with_explained_status' "$first")" "FIXED_CLAIM_TOTAL_MISMATCH" \
  '.summary.total_claims == 31 and .summary.claims_with_explained_status == 31'

run_gate "claim-status-distribution" "claim-ledger" "count-persistent-statuses" \
  "$expected_status" "$observed_status" "CLAIM_STATUS_DISTRIBUTION_MISMATCH" \
  '.summary.open_claims == 10 and .summary.discharged_claims == 18 and .summary.refuted_claims == 3'

run_gate "claim-status-coverage" "claim-ledger" "validate-transition-explanations" \
  'coverage_bps=10000, transition_chains=31' \
  "coverage_bps=$(jq -r '.summary.claim_status_coverage_bps' "$first"), transition_chains=$(jq -r '.summary.transition_chains' "$first")" "CLAIM_STATUS_COVERAGE_MISMATCH" \
  '.summary.claim_status_coverage_bps == 10000 and .summary.transition_chains == 31'

run_gate "delta-evidence-counts" "semantic-classification" "count-delta-evidence" \
  'textual_changes=5, claim_transition_cases=5, adjudicated_cases=5, distinct_propositions=24, added=0, removed=0, changed=1, ambiguous=1' \
  "textual_changes=$(jq -r '.summary.textual_changes' "$first"), claim_transition_cases=$(jq -r '.summary.claim_transition_cases' "$first"), adjudicated_cases=$(jq -r '.summary.adjudicated_cases' "$first"), distinct_propositions=$(jq -r '.summary.distinct_propositions' "$first"), added=$(jq -r '.summary.added_claims' "$first"), removed=$(jq -r '.summary.removed_claims' "$first"), changed=$(jq -r '.summary.changed_claims' "$first"), ambiguous=$(jq -r '.summary.ambiguous_cases' "$first")" "DELTA_EVIDENCE_COUNT_MISMATCH" \
  '.summary.textual_changes == 5 and .summary.claim_transition_cases == 5 and .summary.adjudicated_cases == 5 and .summary.distinct_propositions == 24 and .summary.added_claims == 0 and .summary.removed_claims == 0 and .summary.changed_claims == 1 and .summary.ambiguous_cases == 1'

run_gate "semantic-case-classification" "semantic-classification" "count-fixed-case-outcomes" \
  'preserved=1, changed=2, indeterminate=2, unknown_paths=2' \
  "preserved=$(jq -r '.summary.semantic_preserved' "$first"), changed=$(jq -r '.summary.semantic_changed' "$first"), indeterminate=$(jq -r '.summary.indeterminate' "$first"), unknown_paths=$(jq -r '.summary.unknown_paths' "$first")" "SEMANTIC_CASE_CLASSIFICATION_MISMATCH" \
  '.summary.semantic_preserved == 1 and .summary.semantic_changed == 2 and .summary.indeterminate == 2 and .summary.unknown_paths == 2'

run_gate "claim-identity-artifact" "claim-identity" "bind-expectation-artifact" \
  'expectation_artifact_digest matches copied raw artifact for 5/5 cases' "matched_cases=$observed_artifact_pass/5" "CLAIM_EXPECTATION_ARTIFACT_DIGEST_MISMATCH" \
  '[.cases[].report.claim_identity | select(.expectation_artifact_digest == $expectation_digest)] | length == 5'

run_gate "claim-identity-case-row" "claim-identity" "bind-expectation-case-rows" \
  'case_row_digest matches fixed artifact for 5/5 cases' "matched_cases=$observed_case_row_pass/5" "CLAIM_EXPECTATION_CASE_ROW_DIGEST_MISMATCH" \
  '($expectation[0].cases | map({id,case_row_digest}) | sort_by(.id)) as $fixed | ([.cases[] | {id:.report.case_id,case_row_digest:.report.claim_identity.expectation_case_row_digest}] | sort_by(.id)) as $observed | $fixed == $observed'

run_gate "claim-identity-expectation-binding" "claim-identity" "bind-fixed-expectation-values" \
  'artifact stable identity/evidence rows, IDs, transition digests, counts, and row digests match report for 5/5 cases' \
  'report expectation rows compared with copied artifact rows' "CLAIM_EXPECTATION_REPORT_BINDING_MISMATCH" \
  '($expectation[0].cases | map({id,expected_claim_ids:(.expected_claim_ids|sort),expected_claims:(.expected_claims|sort_by(.stable_id)),expected_transition_identity_digest,expected_claim_total,case_row_digest}) | sort_by(.id)) as $fixed | ([.cases[] | {id:.report.case_id,expected_claim_ids:(.report.claim_identity.expected_claim_ids|sort),expected_claims:(.report.claim_identity.expected_claims|sort_by(.stable_id)),expected_transition_identity_digest:.report.claim_identity.expected_transition_identity_digest,expected_claim_total:.report.claim_identity.expected_claim_total,case_row_digest:.report.claim_identity.expectation_case_row_digest}] | sort_by(.id)) as $observed | $fixed == $observed'

run_gate "claim-identity-inventory" "claim-identity" "compare-fixed-claim-inventory" \
  'unique expected/observed claim ID sets canonically equal for 5/5 cases' "matched_cases=$observed_inventory_pass/5; diff=$out/claim-identity-diff.json" "CLAIM_ID_INVENTORY_MISMATCH" \
  '($diff[0] | map(select(.expected_unique and .observed_unique and (.expected_only | length) == 0 and (.observed_only | length) == 0)) | length) == 5'

run_gate "claim-identity-transition-digest" "claim-identity" "compare-transition-identity-digest" \
  'expected transition identity digest equals observed raw-pair digest for 5/5 cases' "matched_cases=$observed_transition_pass/5" "CLAIM_TRANSITION_IDENTITY_DIGEST_MISMATCH" \
  '[.cases[].report.claim_identity | select(.expected_transition_identity_digest == .observed_transition_identity_digest)] | length == 5'

run_gate "claim-identity-records" "claim-identity" "validate-case-records" \
  'case ID, denominator, and complete stable identity rows for 5/5 cases' \
  "valid_records=$observed_claim_identity_record_pass/5" "CLAIM_IDENTITY_RECORD_BINDING_MISMATCH" \
  '[.cases[].report | select(.claim_identity.case_id == .case_id and .claim_identity.denominator_id == "gooo://semantic-delta-receipt-denominator/v2" and (.claim_identity.expected_claim_ids | length) > 0 and (.claim_identity.observed_claim_ids | length) == (.claim_identity.expected_claim_ids | length) and (.claim_identity.expected_claims | length) == (.claim_identity.observed_claims | length) and (.claim_identity.expected_claims | sort_by(.stable_id)) == (.claim_identity.observed_claims | sort_by(.stable_id)))] | length == 5'

run_gate "claim-identity-decision" "claim-identity" "adjudicate-fixed-expectation" \
  'decision=PASS, resolution=EXACT, exact reason for 5/5 cases' "matched_cases=$observed_identity_pass/5; records=$observed_identity" "CLAIM_IDENTITY_DECISION_NOT_EXACT" \
  '[.cases[].report.claim_identity | select(.decision == "PASS" and .resolution == "EXACT" and .stage == "claim-identity" and .step == "compare-fixed-expectation" and .reason == "FIXED_CLAIM_IDENTITY_EXACT" and .coverage_bps == 10000)] | length == 5'

run_gate "claim-identity" "claim-identity" "compare-fixed-expectation" \
  '5/5 PASS/EXACT identities with artifact, row, inventory, digest, and coverage bindings' "matched_cases=$observed_identity_pass/5; records=$observed_identity" "FIXED_CLAIM_IDENTITY_EXPECTATION_MISMATCH" \
  '[.cases[].report.claim_identity | select(.decision == "PASS" and .resolution == "EXACT" and .stage == "claim-identity" and .step == "compare-fixed-expectation" and .reason == "FIXED_CLAIM_IDENTITY_EXACT" and .expectation_artifact_digest == $expectation_digest and (.expected_claim_ids | length) == .fixed_total and ((.expected_claim_ids | sort) == (.observed_claim_ids | sort)) and .expected_transition_identity_digest == .observed_transition_identity_digest and .coverage_bps == 10000)] | length == 5'

run_gate "source-pair-binding" "claim-identity" "bind-observed-source-pair" \
  'raw/semantic source-pair addresses and digests match receipt for 5/5 cases' \
  "matched_cases=$(jq -r '[.cases[].report | select(.claim_identity.observed_source_pair.before_path == .source_paths[0] and .claim_identity.observed_source_pair.after_path == .source_paths[1] and .claim_identity.observed_source_pair.before_raw_digest == .receipt.before.source_digest and .claim_identity.observed_source_pair.after_raw_digest == .receipt.after.source_digest and .claim_identity.observed_source_pair.before_semantic_digest == .receipt.before.semantic_digest and .claim_identity.observed_source_pair.after_semantic_digest == .receipt.after.semantic_digest)] | length' "$first")/5" "SOURCE_PAIR_BINDING_MISMATCH" \
  '[.cases[].report | select(.claim_identity.observed_source_pair.before_path == .source_paths[0] and .claim_identity.observed_source_pair.after_path == .source_paths[1] and .claim_identity.observed_source_pair.before_raw_digest == .receipt.before.source_digest and .claim_identity.observed_source_pair.after_raw_digest == .receipt.after.source_digest and .claim_identity.observed_source_pair.before_semantic_digest == .receipt.before.semantic_digest and .claim_identity.observed_source_pair.after_semantic_digest == .receipt.after.semantic_digest)] | length == 5'

run_gate "independent-consumer" "independent-adjudication" "reconstruct-raw-pair" \
  'semanticdeltareceiptconsumer.AdjudicateFiles for 5/5 cases' "matched_cases=$observed_consumer/5" "INDEPENDENT_CONSUMER_MISSING" \
  '[.cases[].report.independent_verdict | select(.consumer == "semanticdeltareceiptconsumer.AdjudicateFiles")] | length == 5'

run_gate "subject-effects-binding" "observe-effects" "bind-checkout-and-workspace" \
  "subject/effects binding exact for 5/5 cases; expected=$expected_sha observed=$observed_checkout_sha" "matched_cases=$observed_effects/5" "SUBJECT_OR_EFFECTS_BINDING_MISMATCH" \
  '[.cases[].report.receipt | select(.expected_subject_sha == $expected_sha and .observed_checkout_sha == $observed_checkout_sha and .effects.status == "NET_REPOSITORY_STATE_UNCHANGED" and .effects.changed_path_or_content_count == 0 and .effects.output_location == "OUTSIDE_REPOSITORY")] | length == 5'

run_gate "case-results" "suite-contract" "count-passed-cases" \
  'cases_total=5, cases_passed=5, coverage_bps=10000' \
  "cases_total=$(jq -r '.summary.cases_total' "$first"), cases_passed=$(jq -r '.summary.cases_passed' "$first"), coverage_bps=$(jq -r '.case_contract_coverage_bps' "$first")" "FIXED_CASE_RESULT_CONTRACT_INCOMPLETE" \
  '.summary.cases_total == 5 and .summary.cases_passed == 5 and .case_contract_coverage_bps == 10000 and ([.cases[] | select(.passed == true)] | length) == 5'

run_gate "proof-choice-distribution" "proof-selection" "count-proof-choices" \
  'FOUNDATION=3, COHERENCE=2, REGRESSION=1' \
  "$(jq -c '[.cases[0].report.indicators[] | .proof_choice] | group_by(.) | map({proof_choice:.[0],count:length})' "$first")" "PROOF_CHOICE_DISTRIBUTION_MISMATCH" \
  '([.cases[0].report.indicators[] | select(.proof_choice == "FOUNDATION")] | length) == 3 and ([.cases[0].report.indicators[] | select(.proof_choice == "COHERENCE")] | length) == 2 and ([.cases[0].report.indicators[] | select(.proof_choice == "REGRESSION")] | length) == 1'

artifact_observed=$(jq -c '{schema,denominator_id,claim_count_contract_version,fixed_claim_total,denominator_evolution_receipt,fixed_case_total,cases:(.cases | map({id,expected_claim_total,claim_id_count:(.expected_claim_ids|length),claim_record_count:(.expected_claims|length),case_row_digest}))}' "$expectation")
run_gate "expectation-contract" "claim-identity" "validate-expectation-artifact" \
  'strict fixed artifact v2, complete stable rows, 5 cases, counts 7/7/7/3/7, total 31' "$artifact_observed" "CLAIM_EXPECTATION_CONTRACT_MISMATCH" \
  '$expectation[0].schema == "gooo/semantic-delta-claim-transition-expectations/v2" and $expectation[0].denominator_id == "gooo://semantic-delta-receipt-denominator/v2" and $expectation[0].fixed_case_total == 5 and $expectation[0].claim_count_contract_version == "v1" and $expectation[0].fixed_claim_total == 31 and $expectation[0].denominator_evolution_receipt == "REQUIRED_FOR_FIXED_CLAIM_COUNT_CHANGE" and ([$expectation[0].cases[].id] | sort) == ["ambiguous-match","equivalent","indeterminate","semantic-change","value-program-change"] and ([$expectation[0].cases[].id] | unique | length) == 5 and ([$expectation[0].cases | sort_by(.id)[] | .expected_claim_total]) == [7,7,3,7,7] and ([$expectation[0].cases[].expected_claim_total] | add) == $expectation[0].fixed_claim_total and ([$expectation[0].cases[] | select((.expected_claim_ids | length) == .expected_claim_total and (.expected_claims | length) == .expected_claim_total and (([.expected_claims[].stable_id] | sort) == (.expected_claim_ids | sort)) and (.case_row_digest | test("^sha256:[0-9a-f]{64}$")))] | length) == 5'

evolution_observed=$(jq -c '{schema,authority,old_artifact_digest,new_artifact_digest,persistence_manifest_digest,denominator_id,denominator_unchanged,fixed_claim_total_before,fixed_claim_total_after,change_kind,historical_migration_removed,historical_migration_added,historical_migration_mapping_rows,historical_migration_decision,historical_migration_resolution,expectation_conformance_rows,expectation_conformance_rows_total,expectation_conformance_claim_rows,expectation_conformance_claim_rows_total,v3_observation_pairs_reconstructed,v3_observation_pairs_total,v3_producer_consumer_exact,v3_producer_consumer_exact_total,stable_identity_preserved,stable_identity_total,evidence_only_changes,evidence_only_total,raw_evidence_changed_on_nonsemantic,raw_evidence_nonsemantic_total,semantic_target_preserved_on_nonsemantic,semantic_target_nonsemantic_total,claim_recreated_due_only_to_raw_digest,claim_recreated_due_only_to_raw_digest_total,persistence_decision,persistence_resolution,cases:(.cases | map({case_id,old_expected_count:(.old_expected_ids|length),producer_old_observed_count:(.producer_old_observed_ids|length),consumer_old_observed_count:(.consumer_old_observed_ids|length),old_artifact_exact,old_producer_consumer_exact,new_expected_count:(.new_expected_ids|length),producer_observed_count:(.producer_observed_ids|length),consumer_observed_count:(.consumer_observed_ids|length),new_expectation_producer_exact,new_expectation_consumer_exact,producer_consumer_exact,stable_identity_preserved,stable_identity_total,evidence_only_changes,evidence_only_total,claim_recreated_due_only_to_raw_digest,claim_recreated_due_only_to_raw_digest_total,persistence_exact,producer_persistence_decision:.producer_persistence.decision,consumer_persistence_decision:.consumer_persistence.decision,removed_id_count:(.removed_ids|length),added_id_count:(.added_ids|length),proposition_target_change_count:(.proposition_target_changes|length),consumer_proposition_target_change_count:(.consumer_proposition_target_changes|length)}))}' "$evolution")
run_gate "expectation-evolution" "claim-identity" "validate-expectation-evolution" \
  'historical v1-to-v3 migration is explicit; old/new bytes and source rows are exact, but migration is not persistence' "$evolution_observed" "HISTORICAL_SCHEMA_MIGRATION_INVALID" \
  '$evolution[0].schema == "gooo/semantic-delta-claim-expectation-evolution/v2" and $evolution[0].authority == "SOURCE_DERIVED_SEMANTIC_CLAIM_CONTRACT" and ($evolution[0].old_artifact_digest | test("^sha256:[0-9a-f]{64}$")) and ($evolution[0].new_artifact_digest | test("^sha256:[0-9a-f]{64}$")) and ($evolution[0].persistence_manifest_digest | test("^sha256:[0-9a-f]{64}$")) and $evolution[0].old_artifact_digest != $evolution[0].new_artifact_digest and $evolution[0].new_artifact_digest == $expectation_digest and $evolution[0].old_artifact_bytes > 0 and $evolution[0].new_artifact_bytes > 0 and $evolution[0].denominator_id == "gooo://semantic-delta-receipt-denominator/v2" and $evolution[0].denominator_unchanged == true and $evolution[0].fixed_claim_total_before == 31 and $evolution[0].fixed_claim_total_after == 31 and $evolution[0].change_kind == "HISTORICAL_SCHEMA_MIGRATION" and $evolution[0].historical_migration_decision == "FIXED_POINT" and $evolution[0].historical_migration_resolution == "EXACT" and $evolution[0].evolution_rows_independently_reconstructed == 5 and $evolution[0].evolution_rows_total == 5 and $evolution[0].historical_migration_removed > 0 and $evolution[0].historical_migration_added > 0 and ([$evolution[0].cases[].case_id] | sort) == ["ambiguous-match","equivalent","indeterminate","semantic-change","value-program-change"] and ([$evolution[0].cases[] | select(.old_artifact_exact == true and .old_producer_consumer_exact == true and .producer_consumer_exact == true and ((.new_expected_ids | sort) == (.producer_observed_ids | sort)) and ((.new_expected_ids | sort) == (.consumer_observed_ids | sort)))] | length) == 5'

run_gate "evolution-expectation-binding" "claim-identity" "compare-expectation-with-independent-runtime" \
  'new expectation stable/evidence rows equal producer and consumer reconstructions for 5/5 cases' \
  "producer_exact=$(jq -r '[.cases[] | select(.new_expectation_producer_exact == true)] | length' "$evolution")/5, consumer_exact=$(jq -r '[.cases[] | select(.new_expectation_consumer_exact == true)] | length' "$evolution")/5" "EXPECTATION_RUNTIME_RECORD_MISMATCH" \
  '$evolution[0].expectation_conformance_rows == 5 and $evolution[0].expectation_conformance_claim_rows == 31 and ([$evolution[0].cases[] | select(.new_expectation_producer_exact == true and .new_expectation_consumer_exact == true)] | length) == 5'

persistent_identity_observed=$(jq -c '{stable_identity_preserved,stable_identity_total,persistent_claim_identity,persistent_claim_identity_total,evidence_only_changes,evidence_only_total,raw_evidence_changed_on_nonsemantic,raw_evidence_nonsemantic_total,semantic_target_preserved_on_nonsemantic,semantic_target_nonsemantic_total,claim_recreated_due_only_to_raw_digest,claim_recreated_due_only_to_raw_digest_total,v3_observation_pairs_reconstructed,v3_observation_pairs_total,v3_producer_consumer_exact,v3_producer_consumer_exact_total,persistence_decision,persistence_resolution}' "$evolution")
run_gate "persistent-claim-identity" "claim-identity-persistence" "validate-baseline-alternate-mapping" \
  'actual v3 baseline/alternate source observations preserve stable identity 31/31 and change raw evidence 31/31' "$persistent_identity_observed" "PERSISTENT_CLAIM_IDENTITY_MISMATCH" \
  '$evolution[0].stable_identity_preserved == 31 and $evolution[0].stable_identity_total == 31 and $evolution[0].persistent_claim_identity == 31 and $evolution[0].persistent_claim_identity_total == 31 and $evolution[0].evidence_only_changes == 31 and $evolution[0].evidence_only_total == 31 and $evolution[0].raw_evidence_changed_on_nonsemantic == 31 and $evolution[0].raw_evidence_nonsemantic_total == 31 and $evolution[0].semantic_target_preserved_on_nonsemantic == 31 and $evolution[0].semantic_target_nonsemantic_total == 31 and $evolution[0].claim_recreated_due_only_to_raw_digest == 0 and $evolution[0].claim_recreated_due_only_to_raw_digest_total == 31 and $evolution[0].v3_observation_pairs_reconstructed == 5 and $evolution[0].v3_observation_pairs_total == 5 and $evolution[0].v3_producer_consumer_exact == 5 and $evolution[0].v3_producer_consumer_exact_total == 5 and $evolution[0].persistence_decision == "FIXED_POINT" and $evolution[0].persistence_resolution == "EXACT" and ([$evolution[0].cases[] | select(.persistence_exact == true and .producer_persistence.decision == "FIXED_POINT" and .consumer_persistence.decision == "FIXED_POINT" and .stable_identity_preserved == .stable_identity_total and .evidence_only_changes == .evidence_only_total)] | length) == 5'

inventory_root=$(jq -c '[.[] | select((.expected_only | length) > 0 or (.observed_only | length) > 0 or (.expected_unique | not) or (.observed_unique | not))] | length' "$out/claim-identity-diff.json")
expected_only_cases=$(jq -c '[.[] | select((.expected_only | length) > 0)] | length' "$out/claim-identity-diff.json")
observed_only_cases=$(jq -c '[.[] | select((.observed_only | length) > 0)] | length' "$out/claim-identity-diff.json")
order_only_cases=$(jq -c '[.[] | select(.same_set_different_order)] | length' "$out/claim-identity-diff.json")
duplicate_cases=$(jq -c '[.[] | select((.expected_unique | not) or (.observed_unique | not))] | length' "$out/claim-identity-diff.json")
downstream_failures=$(jq -c '[.gates[] | select(.passed == false and (.gate_id == "claim-identity-transition-digest" or .gate_id == "claim-identity-decision" or .gate_id == "claim-identity")) | {gate_id,stage,step,reason}]' "$gates")
jq -n --arg schema "gooo/semantic-delta-causal-receipt/v1" --arg decision "$(if (( inventory_root > 0 )); then printf FAIL_CLOSED; else printf PASS; fi)" --arg resolution "$(if (( inventory_root > 0 )); then printf LOWER_RESOLUTION; else printf EXACT; fi)" --argjson root_cases "$inventory_root" --argjson expected_only_cases "$expected_only_cases" --argjson observed_only_cases "$observed_only_cases" --argjson order_only_cases "$order_only_cases" --argjson duplicate_cases "$duplicate_cases" --argjson downstream "$downstream_failures" '{schema:$schema,decision:$decision,resolution:$resolution,root_causes:[{root_id:"claim-inventory-set-mismatch",gate_id:"claim-identity-inventory",affected_case_count:$root_cases,expected_only_case_count:$expected_only_cases,observed_only_case_count:$observed_only_cases,same_set_different_order_case_count:$order_only_cases,duplicate_case_count:$duplicate_cases,artifact:"claim-identity-diff.json"}],downstream_failures:$downstream,root_failure_count:(if $root_cases > 0 then 1 else 0 end),downstream_failure_count:($downstream|length),causal_rule:"inventory mismatch is one root; transition/decision failures are downstream"}' > "$out/semantic-delta-causal-receipt.json"

jq --argjson failed "$failed" '. + {schema:"gooo/semantic-delta-receipt-contract-gates/v1",decision:(if $failed == 0 then "PASS" else "FAIL_CLOSED" end),resolution:(if $failed == 0 then "EXACT" else "LOWER_RESOLUTION" end),failed_count:$failed,gate_count:(.gates | length)}' "$gates" > "$gates.next"
mv "$gates.next" "$gates"
jq . "$gates"
if (( failed != 0 )); then
  printf 'contract gates failed: %d; decision=FAIL_CLOSED resolution=LOWER_RESOLUTION\n' "$failed" >&2
  exit 1
fi
