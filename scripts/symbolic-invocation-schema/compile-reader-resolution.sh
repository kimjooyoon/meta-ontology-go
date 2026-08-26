#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${{BASH_SOURCE[0]}")/../.." && pwd)"
OUTPUT_DIR="${{RUNNER_TEMP:-/tmp}/symbolic-invocation-schema"
INPUT="${{OUTPUT_DIR}/symbolic-value-reachability.json"
OUTPUT="${{OUTPUT_DIR}/symbolic-value-reader-projection.json"

: "${{HEAD_SHA:?HEAD_SHA is required}"
mkdir -p "${{OUTPUT_DIR}"

cd "${{ROOT_DIR}"
go run ./scripts/symbolic-invocation-schema/valueresolution \
	"${{INPUT}" "${{HEAD_SHA}" "${{OUTPUT}"

jq -e '
	.schema == "gooo/symbolic-value-reader-projection/v1" and
	.metric_id == "gooo.metric.compiler.symbolic-value-reader-projection.v1" and
	.decision == "PASS" and .resolution == "READER_PROJECTION_ONLY" and
	.coordinates == {"satisfied":18,"total":18,"basis_points":10000} and
	([.readers[].coordinates.total] == [5,9,11]) and
	([.views[].total] == [5,14,18]) and
	.effects.repository_writes == 0 and
	.effects.mutation_authority == false and
	.promotion_credit_bps == 0
' "${{OUTPUT}" >/dev/null

printf 'symbolic-value-reader-projection.json: OK\n'
