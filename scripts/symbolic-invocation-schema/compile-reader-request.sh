#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUTPUT_DIR="${RUNNER_TEMP:-/tmp}/symbolic-invocation-schema"
REQUEST="${ROOT_DIR}/examples/symbolic-invocation-schema/reader-request.gooo"
INPUT="${OUTPUT_DIR}/symbolic-value-reader-projection.json"
OUTPUT="${OUTPUT_DIR}/symbolic-reader-request-result.json"

: "${HEAD_SHA:?HEAD_SHA is required}"
mkdir -p "${OUTPUT_DIR}"

cd "${ROOT_DIR}"
go run ./scripts/symbolic-invocation-schema/readerrequest \
	"${REQUEST}" "${INPUT}" "${HEAD_SHA}" "${OUTPUT}"

jq -e '
	.schema == "gooo/symbolic-reader-request-result/v1" and
	.metric_id == "gooo.metric.compiler.symbolic-reader-request-result.v1" and
	.decision == "PASS" and .resolution == "GOOO_REQUEST_BOUND_ONLY" and
	.request.audience == "USER" and
	.request.expected_resolution == "DECISION_AND_COUNTS_ONLY" and
	.coordinates == {"satisfied":12,"total":12,"basis_points":10000} and
	([.classes[] | [.class,.satisfied,.total]] ==
	 [["OUTCOME",3,3],["DRIVER",4,4],["GUARDRAIL",5,5]]) and
	([.proofs[] | [.proof_choice,.satisfied,.total]] ==
	 [["FOUNDATION",5,5],["COHERENCE",4,4],["REGRESSION",3,3]]) and
	.view.audience == "USER" and .view.coordinates.total == 5 and
	(.view.indicator_ids | length) == 5 and
	.effects == {"repository_writes":0,"mutation_authority":false} and
	.promotion_credit_bps == 0
' "${OUTPUT}" >/dev/null

printf 'symbolic-reader-request-result.json: OK\n'
