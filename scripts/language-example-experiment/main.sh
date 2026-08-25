#!/usr/bin/env bash
set -euo pipefail

: "${HEAD_SHA:?HEAD_SHA is required}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is required}"

example=examples/billing-package
contract="$example/operation-manifest.contract.json"
golden="$example/operation-manifest.golden.json"
output="${RUNNER_TEMP:-/tmp}/language-example-experiment"
mkdir -p "$output"

go fix ./cmd/gooo ./cmd/language-example-experiment ./internal/packageruntime/artifactemit ./internal/meta/languageexampleexperiment
git diff --exit-code -- cmd/gooo cmd/language-example-experiment internal/packageruntime/artifactemit internal/meta/languageexampleexperiment
mapfile -t go_files < <(find cmd/gooo cmd/language-example-experiment internal/packageruntime/artifactemit internal/meta/languageexampleexperiment -name '*.go' -type f | sort)
gofmt -w "${go_files[@]}"
git diff --exit-code -- cmd/gooo cmd/language-example-experiment internal/packageruntime/artifactemit internal/meta/languageexampleexperiment
go test ./cmd/gooo ./cmd/language-example-experiment ./internal/packageruntime/artifactemit ./internal/meta/languageexampleexperiment
go build -o "$output/gooo" ./cmd/gooo

emit=("$output/gooo" emit --kind operation-manifest --entry PayOrder "$example")
"${emit[@]}" > "$output/first.json"
"${emit[@]}" > "$output/replay.json"
cmp -s "$output/first.json" "$output/replay.json"

set +e
"$output/gooo" emit --kind not-registered --entry PayOrder "$example" > "$output/unknown-emitter.json"
unknown_emitter_code=$?
set -e
[[ "$unknown_emitter_code" == "1" ]]
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .reason=="EMITTER_UNKNOWN"' "$output/unknown-emitter.json"

: > "$output/samples.jsonl"
for sequence in 1 2 3 4 5; do
	start_ns=$(date +%s%N)
	/usr/bin/time -f '%M' -o "$output/rss-$sequence.txt" "${emit[@]}" > "$output/sample-$sequence.json"
	end_ns=$(date +%s%N)
	cmp -s "$output/first.json" "$output/sample-$sequence.json"
	wall_ms=$(((end_ns - start_ns + 999999) / 1000000))
	rss_kib=$(tr -d '[:space:]' < "$output/rss-$sequence.txt")
	jq -nc --argjson sequence "$sequence" --argjson wall "$wall_ms" --argjson rss "$rss_kib" \
		'{sequence:$sequence,wall_ms:$wall,rss_kib:$rss}' >> "$output/samples.jsonl"
done

gooo_files=$(find "$example" -maxdepth 1 -type f -name '*.gooo' | wc -l | tr -d '[:space:]')
go_files=$(find "$example" -maxdepth 1 -type f -name '*.go' | wc -l | tr -d '[:space:]')
binary_bytes=$(stat -c '%s' "$output/gooo")
binary_digest="sha256:$(sha256sum "$output/gooo" | cut -d' ' -f1)"
jq -s --arg head "$HEAD_SHA" --arg executable "$binary_digest" --argjson gooo "$gooo_files" \
	--argjson gofiles "$go_files" --argjson binary "$binary_bytes" \
	'{schema:"gooo/language-example-experiment-profile/v1",subject_sha:$head,executable_digest:$executable,
	gooo_files:$gooo,go_files:$gofiles,primary_artifacts:1,binary_bytes:$binary,samples:.,
	effects:{repository_writes:0,mutation_authority:false}}' "$output/samples.jsonl" > "$output/profile.json"

reduce() {
	local selected_golden=$1 selected_artifact=$2 selected_replay=$3 selected_profile=$4 report=$5
	go run ./cmd/language-example-experiment --expected-head "$HEAD_SHA" --contract "$contract" \
		--golden "$selected_golden" --artifact "$selected_artifact" --replay "$selected_replay" \
		--unknown-emitter "$output/unknown-emitter.json" --profile "$selected_profile" --out "$report"
}

reduce "$golden" "$output/first.json" "$output/replay.json" "$output/profile.json" "$output/report.json"
jq -e '.schema=="gooo/language-example-experiment-report/v2" and .decision=="PASS" and .resolution=="EXACT" and .interpretation=="MINIMAL_VALUE_OBSERVED"' "$output/report.json"
jq -e '.summary.coordinates=={satisfied:15,total:15,basis_points:10000}' "$output/report.json"
jq -e '[.views[]|[.audience,.satisfied,.total]]==[["USER",6,6],["TOOL_AUTHOR",12,12],["GOVERNOR",15,15]]' "$output/report.json"
jq -e '.summary.compiler=={source_files:2,gooo_files:2,go_files:0,gooo_definition_basis_points:10000,registered_emitters:3}' "$output/report.json"
jq -e '.summary.value.artifact_digest_checks==3 and .summary.resources.samples==5 and .summary.resources.valid_samples==5 and .summary.resources.wall_violations==0 and .summary.resources.rss_violations==0 and .summary.resources.binary_violations==0' "$output/report.json"
jq -e '.summary.counterexamples.unknown_emitter_rejections==1 and .summary.effects.repository_writes==0 and .summary.effects.mutation_authority==false and (.not_claimed|length)==5' "$output/report.json"

source scripts/language-example-experiment/counterexamples.sh
observe_counterexamples

git diff --exit-code
{
	echo '## Gooo operation manifest experiment'
	echo
	echo '| Reader | Observed | Fixed total | Basis points |'
	echo '|---|---:|---:|---:|'
	jq -r '.views[]|"| \(.audience) | \(.satisfied) | \(.total) | \(.basis_points) |"' "$output/report.json"
	echo
	jq -r '"Gooo definitions: **\(.summary.compiler.gooo_files)**  Go definitions: **\(.summary.compiler.go_files)**  Registered emitters: **\(.summary.compiler.registered_emitters)**"' "$output/report.json"
	jq -r '"Samples: **\(.summary.resources.samples)**  Max wall: **\(.summary.resources.max_wall_ms) ms**  Max RSS: **\(.summary.resources.max_rss_kib) KiB**  Binary: **\(.summary.resources.binary_bytes) bytes**"' "$output/report.json"
	jq -r '"Digest checks: **\(.summary.value.artifact_digest_checks)/3**  Valid samples: **\(.summary.resources.valid_samples)/5**"' "$output/report.json"
	jq -r '"Evidence counterexamples: **\(.satisfied)/\(.total)**"' "$output/counterexamples.json"
	echo
	echo 'Interpretation: **MINIMAL_VALUE_OBSERVED**. Language quality and production readiness are not claimed.'
} >> "$GITHUB_STEP_SUMMARY"
