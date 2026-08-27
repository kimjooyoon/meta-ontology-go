#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$root"
out=${WORKGRAPH_OUTPUT:-workgraph-output}
head_sha=${EXACT_SHA:-${GITHUB_SHA:-0000000000000000000000000000000000000000}}
source_path=examples/workgraph/main.gooo
contract=examples/workgraph/project.json
mkdir -p "$out/generated/first" "$out/generated/replay" "$out/reports"

go fix ./internal/meta/workgraph ./cmd/workgraph-witness
git diff --exit-code -- internal/meta/workgraph cmd/workgraph-witness
test -z "$(gofmt -l internal/meta/workgraph cmd/workgraph-witness)"
go test ./internal/meta/workgraph ./cmd/workgraph-witness
go vet ./internal/meta/workgraph ./cmd/workgraph-witness

go run ./cmd/gooo check "$source_path" > "$out/check.stdout"
common=(-contract "$contract" -source "$source_path" -check-receipt "$out/check.stdout" -head "$head_sha")
go run ./cmd/workgraph-witness "${common[@]}" -out "$out/reports/before.json" -expect FAIL_CLOSED

go run ./cmd/gooo generate "$source_path" --out "$out/generated/first"
go run ./cmd/gooo generate "$source_path" --out "$out/generated/replay"
first="$out/generated/first/semantic.gooo.go"
replay="$out/generated/replay/semantic.gooo.go"
test -s "$first"
cmp "$first" "$replay"

evidence=(-generated "$first" -replay "$replay" -predecessor "$out/reports/before.json")
go run ./cmd/workgraph-witness "${common[@]}" "${evidence[@]}" -resource-out "$out/resource.json" -out "$out/reports/report.json" -expect VERTICAL_SLICE_CLOSED
go run ./cmd/workgraph-witness "${common[@]}" "${evidence[@]}" -resource "$out/resource.json" -out "$out/reports/replay.json" -expect VERTICAL_SLICE_CLOSED
cmp "$out/reports/report.json" "$out/reports/replay.json"

cp "$replay" "$out/generated/mismatch.go"
printf '\n// deterministic counterexample\n' >> "$out/generated/mismatch.go"
go run ./cmd/workgraph-witness "${common[@]}" -generated "$first" -replay "$out/generated/mismatch.go" -predecessor "$out/reports/before.json" -resource "$out/resource.json" -out "$out/reports/refuted.json" -expect FAIL_CLOSED

jq -e '.decision == "FAIL_CLOSED" and .resolution == "OPERATION_CLASS" and .summary.unknown_gates == 3 and .next_operation == "RUN_GOOO_GENERATE_REPLAY"' "$out/reports/before.json" >/dev/null
jq -e '.decision == "VERTICAL_SLICE_CLOSED" and .summary.total_gates == 7 and .summary.closed_gates == 7 and .summary.unknown_gates == 0 and .summary.refuted_gates == 0 and .summary.active_claims == 0 and .summary.discharged_claims == 1 and .claim.before.state == "UNKNOWN" and .claim.after.status == "DISCHARGED" and .claim.trace_retained and (.indicators | length) == 6 and ([.indicators[].state] | all(. == "SATISFIED"))' "$out/reports/report.json" >/dev/null
jq -e '.decision == "FAIL_CLOSED" and .resolution == "EXACT" and .summary.refuted_gates >= 1' "$out/reports/refuted.json" >/dev/null

if [[ -n ${GITHUB_STEP_SUMMARY:-} ]]; then
  jq -r '"### Gooo Workgraph vertical slice\n- decision: `\(.decision)` (`\(.reason)`)\n- gates: `\(.summary.closed_gates)/\(.summary.total_gates)`\n- claim: `\(.claim.before.state) -> \(.claim.after.status)`\n- resources: wall `\(.resource.wall_nanoseconds)` ns, heap sys `\(.resource.heap_sys_bytes)` bytes, allocated `\(.resource.total_alloc_bytes)` bytes\n- repository writes: `\(.summary.repository_writes)`"' "$out/reports/report.json" >> "$GITHUB_STEP_SUMMARY"
fi
