#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"

root="${GITHUB_WORKSPACE:-$(pwd)}"
example="$root/examples/symbolic-invocation-schema"
work="${RUNNER_TEMP:-/tmp}/symbolic-invocation-schema"
binary="$work/gooo"
validator="$work/bin/jv"
mkdir -p "$work/bin"

before_workspace=$(git status --porcelain=v1 --untracked-files=all)
go fix ./cmd/gooo ./internal/packageruntime/artifactemit
git diff --exit-code -- cmd/gooo internal/packageruntime/artifactemit
mapfile -t go_files < <(find cmd/gooo internal/packageruntime/artifactemit -name '*.go' -type f | sort)
gofmt -w "${go_files[@]}"
git diff --exit-code -- cmd/gooo internal/packageruntime/artifactemit
go test ./cmd/gooo ./internal/packageruntime/artifactemit
go build -trimpath -o "$binary" ./cmd/gooo
GOBIN="$work/bin" go install github.com/santhosh-tekuri/jsonschema/cmd/jv@v0.7.0

emit=("$binary" emit --kind symbolic-invocation-schema --entry Checkout "$example")
"${emit[@]}" > "$work/artifact.json"
"${emit[@]}" > "$work/replay.json"
cmp -s "$work/artifact.json" "$work/replay.json"
jq -e '.decision=="PASS" and .resolution=="SYMBOLIC_ONLY" and
  .reason=="SYMBOLIC_INVOCATION_SCHEMA_EMITTED" and .kind=="symbolic-invocation-schema" and
  .schema=="gooo/symbolic-invocation-schema-artifact/v1" and
  .json_schema."$schema"=="https://json-schema.org/draft/2020-12/schema" and
  (.json_schema.examples|length)==1 and
  .extensions.registered_emitters==3 and .effects.repository_writes==0 and
  .effects.mutation_authority==false' "$work/artifact.json"
jq '.json_schema' "$work/artifact.json" > "$work/schema.json"
jq -S '.json_schema.examples[0]' "$work/artifact.json" > "$work/generated-accepted.json"
jq -S . "$example/accepted.json" > "$work/independent-accepted.json"
cmp -s "$work/generated-accepted.json" "$work/independent-accepted.json"

"$validator" "$work/schema.json" "$work/generated-accepted.json" > "$work/accepted-validation.txt"
set +e
"$validator" "$work/schema.json" "$example/rejected.json" > "$work/rejected-validation.txt" 2>&1
rejected_code=$?
set -e
test "$rejected_code" -eq 1

: > "$work/samples.jsonl"
for sequence in 1 2 3 4 5; do
	start_ns=$(date +%s%N)
	/usr/bin/time -f '%M' -o "$work/rss-$sequence.txt" "${emit[@]}" > "$work/sample-$sequence.json"
	end_ns=$(date +%s%N)
	cmp -s "$work/artifact.json" "$work/sample-$sequence.json"
	wall_ms=$(((end_ns - start_ns + 999999) / 1000000))
	rss_kib=$(tr -d '[:space:]' < "$work/rss-$sequence.txt")
	jq -nc --argjson sequence "$sequence" --argjson wall "$wall_ms" --argjson rss "$rss_kib" \
		'{sequence:$sequence,wall_ms:$wall,rss_kib:$rss}' >> "$work/samples.jsonl"
done

gooo_files=$(find "$example" -maxdepth 1 -type f -name '*.gooo' | wc -l | tr -d '[:space:]')
go_files_count=$(find "$example" -maxdepth 1 -type f -name '*.go' | wc -l | tr -d '[:space:]')
gooo_lines=$(cat "$example"/*.gooo | wc -l | tr -d '[:space:]')
files=$(find "$example" -maxdepth 1 -type f | wc -l | tr -d '[:space:]')
directories=$(find "$example" -mindepth 1 -type d | wc -l | tr -d '[:space:]')
binary_bytes=$(stat -c '%s' "$binary")
binary_digest="sha256:$(sha256sum "$binary" | cut -d' ' -f1)"
validator_digest="sha256:$(sha256sum "$validator" | cut -d' ' -f1)"
schema_digest="sha256:$(sha256sum "$work/schema.json" | cut -d' ' -f1)"
generated_instance_digest="sha256:$(sha256sum "$work/generated-accepted.json" | cut -d' ' -f1)"

jq -s --arg head "$HEAD_SHA" --arg binary_digest "$binary_digest" \
	--arg validator_digest "$validator_digest" --arg schema_digest "$schema_digest" \
	--arg generated_instance_digest "$generated_instance_digest" \
	--argjson binary_bytes "$binary_bytes" --argjson gooo_files "$gooo_files" \
	--argjson go_files "$go_files_count" --argjson gooo_lines "$gooo_lines" \
	--argjson files "$files" --argjson directories "$directories" \
	--slurpfile artifact "$work/artifact.json" \
	'{schema:"gooo/symbolic-invocation-schema-receipt/v1",decision:"PASS",resolution:"EXACT",
	reason:"EXTERNAL_SCHEMA_VALIDATION_OBSERVED",subject_sha:$head,
	compiler:{go_version:"1.27.0",binary_digest:$binary_digest,binary_bytes:$binary_bytes,
	registered_emitters:$artifact[0].extensions.registered_emitters},
	source:{gooo_files:$gooo_files,go_files:$go_files,gooo_lines:$gooo_lines,files:$files,directories:$directories},
	artifact:{kind:$artifact[0].kind,artifact_schema:$artifact[0].schema,digest:$artifact[0].digest,
	json_schema_dialect:$artifact[0].json_schema."$schema",json_schema_digest:$schema_digest},
	validation:{tool:"github.com/santhosh-tekuri/jsonschema/cmd/jv@v0.7.0",tool_digest:$validator_digest,
	accepted_instances:1,rejected_instances:1,generated_instances:1,generated_golden_matches:1,
	generated_instance_digest:$generated_instance_digest},deterministic_replays:1,
	resources:{samples:.,sample_count:length,max_wall_ms:(map(.wall_ms)|max),max_rss_kib:(map(.rss_kib)|max)},
	effects:{repository_writes:0,mutation_authority:false},
	not_claimed:["value-level types","domain correctness","production readiness","performance beyond this runner and fixed samples"]}' \
	"$work/samples.jsonl" > "$work/receipt.json"

after_workspace=$(git status --porcelain=v1 --untracked-files=all)
test "$before_workspace" = "$after_workspace"
jq -e '.decision=="PASS" and .source=={gooo_files:3,go_files:0,gooo_lines:16,files:6,directories:0} and
  .validation.accepted_instances==1 and .validation.rejected_instances==1 and
  .validation.generated_instances==1 and .validation.generated_golden_matches==1 and
  .deterministic_replays==1 and .resources.sample_count==5 and
  .effects=={repository_writes:0,mutation_authority:false}' "$work/receipt.json"

{
	echo '## Gooo symbolic invocation schema producer'
	echo
	echo '| Observation | Exact value |'
	echo '|---|---:|'
	jq -r '"| Gooo files / lines | \(.source.gooo_files) / \(.source.gooo_lines) |",
		"| Go files / nested directories | \(.source.go_files) / \(.source.directories) |",
		"| Generated invocation / golden match | \(.validation.generated_instances) / \(.validation.generated_golden_matches) |",
		"| Deterministic replays | \(.deterministic_replays) / 1 |",
		"| External accepted / rejected | \(.validation.accepted_instances) / \(.validation.rejected_instances) |",
		"| Runner samples | \(.resources.sample_count) |",
		"| Max RSS KiB | \(.resources.max_rss_kib) |",
		"| Repository writes | \(.effects.repository_writes) |"' "$work/receipt.json"
	echo
	echo 'This producer emits observations only; no language-quality or delivery score is promoted here.'
} >> "$GITHUB_STEP_SUMMARY"

bash "$(dirname "$0")/compile-value-contract.sh"
bash "$(dirname "$0")/compile-value-reachability.sh"

bash "$(dirname "$0")/compile-reader-resolution.sh"
bash "$(dirname "$0")/compile-reader-request.sh"
