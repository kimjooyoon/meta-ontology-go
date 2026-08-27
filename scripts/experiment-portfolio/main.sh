#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
work="${RUNNER_TEMP:-/tmp}/experiment-portfolio"
build="${RUNNER_TEMP:-/tmp}/experiment-portfolio-build"
gooo="$build/gooo"
producer="$build/experiment-portfolio-receipt"
evaluator="$build/experiment-portfolio-evaluate"
causal_source="$build/experiment-portfolio-causal-source"
causal_evaluator="$build/experiment-portfolio-causality"
contract="$root/examples/experiment-portfolio/contract.json"
manifest="$root/examples/experiment-portfolio/causality-manifest.json"
sources=(derive replay reflect)
mkdir -p "$work" "$build"

go build -trimpath -o "$gooo" ./cmd/gooo
go build -trimpath -o "$producer" ./cmd/experiment-portfolio-receipt
go build -trimpath -o "$evaluator" ./cmd/experiment-portfolio-evaluate
go build -trimpath -o "$causal_source" ./cmd/experiment-portfolio-causal-source
go build -trimpath -o "$causal_evaluator" ./cmd/experiment-portfolio-causality

for candidate in "${sources[@]}"; do
	source="$root/examples/experiment-portfolio/alternatives/$candidate.gooo"
	semantic_source="$work/semantic/examples/experiment-portfolio/alternatives/$candidate.gooo"
	nonsemantic_source="$work/nonsemantic/examples/experiment-portfolio/alternatives/$candidate.gooo"
	mkdir -p "$(dirname "$semantic_source")" "$(dirname "$nonsemantic_source")"
	case "$candidate" in
		derive)
			before="meta.portfolio.derive-coordinate"
			after="meta.portfolio.derive-coordinate:semantic-intervention"
			comment="source-semantic-causality non-semantic derive"
			;;
		replay)
			before="meta.portfolio.replay-independent"
			after="meta.portfolio.replay-independent:semantic-intervention"
			comment="source-semantic-causality non-semantic replay"
			;;
		reflect)
			before="meta.portfolio.reflect-counterexample"
			after="meta.portfolio.reflect-counterexample:semantic-intervention"
			comment="source-semantic-causality non-semantic reflect"
			;;
		*)
			echo "unknown candidate $candidate" >&2
			exit 1
		;;
	esac
	sed "s/computes \"$before\"/computes \"$after\"/" "$source" > "$semantic_source"
		{
			printf '// %s\n\n' "$comment"
			cat "$source"
		} > "$nonsemantic_source"
	"$gooo" check --semantic --json "$source" > "$work/$candidate-baseline-check.json"
	"$gooo" check --semantic --json "$semantic_source" > "$work/$candidate-semantic-check.json"
	"$gooo" check --semantic --json "$nonsemantic_source" > "$work/$candidate-nonsemantic-check.json"

	"$producer" -candidate "$candidate" -subject-sha "$HEAD_SHA" \
		-source "$source" \
		-output "$work/$candidate-baseline-first.json"
	"$producer" -candidate "$candidate" -subject-sha "$HEAD_SHA" \
		-source "$source" \
		-output "$work/$candidate-baseline-replay.json"
	cmp "$work/$candidate-baseline-first.json" "$work/$candidate-baseline-replay.json"

	"$causal_source" -source "$source" -output "$work/$candidate-baseline-observation.json"
	"$causal_source" -source "$semantic_source" -output "$work/$candidate-semantic-observation.json"
	"$causal_source" -source "$nonsemantic_source" -output "$work/$candidate-nonsemantic-observation.json"
	"$producer" -candidate "$candidate" -subject-sha "$HEAD_SHA" \
		-source "$semantic_source" \
		-output "$work/$candidate-semantic-first.json"
	"$producer" -candidate "$candidate" -subject-sha "$HEAD_SHA" \
		-source "$semantic_source" \
		-output "$work/$candidate-semantic-replay.json"
	cmp "$work/$candidate-semantic-first.json" "$work/$candidate-semantic-replay.json"
	"$producer" -candidate "$candidate" -subject-sha "$HEAD_SHA" \
		-source "$nonsemantic_source" \
		-output "$work/$candidate-nonsemantic-first.json"
	"$producer" -candidate "$candidate" -subject-sha "$HEAD_SHA" \
		-source "$nonsemantic_source" \
		-output "$work/$candidate-nonsemantic-replay.json"
	cmp "$work/$candidate-nonsemantic-first.json" "$work/$candidate-nonsemantic-replay.json"
done

jq -n --arg subject "$HEAD_SHA" --slurpfile contract "$contract" \
	--slurpfile derive "$work/derive-baseline-first.json" --slurpfile replay "$work/replay-baseline-first.json" \
	--slurpfile reflect "$work/reflect-baseline-first.json" \
	'{subject_sha:$subject,contract:$contract[0],receipts:[$derive[0],$replay[0],$reflect[0]]}' > "$work/input.json"

"$evaluator" -input "$work/input.json" -output "$work/report.json"
"$evaluator" -input "$work/input.json" -check "$work/report.json"

jq -e '
  .decision == "PORTFOLIO_PRESERVED" and
  .resolution == "EXACT" and
  .reason == "NO_WINNER_NO_AGGREGATE" and
  .summary.candidates == 3 and
  .summary.coordinates_per_candidate == 7 and
  .summary.counterexample_counts == {derive:2,replay:1,reflect:0} and
  (.summary.unknown_location_ids.derive|length) == 2 and
  (.summary.unknown_location_ids.replay|length) == 1 and
  (.summary.unknown_location_ids.reflect|length) == 0 and
  .summary.repository_writes == 0 and
  .summary.mutation_authority == false and
  (.candidates | all(.receipt.coordinate_vector == .coordinate_vector)) and
  (.candidates | all(.receipt.counterexamples == .counterexamples)) and
  ([.candidates[].coordinate_vector[].status] | index("OPEN")) != null and
  ([.candidates[].coordinate_vector[].status] | index("REFUTED")) != null and
  ([.candidates[].coordinate_vector[].status] | index("DISCHARGED")) != null and
  ([.proofs[] | select(.choice == "NO_AGGREGATION") | .passed] | . == [true]) and
  (.candidates | all(.coordinate_vector[0].denominator == 1 and .coordinate_vector[1].denominator == 1 and .coordinate_vector[2].denominator == 2 and .coordinate_vector[3].denominator == 2 and .coordinate_vector[4].denominator == 1 and .coordinate_vector[5].denominator == 1 and .coordinate_vector[6].id == "source-semantic-causality" and .coordinate_vector[6].denominator == 3 and .coordinate_vector[6].status == "OPEN")) and
  (has("score") | not) and (has("winner") | not)
' "$work/report.json"

jq '.receipts[0].coordinate_vector[0].numerator = 0' "$work/input.json" > "$work/forged-input.json"
if "$evaluator" -input "$work/forged-input.json" -output "$work/forged-report.json"; then
	echo "forged receipt unexpectedly passed" >&2
	exit 1
fi
jq -e '.decision == "FAIL_CLOSED" and .reason == "PORTFOLIO_RECEIPT_DIGEST_INVALID" and .summary.unknowns == 1' "$work/forged-report.json"

jq -n --arg subject "$HEAD_SHA" --slurpfile contract "$contract" --slurpfile manifest "$manifest" \
	--slurpfile derive_baseline_observation "$work/derive-baseline-observation.json" \
	--slurpfile derive_semantic_observation "$work/derive-semantic-observation.json" \
	--slurpfile derive_nonsemantic_observation "$work/derive-nonsemantic-observation.json" \
	--slurpfile derive_baseline_receipt "$work/derive-baseline-first.json" \
	--slurpfile derive_semantic_receipt "$work/derive-semantic-first.json" \
	--slurpfile derive_nonsemantic_receipt "$work/derive-nonsemantic-first.json" \
	--slurpfile replay_baseline_observation "$work/replay-baseline-observation.json" \
	--slurpfile replay_semantic_observation "$work/replay-semantic-observation.json" \
	--slurpfile replay_nonsemantic_observation "$work/replay-nonsemantic-observation.json" \
	--slurpfile replay_baseline_receipt "$work/replay-baseline-first.json" \
	--slurpfile replay_semantic_receipt "$work/replay-semantic-first.json" \
	--slurpfile replay_nonsemantic_receipt "$work/replay-nonsemantic-first.json" \
	--slurpfile reflect_baseline_observation "$work/reflect-baseline-observation.json" \
	--slurpfile reflect_semantic_observation "$work/reflect-semantic-observation.json" \
	--slurpfile reflect_nonsemantic_observation "$work/reflect-nonsemantic-observation.json" \
	--slurpfile reflect_baseline_receipt "$work/reflect-baseline-first.json" \
	--slurpfile reflect_semantic_receipt "$work/reflect-semantic-first.json" \
	--slurpfile reflect_nonsemantic_receipt "$work/reflect-nonsemantic-first.json" \
	'{subject_sha:$subject,contract:$contract[0],manifest:$manifest[0],samples:[
		{candidate_id:"derive",baseline:{case_id:"derive-baseline",kind:"BASELINE",observation:$derive_baseline_observation[0],receipt:$derive_baseline_receipt[0]},semantic:{case_id:"derive-semantic",kind:"SEMANTIC",observation:$derive_semantic_observation[0],receipt:$derive_semantic_receipt[0]},nonsemantic:{case_id:"derive-nonsemantic",kind:"NON_SEMANTIC",observation:$derive_nonsemantic_observation[0],receipt:$derive_nonsemantic_receipt[0]}},
		{candidate_id:"replay",baseline:{case_id:"replay-baseline",kind:"BASELINE",observation:$replay_baseline_observation[0],receipt:$replay_baseline_receipt[0]},semantic:{case_id:"replay-semantic",kind:"SEMANTIC",observation:$replay_semantic_observation[0],receipt:$replay_semantic_receipt[0]},nonsemantic:{case_id:"replay-nonsemantic",kind:"NON_SEMANTIC",observation:$replay_nonsemantic_observation[0],receipt:$replay_nonsemantic_receipt[0]}},
		{candidate_id:"reflect",baseline:{case_id:"reflect-baseline",kind:"BASELINE",observation:$reflect_baseline_observation[0],receipt:$reflect_baseline_receipt[0]},semantic:{case_id:"reflect-semantic",kind:"SEMANTIC",observation:$reflect_semantic_observation[0],receipt:$reflect_semantic_receipt[0]},nonsemantic:{case_id:"reflect-nonsemantic",kind:"NON_SEMANTIC",observation:$reflect_nonsemantic_observation[0],receipt:$reflect_nonsemantic_receipt[0]}}
	]}' > "$work/causality-input.json"

"$causal_evaluator" -input "$work/causality-input.json" -output "$work/causality-report.json"
"$causal_evaluator" -input "$work/causality-input.json" -check "$work/causality-report.json"

jq -e '
  .decision == "REFUTED" and
  .resolution == "EXACT" and
  .reason == "DIGEST_ONLY_BINDING" and
  .summary.causal_cases == {observed:0,total:3} and
  .summary.digest_only_cases == 3 and
  .summary.hardcoded_fixture_cases == 3 and
  .summary.unknowns == 0 and
  (.samples | length) == 3 and
  (.samples | all(
    .semantic.status == "REFUTED" and
    .semantic.reason == "DIGEST_ONLY_BINDING" and
    .source_semantic_value_changed == true and
    .source_digest_changed == true and
    .nonsemantic_source_digest_changed == true and
    .nonsemantic_semantic_value_preserved == true and
    .semantic_projection_changed == false and
    .decision_changed == false and
    .claim_transitions_changed == false and
    .nonsemantic_decision_changed == false and
    .hardcoded_fixture == true and
    .nonsemantic.status == "DISCHARGED" and
    .nonsemantic.reason == "SEMANTIC_PROJECTION_PRESERVED" and
    .nonsemantic.source_digest != .baseline.source_digest and
    .nonsemantic.semantic_value == .baseline.semantic_value and
    .nonsemantic.claim_transitions == .baseline.claim_transitions and
    .semantic.claim_transitions == .baseline.claim_transitions and
    (.changed_fields | length) == 0
  )) and
  (has("score") | not) and (has("winner") | not)
' "$work/causality-report.json"

jq -r '
  "### Meta-programming experiment portfolio",
  "- decision: \(.decision) / \(.resolution)",
  "- candidates: \(.summary.candidates)",
  "- vectors: \(.summary.coordinates_per_candidate) fixed coordinates per candidate (v1 six denominators preserved; v2 adds source-semantic-causality 3-case denominator)",
  "- counterexamples: derive=\(.summary.counterexample_counts.derive), replay=\(.summary.counterexample_counts.replay), reflect=\(.summary.counterexample_counts.reflect)",
  "- unknown locations: derive=\(.summary.unknown_location_ids.derive|length), replay=\(.summary.unknown_location_ids.replay|length), reflect=\(.summary.unknown_location_ids.reflect|length)",
  "- aggregate score/winner: not emitted",
  "- receipt: \(.digest)"
' "$work/report.json" >> "$GITHUB_STEP_SUMMARY"
jq -r '
  "",
  "### Source-semantic causality audit",
  "- causal_cases: \(.summary.causal_cases.observed)/\(.summary.causal_cases.total)",
  "- digest_only_cases: \(.summary.digest_only_cases)",
  "- hardcoded_fixture_cases: \(.summary.hardcoded_fixture_cases)",
  "- unknowns: \(.summary.unknowns)",
  ("- claim transitions: " + ([.samples[] | "\(.candidate_id): baseline=\(.baseline.claim_transitions|length), semantic=\(.semantic.claim_transitions|length), nonsemantic=\(.nonsemantic.claim_transitions|length)"] | join("; "))),
  ("- semantic intervention: " + ([.samples[] | "\(.candidate_id)=\(.semantic.status)/\(.semantic.reason)"] | join("; "))),
  ("- non-semantic intervention: " + ([.samples[] | "\(.candidate_id)=\(.nonsemantic.status)/\(.nonsemantic.reason)"] | join("; ")))
' "$work/causality-report.json" >> "$GITHUB_STEP_SUMMARY"
