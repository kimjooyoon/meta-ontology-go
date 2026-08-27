#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${PROOF_CARRYING_OUTPUT:?PROOF_CARRYING_OUTPUT is required}"
output="$PROOF_CARRYING_OUTPUT"
case "$output" in
  /*) ;;
  *) echo "proof-carrying output must be an absolute path"; exit 1 ;;
esac
case "$output" in
  "$PWD"|"$PWD"/*) echo "proof-carrying output must be outside the repository"; exit 1 ;;
esac
mkdir -p "$output"

workflow_failure() {
  local status="$?" failure_line="${BASH_LINENO[0]:-${LINENO}}" command="${BASH_COMMAND:-unknown}"
  printf 'line=%s\nstatus=%s\ncommand=%s\n' "$failure_line" "$status" "$command" > "$output/ci-failure.txt" || true
  echo "proof-carrying artifact workflow failed: line=$failure_line status=$status command=$command" >&2
  exit "$status"
}
trap workflow_failure ERR

source_repo="examples/language-proof-carrying-artifact/main.gooo"
recipe_projection="examples/language-proof-carrying-artifact/recipe.json"
contract_repo="examples/language-proof-carrying-artifact/contract.json"

snapshot_repo() {
  local target="$1" entries_file path kind digest
  entries_file="$target.entries.json"
  git ls-files -co --exclude-standard -z | while IFS= read -r -d '' path; do
    if git ls-files --error-unmatch -- "$path" >/dev/null 2>&1; then kind="TRACKED"; else kind="UNTRACKED"; fi
    if [ -f "$path" ]; then digest="sha256:$(sha256sum -- "$path" | awk '{print $1}')"; else digest="MISSING"; fi
    jq -cn --arg kind "$kind" --arg path "$path" --arg digest "$digest" '{kind:$kind,path:$path,digest:$digest}'
  done | jq -s 'sort_by(.path)' > "$entries_file"
  jq -n --arg schema "gooo/repository-write-set-snapshot/v1" --slurpfile entries "$entries_file" '{schema:$schema,version:1,entries:$entries[0]}' > "$target"
}

snapshot_repo "$output/repository-before.json"

# This workflow intentionally has no local test invocation. The only Go
# execution below is the GitHub Actions execution of the producer and kernel.
printf '%s\n' 'local_tests=0' > "$output/local-test-policy.txt"
toolchain_version="$(go version)"
case "$toolchain_version" in
  "go version go1.27.0 "*) ;;
  *) echo "unexpected Go toolchain: $toolchain_version" >&2; exit 1 ;;
esac
printf '%s\n' "$toolchain_version" > "$output/go-version.txt"
actual_head="$(git rev-parse HEAD)"
test "$actual_head" = "$HEAD_SHA"

cp "$source_repo" "$output/source.gooo"
cp "$contract_repo" "$output/contract.json"
go run ./cmd/language-proof-carrying-artifact -derive-recipe "$output/source.gooo" -out "$output/recipe.json"
test "$(jq -cS . "$output/recipe.json")" = "$(jq -cS . "$recipe_projection")" || { echo "checked-in recipe is not the source-derived projection" >&2; exit 1; }

go run ./cmd/gooo check "$source_repo"
go run ./cmd/gooo run --json --entry GenerateProofCarryingArtifact "$source_repo" > "$output/operation-receipt.json"

internal_dependencies="$(go list -deps ./cmd/language-proof-carrying-artifact-verifier | sort | grep -E '^github.com/kimjooyoon/meta-ontology-go/internal/' || true)"
producer_import_numerator="$(printf '%s\n' "$internal_dependencies" | grep -E '/internal/(meta/languageproofartifact|sourceexecution)(/|$)' | sed '/^$/d' | wc -l | tr -d ' ' || true)"
producer_import_denominator="$(printf '%s\n' "$internal_dependencies" | sed '/^$/d' | wc -l | tr -d ' ' || true)"
core_parser_dependencies="$(printf '%s\n' "$internal_dependencies" | grep -E '/internal/(syntax|bidir)(/|$)' | sed '/^$/d' | wc -l | tr -d ' ' || true)"
jq -n --arg schema "gooo/language-proof-carrying-artifact-independence/v1" --argjson producer_dependencies "$producer_import_numerator" --argjson producer_import_numerator "$producer_import_numerator" --argjson producer_import_denominator "$producer_import_denominator" --argjson core_parser_dependencies "$core_parser_dependencies" '{schema:$schema,producer_dependencies:$producer_dependencies,producer_import_numerator:$producer_import_numerator,producer_import_denominator:$producer_import_denominator,core_parser_dependencies:$core_parser_dependencies}' > "$output/independence.json"
test "$producer_import_numerator" -eq 0
test "$producer_import_denominator" -gt 0
test "$core_parser_dependencies" -gt 0

make_write_set() {
  local before_entries before_digest after_entries after_digest changed_file
  snapshot_repo "$output/repository-after.json"
  before_entries="$(jq -c '.entries' "$output/repository-before.json")"
  before_digest="sha256:$(printf '%s' "$before_entries" | sha256sum | awk '{print $1}')"
  after_entries="$(jq -c '.entries' "$output/repository-after.json")"
  after_digest="sha256:$(printf '%s' "$after_entries" | sha256sum | awk '{print $1}')"
  changed_file="$output/repository-write-set-changed.json"
  jq -n --slurpfile before "$output/repository-before.json" --slurpfile after "$output/repository-after.json" 'def by_path: reduce .[] as $item ({}; .[$item.path] = $item); ($before[0].entries | by_path) as $b | ($after[0].entries | by_path) as $a | (($b | keys) + ($a | keys) | unique | sort) as $paths | [$paths[] as $path | if ($b[$path] != null and $a[$path] != null and $b[$path] == $a[$path]) then empty else {path:$path,before_digest:($b[$path].digest // ""),after_digest:($a[$path].digest // ""),before_kind:($b[$path].kind // ""),after_kind:($a[$path].kind // "")} end]' > "$changed_file"
  jq -n --slurpfile before "$output/repository-before.json" --slurpfile after "$output/repository-after.json" --slurpfile changed "$changed_file" --arg before_digest "$before_digest" --arg after_digest "$after_digest" '{schema:"gooo/repository-write-set-observation/v1",version:1,before:$before[0].entries,after:$after[0].entries,changed:$changed[0],before_digest:$before_digest,after_digest:$after_digest,net_changed_paths:($changed[0]|length),capability_mutation_granted:false,observed_scope:"NET_BEFORE_AFTER_TRACKED_AND_UNTRACKED",net_repository_state_unchanged:(($changed[0]|length)==0),transient_writes_unknown:true,actual_writes_observation:"UNKNOWN",global_mutation_authority:"UNKNOWN",authority_observation:"DECLARATION_ONLY_NOT_GLOBAL_AUTHORITY",digest:""}' > "$output/write-set.json"
}

make_write_set

produce_artifact() {
  local source_bytes="$1" source_name="$2" operation="$3" artifact="$4"
  go run ./cmd/language-proof-carrying-artifact -head "$HEAD_SHA" -source-path "$source_name" -source "$source_bytes" -operation "$operation" -recipe "$output/recipe.json" -write-set "$output/write-set.json" -out "$artifact"
}

run_receipt() {
  local source="$1" receipt="$2"
  go run ./cmd/gooo run --json --entry GenerateProofCarryingArtifact "$source" > "$receipt"
}

produce_artifact "$output/source.gooo" "$source_repo" "$output/operation-receipt.json" "$output/artifact.json"

sed 's|gooo://proof/source-evidence|gooo://proof/source-evidence-v2|' "$source_repo" > "$output/semantic-intervention.gooo"
{
  printf '%s\n' '// comment-only intervention: comments are lexically ignored by the consumer projection.'
  cat "$source_repo"
} > "$output/comment-only-intervention.gooo"
run_receipt "$output/semantic-intervention.gooo" "$output/semantic-operation-receipt.json"
run_receipt "$output/comment-only-intervention.gooo" "$output/comment-operation-receipt.json"
produce_artifact "$output/semantic-intervention.gooo" "$output/semantic-intervention.gooo" "$output/semantic-operation-receipt.json" "$output/semantic-intervention-artifact.json"
produce_artifact "$output/comment-only-intervention.gooo" "$output/comment-only-intervention.gooo" "$output/comment-operation-receipt.json" "$output/comment-only-intervention-artifact.json"

fake_source_digest="sha256:0000000000000000000000000000000000000000000000000000000000000000"
coherent_operation_body="$(jq -c --arg fake "$fake_source_digest" '.source_digest=$fake | .events |= map(if .kind == "SOURCE_PARSED" then .subject=$fake else . end) | .digest=""' "$output/operation-receipt.json")"
coherent_operation_digest="sha256:$(printf '%s' "$coherent_operation_body" | sha256sum | awk '{print $1}')"
jq -c --arg digest "$coherent_operation_digest" '.digest=$digest' <<< "$coherent_operation_body" > "$output/coherent-operation-receipt.json"

reseal() {
  local input="$1" target="$2" filter="$3" body digest
  body="$(jq -c "$filter | .digest = \"\"" "$input")"
  digest="sha256:$(printf '%s' "$body" | sha256sum | awk '{print $1}')"
  jq -c --arg digest "$digest" '.digest = $digest' <<< "$body" > "$target"
}

reseal "$output/artifact.json" "$output/tampered.json" '(.evidence[] | select(.kind == "SOURCE") | .source_digest) = "sha256:0000000000000000000000000000000000000000000000000000000000000000"'
reseal "$output/artifact.json" "$output/missing-operation.json" 'del(.evidence[] | select(.kind == "OPERATION"))'
cp "$output/artifact.json" "$output/byte-only.json"
cp "$output/artifact.json" "$output/recipe-only.json"
cp "$output/artifact.json" "$output/missing-attachment.json"
jq -c '.steps[0].meta_operation = "accept-source-claim-without-recheck"' "$output/recipe.json" > "$output/wrong-recipe.json"
jq -c '.events[2].subject = "wrong-attachment-activity"' "$output/operation-receipt.json" > "$output/wrong-attachment-digest.json"

# Coherent tamper: reseal the operation, its evidence, every linked claim,
# and the complete prior-ledger chain. The consumer must still reject the
# source/operation relation rather than relying on an inconsistent envelope.
old_operation_evidence_digest="$(jq -r '.evidence[] | select(.kind == "OPERATION") | .evidence_digest' "$output/artifact.json")"
coherent_evidence_body="$(jq -c --arg receipt_digest "$coherent_operation_digest" '.evidence[] | select(.kind == "OPERATION") | .receipt_digest=$receipt_digest | .evidence_digest=""' "$output/artifact.json")"
coherent_evidence_digest="sha256:$(printf '%s' "$coherent_evidence_body" | sha256sum | awk '{print $1}')"
coherent_body="$(jq -c --arg receipt_digest "$coherent_operation_digest" --arg old "$old_operation_evidence_digest" --arg new "$coherent_evidence_digest" '.evidence |= map(if .kind == "OPERATION" then .receipt_digest=$receipt_digest | .evidence_digest=$new else . end) | .claims |= map(.evidence_digests |= map(if . == $old then $new else . end) | .digest="") | .prior_ledger.entries |= map(.evidence_digests |= map(if . == $old then $new else . end)) | .digest=""' "$output/artifact.json")"
for i in 0 1 2 3 4; do
  claim_body="$(jq -c --argjson i "$i" '.claims[$i].digest=""' <<< "$coherent_body")"
  claim_canonical="$(jq -c --argjson i "$i" '.claims[$i]' <<< "$claim_body")"
  claim_digest="sha256:$(printf '%s' "$claim_canonical" | sha256sum | awk '{print $1}')"
  coherent_body="$(jq -c --argjson i "$i" --arg digest "$claim_digest" '.claims[$i].digest=$digest' <<< "$coherent_body")"
done
previous=""
for i in 0 1 2 3 4; do
  entry_body="$(jq -c --argjson i "$i" --arg previous "$previous" '.prior_ledger.entries[$i].previous_digest=$previous | .prior_ledger.entries[$i].digest=""' <<< "$coherent_body")"
  entry_canonical="$(jq -c --argjson i "$i" '.prior_ledger.entries[$i]' <<< "$entry_body")"
  entry_digest="sha256:$(printf '%s' "$entry_canonical" | sha256sum | awk '{print $1}')"
  coherent_body="$(jq -c --argjson i "$i" --arg previous "$previous" --arg digest "$entry_digest" '.prior_ledger.entries[$i].previous_digest=$previous | .prior_ledger.entries[$i].digest=$digest' <<< "$coherent_body")"
  previous="$entry_digest"
done
ledger_body="$(jq -c '.prior_ledger.digest=""' <<< "$coherent_body")"
ledger_canonical="$(jq -c '.prior_ledger' <<< "$ledger_body")"
ledger_digest="sha256:$(printf '%s' "$ledger_canonical" | sha256sum | awk '{print $1}')"
coherent_body="$(jq -c --arg digest "$ledger_digest" '.prior_ledger.digest=$digest' <<< "$coherent_body")"
coherent_artifact_canonical="$(jq -c '.digest=""' <<< "$coherent_body")"
coherent_artifact_digest="sha256:$(printf '%s' "$coherent_artifact_canonical" | sha256sum | awk '{print $1}')"
jq -c --arg digest "$coherent_artifact_digest" '.digest=$digest' <<< "$coherent_body" > "$output/coherent-tamper.json"

# Claim-structure tampering is coherent at the byte envelope and prior-ledger
# level, but not coherent with the verifier's canonical claim proposition.
claim_structure_tamper() {
  local index="$1" target="$2" mutation="$3" body claim_body claim_canonical claim_digest entry_body entry_canonical entry_digest previous ledger_body ledger_canonical ledger_digest artifact_canonical artifact_digest
  body="$(jq -c --argjson index "$index" "$mutation | .claims[\$index].digest=\"\" | .prior_ledger.entries[\$index].digest=\"\" | .prior_ledger.digest=\"\" | .digest=\"\"" "$output/artifact.json")"
  for i in 0 1 2 3 4; do
    claim_body="$(jq -c --argjson i "$i" '.claims[$i].digest=""' <<< "$body")"
    claim_canonical="$(jq -c --argjson i "$i" '.claims[$i]' <<< "$claim_body")"
    claim_digest="sha256:$(printf '%s' "$claim_canonical" | sha256sum | awk '{print $1}')"
    body="$(jq -c --argjson i "$i" --arg digest "$claim_digest" '.claims[$i].digest=$digest' <<< "$body")"
  done
  previous=""
  for i in 0 1 2 3 4; do
    entry_body="$(jq -c --argjson i "$i" --arg previous "$previous" '.prior_ledger.entries[$i].previous_digest=$previous | .prior_ledger.entries[$i].digest=""' <<< "$body")"
    entry_canonical="$(jq -c --argjson i "$i" '.prior_ledger.entries[$i]' <<< "$entry_body")"
    entry_digest="sha256:$(printf '%s' "$entry_canonical" | sha256sum | awk '{print $1}')"
    body="$(jq -c --argjson i "$i" --arg previous "$previous" --arg digest "$entry_digest" '.prior_ledger.entries[$i].previous_digest=$previous | .prior_ledger.entries[$i].digest=$digest' <<< "$body")"
    previous="$entry_digest"
  done
  ledger_body="$(jq -c '.prior_ledger.digest=""' <<< "$body")"
  ledger_canonical="$(jq -c '.prior_ledger' <<< "$ledger_body")"
  ledger_digest="sha256:$(printf '%s' "$ledger_canonical" | sha256sum | awk '{print $1}')"
  body="$(jq -c --arg digest "$ledger_digest" '.prior_ledger.digest=$digest' <<< "$body")"
  artifact_canonical="$(jq -c '.digest=""' <<< "$body")"
  artifact_digest="sha256:$(printf '%s' "$artifact_canonical" | sha256sum | awk '{print $1}')"
  jq -c --arg digest "$artifact_digest" '.digest=$digest' <<< "$body" > "$target"
}

claim_structure_tamper 0 "$output/claim-proposition-tamper.json" '.claims[$index].proposition="source-bytes-do-not-match" | .prior_ledger.entries[$index].proposition=.claims[$index].proposition'
claim_structure_tamper 1 "$output/claim-dependency-tamper.json" '.claims[$index].dependencies=["no-byte-authority"] | .prior_ledger.entries[$index].dependencies=.claims[$index].dependencies'
claim_structure_tamper 0 "$output/claim-proof-choice-tamper.json" '.claims[$index].proof_choice="COHERENCE" | .prior_ledger.entries[$index].proof_choice=.claims[$index].proof_choice'
claim_structure_tamper 0 "$output/claim-target-tamper.json" '.claims[$index].target_digest="sha256:0000000000000000000000000000000000000000000000000000000000000000" | .prior_ledger.entries[$index].target_digest=.claims[$index].target_digest'

# An unrelated evidence tamper reseals the inner and outer digests while
# leaving source and operation attachments untouched.
unrelated_evidence_body="$(jq -c '(.evidence[] | select(.kind == "INVARIANT") | .predicate) = "unrelated-observation" | (.evidence[] | select(.kind == "INVARIANT") | .evidence_digest) = "" | .digest=""' "$output/artifact.json")"
unrelated_evidence_digest="sha256:$(printf '%s' "$(jq -c '.evidence[] | select(.kind == "INVARIANT")' <<< "$unrelated_evidence_body")" | sha256sum | awk '{print $1}')"
unrelated_body="$(jq -c --arg digest "$unrelated_evidence_digest" '.evidence |= map(if .kind == "INVARIANT" then .evidence_digest=$digest else . end)' <<< "$unrelated_evidence_body")"
unrelated_artifact_digest="sha256:$(printf '%s' "$(jq -c '.digest=""' <<< "$unrelated_body")" | sha256sum | awk '{print $1}')"
jq -c --arg digest "$unrelated_artifact_digest" '.digest=$digest' <<< "$unrelated_body" > "$output/unrelated-tamper.json"
reseal "$output/artifact.json" "$output/stale-head.json" '.head_sha = "0000000000000000000000000000000000000000"'
jq -cn --arg schema "gooo/language-proof-carrying-artifact-verifier/v1" --arg consumer "gooo://consumer/unauthorized" '{schema:$schema,consumer:$consumer}' > "$output/unauthorized-consumer.json"

tree_digest="sha256:$(git ls-tree -r HEAD | sha256sum | awk '{print $1}')"
source_digest="sha256:$(sha256sum "$output/source.gooo" | awk '{print $1}')"
operation_digest="sha256:$(sha256sum "$output/operation-receipt.json" | awk '{print $1}')"
recipe_digest="sha256:$(sha256sum "$output/recipe.json" | awk '{print $1}')"
contract_digest="sha256:$(sha256sum "$output/contract.json" | awk '{print $1}')"
jq -n --arg head "$HEAD_SHA" --arg actual "$actual_head" --arg tree "$tree_digest" --arg source "$source_digest" --arg operation "$operation_digest" --arg recipe "$recipe_digest" --arg contract "$contract_digest" '{head_sha:$head,actual_head_sha:$actual,tree_digest:$tree,source_digest:$source,operation_digest:$operation,recipe_digest:$recipe,contract_digest:$contract}' > "$output/checkout.json"

run_verifier() {
  local report="$1" unauthorized_bundle="${2:-$output/bundle.json}"
  go run ./cmd/language-proof-carrying-artifact-verifier -head "$HEAD_SHA" -contract "$output/contract.json" -valid "$output/artifact.json" -tampered "$output/tampered.json" -coherent-tamper "$output/coherent-tamper.json" -missing "$output/missing-operation.json" -byte-only "$output/byte-only.json" -wrong-recipe "$output/wrong-recipe.json" -recipe-only "$output/recipe-only.json" -missing-attachment "$output/missing-attachment.json" -wrong-attachment-digest "$output/wrong-attachment-digest.json" -unrelated-tamper "$output/unrelated-tamper.json" -stale-head "$output/stale-head.json" -claim-proposition-tamper "$output/claim-proposition-tamper.json" -claim-dependency-tamper "$output/claim-dependency-tamper.json" -claim-proof-choice-tamper "$output/claim-proof-choice-tamper.json" -claim-target-tamper "$output/claim-target-tamper.json" -unauthorized-consumer "$output/unauthorized-consumer.json" -unauthorized-bundle "$unauthorized_bundle" -source "$output/source.gooo" -operation "$output/operation-receipt.json" -recipe "$output/recipe.json" -independence "$output/independence.json" -write-set "$output/write-set.json" -coherent-operation "$output/coherent-operation-receipt.json" -semantic-artifact "$output/semantic-intervention-artifact.json" -semantic-source "$output/semantic-intervention.gooo" -semantic-operation "$output/semantic-operation-receipt.json" -comment-artifact "$output/comment-only-intervention-artifact.json" -comment-source "$output/comment-only-intervention.gooo" -comment-operation "$output/comment-operation-receipt.json" -checkout "$output/checkout.json" -output "$report"
}

bundle_inputs="$output/bundle-inputs.json"
paths=(artifact.json tampered.json coherent-tamper.json missing-operation.json byte-only.json wrong-recipe.json recipe-only.json missing-attachment.json wrong-attachment-digest.json unrelated-tamper.json stale-head.json unauthorized-consumer.json claim-proposition-tamper.json claim-dependency-tamper.json claim-proof-choice-tamper.json claim-target-tamper.json source.gooo operation-receipt.json recipe.json contract.json independence.json write-set.json coherent-operation-receipt.json checkout.json semantic-intervention-artifact.json semantic-intervention.gooo semantic-operation-receipt.json comment-only-intervention-artifact.json comment-only-intervention.gooo comment-operation-receipt.json)
for path in "${paths[@]}"; do
  role="case-attachment"
  case "$path" in
    source.gooo) role="source-bytes";;
    operation-receipt.json|coherent-operation-receipt.json|semantic-operation-receipt.json|comment-operation-receipt.json) role="operation-receipt";;
    recipe.json) role="source-derived-recipe";;
    contract.json) role="validator-expectation";;
    checkout.json) role="checkout-binding";;
    independence.json) role="independence";;
    write-set.json) role="net-repository-observation";;
    *.gooo) role="intervention-source";;
  esac
  jq -cn --arg path "$path" --arg file "$output/$path" --arg role "$role" '{path:$path,file:$file,role:$role}'
done | jq -s '.' > "$bundle_inputs"
go run ./cmd/language-proof-carrying-artifact-verifier -head "$HEAD_SHA" -checkout "$output/checkout.json" -bundle-inputs "$bundle_inputs" -pack-bundle "$output/bundle.json"
go run ./cmd/language-proof-carrying-artifact-verifier -bundle "$output/bundle.json" -output "$output/bundle-preliminary-report.json"
go run ./cmd/language-proof-carrying-artifact-consumer -bundle "$output/bundle.json" -report "$output/bundle-preliminary-report.json" -target artifact.json -out "$output/consumer-receipt.json"
go run ./cmd/language-proof-carrying-artifact-verifier -bundle "$output/bundle.json" -consumer-receipt "$output/consumer-receipt.json" -output "$output/bundle-report.json"
jq -e '.conformance_decision == "FAIL_CLOSED" and .conformance_resolution == "LOWER_RESOLUTION" and .conformance_reason == "CONSUMER_RECHECK_NOT_OBSERVED" and .conformance_coordinate.stage == "CONSUME_BUNDLE" and .conformance_coordinate.step == "consumer-recheck" and .summary.bundle_only_verification == 1 and .summary.consumer_rechecks == 0 and .proof_summary == {phase:"PRELIMINARY",proofs:3,evidence_validated:3,evidence_validated_total:3,observed_state:1,observed_state_total:3,open:2,open_total:3,discharged:0,discharged_total:3,authority:0,authority_total:1} and ([.indicators[] | select(.satisfied == false)] | length) == 1' "$output/bundle-preliminary-report.json"
jq -e '.consumer_receipt.output_exists == false and .consumer_receipt.target_digest == "" and .consumer_receipt.output_digest == "" and .artifact_use_authority == "" and ([.proofs[] | select(.phase == "PRELIMINARY" and .state == "OBSERVED" and .evidence_validated == true and .passed == false and .consumer_gate_open == false)] | length) == 1 and ([.proofs[] | select(.phase == "PRELIMINARY" and .state == "OPEN" and .evidence_validated == true and .passed == false and .consumer_gate_open == true)] | length) == 2' "$output/bundle-preliminary-report.json"

preliminary_negative_fixture_total=14
preliminary_negative_fixture_observed=0
assert_preliminary_rejected() {
  local name="$1" filter="$2" expected="$3" raw_fixture="$output/$1-mutated.json" fixture="$output/$1.json" check_log="$output/$1-check.log" consumer_log="$output/$1-consumer.log"
  jq -c "$filter" "$output/bundle-preliminary-report.json" > "$raw_fixture"
  go run ./cmd/language-proof-carrying-artifact-verifier -reseal-report "$raw_fixture" -output "$fixture"
  if go run ./cmd/language-proof-carrying-artifact-verifier -check "$fixture" > "$check_log" 2>&1; then
    echo "$name preliminary unexpectedly validated by CLI" >&2
    exit 1
  fi
  rg -F "$expected" "$check_log"
  if go run ./cmd/language-proof-carrying-artifact-consumer -bundle "$output/bundle.json" -report "$fixture" -target artifact.json -out "$output/$name-consumer-receipt.json" > "$consumer_log" 2>&1; then
    echo "$name preliminary unexpectedly lifted by consumer" >&2
    exit 1
  fi
  rg -F 'ATTESTATION_MISMATCH' "$consumer_log"
  preliminary_negative_fixture_observed=$((preliminary_negative_fixture_observed + 1))
}

assert_preliminary_rejected "preliminary-38-of-40" '.indicators[2].value = 38 | .indicators[2].target = 40 | .indicators[2].satisfied = false' 'stage=EVALUATE step=preliminary-inventory reason=PRELIMINARY_INDICATOR_INVENTORY_MISMATCH'
assert_preliminary_rejected "preliminary-duplicate-metric" '.indicators[1].metric_id = .indicators[0].metric_id' 'stage=EVALUATE step=preliminary-inventory reason=PRELIMINARY_INDICATOR_INVENTORY_MISMATCH'
assert_preliminary_rejected "preliminary-missing-metric" 'del(.indicators[1])' 'stage=EVALUATE step=preliminary-inventory reason=PRELIMINARY_INDICATOR_INVENTORY_MISMATCH'
assert_preliminary_rejected "preliminary-proof-false" '.proofs[0].passed = false' 'stage=VERIFY_PROOF step=preliminary-proof-gate reason=PRELIMINARY_PROOF_NOT_SATISFIED'
assert_preliminary_rejected "preliminary-broken-transition" '.claim_transitions[0].digest = ""' 'stage=CONSUME step=preliminary-transition-chain reason=PRELIMINARY_TRANSITION_CHAIN_MISMATCH'
assert_preliminary_rejected "preliminary-invalid-claim-ledger" '.prior_ledger.entries[0].status = "DISCHARGED"' 'stage=CONSUME_LEDGER step=preliminary-ledger reason=PRELIMINARY_CLAIM_LEDGER_MISMATCH'
assert_preliminary_rejected "preliminary-final-authority" '.conformance_decision = "PASS" | .conformance_resolution = "EXACT" | .conformance_reason = "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED" | .conformance_coordinate = {stage:"CONSUME_AUTHORITY",step:"grant-read-only-consumption",reason:"PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED"} | .artifact_use_authority = "READ_ONLY_CONSUMPTION"' 'stage=CONSUME_AUTHORITY step=preliminary-report reason=PRELIMINARY_FINAL_AUTHORITY_PRESENT'
assert_preliminary_rejected "preliminary-open-proof-passed" '.proofs[1].passed = true' 'stage=VERIFY_PROOF step=preliminary-proof-gate reason=PRELIMINARY_PROOF_NOT_SATISFIED'
assert_preliminary_rejected "preliminary-open-gate-closed" '.proofs[1].consumer_gate_open = false' 'stage=VERIFY_PROOF step=preliminary-proof-gate reason=PRELIMINARY_PROOF_NOT_SATISFIED'
assert_preliminary_rejected "preliminary-discharged-without-receipt" '.proofs[1].state = "DISCHARGED"' 'stage=VERIFY_PROOF step=preliminary-proof-gate reason=PRELIMINARY_PROOF_NOT_SATISFIED'
assert_preliminary_rejected "preliminary-summary-numerator" '.proof_summary.evidence_validated = 2' 'stage=VERIFY_PROOF step=preliminary-proof-summary reason=PRELIMINARY_PROOF_SUMMARY_MISMATCH'
assert_preliminary_rejected "preliminary-summary-denominator" '.proof_summary.open_total = 2' 'stage=VERIFY_PROOF step=preliminary-proof-summary reason=PRELIMINARY_PROOF_SUMMARY_MISMATCH'
assert_preliminary_rejected "preliminary-summary-phase" '.proof_summary.phase = "FINAL"' 'stage=VERIFY_PROOF step=preliminary-proof-summary reason=PRELIMINARY_PROOF_SUMMARY_MISMATCH'
assert_preliminary_rejected "preliminary-summary-authority" '.proof_summary.authority = 1' 'stage=VERIFY_PROOF step=preliminary-proof-summary reason=PRELIMINARY_PROOF_SUMMARY_MISMATCH'
test "$preliminary_negative_fixture_observed" -eq "$preliminary_negative_fixture_total"
jq -n --argjson total "$preliminary_negative_fixture_total" --argjson observed "$preliminary_negative_fixture_observed" '{schema:"gooo/preliminary-negative-fixture-ledger/v1",total:$total,observed:$observed,stale_digest_only_rejections:0,validator_and_consumer_checked:true}' > "$output/preliminary-negative-fixture-ledger.json"

jq '.digest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"' "$output/bundle.json" > "$output/corrupt-bundle.json"
if run_verifier "$output/missing-bundle-report.json" "$output/no-such-bundle.json" > "$output/missing-bundle.log" 2>&1; then
  echo "missing bundle unexpectedly produced a liftable preliminary" >&2
  exit 1
fi
if run_verifier "$output/corrupt-bundle-report.json" "$output/corrupt-bundle.json" > "$output/corrupt-bundle.log" 2>&1; then
  echo "corrupt bundle unexpectedly produced a liftable preliminary" >&2
  exit 1
fi
rg -F 'BUNDLE_CONSUMPTION_NOT_OBSERVED' "$output/missing-bundle.log" "$output/corrupt-bundle.log"
run_verifier "$output/report.json"

go run ./cmd/language-proof-carrying-artifact-verifier -check "$output/report.json"
jq -e '.conformance_decision == "FAIL_CLOSED" and .conformance_resolution == "LOWER_RESOLUTION" and .conformance_reason == "BUNDLE_CONSUMPTION_NOT_OBSERVED" and .conformance_coordinate.stage == "CONSUME_BUNDLE" and .conformance_coordinate.step == "consumer-recheck" and .summary.bundle_only_verification == 0 and .summary.consumer_rechecks == 0 and ([.indicators[] | select(.satisfied == false)] | length) == 2' "$output/report.json"
go run ./cmd/language-proof-carrying-artifact-verifier -check "$output/report.json"
go run ./cmd/language-proof-carrying-artifact-verifier -check "$output/bundle-preliminary-report.json"
go run ./cmd/language-proof-carrying-artifact-verifier -check "$output/bundle-report.json"
fake_preliminary_digest="sha256:0000000000000000000000000000000000000000000000000000000000000000"
go run ./cmd/language-proof-carrying-artifact-verifier -coherent-preliminary-tamper "$output/bundle-report.json" -preliminary-digest "$fake_preliminary_digest" -output "$output/coherent-preliminary-binding-tamper.json"
if go run ./cmd/language-proof-carrying-artifact-verifier -check "$output/coherent-preliminary-binding-tamper.json" > "$output/coherent-preliminary-binding-tamper-check.log" 2>&1; then
  echo "coherent preliminary binding tamper unexpectedly validated" >&2
  exit 1
fi
rg -F 'stage=CONSUME_BUNDLE step=consumer-receipt reason=PRELIMINARY_BINDING_MISMATCH' "$output/coherent-preliminary-binding-tamper-check.log"
if go run ./cmd/language-proof-carrying-artifact-consumer -bundle "$output/bundle.json" -report "$output/coherent-preliminary-binding-tamper.json" -target artifact.json -out "$output/coherent-preliminary-binding-tamper-receipt.json" > "$output/coherent-preliminary-binding-tamper-consumer.log" 2>&1; then
  echo "coherent preliminary binding tamper unexpectedly consumed" >&2
  exit 1
fi
rg -F 'ATTESTATION_MISMATCH' "$output/coherent-preliminary-binding-tamper-consumer.log"
if go run ./cmd/language-proof-carrying-artifact-consumer -bundle "$output/bundle.json" -report "$output/bundle-report.json" -target missing-target.json -out "$output/target-missing-receipt.json" > "$output/target-missing.log" 2>&1; then
  echo "target-missing consumer unexpectedly succeeded" >&2
  exit 1
fi
rg -F 'TARGET_MISSING' "$output/target-missing.log"
jq '.consumer_receipt.output_exists = false | .consumer_receipt.output_digest = ""' "$output/bundle-report.json" > "$output/output-absent-report.json"
if go run ./cmd/language-proof-carrying-artifact-consumer -bundle "$output/bundle.json" -report "$output/output-absent-report.json" -target artifact.json -out "$output/output-absent-receipt.json" > "$output/output-absent.log" 2>&1; then
  echo "output-absent consumer unexpectedly succeeded" >&2
  exit 1
fi
rg -F 'RECEIPT_MISMATCH' "$output/output-absent.log"
jq '.proofs[0].passed = false' "$output/bundle-report.json" > "$output/proof-false-report.json"
if go run ./cmd/language-proof-carrying-artifact-verifier -check "$output/proof-false-report.json"; then
  echo "proof-false report unexpectedly validated" >&2
  exit 1
fi
jq '(.indicators[] | select(.metric_id == "gooo.metric.language.proof-carrying-artifact-cases.v3") | .value) = 38 | (.indicators[] | select(.metric_id == "gooo.metric.language.proof-carrying-artifact-cases.v3") | .satisfied) = false' "$output/bundle-report.json" > "$output/main-38-of-40-report.json"
if go run ./cmd/language-proof-carrying-artifact-verifier -check "$output/main-38-of-40-report.json"; then
  echo "38-of-40 report unexpectedly validated" >&2
  exit 1
fi
jq -e --slurpfile receipt "$output/consumer-receipt.json" '.consumer_receipt == $receipt[0] and .conformance_decision == "PASS" and .conformance_resolution == "EXACT" and .checkout_binding_scope == "BUNDLE_HISTORICAL_SUBJECT_BINDING" and .summary.bundle_only_verification == 1 and .summary.consumer_rechecks == 1 and .summary.claim_templates == 5 and .summary.claim_instances == 80 and .summary.accepted_transitions == 5 and .summary.case_discharged_claims == 43 and .summary.case_open_claims == 20 and .summary.case_refuted_claims == 17 and .summary.coherent_claim_structure_rejections == 4 and .summary.final_ledger_open_claims == 5 and .summary.final_ledger_discharged_claims == 5 and .summary.recipe_rejections == 1 and .summary.recipe_only_rejections == 1 and .summary.missing_attachment_rejections == 1 and .summary.wrong_attachment_rejections == 1 and .summary.unrelated_evidence_rejections == 1 and .summary.stale_head_rejections == 1 and .summary.unauthorized_consumer_denials == 1 and .unauthorized_consumer_target_digest != "" and .unauthorized_consumer_output_exists == false and .unauthorized_consumer_output_digest == "" and .unauthorized_consumer_error_class == "ATTESTATION_MISMATCH" and .unauthorized_consumer_error_digest != "" and .summary.net_repository_state_unchanged == 1 and .summary.unknown_authority_observations == 1 and .write_set.net_changed_paths == 0 and .write_set.observed_scope == "NET_BEFORE_AFTER_TRACKED_AND_UNTRACKED" and .write_set.transient_writes_unknown == true and .write_set.actual_writes_observation == "UNKNOWN" and .write_set.global_mutation_authority == "UNKNOWN" and .capability_mutation_granted == false and .authority_observation == "UNKNOWN_GLOBAL_TRANSIENT_SCOPE" and ([.indicators[] | select(.satisfied == false)] | length) == 0 and ([.claim_transitions[] | select(.from == "CARRIED" and .to == "PRESERVED")] | length) == 4 and ([.claim_transitions[] | select(.claim_id == "consumer-authority" and .capability == "ARTIFACT_USE" and .from == "NONE" and .to == "READ_ONLY_CONSUMPTION")] | length) == 1 and ([.interventions[] | select(.kind == "SEMANTIC" and .semantic_digest_changed == true and .operation_receipt_changed == true and .evidence_links_changed == true and .claim_transitions_changed == true)] | length) == 1 and ([.interventions[] | select(.kind == "NONSEMANTIC" and .raw_digest_changed == true and .semantic_digest_preserved == true and .consumer_decision_preserved == true)] | length) == 1 and ([.cases[] | select(.id == "coherent-tamper-reconstruction" and .observed_reason == "OPERATION_RECONSTRUCTION_MISMATCH")] | length) == 1 and ([.cases[] | select(.id == "recipe-only-mismatch" and .observed_reason == "INDEPENDENT_RECIPE_MISMATCH" and .coordinate.stage == "CONSUME_RECIPE" and .coordinate.step == "recipe")] | length) == 1 and ([.cases[] | select(.id == "coherent-claim-proposition-tamper" and .observed_reason == "PROOF_CLAIM_STATEMENT_MISMATCH")] | length) == 1 and ([.cases[] | select(.id == "coherent-claim-dependency-tamper" and .observed_reason == "PROOF_CLAIM_STATEMENT_MISMATCH")] | length) == 1 and ([.cases[] | select(.id == "coherent-claim-proof-choice-tamper" and .observed_reason == "PROOF_CLAIM_STATEMENT_MISMATCH")] | length) == 1 and ([.cases[] | select(.id == "coherent-claim-target-tamper" and .observed_reason == "PROOF_CLAIM_STATEMENT_MISMATCH")] | length) == 1 and ([.cases[] | select(.id == "unauthorized-consumer" and .observed_reason == "UNAUTHORIZED_CONSUMER_NOT_ATTESTED" and .consumer_output_exists == false and .consumer_output_digest == "" and .consumer_error_class == "ATTESTATION_MISMATCH")] | length) == 1 and ([.counterexamples[] | select(.id == "bundle-not-provided" and .coordinate.reason == "BUNDLE_CONSUMPTION_NOT_OBSERVED" and .to == "OPEN")] | length) == 1 and ([.counterexamples[] | select(.id == "bundle-corrupt" and .error_class == "BUNDLE_INVALID" and .to == "OPEN")] | length) == 1 and ([.counterexamples[] | select(.id == "unauthorized-attestation-mismatch" and .error_class == "ATTESTATION_MISMATCH" and .to == "REFUTED")] | length) == 1 and ([.counterexamples[] | select(.id == "consumer-target-missing" and .error_class == "TARGET_MISSING" and .to == "OPEN")] | length) == 1 and ([.counterexamples[] | select(.id == "consumer-output-absent" and .error_class == "RECEIPT_MISMATCH" and .output_exists == false and .to == "OPEN")] | length) == 1 and ([.counterexamples[] | select(.id == "proof-false" and .coordinate.reason == "PROOF_NOT_SATISFIED" and .to == "OPEN")] | length) == 1 and ([.counterexamples[] | select(.id == "main-indicator-38-of-40" and .coordinate.reason == "INDICATOR_GATE_NOT_SATISFIED" and .to == "OPEN")] | length) == 1' "$output/bundle-report.json"
jq -e '.proof_summary == {phase:"FINAL",proofs:3,evidence_validated:3,evidence_validated_total:3,observed_state:0,observed_state_total:3,open:0,open_total:3,discharged:3,discharged_total:3,authority:1,authority_total:1} and ([.proofs[] | select(.phase == "FINAL" and .state == "DISCHARGED" and .evidence_validated == true and .passed == true and .consumer_gate_open == false)] | length) == 3' "$output/bundle-report.json"

jq '.consumer_receipt = {} | .digest = ""' "$output/bundle-report.json" > "$output/final-receipt-missing-mutated.json"
go run ./cmd/language-proof-carrying-artifact-verifier -reseal-report "$output/final-receipt-missing-mutated.json" -output "$output/final-receipt-missing.json"
if go run ./cmd/language-proof-carrying-artifact-verifier -check "$output/final-receipt-missing.json" > "$output/final-receipt-missing-check.log" 2>&1; then
  echo "final receipt-missing report unexpectedly validated" >&2
  exit 1
fi
rg -F 'stage=CONSUME_BUNDLE step=consumer-receipt reason=CONSUMER_RECEIPT_MISSING' "$output/final-receipt-missing-check.log"
if go run ./cmd/language-proof-carrying-artifact-consumer -bundle "$output/bundle.json" -report "$output/final-receipt-missing.json" -target artifact.json -out "$output/final-receipt-missing-consumer-receipt.json" > "$output/final-receipt-missing-consumer.log" 2>&1; then
  echo "final receipt-missing consumer unexpectedly succeeded" >&2
  exit 1
fi
rg -F 'ATTESTATION_MISMATCH' "$output/final-receipt-missing-consumer.log"

sha256sum "$output"/* > "$output/manifest.sha256"
