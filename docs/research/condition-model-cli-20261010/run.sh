#!/bin/bash
# Fresh macOS observation. Every existing output directory is refused.
set -euo pipefail
test "$#" = 5
task_binary=$1
task_source=$2
task_model=$3
task_output=$4
task_go=$5
task_study=$(cd "$(dirname "$0")" && pwd)
test ! -e "$task_output"
test "$(shasum -a 256 "$task_model" | cut -d ' ' -f1)" = a16696ed44c668f38cc2e7ee1dc6ff3f4716649d57e59df7c9490d1c670edc48
mkdir "$task_output"
printf '%s\n' 'Experiment started. Keep all partial observations. Never repeat this destination.' > "$task_output/started.txt"
cp "$task_source" "$task_output/source.gooo"
cp "$task_model" "$task_output/model.json"
cp "$task_study/runtime-cases.json" "$task_output/runtime-cases.json"
"$task_binary" version --build --json > "$task_output/build.json"
jq -e '.vcs_modified=="false" and .go_version=="go1.27.2" and .decision_runtime.version=="v0.2.30-experimental" and .decision_runtime.replaced==false' "$task_output/build.json" >/dev/null
/usr/bin/time -l "$task_binary" body-context --activity Choose --model "$task_output/model.json" "$task_output/source.gooo" > "$task_output/preflight.json" 2> "$task_output/preflight.time"
/usr/bin/time -l "$task_binary" body-codegen --json --activity Choose "$task_output/source.gooo" > "$task_output/deterministic.json" 2> "$task_output/deterministic.time"
/usr/bin/time -l "$task_binary" body-codegen --json --activity Choose --path-model "$task_output/model.json" --path-step-attempts 1 --path-feedback-rounds 3 "$task_output/source.gooo" > "$task_output/condition.json" 2> "$task_output/condition.time"
shasum -a 256 "$task_output/model.json" | cut -d ' ' -f1 > "$task_output/model.sha256"
rm "$task_output/model.json"
/usr/bin/time -l "$task_binary" body-execute --source "$task_output/source.gooo" --generation "$task_output/condition.json" --cases "$task_output/runtime-cases.json" --go-bin "$task_go" > "$task_output/native.json" 2> "$task_output/native.time"
printf '%s\n' 'Completed once; no training. Copied model removed before saved native execution.' > "$task_output/completed.txt"
