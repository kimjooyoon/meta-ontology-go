#!/bin/bash
# Fixed macOS observation runner. Existing output is never overwritten.
set -euo pipefail
test "$#" = 4 || { echo 'usage: run.sh COMPILER MODEL GO_BINARY NEW_OUTPUT_DIRECTORY' >&2; exit 2; }
task_binary=$1
task_model=$2
task_go=$3
task_dest=$4
for task_path in "$task_binary" "$task_model" "$task_go" "$task_dest"; do
  case "$task_path" in /*) ;; *) echo 'absolute paths required' >&2; exit 2;; esac
done
test "$(uname -s)" = Darwin
task_source=$(git rev-parse --show-toplevel)
task_sha=$(git rev-parse HEAD)
task_model_sha=d5f9c4ffb204852f4e1b3b1b1c491c7d1fb0753d0df2c1e0f0ffead80d9ac47e
test -z "$(git status --porcelain)"
test ! -e "$task_dest"
test "$(shasum -a 256 "$task_model" | cut -d ' ' -f1)" = "$task_model_sha"
mkdir "$task_dest"
cp "$task_source/docs/research/leaky-flow-cli-20261010/protocol.txt" "$task_dest/protocol.txt"
cp "$task_source/docs/research/leaky-flow-cli-20261010/source-selection.json" "$task_dest/source-selection.json"
"$task_binary" version --build --json > "$task_dest/producer-version.json"
jq -e --arg sha "$task_sha" '.compiler_source_sha==$sha and .source_status=="CLEAN_VCS" and .go_version=="go1.27.2" and .decision_runtime.version=="v0.2.35-experimental" and .decision_runtime.replaced==false' "$task_dest/producer-version.json" >/dev/null
shasum -a 256 "$task_binary" > "$task_dest/producer-binary.sha256"
"$task_go" version -m "$task_binary" > "$task_dest/producer-build.txt"
printf '%s\n' "$task_model_sha" > "$task_dest/model.sha256"
date -u +%FT%TZ > "$task_dest/started.txt"
run_case() {
  local task_name=$1
  shift
  test ! -e "$task_name.json"
  /usr/bin/time -l env GOWORK=off GOTOOLCHAIN=go1.27.2 "$task_binary" "$@" > "$task_name.json" 2> "$task_name.time-stderr"
  printf '%s\t%s\n' "$task_form" "$task_name" >> "$task_dest/completed-commands.txt"
}
for task_form in max-direct max-copy min-direct min-copy nested; do
  task_family=${task_form%%-*}
  mkdir "$task_dest/$task_form"
  cp "$task_source/examples/body-codegen/source-leaky-$task_form.gooo.fixture" "$task_dest/$task_form/source.gooo"
  cp "$task_source/examples/body-codegen/source-leaky-$task_family-construction-cases.json" "$task_dest/$task_form/construction-cases.json"
  cp "$task_source/examples/body-codegen/source-leaky-$task_family-evaluation-cases.json" "$task_dest/$task_form/evaluation-cases.json"
  cp "$task_model" "$task_dest/$task_form/model.json"
  cd "$task_dest/$task_form"
  shasum -a 256 source.gooo model.json construction-cases.json evaluation-cases.json > inputs.sha256
  run_case preflight body-context --activity Choose --model model.json source.gooo
  run_case deterministic body-codegen --json --activity Choose source.gooo
  run_case initial body-codegen --json --activity Choose --path-model model.json --path-step-attempts 1 source.gooo
  run_case feedback body-codegen --json --activity Choose --path-model model.json --path-step-attempts 1 --path-feedback-rounds 3 source.gooo
  run_case construction body-construct --source source.gooo --entry Main --model model.json --construction-cases construction-cases.json --cases evaluation-cases.json --attempts 4 --go-bin "$task_go" --out construction
  test "$(shasum -a 256 model.json | cut -d ' ' -f1)" = "$task_model_sha"
  rm model.json
  run_case replay body-construct --source construction/original.gooo --construction construction/construction.json --cases evaluation-cases.json --go-bin "$task_go"
done
date -u +%FT%TZ > "$task_dest/completed.txt"
cat "$task_dest/completed-commands.txt"
