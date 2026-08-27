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
producer_package="github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentportfolio"
consumer_package="github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentportfolio/causalityconsumer"
contract="$root/examples/experiment-portfolio/contract.json"
manifest="$root/examples/experiment-portfolio/causality-manifest.json"
sources=(derive replay reflect)
mkdir -p "$work" "$build"

go build -trimpath -o "$gooo" ./cmd/gooo
go build -trimpath -o "$producer" ./cmd/experiment-portfolio-receipt
go build -trimpath -o "$evaluator" ./cmd/experiment-portfolio-evaluate
go build -trimpath -o "$causal_source" ./cmd/experiment-portfolio-causal-source
go build -trimpath -o "$causal_evaluator" ./cmd/experiment-portfolio-causality

consumer_deps="$(go list -deps ./cmd/experiment-portfolio-causality)"
forbidden_producer_deps="$(printf '%s\n' "$consumer_deps" | grep -Fx "$producer_package" || true)"
if [[ -n "$forbidden_producer_deps" ]]; then
	echo "causality consumer imports producer implementation: $forbidden_producer_deps" >&2
	exit 1
fi
producer_deps="$(go list -deps ./cmd/experiment-portfolio-receipt)"
forbidden_consumer_deps="$(printf '%s\n' "$producer_deps" | grep -Fx "$consumer_package" || true)"
if [[ -n "$forbidden_consumer_deps" ]]; then
	echo "receipt producer imports causality consumer implementation: $forbidden_consumer_deps" >&2
	exit 1
fi
echo "forbidden producer deps observed=0 allowed_max=0"
echo "independence contract=1/1"

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

observer_comment_source="$work/source-observer-comment-prefix.gooo"
observer_quoted_source="$work/source-observer-unrelated-quoted-prefix.gooo"
cat > "$observer_comment_source" <<'EOF'
package observercomment
namespace observercomment

// computes "fake-comment"
entity Source id "gooo://observer/comment/source"
entity Target id "gooo://observer/comment/target"
activity Observe(Source) -> Target computes "observer.canonical.comment"
EOF
cat > "$observer_quoted_source" <<'EOF'
package observerquoted
namespace observerquoted

// unrelated quoted string: "computes \"fake-string\""
entity Source id "gooo://observer/quoted/source"
entity Target id "gooo://observer/quoted/target"
activity Observe(Source) -> Target computes "observer.canonical.string"
EOF
"$causal_source" -source "$observer_comment_source" -output "$work/source-observer-comment.json"
"$causal_source" -source "$observer_quoted_source" -output "$work/source-observer-unrelated-quoted.json"
jq -e '.semantic_value == "observer.canonical.comment" and (.source_digest | startswith("sha256:"))' "$work/source-observer-comment.json"
jq -e '.semantic_value == "observer.canonical.string" and (.source_digest | startswith("sha256:"))' "$work/source-observer-unrelated-quoted.json"
echo "source-observer conformance: 2/2"

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
  (.candidates | all(.coordinate_vector[0].denominator == 1 and .coordinate_vector[1].denominator == 1 and .coordinate_vector[2].denominator == 2 and .coordinate_vector[3].denominator == 2 and .coordinate_vector[4].denominator == 1 and .coordinate_vector[5].denominator == 1 and .coordinate_vector[6].id == "source-semantic-causality" and .coordinate_vector[6].denominator == 3 and .coordinate_vector[6].status == "REFUTED" and .coordinate_vector[6].reason == "three semantic interventions are directly refuted as digest-only bindings")) and
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
  .transition_summary.fixed_denominator == 9 and
  .transition_summary.refuted == {numerator:3,denominator:9} and
  .transition_summary.discharged == {numerator:6,denominator:9} and
  .transition_summary.open == {numerator:0,denominator:9} and
  .transition_summary.reason == "three operations x three intervention claims" and
  (.samples | length) == 3 and
  (.samples | all(
    .baseline.status == "DISCHARGED" and
    .baseline.claim_transitions[0].from == "OPEN" and
    .baseline.claim_transitions[0].to == "DISCHARGED" and
    .baseline.claim_transitions[0].reason == "SOURCE_OBSERVATION_BOUND" and
    .semantic.status == "REFUTED" and
    .semantic.reason == "DIGEST_ONLY_BINDING" and
    .semantic.claim_transitions[0].from == "OPEN" and
    .semantic.claim_transitions[0].to == "REFUTED" and
    .semantic.claim_transitions[0].reason == "DIGEST_ONLY_BINDING" and
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
    .nonsemantic.claim_transitions[0].from == "OPEN" and
    .nonsemantic.claim_transitions[0].to == "DISCHARGED" and
    .nonsemantic.claim_transitions[0].reason == "SEMANTIC_PROJECTION_PRESERVED" and
    .nonsemantic.source_digest != .baseline.source_digest and
    .nonsemantic.semantic_value == .baseline.semantic_value and
    (.baseline.claim_transitions | length) == 1 and
    (.semantic.claim_transitions | length) == 1 and
    (.nonsemantic.claim_transitions | length) == 1 and
    (.changed_fields | length) == 0
  )) and
  (has("score") | not) and (has("winner") | not)
' "$work/causality-report.json"

jq '.samples[0].nonsemantic = (.samples[0].baseline | .case_id = "derive-nonsemantic" | .kind = "NON_SEMANTIC")' \
	"$work/causality-input.json" > "$work/causality-unknown-input.json"
if "$causal_evaluator" -input "$work/causality-unknown-input.json" -output "$work/causality-unknown-report.json"; then
	echo "unknown causality input unexpectedly passed" >&2
	exit 1
fi
jq -e '
  .decision == "FAIL_CLOSED" and
  .resolution == "LOWER_RESOLUTION" and
  .reason == "CAUSALITY_INPUT_UNKNOWN" and
  .summary.unknowns == 1 and
  .unknown_findings == [{candidate_id:"derive",case_id:"derive-nonsemantic",stage:"NON_SEMANTIC_INTERVENTION",step:"observe-source",reason:"NON_SEMANTIC_DIGEST_UNCHANGED"}]
' "$work/causality-unknown-report.json"

jq '(.samples[0].semantic.observation.source_digest = .samples[0].baseline.observation.source_digest) | (.samples[0].semantic.receipt = .samples[0].baseline.receipt)' \
	"$work/causality-input.json" > "$work/causality-semantic-only-unknown-input.json"
if "$causal_evaluator" -input "$work/causality-semantic-only-unknown-input.json" -output "$work/causality-semantic-only-unknown-report.json"; then
	echo "semantic-only unknown causality input unexpectedly passed" >&2
	exit 1
fi
jq -e '
  .decision == "FAIL_CLOSED" and
  .resolution == "LOWER_RESOLUTION" and
  .reason == "CAUSALITY_INPUT_UNKNOWN" and
  .summary.unknowns == 1 and
  .transition_summary.fixed_denominator == 9 and
  .transition_summary.refuted == {numerator:2,denominator:9} and
  .transition_summary.discharged == {numerator:6,denominator:9} and
  .transition_summary.open == {numerator:1,denominator:9} and
  .unknown_findings == [{candidate_id:"derive",case_id:"derive-semantic",stage:"SEMANTIC_INTERVENTION",step:"observe-source",reason:"SEMANTIC_DIGEST_UNCHANGED"}] and
  .samples[0].baseline.status == "DISCHARGED" and
  .samples[0].baseline.claim_transitions[0].to == "DISCHARGED" and
  .samples[0].semantic.status == "UNKNOWN" and
  .samples[0].semantic.claim_transitions[0] == {id:"derive-semantic-claim",from:"OPEN",to:"OPEN",stage:"SEMANTIC_INTERVENTION",step:"observe-source",reason:"SEMANTIC_DIGEST_UNCHANGED"} and
  .samples[0].nonsemantic.status == "DISCHARGED" and
  .samples[0].nonsemantic.claim_transitions[0].to == "DISCHARGED" and
  .samples[0].nonsemantic.claim_transitions[0].reason == "SEMANTIC_PROJECTION_PRESERVED"
' "$work/causality-semantic-only-unknown-report.json"

jq '(.samples[0].nonsemantic.observation.source_digest = .samples[0].baseline.observation.source_digest) | (.samples[0].nonsemantic.receipt = .samples[0].baseline.receipt)' \
	"$work/causality-input.json" > "$work/causality-nonsemantic-only-unknown-input.json"
if "$causal_evaluator" -input "$work/causality-nonsemantic-only-unknown-input.json" -output "$work/causality-nonsemantic-only-unknown-report.json"; then
	echo "nonsemantic-only unknown causality input unexpectedly passed" >&2
	exit 1
fi
jq -e '
  .decision == "FAIL_CLOSED" and
  .resolution == "LOWER_RESOLUTION" and
  .reason == "CAUSALITY_INPUT_UNKNOWN" and
  .summary.unknowns == 1 and
  .transition_summary.fixed_denominator == 9 and
  .transition_summary.refuted == {numerator:3,denominator:9} and
  .transition_summary.discharged == {numerator:5,denominator:9} and
  .transition_summary.open == {numerator:1,denominator:9} and
  .unknown_findings == [{candidate_id:"derive",case_id:"derive-nonsemantic",stage:"NON_SEMANTIC_INTERVENTION",step:"observe-source",reason:"NON_SEMANTIC_DIGEST_UNCHANGED"}] and
  .samples[0].semantic.status == "REFUTED" and
  .samples[0].semantic.claim_transitions[0].to == "REFUTED" and
  .samples[0].semantic.claim_transitions[0].reason == "DIGEST_ONLY_BINDING" and
  .samples[0].nonsemantic.status == "UNKNOWN" and
  .samples[0].nonsemantic.claim_transitions[0] == {id:"derive-nonsemantic-claim",from:"OPEN",to:"OPEN",stage:"NON_SEMANTIC_INTERVENTION",step:"observe-source",reason:"NON_SEMANTIC_DIGEST_UNCHANGED"}
' "$work/causality-nonsemantic-only-unknown-report.json"

jq '(.samples[0].semantic.observation.source_digest = .samples[0].baseline.observation.source_digest) | (.samples[0].semantic.receipt = .samples[0].baseline.receipt) | (.samples[0].nonsemantic.observation.source_digest = .samples[0].baseline.observation.source_digest) | (.samples[0].nonsemantic.receipt = .samples[0].baseline.receipt)' \
	"$work/causality-input.json" > "$work/causality-both-unknown-input.json"
if "$causal_evaluator" -input "$work/causality-both-unknown-input.json" -output "$work/causality-both-unknown-report.json"; then
	echo "both-unknown causality input unexpectedly passed" >&2
	exit 1
fi
jq -e '
  .decision == "FAIL_CLOSED" and
  .resolution == "LOWER_RESOLUTION" and
  .reason == "CAUSALITY_INPUT_UNKNOWN" and
  .summary.unknowns == 2 and
  .transition_summary.fixed_denominator == 9 and
  .transition_summary.refuted == {numerator:2,denominator:9} and
  .transition_summary.discharged == {numerator:5,denominator:9} and
  .transition_summary.open == {numerator:2,denominator:9} and
  .unknown_findings == [
    {candidate_id:"derive",case_id:"derive-semantic",stage:"SEMANTIC_INTERVENTION",step:"observe-source",reason:"SEMANTIC_DIGEST_UNCHANGED"},
    {candidate_id:"derive",case_id:"derive-nonsemantic",stage:"NON_SEMANTIC_INTERVENTION",step:"observe-source",reason:"NON_SEMANTIC_DIGEST_UNCHANGED"}
  ] and
  .samples[0].baseline.status == "DISCHARGED" and
  .samples[0].semantic.status == "UNKNOWN" and
  .samples[0].semantic.claim_transitions[0].to == "OPEN" and
  .samples[0].semantic.claim_transitions[0].reason == "SEMANTIC_DIGEST_UNCHANGED" and
  .samples[0].nonsemantic.status == "UNKNOWN" and
  .samples[0].nonsemantic.claim_transitions[0].to == "OPEN" and
  .samples[0].nonsemantic.claim_transitions[0].reason == "NON_SEMANTIC_DIGEST_UNCHANGED"
' "$work/causality-both-unknown-report.json"

jq '.manifest.cases[0].required_change_fields = ["decision"]' \
	"$work/causality-input.json" > "$work/causality-subset-input.json"
"$causal_evaluator" -input "$work/causality-subset-input.json" -output "$work/causality-subset-report.json"
jq -e '
  .decision == "REFUTED" and
  .resolution == "EXACT" and
  .samples[0].required_change_fields == ["decision"] and
  .samples[0].semantic.status == "REFUTED" and
  .samples[0].semantic.reason == "DIGEST_ONLY_BINDING"
' "$work/causality-subset-report.json"

jq '.manifest.cases[0].required_change_fields = []' \
	"$work/causality-input.json" > "$work/causality-empty-subset-input.json"
if "$causal_evaluator" -input "$work/causality-empty-subset-input.json" -output "$work/causality-empty-subset-report.json"; then
	echo "empty required_change_fields unexpectedly passed" >&2
	exit 1
fi
jq -e '.decision == "FAIL_CLOSED" and .resolution == "LOWER_RESOLUTION" and .reason == "CAUSALITY_MANIFEST_CASE_INVALID"' \
	"$work/causality-empty-subset-report.json"

jq '.samples[0].semantic.receipt.coordinate_vector[0].numerator = 0' "$work/causality-input.json" > "$work/causality-forged-input.json"
if "$causal_evaluator" -input "$work/causality-forged-input.json" -output "$work/causality-forged-report.json"; then
	echo "forged causality receipt unexpectedly passed" >&2
	exit 1
fi
jq -e '.decision == "FAIL_CLOSED" and .reason == "PORTFOLIO_RECEIPT_DIGEST_INVALID" and .summary.unknowns == 1' "$work/causality-forged-report.json"

jq 'del(.samples[0].semantic.claim_transitions[0])' "$work/causality-report.json" > "$work/causality-transition-deleted-report.json"
if "$causal_evaluator" -input "$work/causality-input.json" -check "$work/causality-transition-deleted-report.json"; then
	echo "deleted causality transition unexpectedly passed" >&2
	exit 1
fi

jq '.samples[0].semantic.status = "OPEN" | .samples[0].semantic.claim_transitions[0].to = "OPEN"' "$work/causality-report.json" > "$work/causality-open-laundered-report.json"
if "$causal_evaluator" -input "$work/causality-input.json" -check "$work/causality-open-laundered-report.json"; then
	echo "open causality status laundering unexpectedly passed" >&2
	exit 1
fi

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
  "- transitions: refuted=\(.transition_summary.refuted.numerator)/\(.transition_summary.refuted.denominator), discharged=\(.transition_summary.discharged.numerator)/\(.transition_summary.discharged.denominator), open=\(.transition_summary.open.numerator)/\(.transition_summary.open.denominator)",
  ("- claim transitions: " + ([.samples[] | "\(.candidate_id): baseline=\(.baseline.claim_transitions|length), semantic=\(.semantic.claim_transitions|length), nonsemantic=\(.nonsemantic.claim_transitions|length)"] | join("; "))),
  ("- semantic intervention: " + ([.samples[] | "\(.candidate_id)=\(.semantic.status)/\(.semantic.reason)"] | join("; "))),
  ("- non-semantic intervention: " + ([.samples[] | "\(.candidate_id)=\(.nonsemantic.status)/\(.nonsemantic.reason)"] | join("; ")))
' "$work/causality-report.json" >> "$GITHUB_STEP_SUMMARY"
{
	echo "- forbidden producer deps observed=0 allowed_max=0"
	echo "- independence contract=1/1"
	echo "- unknown regressions: semantic-only, nonsemantic-only, and both preserve independent claim transitions"
	echo "- required_change_fields: non-empty arbitrary subset accepted; empty subset rejected"
	echo "- source-observer conformance: 2/2 via parser/lowering semantic projection"
} >> "$GITHUB_STEP_SUMMARY"
