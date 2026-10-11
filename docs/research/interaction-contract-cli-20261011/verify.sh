#!/bin/sh
set -eu
task_root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$task_root"
shasum -a 256 -c SHA256SUMS
task_tmp=$(mktemp -d)
trap 'rm -rf "$task_tmp"' EXIT HUP INT TERM
for task_step in context deterministic model native replay delta; do
  gzip -dc "$task_step.json.gz" > "$task_tmp/$task_step.json"
done
(cd "$task_tmp" && shasum -a 256 -c "$task_root/ORIGINAL-SHA256SUMS")
jq -n --arg source_sha "$(cat source-sha.txt)" \
  --slurpfile context "$task_tmp/context.json" \
  --slurpfile deterministic "$task_tmp/deterministic.json" \
  --slurpfile model "$task_tmp/model.json" \
  --slurpfile native "$task_tmp/native.json" \
  --slurpfile replay "$task_tmp/replay.json" \
  --slurpfile delta "$task_tmp/delta.json" \
  -f audit.jq > "$task_tmp/audit-summary.json"
cmp audit-summary.json "$task_tmp/audit-summary.json"
printf '%s\n' 'Saved interaction/native observations verified; no model or program executed.'
