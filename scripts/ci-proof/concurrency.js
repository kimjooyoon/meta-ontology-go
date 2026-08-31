'use strict';

const crypto = require('crypto');

const EXACT = 'EXACT';
const LOWER_RESOLUTION = 'LOWER_RESOLUTION';
const FAIL_CLOSED = 'FAIL_CLOSED';
const EXPECTED_GROUP_EXPRESSION = "ci-${{ github.workflow }}-${{ github.event.pull_request.number || github.ref }}-${{ github.run_attempt > 1 && format('replay-{0}', github.run_id) || 'head' }}";
const LEGACY_GROUP_EXPRESSION = 'ci-${{ github.workflow }}-${{ github.event.pull_request.number || github.ref }}';

function indicator(metricId, kind, target, unit, relation, proofChoice, metaOperation, activity, value) {
  const satisfied = relation === 'greater_or_equal' ? value >= target : value <= target;
  return {
    metric_id: metricId,
    class: kind,
    target,
    unit,
    relation,
    proof_choice: proofChoice,
    producer: 'ciConcurrencyLane.select',
    consumer: 'CI workflow scheduler and proof bundle',
    meta_operation: metaOperation,
    activity,
    value,
    satisfied,
  };
}

function topLevelConcurrencyGroups(source) {
  if (typeof source !== 'string') throw new Error('workflow source is required');
  const groups = [];
  const lines = source.split(/\r?\n/);
  let inConcurrency = false;
  for (const line of lines) {
    if (/^concurrency:\s*$/.test(line)) {
      inConcurrency = true;
      continue;
    }
    if (!inConcurrency) continue;
    if (/^[^\s#]/.test(line)) break;
    const match = line.match(/^\s{2}group:\s*(.+?)\s*$/);
    if (match) groups.push(match[1]);
  }
  return groups;
}

function validateInput(input) {
  if (!input || typeof input.workflow !== 'string' || input.workflow.length === 0 || typeof input.scope !== 'string' || input.scope.length === 0 || !Number.isSafeInteger(input.run_id) || input.run_id <= 0 || !Number.isSafeInteger(input.run_attempt) || input.run_attempt <= 0) {
    throw new Error('concurrency lane input is incomplete or invalid');
  }
}

function resolvedGroup(input) {
  const lane = input.run_attempt > 1 ? `replay-${input.run_id}` : 'head';
  return `ci-${input.workflow}-${input.scope}-${lane}`;
}

function finalize(report) {
  const contractCoverage = report.decision === EXACT ? 10000 : 0;
  const isolationCoverage = report.replay_isolation_satisfied ? 10000 : 0;
  const cancellationCoverage = report.head_cancellation_preserved ? 10000 : 0;
  report.indicators = [
    indicator('gooo.metric.ci.concurrency-lane-binding.coverage-bps.v1', 'outcome', 10000, 'basis_points', 'greater_or_equal', 'coherence', 'bind-ci-concurrency-lane', 'BindCIConcurrencyLane', contractCoverage),
    indicator('gooo.metric.ci.rerun-isolation.coverage-bps.v1', 'outcome', 10000, 'basis_points', 'greater_or_equal', 'coherence', 'isolate-rerun-from-head', 'IsolateRerunFromHead', isolationCoverage),
    indicator('gooo.metric.ci.head-cancellation-preservation.coverage-bps.v1', 'guardrail', 10000, 'basis_points', 'greater_or_equal', 'regression', 'preserve-head-cancellation', 'PreserveHeadCancellation', cancellationCoverage),
    indicator('gooo.metric.ci.head-preemption-risk.guardrail.v1', 'guardrail', 0, 'at_risk_runs', 'less_or_equal', 'regression', 'prevent-rerun-head-preemption', 'PreventRerunHeadPreemption', report.summary.head_preemption_risk),
    indicator('gooo.metric.ci.concurrency-lane-observer-writes.guardrail.v1', 'guardrail', 0, 'repository_writes', 'less_or_equal', 'foundation', 'preserve-read-only-concurrency-observer', 'PreserveReadOnlyConcurrencyObserver', report.summary.repository_writes),
  ];
  report.report_digest = `sha256:${crypto.createHash('sha256').update(JSON.stringify(report)).digest('hex')}`;
  return report;
}

function buildConcurrencyReceipt(source, input) {
  validateInput(input);
  const groups = topLevelConcurrencyGroups(source);
  const observed = groups.length === 1 ? groups[0] : '';
  const replay = input.run_attempt > 1;
  let decision = EXACT;
  let reason = replay ? 'RERUN_ISOLATED_FROM_HEAD' : 'HEAD_LANE_BOUND';
  let resolution = 'run_attempt';
  let nextOperation = 'continue-ci-proof';

  if (groups.length === 0) {
    decision = FAIL_CLOSED;
    reason = 'CONCURRENCY_LANE_CONTRACT_MISSING';
    resolution = 'none';
    nextOperation = 'repair-ci-concurrency-contract';
  } else if (groups.length > 1) {
    decision = FAIL_CLOSED;
    reason = 'CONCURRENCY_LANE_CONTRACT_AMBIGUOUS';
    resolution = 'none';
    nextOperation = 'repair-ci-concurrency-contract';
  } else if (observed === LEGACY_GROUP_EXPRESSION) {
    decision = LOWER_RESOLUTION;
    reason = replay ? 'RERUN_SHARES_HEAD_LANE' : 'HEAD_LANE_LEGACY_UNVERSIONED';
    resolution = 'workflow_scope';
    nextOperation = 'bind-rerun-to-isolated-lane';
  } else if (observed !== EXPECTED_GROUP_EXPRESSION) {
    decision = FAIL_CLOSED;
    reason = 'CONCURRENCY_LANE_CONTRACT_UNKNOWN';
    resolution = 'none';
    nextOperation = 'repair-ci-concurrency-contract';
  }

  const exact = decision === EXACT;
  const legacy = observed === LEGACY_GROUP_EXPRESSION;
  return finalize({
    schema: 'gooo/ci-concurrency-lane/v1',
    workflow: input.workflow,
    scope: input.scope,
    run_id: input.run_id,
    run_attempt: input.run_attempt,
    lane_class: replay ? 'replay' : 'head',
    expected_group_expression: EXPECTED_GROUP_EXPRESSION,
    observed_group_expression: observed,
    resolved_group: exact ? resolvedGroup(input) : (legacy ? `ci-${input.workflow}-${input.scope}` : ''),
    decision,
    reason,
    resolution,
    next_operation: nextOperation,
    execution_authorized: exact,
    replay_isolation_required: replay,
    replay_isolation_satisfied: !replay || exact,
    head_cancellation_preserved: exact || legacy,
    summary: {
      observed_group_candidates: groups.length,
      ambiguous_group_candidates: groups.length > 1 ? groups.length : 0,
      head_preemption_risk: replay && !exact ? 1 : 0,
      repository_writes: 0,
    },
  });
}

module.exports = {EXPECTED_GROUP_EXPRESSION, LEGACY_GROUP_EXPRESSION, buildConcurrencyReceipt, resolvedGroup, topLevelConcurrencyGroups};
