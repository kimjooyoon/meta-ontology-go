'use strict';

const assert = require('assert');
const fs = require('fs');
const path = require('path');
const {EXPECTED_GROUP_EXPRESSION, LEGACY_GROUP_EXPRESSION, buildConcurrencyReceipt} = require('./concurrency');

function workflowSource(kind) {
  const groups = {
    expected: [EXPECTED_GROUP_EXPRESSION],
    legacy: [LEGACY_GROUP_EXPRESSION],
    missing: [],
    unknown: ['ci-${{ github.workflow }}-${{ github.event.pull_request.number || github.ref }}-unknown'],
    duplicate: [EXPECTED_GROUP_EXPRESSION, EXPECTED_GROUP_EXPRESSION],
  }[kind];
  return ['name: CI', 'concurrency:', ...groups.map(group => `  group: ${group}`), '  cancel-in-progress: true', 'jobs:', '  test:', '    runs-on: ubuntu-latest'].join('\n');
}

const fixture = JSON.parse(fs.readFileSync(path.join(__dirname, '../../examples/ci-concurrency-lane/usecases.json'), 'utf8'));
for (const usecase of fixture.cases) {
  const input = {workflow: 'CI', scope: 'refs/heads/dev', run_id: 17, run_attempt: usecase.run_attempt};
  const first = buildConcurrencyReceipt(workflowSource(usecase.contract), input);
  const replay = buildConcurrencyReceipt(workflowSource(usecase.contract), input);
  assert.strictEqual(first.report_digest, replay.report_digest, `${usecase.id} replay digest`);
  assert.strictEqual(first.decision, usecase.expected_decision, `${usecase.id} decision`);
  assert.strictEqual(first.reason, usecase.expected_reason, `${usecase.id} reason`);
  assert.strictEqual(first.resolution, usecase.expected_resolution, `${usecase.id} resolution`);
  assert.strictEqual(first.next_operation, usecase.expected_next_operation, `${usecase.id} next operation`);
  assert.strictEqual(first.execution_authorized, usecase.expected_execution_authorized, `${usecase.id} authorization`);
  assert.strictEqual(first.lane_class, usecase.expected_lane_class, `${usecase.id} lane class`);
  assert.strictEqual(first.summary.head_preemption_risk, usecase.expected_head_preemption_risk, `${usecase.id} preemption risk`);
  assert.strictEqual(first.summary.repository_writes, 0, `${usecase.id} writes`);
  assert.strictEqual(first.indicators.length, 5, `${usecase.id} indicator count`);
}

const repositorySource = fs.readFileSync(path.join(__dirname, '../../.github/workflows/ci.yml'), 'utf8');
const head = buildConcurrencyReceipt(repositorySource, {workflow: 'CI', scope: 'refs/heads/dev', run_id: 17, run_attempt: 1});
const rerun = buildConcurrencyReceipt(repositorySource, {workflow: 'CI', scope: 'refs/heads/dev', run_id: 17, run_attempt: 2});
assert.strictEqual(head.decision, 'EXACT');
assert.strictEqual(rerun.decision, 'EXACT');
assert.notStrictEqual(head.resolved_group, rerun.resolved_group);
assert.match(rerun.resolved_group, /replay-17$/);
assert.throws(() => buildConcurrencyReceipt(repositorySource, {workflow: 'CI', scope: '', run_id: 17, run_attempt: 0}), /input is incomplete or invalid/);

console.log('concurrency lane tests passed');
