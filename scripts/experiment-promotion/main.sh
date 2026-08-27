#!/usr/bin/env bash
set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

head_sha=${HEAD_SHA:-$(git rev-parse HEAD)}
if [[ ! "$head_sha" =~ ^[0-9a-f]{40}$ ]]; then
  echo "HEAD_SHA must be an exact 40-character commit SHA" >&2
  exit 1
fi

snapshot_repository() {
  git ls-files -s
  git status --porcelain=v1
  git diff --no-ext-diff
}

snapshot_paths() {
  git status --porcelain=v1 | awk 'NF {print $NF}' | sort -u
}

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT
before_snapshot="$tmp_dir/before.snapshot"
after_snapshot="$tmp_dir/after.snapshot"
changed_paths="$tmp_dir/changed.paths"
snapshot_repository > "$before_snapshot"
go run ./cmd/gooo check examples/experiment-promotion/main.gooo
snapshot_repository > "$after_snapshot"
snapshot_paths > "$changed_paths"

if [[ "$(go list -deps ./cmd/experiment-promotion-verify | grep -F -x -c 'github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotion' || true)" != "0" ]]; then
  echo "independent consumer imports producer package" >&2
  exit 1
fi
echo "producer_imports=0/0"
echo "consumer_procedure=experimentpromotionverify.reconstruct/v2"

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
    -snapshot-before "$before_snapshot" \
    -snapshot-after "$after_snapshot" \
    -snapshot-paths "$changed_paths"
  go run ./cmd/experiment-promotion-verify \
    -source examples/experiment-promotion/main.gooo \
    -observations "$fixture" \
    -contract examples/experiment-promotion/contract.json \
    -report "$report" \
    -out "$verification" \
    -subject-sha "$head_sha" \
    -snapshot-before "$before_snapshot" \
    -snapshot-after "$after_snapshot" \
    -snapshot-paths "$changed_paths"
  jq -e '.source_projection.experiments | length == 30' "$report" >/dev/null
  jq -e '.source_projection.gates | length == 5' "$report" >/dev/null
  jq -e '.summary.declared_experiments_numerator == 30 and .summary.declared_experiments_denominator == 30 and .summary.materialized_claim_slots_numerator == 150 and .summary.materialized_claim_slots_denominator == 150' "$report" >/dev/null
  jq -e '.aggregate_metrics | length == 0 and .mutation_authority == false and .repository_writes == 0' "$report" >/dev/null
  jq -e '.checks | all(.status == "PASS")' "$verification" >/dev/null
  echo "$name $(jq -r '"declared_experiments=\(.summary.declared_experiments_numerator)/\(.summary.declared_experiments_denominator) materialized_claim_slots=\(.summary.materialized_claim_slots_numerator)/\(.summary.materialized_claim_slots_denominator) experiment_states=\(.summary.experiment_states | to_entries | map("\(.key):\(.value)") | join(",")) gate_states=\(.summary.gate_states | to_entries | map("\(.key):\(.value)") | join(",")) fixture_experiment_states=\(.summary.fixture_experiment_states | to_entries | map("\(.key):\(.value)") | join(",")) fixture_gate_states=\(.summary.fixture_gate_states | to_entries | map("\(.key):\(.value)") | join(","))"' "$report")"
}

run_fixture fully-proven
run_fixture missing-common-gate
run_fixture malformed-evidence
run_fixture contradictory-semantic

snapshot_repository > "$tmp_dir/after.final.snapshot"
if ! cmp -s "$after_snapshot" "$tmp_dir/after.final.snapshot"; then
  echo "repository snapshot changed during read-only replay" >&2
  exit 1
fi

full="$tmp_dir/fully-proven.report.json"
missing="$tmp_dir/missing-common-gate.report.json"
malformed="$tmp_dir/malformed-evidence.report.json"
contradictory="$tmp_dir/contradictory-semantic.report.json"
jq -e '.summary.experiment_states.PROVEN == 0 and .summary.experiment_states.UNKNOWN == 30 and .summary.fixture_experiment_states.PROVEN == 1 and .summary.fixture_gate_states.PROVEN == 5 and .summary.counterexamples_detected_denominator == 9' "$full" >/dev/null
jq -e '.summary.experiment_states.PROVEN == 0 and .summary.experiment_states.UNKNOWN == 30 and .summary.fixture_experiment_states.OPEN == 1 and .summary.fixture_gate_states.OPEN == 1' "$missing" >/dev/null
jq -e '.summary.experiment_states.UNKNOWN == 30 and .summary.fixture_experiment_states.REFUTED == 1 and .summary.fixture_gate_states.REFUTED == 1' "$malformed" >/dev/null
jq -e '.summary.experiment_states.UNKNOWN == 30 and .summary.fixture_experiment_states.REFUTED == 1 and .summary.fixture_gate_states.REFUTED == 1' "$contradictory" >/dev/null

echo "declared_experiments=30/30"
echo "materialized_claim_slots=150/150"
echo "persistent_claims=150/150"
echo "guardrail_forbidden_aggregate_claim observed=$(jq -r '.guardrails[] | select(.id == "forbidden-aggregate-claims") | .observed' "$full") allowed_max=$(jq -r '.guardrails[] | select(.id == "forbidden-aggregate-claims") | .allowed_max' "$full") conformance=$(jq -r '.guardrails[] | select(.id == "forbidden-aggregate-claims") | "\(.conformance_numerator)/\(.conformance_denominator)"' "$full")"
echo "guardrail_repository_writes observed=$(jq -r '.guardrails[] | select(.id == "repository-writes") | .observed' "$full") allowed_max=$(jq -r '.guardrails[] | select(.id == "repository-writes") | .allowed_max' "$full") conformance=$(jq -r '.guardrails[] | select(.id == "repository-writes") | "\(.conformance_numerator)/\(.conformance_denominator)"' "$full")"
echo "fixture_corpus=4/4 actual_promotion_authority=0/4"
echo "head_sha=$head_sha"
