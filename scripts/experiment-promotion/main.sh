#!/usr/bin/env bash
set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

head_sha=${HEAD_SHA:-$(git rev-parse HEAD)}
if [[ ! "$head_sha" =~ ^[0-9a-f]{40}$ ]]; then
  echo "HEAD_SHA must be an exact 40-character commit SHA" >&2
  exit 1
fi

snapshot_digest() {
  {
    git ls-files -s
    git status --porcelain=v1
  } | sha256sum | awk '{print "sha256:" $1}'
}

snapshot_paths() {
  git status --porcelain=v1 | awk 'NF {print $NF}' | sort -u | wc -l | tr -d ' '
}

before_digest=$(snapshot_digest)
before_paths=$(snapshot_paths)
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

if [[ "$(go list -deps ./cmd/experiment-promotion-verify | grep -F -x -c 'github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotion' || true)" != "0" ]]; then
  echo "independent consumer imports producer package" >&2
  exit 1
fi
echo "producer_imports=0/0"

go test ./cmd/experiment-promotion-witness ./cmd/experiment-promotion-verify ./internal/meta/experimentpromotion ./internal/meta/experimentpromotionverify
go run ./cmd/gooo check examples/experiment-promotion/main.gooo

run_fixture() {
  local name=$1
  local fixture="examples/experiment-promotion/fixtures/${name}.json"
  local report="$tmp_dir/${name}.report.json"
  local verification="$tmp_dir/${name}.verification.json"
  go run ./cmd/experiment-promotion-witness \
    -source examples/experiment-promotion/main.gooo \
    -observations "$fixture" \
    -contract examples/experiment-promotion/contract.json \
    -out "$report" \
    -subject-sha "$head_sha" \
    -before-digest "$before_digest" \
    -after-digest "$before_digest" \
    -changed-paths 0
  go run ./cmd/experiment-promotion-verify \
    -source examples/experiment-promotion/main.gooo \
    -observations "$fixture" \
    -contract examples/experiment-promotion/contract.json \
    -report "$report" \
    -out "$verification" \
    -subject-sha "$head_sha" \
    -before-digest "$before_digest" \
    -after-digest "$before_digest" \
    -changed-paths 0
  jq -e '.source_projection.experiments | length == 30' "$report" >/dev/null
  jq -e '.source_projection.gates | length == 5' "$report" >/dev/null
  jq -e '.summary.experiments_numerator == 30 and .summary.experiments_denominator == 30 and .summary.gate_slots_numerator == 150 and .summary.gate_slots_denominator == 150' "$report" >/dev/null
  jq -e '.aggregate_metrics | length == 0 and .mutation_authority == false and .repository_writes == 0' "$report" >/dev/null
  jq -e '.checks | all(.status == "PASS")' "$verification" >/dev/null
  echo "$name $(jq -r '"experiments=\(.summary.experiments_numerator)/\(.summary.experiments_denominator) gate_slots=\(.summary.gate_slots_numerator)/\(.summary.gate_slots_denominator) experiment_states=\(.summary.experiment_states | to_entries | map("\(.key):\(.value)") | join(",")) gate_states=\(.summary.gate_states | to_entries | map("\(.key):\(.value)") | join(","))"' "$report")"
}

run_fixture fully-proven
run_fixture missing-common-gate
run_fixture malformed-evidence
run_fixture contradictory-semantic

after_digest=$(snapshot_digest)
after_paths=$(snapshot_paths)
if [[ "$before_digest" != "$after_digest" || "$before_paths" != "$after_paths" ]]; then
  echo "repository snapshot changed during read-only replay" >&2
  exit 1
fi

full="$tmp_dir/fully-proven.report.json"
missing="$tmp_dir/missing-common-gate.report.json"
malformed="$tmp_dir/malformed-evidence.report.json"
contradictory="$tmp_dir/contradictory-semantic.report.json"
jq -e '.summary.experiment_states.PROVEN == 1 and .summary.experiment_states.OPEN == 0 and .summary.experiment_states.UNKNOWN == 29 and .summary.experiment_states.REFUTED == 0 and .summary.gate_states.PROVEN == 5 and .summary.gate_states.UNKNOWN == 145' "$full" >/dev/null
jq -e '.summary.experiment_states.PROVEN == 0 and .summary.experiment_states.OPEN == 1 and .summary.experiment_states.UNKNOWN == 29 and .summary.gate_states.PROVEN == 4 and .summary.gate_states.OPEN == 1 and .summary.gate_states.UNKNOWN == 145' "$missing" >/dev/null
jq -e '.summary.experiment_states.REFUTED == 1 and .summary.experiment_states.UNKNOWN == 29 and .summary.gate_states.REFUTED == 1 and .summary.gate_states.UNKNOWN == 149' "$malformed" >/dev/null
jq -e '.summary.experiment_states.REFUTED == 1 and .summary.experiment_states.UNKNOWN == 29 and .summary.gate_states.REFUTED == 1 and .summary.gate_states.UNKNOWN == 149' "$contradictory" >/dev/null

echo "experiments=30/30 gate_slots=150/150"
echo "persistent_claims=150/150"
echo "guardrail_forbidden_aggregate_claim observed=0 allowed_max=0 conformance=1/1"
echo "guardrail_repository_writes observed=0 allowed_max=0 conformance=1/1"
echo "fixtures=4/4 semantic_causality=1/1 nonsemantic_missing=1/1"
echo "head_sha=$head_sha"
