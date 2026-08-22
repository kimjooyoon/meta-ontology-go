#!/usr/bin/env bash
set -euo pipefail

source_dir="${1:?source metrics directory required}"
output_dir="$source_dir/operation-artifact"
head="${METRICS_COMMIT_SHA:?metrics commit SHA required}"
mkdir -p "$output_dir"
actionability="$source_dir/meta-actionability-report.json"
binding="$source_dir/meta-binding-report.json"
binding_replay="$source_dir/meta-binding-report-replay.json"
partition="$source_dir/directory-partition-report.json"
partition_replay="$source_dir/directory-partition-report-replay.json"
kind="$source_dir/directory-kind-report.json"
kind_replay="$source_dir/directory-kind-report-replay.json"
generation="$source_dir/self-improvement-plan.json"
generation_replay="$source_dir/self-improvement-replay.json"
for report in "$actionability" "$binding" "$binding_replay" "$partition" \
  "$partition_replay" "$kind" "$kind_replay" "$generation" "$generation_replay"; do
  test -s "$report"
done

digest() {
  printf 'sha256:%s' "$(sha256sum "$1" | cut -d' ' -f1)"
}
jq -n --arg schema 'gooo/meta-operation-artifact-observations/v1' \
  --arg repository "$GITHUB_REPOSITORY" --arg head "$head" \
  --arg binding "$(digest "$binding")" \
  --arg binding_replay "$(digest "$binding_replay")" \
  --arg partition "$(digest "$partition")" \
  --arg partition_replay "$(digest "$partition_replay")" \
  --arg kind "$(digest "$kind")" \
  --arg kind_replay "$(digest "$kind_replay")" \
  --arg generation "$(digest "$generation")" \
  --arg generation_replay "$(digest "$generation_replay")" \
  --argjson run_id "$GITHUB_RUN_ID" --argjson run_attempt "$GITHUB_RUN_ATTEMPT" '
  def observed($name; $digest; $replay; $evidence): {
    name: $name, head_sha: $head, digest: $digest,
    replay_digest: $replay, evidence_keys: $evidence
  };
  {schema: $schema, repository: $repository, commit_sha: $head,
   run_id: $run_id, run_attempt: $run_attempt, repository_writes: 0,
   artifacts: [
    observed("meta-binding-coverage-" + $head; $binding; $binding_replay;
      ["meta-binding.report"]),
    observed("directory-partition-" + $head; $partition; $partition_replay;
      ["directory-partition.report"]),
    observed("directory-kind-separation-" + $head; $kind; $kind_replay;
      ["directory-kind.report"]),
    observed("self-improvement-generation-" + $head; $generation; $generation_replay;
      ["operation.split-go-declarations", "operation.split-gooo-sections"])
   ]}
' >"$output_dir/observations.json"
