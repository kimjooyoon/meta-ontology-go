#!/usr/bin/env bash
set -euo pipefail

root="${GITHUB_WORKSPACE:-$(pwd)}"
work="${RUNNER_TEMP:-/tmp}/artifact-resolution-experiment"
binary="$work/gooo"
reducer="$work/artifact-resolution-experiment"
contract="$root/examples/billing-package/artifact-resolution.contract.json"
manifest_golden="$root/examples/billing-package/operation-manifest.golden.json"
interface_golden="$root/examples/billing-package/operation-interface.golden.json"
mkdir -p "$work"

go fix ./...
git diff --exit-code
test -z "$(gofmt -l cmd/gooo cmd/artifact-resolution-experiment internal/packageruntime/artifactemit internal/meta/artifactresolutionexperiment)"
go test ./cmd/gooo ./cmd/artifact-resolution-experiment ./internal/packageruntime/artifactemit ./internal/meta/artifactresolutionexperiment
go build -trimpath -o "$binary" ./cmd/gooo
go build -trimpath -o "$reducer" ./cmd/artifact-resolution-experiment

"$binary" emit --kind operation-manifest --entry PayOrder examples/billing-package > "$work/manifest.json"
"$binary" emit --kind operation-manifest --entry PayOrder examples/billing-package > "$work/manifest-replay.json"
"$binary" emit --kind operation-interface --entry PayOrder examples/billing-package > "$work/interface.json"
"$binary" emit --kind operation-interface --entry PayOrder examples/billing-package > "$work/interface-replay.json"
if "$binary" emit --kind not-registered --entry PayOrder examples/billing-package > "$work/unknown-emitter.json"; then
  echo "unknown emitter unexpectedly passed" >&2
  exit 1
fi

jq -n --arg subject "$HEAD_SHA" \
  --slurpfile contract "$contract" \
  --slurpfile manifest "$work/manifest.json" \
  --slurpfile manifest_replay "$work/manifest-replay.json" \
  --slurpfile manifest_golden "$manifest_golden" \
  --slurpfile interface "$work/interface.json" \
  --slurpfile interface_replay "$work/interface-replay.json" \
  --slurpfile interface_golden "$interface_golden" \
  --slurpfile unknown "$work/unknown-emitter.json" \
  '{subject_sha:$subject,contract:$contract[0],manifest:$manifest[0],manifest_replay:$manifest_replay[0],manifest_golden:$manifest_golden[0],interface:$interface[0],interface_replay:$interface_replay[0],interface_golden:$interface_golden[0],unknown_emitter:$unknown[0]}' \
  > "$work/input.json"

"$reducer" -input "$work/input.json" -output "$work/report.json"
"$reducer" -input "$work/input.json" -check "$work/report.json"
jq -e '.decision=="PASS" and .resolution=="EXACT" and .summary.coordinates.satisfied==13 and .summary.coordinates.total==13 and (.views[]|select(.audience=="USER")|.satisfied==5 and .total==5) and (.views[]|select(.audience=="TOOL_AUTHOR")|.satisfied==10 and .total==10) and (.views[]|select(.audience=="GOVERNOR")|.satisfied==13 and .total==13) and .summary.resolution.manifest_definitions==2 and .summary.resolution.interface_definitions==0 and .summary.resolution.registered_emitters==2 and .summary.unknowns==0' "$work/report.json"

jq '.interface.decision="UNKNOWN"' "$work/input.json" > "$work/unknown-top-input.json"
if "$reducer" -input "$work/unknown-top-input.json" -output "$work/unknown-top-report.json"; then
  echo "unknown top decision unexpectedly passed" >&2
  exit 1
fi
jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .summary.coordinates.satisfied==0 and .summary.coordinates.total==13 and .summary.unknowns==1' "$work/unknown-top-report.json"

jq '.interface_golden.operation.activity="Other"' "$work/input.json" > "$work/drift-input.json"
if "$reducer" -input "$work/drift-input.json" -output "$work/drift-report.json"; then
  echo "golden drift unexpectedly passed" >&2
  exit 1
fi
jq -e '.decision=="FAIL_CLOSED" and .resolution=="EXACT" and .summary.coordinates.satisfied==12 and .summary.coordinates.total==13' "$work/drift-report.json"
git diff --exit-code

jq -r '"### Gooo artifact resolution experiment\n- decision: \(.decision) / \(.resolution)\n- indicators: \(.summary.coordinates.satisfied)/\(.summary.coordinates.total)\n- USER: \(.views[0].satisfied)/\(.views[0].total)\n- TOOL_AUTHOR: \(.views[1].satisfied)/\(.views[1].total)\n- GOVERNOR: \(.views[2].satisfied)/\(.views[2].total)\n- definitions: full=\(.summary.resolution.manifest_definitions), interface=\(.summary.resolution.interface_definitions)\n- emitters: \(.summary.resolution.registered_emitters)\n- unknowns: \(.summary.unknowns)\n- receipt: \(.digest)"' "$work/report.json" >> "$GITHUB_STEP_SUMMARY"
