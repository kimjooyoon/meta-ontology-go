#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "$PROOF_CARRYING_OUTPUT"
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

source_path="examples/language-proof-carrying-artifact/main.gooo"
recipe="examples/language-proof-carrying-artifact/recipe.json"
contract="examples/language-proof-carrying-artifact/contract.json"

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

go fix ./cmd/language-proof-carrying-artifact ./internal/meta/languageproofartifact ./cmd/language-proof-carrying-artifact-verifier ./internal/meta/languageproofartifactverifier
test -z "$(git diff -- cmd/language-proof-carrying-artifact internal/meta/languageproofartifact cmd/language-proof-carrying-artifact-verifier internal/meta/languageproofartifactverifier)"
test -z "$(gofmt -l cmd/language-proof-carrying-artifact internal/meta/languageproofartifact cmd/language-proof-carrying-artifact-verifier internal/meta/languageproofartifactverifier)"
go test ./cmd/language-proof-carrying-artifact ./internal/meta/languageproofartifact ./cmd/language-proof-carrying-artifact-verifier ./internal/meta/languageproofartifactverifier

go run ./cmd/gooo check "$source_path"
go run ./cmd/gooo run --json --entry GenerateProofCarryingArtifact "$source_path" > "$output/operation-receipt.json"

before_entries="$(jq -c '.entries' "$output/repository-before.json")"
before_digest="sha256:$(printf '%s' "$before_entries" | sha256sum | awk '{print $1}')"

go list -deps ./cmd/language-proof-carrying-artifact-verifier | sort > "$output/verifier-dependencies.txt"
internal_dependencies="$(grep -E '^github.com/kimjooyoon/meta-ontology-go/internal/' "$output/verifier-dependencies.txt" || true)"
producer_import_numerator="$(printf '%s\n' "$internal_dependencies" | grep -E '/internal/(meta/languageproofartifact|sourceexecution)(/|$)' | sed '/^$/d' | wc -l | tr -d ' ' || true)"
producer_import_denominator="$(printf '%s\n' "$internal_dependencies" | sed '/^$/d' | wc -l | tr -d ' ' || true)"
core_parser_dependencies="$(printf '%s\n' "$internal_dependencies" | grep -E '/internal/(syntax|bidir)(/|$)' | sed '/^$/d' | wc -l | tr -d ' ' || true)"
jq -n --arg schema "gooo/language-proof-carrying-artifact-independence/v1" --argjson producer_dependencies "$producer_import_numerator" --argjson producer_import_numerator "$producer_import_numerator" --argjson producer_import_denominator "$producer_import_denominator" --argjson core_parser_dependencies "$core_parser_dependencies" '{schema:$schema,producer_dependencies:$producer_dependencies,producer_import_numerator:$producer_import_numerator,producer_import_denominator:$producer_import_denominator,core_parser_dependencies:$core_parser_dependencies}' > "$output/independence.json"
test "$producer_import_numerator" -eq 0
test "$producer_import_denominator" -gt 0

make_write_set() {
  local after_entries after_digest changed changed_file
  snapshot_repo "$output/repository-after.json"
  after_entries="$(jq -c '.entries' "$output/repository-after.json")"
  after_digest="sha256:$(printf '%s' "$after_entries" | sha256sum | awk '{print $1}')"
  changed_file="$output/repository-write-set-changed.json"
  jq -n --slurpfile before "$output/repository-before.json" --slurpfile after "$output/repository-after.json" 'def by_path: reduce .[] as $item ({}; .[$item.path] = $item); ($before[0].entries | by_path) as $b | ($after[0].entries | by_path) as $a | (($b | keys) + ($a | keys) | unique | sort) as $paths | [$paths[] as $path | if ($b[$path] != null and $a[$path] != null and $b[$path] == $a[$path]) then empty else {path:$path,before_digest:($b[$path].digest // ""),after_digest:($a[$path].digest // ""),before_kind:($b[$path].kind // ""),after_kind:($a[$path].kind // "")} end]' > "$changed_file"
  jq -n --arg schema "gooo/repository-write-set-observation/v1" --slurpfile before "$output/repository-before.json" --slurpfile after "$output/repository-after.json" --slurpfile changed "$changed_file" --arg before_digest "$before_digest" --arg after_digest "$after_digest" '{schema:$schema,version:1,before:$before[0].entries,after:$after[0].entries,changed:$changed[0],before_digest:$before_digest,after_digest:$after_digest,repository_writes:($changed[0]|length),mutation_authority:false,digest:""}' > "$output/write-set.json"
}

produce_artifact() {
  local source="$1" operation="$2" artifact="$3"
  go run ./cmd/language-proof-carrying-artifact -head "$HEAD_SHA" -source-path "$source" -source "$source" -operation "$operation" -recipe "$recipe" -write-set "$output/write-set.json" -out "$artifact"
}

run_receipt() {
  local source="$1" receipt="$2"
  go run ./cmd/gooo run --json --entry GenerateProofCarryingArtifact "$source" > "$receipt"
}

sed 's|gooo://proof/source-evidence|gooo://proof/source-evidence-v2|' "$source_path" > "$output/semantic-intervention.gooo"
{
  printf '%s\n' '// comment-only intervention: syntax comments are not semantic declarations.'
  cat "$source_path"
} > "$output/comment-only-intervention.gooo"
run_receipt "$output/semantic-intervention.gooo" "$output/semantic-operation-receipt.json"
run_receipt "$output/comment-only-intervention.gooo" "$output/comment-operation-receipt.json"

make_write_set
produce_artifact "$source_path" "$output/operation-receipt.json" "$output/artifact.json"
produce_artifact "$output/semantic-intervention.gooo" "$output/semantic-operation-receipt.json" "$output/semantic-intervention-artifact.json"
produce_artifact "$output/comment-only-intervention.gooo" "$output/comment-operation-receipt.json" "$output/comment-only-intervention-artifact.json"

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
jq '.steps[0].meta_operation = "accept-source-claim-without-recheck"' "$recipe" > "$output/wrong-recipe.json"

coherent_body="$(jq -c --arg receipt_digest "$coherent_operation_digest" '.evidence |= map(if .kind == "OPERATION" then .receipt_digest=$receipt_digest else . end) | .digest=""' "$output/artifact.json")"
coherent_evidence_body="$(jq -c '.evidence[] | select(.kind == "OPERATION") | .evidence_digest=""' <<< "$coherent_body")"
coherent_evidence_digest="sha256:$(printf '%s' "$coherent_evidence_body" | sha256sum | awk '{print $1}')"
coherent_body="$(jq -c --arg evidence_digest "$coherent_evidence_digest" '.evidence |= map(if .kind == "OPERATION" then .evidence_digest=$evidence_digest else . end)' <<< "$coherent_body")"
coherent_ledger="$(jq -c --arg evidence_digest "$coherent_evidence_digest" '.entries |= map(if .claim_id == "operation-receipt-bound" then .evidence_digests=[$evidence_digest] else . end) | .digest=""' <<< "$(jq -c '.prior_ledger' <<< "$coherent_body")")"
coherent_entry_body="$(jq -c '.entries[1] | .digest=""' <<< "$coherent_ledger")"
coherent_entry_digest="sha256:$(printf '%s' "$coherent_entry_body" | sha256sum | awk '{print $1}')"
coherent_ledger="$(jq -c --arg digest "$coherent_entry_digest" '.entries[1].digest=$digest | .entries[2].previous_digest=$digest | .digest=""' <<< "$coherent_ledger")"
coherent_entry_body="$(jq -c '.entries[2] | .digest=""' <<< "$coherent_ledger")"
coherent_entry_digest="sha256:$(printf '%s' "$coherent_entry_body" | sha256sum | awk '{print $1}')"
coherent_ledger="$(jq -c --arg digest "$coherent_entry_digest" '.entries[2].digest=$digest | .digest=""' <<< "$coherent_ledger")"
coherent_ledger_digest="sha256:$(printf '%s' "$coherent_ledger" | sha256sum | awk '{print $1}')"
coherent_ledger="$(jq -c --arg digest "$coherent_ledger_digest" '.digest=$digest' <<< "$coherent_ledger")"
coherent_body="$(jq -c --argjson prior_ledger "$coherent_ledger" '.prior_ledger=$prior_ledger' <<< "$coherent_body")"
coherent_artifact_digest="sha256:$(printf '%s' "$coherent_body" | sha256sum | awk '{print $1}')"
jq -c --arg digest "$coherent_artifact_digest" '.digest=$digest' <<< "$coherent_body" > "$output/coherent-tamper.json"

run_verifier() {
  local report="$1"
  go run ./cmd/language-proof-carrying-artifact-verifier -head "$HEAD_SHA" -contract "$contract" -valid "$output/artifact.json" -tampered "$output/tampered.json" -coherent-tamper "$output/coherent-tamper.json" -missing "$output/missing-operation.json" -byte-only "$output/byte-only.json" -wrong-recipe "$output/wrong-recipe.json" -source "$source_path" -operation "$output/operation-receipt.json" -recipe "$recipe" -independence "$output/independence.json" -write-set "$output/write-set.json" -coherent-operation "$output/coherent-operation-receipt.json" -semantic-artifact "$output/semantic-intervention-artifact.json" -semantic-source "$output/semantic-intervention.gooo" -semantic-operation "$output/semantic-operation-receipt.json" -comment-artifact "$output/comment-only-intervention-artifact.json" -comment-source "$output/comment-only-intervention.gooo" -comment-operation "$output/comment-operation-receipt.json" -output "$report"
}

run_verifier "$output/report.json"
run_verifier "$output/replay.json"
cmp -s "$output/report.json" "$output/replay.json"
go run ./cmd/language-proof-carrying-artifact-verifier -check "$output/report.json"

jq -e '.conformance_decision == "PASS" and .conformance_resolution == "EXACT" and .subject_artifact_decision == "CARRIED" and .artifact_use_authority == "READ_ONLY_CONSUMPTION" and .summary.cases_satisfied == 6 and .summary.cases_total == 6 and .summary.valid_artifacts == 1 and .summary.evidence_kinds_carried == 3 and .summary.exact_evidence_links == 3 and .summary.recipe_matches == 1 and .summary.preserved_transitions == 3 and .summary.transition_total == 4 and .summary.tampered_rejections == 1 and .summary.coherent_tamper_rejections == 1 and .summary.missing_evidence_rejections == 1 and .summary.byte_only_denials == 1 and .summary.recipe_rejections == 1 and .summary.ledger_discharged_claims == 3 and .summary.ledger_open_claims == 6 and .summary.ledger_refuted_claims == 9 and .summary.semantic_interventions == 1 and .summary.nonsemantic_interventions == 1 and .summary.read_only_authorities == 1 and .summary.producer_dependencies == 0 and .summary.producer_import_numerator == 0 and .summary.producer_import_denominator > 0 and .summary.core_parser_dependencies > 0 and .summary.generated_authority == 0 and .summary.semantic_claims == 0 and .summary.repository_writes == 0 and .summary.mutation_authorities == 0 and .summary.promotion_authorities == 0 and .summary.semantic_authorities == 0 and .write_set.repository_writes == 0 and .write_set.mutation_authority == false and ([.indicators[] | select(.satisfied == false)] | length) == 0 and ([.claim_transitions[] | select(.from == "CARRIED" and .to == "PRESERVED")] | length) == 3 and ([.claim_transitions[] | select(.claim_id == "consumer-authority" and .capability == "ARTIFACT_USE" and .from == "NONE" and .to == "READ_ONLY_CONSUMPTION")] | length) == 1 and ([.interventions[] | select(.status == "SATISFIED")] | length) == 2 and ([.interventions[] | select(.kind == "SEMANTIC" and .semantic_digest_changed == true and .operation_receipt_changed == true and .evidence_links_changed == true and .claim_transitions_changed == true)] | length) == 1 and ([.interventions[] | select(.kind == "NONSEMANTIC" and .raw_digest_changed == true and .semantic_digest_preserved == true and .consumer_decision_preserved == true)] | length) == 1 and ([.cases[] | select(.id == "coherent-tamper-reconstruction" and .observed_reason == "OPERATION_RECONSTRUCTION_MISMATCH")] | length) == 1' "$output/report.json"

sha256sum "$output"/*.json > "$output/manifest.sha256"
