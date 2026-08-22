#!/usr/bin/env bash
set -euo pipefail

source_dir="${1:?source metrics directory required}"
output_dir="$source_dir/operation-artifact"
head="${METRICS_COMMIT_SHA:?metrics commit SHA required}"
mkdir -p "$output_dir"
actionability="$source_dir/meta-actionability-report.json"
binding="$source_dir/meta-binding-report.json"
partition="$source_dir/directory-partition-report.json"
kind="$source_dir/directory-kind-report.json"
generation="$source_dir/self-improvement-plan.json"
for report in "$actionability" "$binding" "$partition" "$kind" "$generation"; do
  test -s "$report"
done

digest() {
  printf 'sha256:%s' "$(sha256sum "$1" | cut -d' ' -f1)"
}
operations="$(jq -c '[.operations[] |
  (.operation_id // .id // .operation // .meta_operation)]' "$actionability")"
common='["meta-binding.report","directory-partition.report",
  "directory-kind.report","self-improvement-generation.plan",
  "split-go-declarations.report","split-gooo-sections.report"]'

jq -n --arg schema 'meta-operation-artifact-observation/v1' \
  --arg repository "$GITHUB_REPOSITORY" --arg head "$head" \
  --arg binding "$(digest "$binding")" \
  --arg partition "$(digest "$partition")" \
  --arg kind "$(digest "$kind")" \
  --arg generation "$(digest "$generation")" \
  --argjson operations "$operations" --argjson evidence "$common" '
  def observed($id; $name; $digest): {
    operation_id: $id, artifact_name: $name, commit_sha: $head,
    digest: $digest, replay_digest: $digest,
    evidence_keys: $evidence, repository_writes: 0
  };
  {schema: $schema, repository: $repository, commit_sha: $head,
   artifacts: ($operations | map(. as $id |
    if test("binding"; "i") then
      observed($id; "meta-binding-coverage-" + $head; $binding)
    elif test("partition"; "i") then
      observed($id; "directory-partition-" + $head; $partition)
    elif test("kind"; "i") then
      observed($id; "directory-kind-separation-" + $head; $kind)
    else observed($id; "self-improvement-generation-" + $head; $generation)
    end))}
' >"$output_dir/observations.json"
