#!/usr/bin/env bash
set -euo pipefail

: "$HEAD_SHA"
work="${RUNNER_TEMP:-/tmp}/denominator-evolution-$HEAD_SHA"
mkdir -p "$work"

snapshot_repository() {
	{
		git ls-files -s
		git status --porcelain=v1 --untracked-files=all
		git diff --binary --no-ext-diff
	} > "$1"
}

run_producer() {
	go run ./cmd/denominator-evolution-witness \
		--head "$HEAD_SHA" \
		--contract examples/denominator-evolution/contract.json \
		--source "$1" \
		--repository-writes "$repository_writes" \
		--snapshot-before "$snapshot_before" \
		--snapshot-after "$snapshot_after" \
		--out "$2"
}

run_consumer() {
	go run ./cmd/denominator-evolution-verify \
		--head "$HEAD_SHA" \
		--contract examples/denominator-evolution/contract.json \
		--source "$1" \
		--report "$2" \
		--repository-writes "$repository_writes" \
		--snapshot-before "$snapshot_before" \
		--snapshot-after "$snapshot_after" \
		--out "$3"
}

snapshot_repository "$work/repository-before.txt"

go test ./cmd/denominator-evolution-witness ./cmd/denominator-evolution-verify \
	./internal/meta/denominatorevolution ./internal/meta/denominatorevolutionverify
go run ./cmd/gooo check examples/denominator-evolution/main.gooo

go list -deps ./cmd/denominator-evolution-verify > "$work/consumer-dependencies.txt"
producer_imports=0
if grep -q '^github.com/kimjooyoon/meta-ontology-go/internal/meta/denominatorevolution$' "$work/consumer-dependencies.txt"; then
	echo "independent consumer imports producer" >&2
	exit 1
fi

snapshot_repository "$work/repository-after-build.txt"
if cmp -s "$work/repository-before.txt" "$work/repository-after-build.txt"; then
	repository_writes=0
else
	repository_writes=$( { git diff --name-only; git diff --cached --name-only; git ls-files --others --exclude-standard; } | sort -u | wc -l | tr -d ' ' )
fi
snapshot_before=$(sha256sum "$work/repository-before.txt" | awk '{print "sha256:" $1}')
snapshot_after=$(sha256sum "$work/repository-after-build.txt" | awk '{print "sha256:" $1}')

run_producer examples/denominator-evolution/main.gooo "$work/report-a.json"
run_producer examples/denominator-evolution/main.gooo "$work/report-b.json"
cmp -s "$work/report-a.json" "$work/report-b.json"
run_consumer examples/denominator-evolution/main.gooo "$work/report-a.json" "$work/verification-a.json"
run_consumer examples/denominator-evolution/main.gooo "$work/report-a.json" "$work/verification-b.json"
cmp -s "$work/verification-a.json" "$work/verification-b.json"

jq -e '
  .decision == "PASS" and .resolution == "EXACT" and
  .summary.cases_satisfied == 3 and .summary.cases_total == 3 and
  .denominator.version == "gooo/measurement-denominator/v1" and (.denominator.obligations | length) == 5 and
  .summary.fixed_denominator_numerator == 5 and .summary.fixed_denominator_denominator == 5 and
  .summary.source_cases_numerator == 3 and .summary.source_cases_denominator == 3 and
  .summary.persistent_claims_numerator == 3 and .summary.persistent_claims_denominator == 3 and
  .summary.guardrail_observations_numerator == 2 and .summary.guardrail_observations_denominator == 2 and
  .summary.legal_advance_numerator == 1 and .summary.legal_advance_denominator == 1 and
  .summary.unauthorized_rejection_numerator == 1 and .summary.unauthorized_rejection_denominator == 1 and
  .summary.unknown_predecessor_numerator == 1 and .summary.unknown_predecessor_denominator == 1 and
  .summary.addition_reason_numerator == 1 and .summary.addition_reason_denominator == 1 and
  .summary.deletion_reason_numerator == 1 and .summary.deletion_reason_denominator == 1 and
  .repository_writes == 0 and .mutation_authority == false and .repository_snapshot.changed_paths == 0 and
  ([.summary.guardrails[] | select(.id == "gooo.guardrail.denominator.forbidden-estimate.v1" and .direction == "AT_MOST" and .observed == 0 and .allowed_max == 0 and .conformance_numerator == 1 and .conformance_denominator == 1 and .conforms == true)] | length) == 1 and
  ([.summary.guardrails[] | select(.id == "gooo.guardrail.denominator.repository-writes.v1" and .direction == "AT_MOST" and .observed == 0 and .allowed_max == 0 and .conformance_numerator == 1 and .conformance_denominator == 1 and .conforms == true)] | length) == 1 and
  ([.indicators[] | select(.guardrail != null and .guardrail.conforms == true and .guardrail.observed == 0 and .guardrail.allowed_max == 0 and .guardrail.conformance_numerator == 1 and .guardrail.conformance_denominator == 1)] | length) == 2 and
  ([.cases[] | select(.id == "legal-advance") | .receipt.guardrails[] | select(.direction == "AT_MOST" and .observed == 0 and .allowed_max == 0 and .conformance_numerator == 1 and .conformance_denominator == 1 and .conforms == true)] | length) == 2 and
  ([.claim_ledger[] | select(.prior_state == "OPEN")] | length) == 3 and
  ([.cases[] | select(.id == "unknown-predecessor" and .observed_decision == "FAIL_CLOSED" and .observed_resolution == "LOWER_RESOLUTION" and .from_claim == "OPEN" and .to_claim == "OPEN")] | length) == 1 and
  ([.cases[] | select(.status != "SATISFIED")] | length) == 0 and
  ([.indicators[] | select(.satisfied != true)] | length) == 0
' "$work/report-a.json"

jq -e '
  .decision == "PASS" and .resolution == "EXACT" and .repository_writes == 0 and .mutation_authority == false and
  ([.guardrails[] | select(.id == "gooo.guardrail.denominator.forbidden-estimate.v1" and .direction == "AT_MOST" and .observed == 0 and .allowed_max == 0 and .conformance_numerator == 1 and .conformance_denominator == 1 and .conforms == true)] | length) == 1 and
  ([.guardrails[] | select(.id == "gooo.guardrail.denominator.repository-writes.v1" and .direction == "AT_MOST" and .observed == 0 and .allowed_max == 0 and .conformance_numerator == 1 and .conformance_denominator == 1 and .conforms == true)] | length) == 1 and
  ([.claim_ledger[] | select(.prior_state == "OPEN")] | length) == 3 and
  ([.checks[] | select(.status != "PASS")] | length) == 0
' "$work/verification-a.json"

# Semantic intervention mutates source case and ledger values. Both judgments
# must fail closed and source-derived semantic evidence must change.
semantic_source="$work/semantic.gooo"
sed -e 's/AUTHORIZED_MIGRATION\^ADVANCE\^EXACT\^DENOMINATOR_ADVANCE_AUTHORIZED\^OPEN\^DISCHARGED/AUTHORIZED_MIGRATION^BLOCK^INVARIANT_ONLY^MIGRATION_RECEIPT_MISSING^OPEN^REFUTED/' \
	-e 's/event=1\^legal-advance\^OPEN\^DISCHARGED\^DECIDE\^apply-migration-receipt/event=1^legal-advance^OPEN^REFUTED^DECIDE^apply-migration-receipt/' \
	-e 's/case=legal-advance\^ADVANCE\^EXACT\^DENOMINATOR_ADVANCE_AUTHORIZED/case=legal-advance^BLOCK^INVARIANT_ONLY^MIGRATION_RECEIPT_MISSING/' \
	-e 's/The migration receipt replays to the same successor digest\./The migration receipt replays to a changed successor digest./' \
	examples/denominator-evolution/main.gooo > "$semantic_source"
semantic_producer_status=0
run_producer "$semantic_source" "$work/report-semantic.json" || semantic_producer_status=$?
semantic_consumer_status=0
run_consumer "$semantic_source" "$work/report-semantic.json" "$work/verification-semantic.json" || semantic_consumer_status=$?

# Nonsemantic intervention: comments alter raw bytes but not lowered meaning.
comment_source="$work/comment-only.gooo"
sed '1i\/* comment-only intervention */' examples/denominator-evolution/main.gooo > "$comment_source"
run_producer "$comment_source" "$work/report-comment.json"
run_consumer "$comment_source" "$work/report-comment.json" "$work/verification-comment.json"

semantic_causality=0
if [[ "$semantic_producer_status" -ne 0 && "$semantic_consumer_status" -ne 0 ]] && \
	[[ "$(jq -r '.source_projection.semantic_digest' "$work/report-a.json")" != "$(jq -r '.source_projection.semantic_digest' "$work/report-semantic.json")" ]] && \
	[[ "$(jq -r '.claim_ledger[0].digest' "$work/report-a.json")" != "$(jq -r '.claim_ledger[0].digest' "$work/report-semantic.json")" ]] && \
	[[ "$(jq -r '.cases[0].receipt.digest' "$work/report-a.json")" != "$(jq -r '.cases[0].receipt.digest' "$work/report-semantic.json")" ]]; then
	semantic_causality=1
fi
nonsemantic_preservation=0
if [[ "$(jq -S 'del(.source_digest, .digest, .head_sha)' "$work/report-a.json")" == "$(jq -S 'del(.source_digest, .digest, .head_sha)' "$work/report-comment.json")" ]] && \
	[[ "$(jq -r '.source_projection.semantic_digest' "$work/report-a.json")" == "$(jq -r '.source_projection.semantic_digest' "$work/report-comment.json")" ]] && \
	[[ "$(jq -r '.source_digest' "$work/report-a.json")" != "$(jq -r '.source_digest' "$work/report-comment.json")" ]]; then
	nonsemantic_preservation=1
fi

snapshot_repository "$work/repository-final.txt"
if ! cmp -s "$work/repository-before.txt" "$work/repository-final.txt"; then
	echo "denominator evolution wrote to the repository" >&2
	exit 1
fi
git diff --exit-code

{
	echo "## Denominator evolution experiment"
	echo "- source_cases=$(jq -r '.summary.source_cases_numerator|tostring' "$work/report-a.json")/$(jq -r '.summary.source_cases_denominator|tostring' "$work/report-a.json")"
	echo "- producer_imports=$producer_imports/0"
	echo "- semantic_causality=$semantic_causality/1"
	echo "- nonsemantic_preservation=$nonsemantic_preservation/1"
	echo "- persistent_claims=$(jq -r '.summary.persistent_claims_numerator|tostring' "$work/report-a.json")/$(jq -r '.summary.persistent_claims_denominator|tostring' "$work/report-a.json")"
	echo "- guardrail_observations=$(jq -r '.summary.guardrail_observations_numerator|tostring' "$work/report-a.json")/$(jq -r '.summary.guardrail_observations_denominator|tostring' "$work/report-a.json")"
	jq -r '"- fixed denominator: \(.denominator.version) / \(.summary.fixed_denominator_numerator)/\(.summary.fixed_denominator_denominator)\n- legal advance: \(.summary.legal_advance_numerator)/\(.summary.legal_advance_denominator)\n- unauthorized changes rejected: \(.summary.unauthorized_rejection_numerator)/\(.summary.unauthorized_rejection_denominator)\n- unknown predecessors fail closed: \(.summary.unknown_predecessor_numerator)/\(.summary.unknown_predecessor_denominator)\n- producer receipt: \(.digest)"' "$work/report-a.json"
	jq -r '.summary.guardrails[] | "- guardrail \(.id): direction=\(.direction) observed=\(.observed) allowed_max=\(.allowed_max) conformance=\(.conformance_numerator)/\(.conformance_denominator) conforms=\(.conforms)"' "$work/report-a.json"
	jq -r '"- independent decision: \(.decision) / \(.resolution)\n- consumer receipt: \(.digest)"' "$work/verification-a.json"
	jq -r '.guardrails[] | "- independent guardrail \(.id): direction=\(.direction) observed=\(.observed) allowed_max=\(.allowed_max) conformance=\(.conformance_numerator)/\(.conformance_denominator) conforms=\(.conforms)"' "$work/verification-a.json"
} >> "$GITHUB_STEP_SUMMARY"

if [[ "$semantic_causality" != 1 || "$nonsemantic_preservation" != 1 ]]; then
	echo "source intervention predicates did not hold" >&2
	exit 1
fi
