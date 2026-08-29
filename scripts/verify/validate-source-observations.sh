#!/usr/bin/env bash
set -euo pipefail

metrics_dir=${1:?source metrics directory is required}
metrics="$metrics_dir/source-metrics.json"
binding="$metrics_dir/meta-binding-report.json"
actionability="$metrics_dir/meta-actionability-report.json"
plan="$metrics_dir/self-improvement-plan.json"
execution="$metrics_dir/self-improvement-execution.json"
receipts="$metrics_dir/self-improvement-receipts.json"
provenance="$metrics_dir/self-improvement-provenance.json"

for artifact in "$metrics" "$binding" "$actionability" "$plan" "$execution" "$receipts" "$provenance"; do
	test -s "$artifact"
done

jq -e '
  (.root | type == "string") and
  (.files | type == "array") and
  (.directories | type == "array") and
  any(.directories[]; .path == ".") and
  (.meta | type == "object") and
  (.meta.schema == "gooo/source-policy/v1") and
  (.meta.policy.schema == "gooo/source-policy/v1") and
  (.meta.policy.max_file_lines == 75) and
  (.meta.policy.max_function_lines == 75) and
  (.meta.policy.exempt_project_root_readme == true) and
  (.meta.indicators | type == "array") and
  all(.meta.indicators[]; (.metric_id | type == "string") and (.subject | type == "string") and (.value | type == "number") and (.limit | type == "number"))
' "$metrics" >/dev/null
jq -e '
  (.schema == "gooo/meta-binding-report/v1") and
  (.decision == "PASS") and
  (.summary.unbound_indicators == 0) and
  (.witnesses | type == "array") and
  (.report_digest | strings | test("^sha256:[0-9a-f]{64}$"))
' "$binding" >/dev/null
jq -e '
  (.schema == "gooo/meta-actionability-report/v1") and
  (.indicators | type == "array") and
  (.operations | type == "array") and
  (.report_digest | strings | test("^sha256:[0-9a-f]{64}$"))
' "$actionability" >/dev/null
jq -e '
  (.schema_version == "gooo/self-improvement-generation/v6") and
  (.decision == "PLAN" or .decision == "FIXED_POINT" or .decision == "UNKNOWN" or .decision == "REJECTED") and
  (.registry | type == "array") and
  (.selected | type == "array") and
  (.unknown_indicator_ids | type == "array") and
  (.refuted_indicator_ids | type == "array") and
  (.counterexamples | type == "array") and
  (.promotion_authorized == false) and
  (.plan_digest | strings | test("^[0-9a-f]{64}$")) and
  (.replay_digest | strings | test("^[0-9a-f]{64}$"))
' "$plan" >/dev/null
jq -e '
  (.schema_version == "gooo/meta-operation-execution/v6") and
  (.decision == "PROPOSED" or .decision == "FIXED_POINT" or .decision == "UNKNOWN" or .decision == "REJECTED") and
  (.steps | type == "array") and
  (.promotion_authorized == false) and
  (.manifest_digest | strings | test("^[0-9a-f]{64}$")) and
  (.replay_digest | strings | test("^[0-9a-f]{64}$"))
' "$execution" >/dev/null
jq -e '
  (.schema_version == "gooo/meta-operation-receipt-report/v2") and
  (.decision == "CONFORMANT" or .decision == "FIXED_POINT" or .decision == "UNKNOWN" or .decision == "REJECTED" or .decision == "REFUTED") and
  (.receipts | type == "array") and
  (.failures | type == "array") and
  (.unknowns | type == "array") and
  (.promotion_authorized == false) and
  (.report_digest | strings | test("^[0-9a-f]{64}$")) and
  (.replay_digest | strings | test("^[0-9a-f]{64}$"))
' "$receipts" >/dev/null
jq -e '
  (.schema_version == "gooo/meta-artifact-provenance/v1") and
  (.decision == "BOUND") and
  (.summary.fail == 0) and
  (.summary.unknown == 0) and
  (.summary.pass == (.indicators | length)) and
  (.promotion_authorized == false) and
  (.envelope_digest | strings | test("^[0-9a-f]{64}$")) and
  (.replay_digest | strings | test("^[0-9a-f]{64}$"))
' "$provenance" >/dev/null

cmp -s "$metrics" "$metrics_dir/source-metrics-replay.json"
cmp -s "$binding" "$metrics_dir/meta-binding-report-replay.json"
cmp -s "$actionability" "$metrics_dir/meta-actionability-report-replay.json"
cmp -s "$plan" "$metrics_dir/self-improvement-replay.json"
cmp -s "$execution" "$metrics_dir/self-improvement-execution-replay.json"
cmp -s "$receipts" "$metrics_dir/self-improvement-receipts-replay.json"
cmp -s "$provenance" "$metrics_dir/self-improvement-provenance-replay.json"

summary=${GITHUB_STEP_SUMMARY:-/dev/null}
jq -r '
  . as $report |
  (.directories[] | select(.path == ".")) as $root |
  def rows($id): [.meta.indicators[] | select(.metric_id == $id)];
  "### Source-cap observations (DRIVER; observation-only)",
  "- thresholds: go_file=\($report.meta.policy.max_file_lines), gooo_file=\($report.meta.policy.max_file_lines), function=\($report.meta.policy.max_function_lines)",
  "- root counts: direct_files=\($root.direct_files), direct_folders=\($root.direct_folders), recursive_files=\($root.recursive_files), recursive_folders=\($root.recursive_folders)",
  "- Go file candidates: \(rows("gooo.metric.source.go-file-lines.v1") | length)",
  "- Gooo file candidates: \(rows("gooo.metric.source.gooo-file-lines.v1") | length)",
  "- function candidates: \(rows("gooo.metric.source.function-lines.v1") | length)",
  (rows("gooo.metric.source.go-file-lines.v1")[] | "  - \(.subject): actual=\(.value) limit=\(.limit)"),
  (rows("gooo.metric.source.gooo-file-lines.v1")[] | "  - \(.subject): actual=\(.value) limit=\(.limit)"),
  (rows("gooo.metric.source.function-lines.v1")[] | "  - \(.subject): actual=\(.value) limit=\(.limit)")
' "$metrics" >> "$summary"
