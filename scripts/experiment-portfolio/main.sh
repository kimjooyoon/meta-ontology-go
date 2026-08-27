#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
work="${RUNNER_TEMP:-/tmp}/experiment-portfolio"
build="${RUNNER_TEMP:-/tmp}/experiment-portfolio-build"
gooo="$build/gooo"
producer="$build/experiment-portfolio-receipt"
evaluator="$build/experiment-portfolio-evaluate"
contract="$root/examples/experiment-portfolio/contract.json"
sources=(derive replay reflect)
mkdir -p "$work" "$build"

go build -trimpath -o "$gooo" ./cmd/gooo
go build -trimpath -o "$producer" ./cmd/experiment-portfolio-receipt
go build -trimpath -o "$evaluator" ./cmd/experiment-portfolio-evaluate

for candidate in "${sources[@]}"; do
	"$gooo" check --semantic --json "$root/examples/experiment-portfolio/alternatives/$candidate.gooo" > "$work/$candidate-check.json"
	"$producer" -candidate "$candidate" -subject-sha "$HEAD_SHA" \
		-source "$root/examples/experiment-portfolio/alternatives/$candidate.gooo" \
		-output "$work/$candidate-first.json"
	"$producer" -candidate "$candidate" -subject-sha "$HEAD_SHA" \
		-source "$root/examples/experiment-portfolio/alternatives/$candidate.gooo" \
		-output "$work/$candidate-replay.json"
	cmp "$work/$candidate-first.json" "$work/$candidate-replay.json"
done

jq -n --arg subject "$HEAD_SHA" --slurpfile contract "$contract" \
	--slurpfile derive "$work/derive-first.json" --slurpfile replay "$work/replay-first.json" \
	--slurpfile reflect "$work/reflect-first.json" \
	'{subject_sha:$subject,contract:$contract[0],receipts:[$derive[0],$replay[0],$reflect[0]]}' > "$work/input.json"

"$evaluator" -input "$work/input.json" -output "$work/report.json"
"$evaluator" -input "$work/input.json" -check "$work/report.json"

jq -e '
  .decision == "PORTFOLIO_PRESERVED" and
  .resolution == "EXACT" and
  .reason == "NO_WINNER_NO_AGGREGATE" and
  .summary.candidates == 3 and
  .summary.coordinates_per_candidate == 6 and
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
  (has("score") | not) and (has("winner") | not)
' "$work/report.json"

jq '.receipts[0].coordinate_vector[0].numerator = 0' "$work/input.json" > "$work/forged-input.json"
if "$evaluator" -input "$work/forged-input.json" -output "$work/forged-report.json"; then
	echo "forged receipt unexpectedly passed" >&2
	exit 1
fi
jq -e '.decision == "FAIL_CLOSED" and .reason == "PORTFOLIO_RECEIPT_DIGEST_INVALID" and .summary.unknowns == 1' "$work/forged-report.json"

jq -r '
  "### Meta-programming experiment portfolio",
  "- decision: \(.decision) / \(.resolution)",
  "- candidates: \(.summary.candidates)",
  "- vectors: \(.summary.coordinates_per_candidate) fixed coordinates per candidate",
  "- counterexamples: derive=\(.summary.counterexample_counts.derive), replay=\(.summary.counterexample_counts.replay), reflect=\(.summary.counterexample_counts.reflect)",
  "- unknown locations: derive=\(.summary.unknown_location_ids.derive|length), replay=\(.summary.unknown_location_ids.replay|length), reflect=\(.summary.unknown_location_ids.reflect|length)",
  "- aggregate score/winner: not emitted",
  "- receipt: \(.digest)"
' "$work/report.json" >> "$GITHUB_STEP_SUMMARY"
