#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

output="${QUORUM_OUTPUT:-evidence-quorum-output}"
source="examples/evidence-quorum/main.gooo"
policy="examples/evidence-quorum/policy.gooo"
mkdir -p "$output/receipts" "$output/cases" "$output/generated"

test -z "$(gofmt -l internal/meta/evidencequorumconsumer internal/meta/evidencequorumwire internal/meta/evidencequorumpolicy internal/meta/evidencequorumchannel cmd/evidence-quorum-witness cmd/evidence-quorum-source-channel cmd/evidence-quorum-reconstructor cmd/evidence-quorum-artifact-observer cmd/evidence-quorum-counterexample)"

go build -o "$output/gooo" ./cmd/gooo
go build -o "$output/evidence-quorum-witness" ./cmd/evidence-quorum-witness
go build -o "$output/source-channel" ./cmd/evidence-quorum-source-channel
go build -o "$output/reconstructor" ./cmd/evidence-quorum-reconstructor
go build -o "$output/artifact-observer" ./cmd/evidence-quorum-artifact-observer
go build -o "$output/counterexample" ./cmd/evidence-quorum-counterexample

go list -deps ./internal/meta/evidencequorumconsumer | sort -u > "$output/consumer-dependencies.txt"
forbidden="$(grep -E '/internal/meta/evidencequorum$|/internal/meta/evidencequorum/.*$' "$output/consumer-dependencies.txt" || true)"
test -z "$forbidden"
printf '%s\n' "$forbidden" > "$output/consumer-forbidden-dependencies.txt"

"$output/gooo" run --json --entry ProduceEvidence "$source" > "$output/source-execution-receipt.json"
"$output/gooo" run --json --entry ProduceEvidence "$source" > "$output/source-execution-replay.json"
cmp -s "$output/source-execution-receipt.json" "$output/source-execution-replay.json"
"$output/gooo" generate "$source" --out "$output/generated"
cp "$source" "$output/source.gooo"
cp "$policy" "$output/policy.gooo"

go list -deps ./cmd/gooo | sort -u > "$output/deps-gooo.txt"
go list -deps ./cmd/evidence-quorum-reconstructor | sort -u > "$output/deps-reconstructor.txt"
go list -deps ./cmd/evidence-quorum-artifact-observer | sort -u > "$output/deps-observer.txt"

"$output/source-channel" --receipt "$output/source-execution-receipt.json" --source "$source" --policy "$policy" \
  --head "$HEAD_SHA" --source-executable "$output/gooo" --dependencies "$output/deps-gooo.txt" \
  --out "$output/receipts/source-execution.json"
"$output/reconstructor" --source "$source" --policy "$policy" --head "$HEAD_SHA" \
  --dependencies "$output/deps-reconstructor.txt" --out "$output/receipts/raw-reconstruction.json"
"$output/artifact-observer" --source "$source" --policy "$policy" \
  --artifact "$output/generated/semantic.gooo.go" --manifest "$output/generated/semantic.gooo.manifest.jsonl" \
  --head "$HEAD_SHA" --dependencies "$output/deps-observer.txt" --out "$output/receipts/artifact-observation.json"

"$output/counterexample" --mode duplicate --input "$output/receipts/source-execution.json" --out "$output/receipts/synthetic-duplicate.json"
"$output/counterexample" --mode valid-conflict --input "$output/receipts/source-execution.json" --out "$output/receipts/synthetic-valid-conflict.json"
"$output/counterexample" --mode invalid-conflict --input "$output/receipts/source-execution.json" --out "$output/receipts/synthetic-invalid-conflict.json"
"$output/counterexample" --mode unknown --input "$output/receipts/source-execution.json" --out "$output/receipts/synthetic-unknown.json"

source_receipt="$output/receipts/source-execution.json"
reconstruction_receipt="$output/receipts/raw-reconstruction.json"
artifact_receipt="$output/receipts/artifact-observation.json"
duplicate_receipt="$output/receipts/synthetic-duplicate.json"
valid_conflict_receipt="$output/receipts/synthetic-valid-conflict.json"
invalid_conflict_receipt="$output/receipts/synthetic-invalid-conflict.json"
unknown_receipt="$output/receipts/synthetic-unknown.json"
case_spec="current-quorum=$source_receipt,$reconstruction_receipt,$artifact_receipt;synthetic-duplicate=$source_receipt,$reconstruction_receipt,$artifact_receipt,$duplicate_receipt;synthetic-valid-conflict=$source_receipt,$reconstruction_receipt,$artifact_receipt,$valid_conflict_receipt;synthetic-invalid-conflict=$source_receipt,$reconstruction_receipt,$artifact_receipt,$invalid_conflict_receipt;insufficient-current=$source_receipt,$reconstruction_receipt;synthetic-unknown=$source_receipt,$reconstruction_receipt,$artifact_receipt,$unknown_receipt"

"$output/evidence-quorum-witness" --policy "$policy" --source "$source" --head "$HEAD_SHA" \
  --source-path "$source" --cases "$case_spec" --out "$output/report.json"
"$output/evidence-quorum-witness" --check "$output/report.json"

jq -e '.decision == "PASS" and .resolution == "EXACT" and
  .summary.cases_satisfied == 6 and .summary.cases_total == 6 and
  .summary.claims_total == 6 and .summary.discharged_claims == 2 and
  .summary.open_claims == 3 and .summary.refuted_claims == 1 and
  .summary.raw_receipts_total == 21 and .summary.current_evidence_total == 3 and
  .summary.synthetic_evidence_total == 4 and .summary.distinct_provenance_groups == 3 and
  .summary.collapsed_replicas == 4 and .summary.conflict_cases == 1 and
  .summary.quorum_satisfied_cases == 2 and .summary.lower_resolution_cases == 3 and
  .summary.unknown_observation_cases == 1 and .summary.minimum_independent_groups == 3 and
  .summary.source_reconstruction_count == 1 and .summary.source_reconstruction_total == 1 and
  .summary.producer_package_imports == 0 and .summary.producer_package_import_total == 1 and
  .summary.confidence_aggregated == false and .summary.repository_writes == 0 and
  .summary.mutation_authority == false' "$output/report.json"
jq -e '([.cases[] | select(.status == "SATISFIED")] | length) == 6 and
  ([.cases[] | select(.id == "current-quorum") | .independent_groups] | .[0]) == 3 and
  ([.cases[] | select(.id == "synthetic-duplicate") | .collapsed_replicas] | .[0]) == 1 and
  ([.cases[] | select(.id == "synthetic-valid-conflict") | .subject_state] | .[0]) == "REFUTED" and
  ([.cases[] | select(.id == "synthetic-invalid-conflict") | .subject_state] | .[0]) == "OPEN" and
  ([.cases[] | select(.id == "synthetic-invalid-conflict") | .subject_decision] | .[0]) == "FAIL_CLOSED" and
  ([.cases[] | select(.id == "synthetic-unknown") | .subject_state] | .[0]) == "OPEN" and
  ([.cases[] | select(.id == "synthetic-unknown") | .observation_state] | .[0]) == "UNKNOWN" and
  ([.cases[] | select(.id == "synthetic-unknown") | .stage] | .[0]) == "UNKNOWN" and
  ([.cases[] | select(.id == "synthetic-unknown") | .step] | .[0]) == "UNKNOWN" and
  ([.cases[].claims[] | .transitions[]] | length) == 6 and
  ([.cases[].claims[] | .transitions[] | select(.previous_digest != "" and (.evidence_digests | length) > 0 and (.provenance | length) > 0)] | length) == 6' "$output/report.json"

baseline_semantic="$(jq -r '.source_semantic_digest' "$output/report.json")"
baseline_observation="$(jq -r '.cases[0].claims[0].transitions[0].provenance[0].observation_digest' "$output/report.json")"
threshold_policy="$output/policy-threshold.gooo"
sed 's/threshold=3/threshold=4/' "$policy" > "$threshold_policy"
"$output/source-channel" --receipt "$output/source-execution-receipt.json" --source "$source" --policy "$threshold_policy" \
  --head "$HEAD_SHA" --source-executable "$output/gooo" --dependencies "$output/deps-gooo.txt" \
  --out "$output/receipts/threshold-source-execution.json"
"$output/reconstructor" --source "$source" --policy "$threshold_policy" --head "$HEAD_SHA" \
  --dependencies "$output/deps-reconstructor.txt" --out "$output/receipts/threshold-raw-reconstruction.json"
"$output/artifact-observer" --source "$source" --policy "$threshold_policy" \
  --artifact "$output/generated/semantic.gooo.go" --manifest "$output/generated/semantic.gooo.manifest.jsonl" \
  --head "$HEAD_SHA" --dependencies "$output/deps-observer.txt" --out "$output/receipts/threshold-artifact-observation.json"
threshold_report="$output/threshold-intervention-report.json"
threshold_spec="current-quorum=$output/receipts/threshold-source-execution.json,$output/receipts/threshold-raw-reconstruction.json,$output/receipts/threshold-artifact-observation.json"
"$output/evidence-quorum-witness" --policy "$threshold_policy" --source "$source" --head "$HEAD_SHA" \
  --source-path "$source" --cases "$threshold_spec" --out "$threshold_report" || true
threshold_decision="$(jq -r '.decision' "$threshold_report")"
threshold_semantic="$(jq -r '.policy_semantic_digest' "$threshold_report")"
threshold_observation="$(jq -r '.cases[0].claims[0].transitions[0].provenance[0].observation_digest // ""' "$threshold_report")"

comment_policy="$output/policy-comment-only.gooo"
sed '1i// comment-only intervention: presentation text is not semantic policy\n' "$policy" > "$comment_policy"
comment_report="$output/comment-intervention-report.json"
"$output/evidence-quorum-witness" --policy "$comment_policy" --source "$source" --head "$HEAD_SHA" \
  --source-path "$source" --cases "$case_spec" --out "$comment_report"
comment_semantic="$(jq -r '.source_semantic_digest' "$comment_report")"
comment_observation="$(jq -r '.cases[0].claims[0].transitions[0].provenance[0].observation_digest' "$comment_report")"
comment_decision="$(jq -r '.decision' "$comment_report")"

jq -n --arg before "$baseline_semantic" --arg after "$threshold_semantic" --arg beforeObs "$baseline_observation" --arg afterObs "$threshold_observation" \
  --arg thresholdDecision "$threshold_decision" --arg commentDecision "$comment_decision" --arg commentSemantic "$comment_semantic" --arg commentObservation "$comment_observation" \
  --arg reportDecision "$(jq -r '.decision' "$output/report.json")" \
  '{threshold_changed:{before_decision:$reportDecision,after_decision:$thresholdDecision,before_semantic_digest:$before,after_semantic_digest:$after,before_observation_digest:$beforeObs,after_observation_digest:$afterObs,quorum_result_changed:($thresholdDecision != $reportDecision),semantic_digest_changed:($before != $after),effects_before:{repository_writes:0,mutation_authority:false},effects_after:{repository_writes:0,mutation_authority:false}},comment_only:{before_decision:$reportDecision,after_decision:$commentDecision,before_semantic_digest:$before,after_semantic_digest:$commentSemantic,before_observation_digest:$beforeObs,after_observation_digest:$commentObservation,quorum_result_changed:($reportDecision != $commentDecision),semantic_digest_changed:($before != $commentSemantic),effects_before:{repository_writes:0,mutation_authority:false},effects_after:{repository_writes:0,mutation_authority:false}}}' > "$output/interventions.json"
jq -e '.threshold_changed.quorum_result_changed and .threshold_changed.semantic_digest_changed and
  (.threshold_changed.effects_before.repository_writes == 0) and (.threshold_changed.effects_after.repository_writes == 0) and
  (.comment_only.quorum_result_changed == false) and (.comment_only.semantic_digest_changed == false) and
  (.comment_only.effects_before.repository_writes == 0) and (.comment_only.effects_after.repository_writes == 0)' "$output/interventions.json"

{
  jq -r '"- cases: \(.summary.cases_satisfied)/\(.summary.cases_total)\n- current evidence: \(.summary.current_evidence_total)/\(.summary.current_evidence_total)\n- synthetic counterexamples: \(.summary.synthetic_evidence_total)\n- raw receipts: \(.summary.raw_receipts_total)\n- distinct provenance groups: \(.summary.distinct_provenance_groups)\n- collapsed replicas: \(.summary.collapsed_replicas)\n- sufficient quorum: \(.cases[0].independent_groups)/\(.summary.minimum_independent_groups)\n- source reconstruction: \(.summary.source_reconstruction_count)/\(.summary.source_reconstruction_total)\n- consumer producer-package imports: \(.summary.producer_package_imports)/\(.summary.producer_package_import_total)\n- conflicts REFUTED only by valid predicate: \(.summary.conflict_cases)\n- unknown observations: \(.summary.unknown_observation_cases)\n- claim transitions: \([.cases[].claims[].transitions[]] | length)\n- report: \(.digest)"' "$output/report.json"
} >> "${GITHUB_STEP_SUMMARY:-/dev/null}"

sha256sum "$output"/report.json "$output"/interventions.json "$output"/source-execution-receipt.json \
  "$output"/source-execution-replay.json "$output"/source.gooo "$output"/policy.gooo \
  "$output"/receipts/*.json "$output"/cases/*.json "$output"/generated/* 2>/dev/null > "$output/manifest.sha256" || true
