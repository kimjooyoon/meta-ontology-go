'use strict';

const crypto = require('crypto');

const DEFAULT_PAGE_SIZE = 100;
const DEFAULT_PAGE_LIMIT = 1000;
const EXACT = 'EXACT';
const LOWER_RESOLUTION = 'LOWER_RESOLUTION';
const FAIL_CLOSED = 'FAIL_CLOSED';

async function listWorkflowArtifacts(fetchPage, options = {}) {
  const pageSize = options.pageSize || DEFAULT_PAGE_SIZE;
  const pageLimit = options.pageLimit || DEFAULT_PAGE_LIMIT;
  const artifacts = [];
  for (let page = 1; page <= pageLimit; page += 1) {
    let response;
    try {
      response = await fetchPage(page);
    } catch (error) {
      throw new Error(`workflow artifact API read failed: ${error.message}`);
    }
    const rows = response && response.data && response.data.artifacts;
    if (!Array.isArray(rows)) {
      throw new Error('workflow artifact API response is missing an artifacts array');
    }
    artifacts.push(...rows);
    if (rows.length < pageSize) {
      return artifacts;
    }
  }
  throw new Error('workflow artifact API pagination exceeded the fail-closed page limit');
}

function selectUniqueArtifact(artifacts, name) {
  const matches = artifacts.filter(artifact => artifact && artifact.name === name);
  if (matches.length > 1) {
    throw new Error(`workflow artifact inventory contains duplicate ${name} artifacts`);
  }
  return matches[0] || null;
}

function indicator(metricId, kind, target, unit, relation, proofChoice, metaOperation, activity, value) {
  const satisfied = relation === 'greater_or_equal' ? value >= target : value <= target;
  return {
    metric_id: metricId,
    class: kind,
    target,
    unit,
    relation,
    proof_choice: proofChoice,
    producer: 'ciArtifactLineage.select',
    consumer: 'CI proof bundle',
    meta_operation: metaOperation,
    activity,
    value,
    satisfied,
  };
}

function finalizeLineageReport(report) {
  const exact = report.decision === EXACT ? 10000 : 0;
  const continuity = report.selected ? 10000 : 0;
  report.indicators = [
    indicator('gooo.metric.ci.artifact-current-attempt.coverage-bps.v1', 'outcome', 10000, 'basis_points', 'greater_or_equal', 'coherence', 'bind-current-attempt-artifact', 'BindCurrentAttemptArtifact', exact),
    indicator('gooo.metric.ci.artifact-lineage-continuity.coverage-bps.v1', 'outcome', 10000, 'basis_points', 'greater_or_equal', 'coherence', 'observe-artifact-lineage', 'ObserveArtifactLineage', continuity),
    indicator('gooo.metric.ci.artifact-lineage-fallback-distance.attempts.v1', 'guardrail', 0, 'attempts', 'less_or_equal', 'regression', 'bound-artifact-fallback', 'BoundArtifactFallback', report.summary.fallback_distance_attempts),
    indicator('gooo.metric.ci.artifact-lineage-ambiguity.guardrail.v1', 'guardrail', 0, 'candidates', 'less_or_equal', 'coherence', 'reject-ambiguous-artifact-lineage', 'RejectAmbiguousArtifactLineage', report.summary.ambiguous_candidates),
    indicator('gooo.metric.ci.artifact-lineage-observer-writes.guardrail.v1', 'guardrail', 0, 'repository_writes', 'less_or_equal', 'foundation', 'preserve-read-only-artifact-lineage', 'PreserveReadOnlyArtifactLineage', report.summary.repository_writes),
  ];
  const canonical = JSON.stringify(report);
  report.report_digest = `sha256:${crypto.createHash('sha256').update(canonical).digest('hex')}`;
  return report;
}

function selectArtifactLineage(artifacts, stem, runId, currentAttempt) {
  if (!Array.isArray(artifacts) || !/^[a-z0-9-]+$/.test(stem || '') || !Number.isSafeInteger(runId) || runId <= 0 || !Number.isSafeInteger(currentAttempt) || currentAttempt <= 0) {
    throw new Error('artifact lineage input is incomplete or invalid');
  }

  const prefix = `${stem}-${runId}-`;
  const observed = artifacts.filter(artifact => artifact && typeof artifact.name === 'string' && artifact.name.startsWith(prefix));
  const parsed = [];
  let malformed = 0;
  let invalid = 0;
  let future = 0;
  for (const artifact of observed) {
    const suffix = artifact.name.slice(prefix.length);
    if (!/^[1-9][0-9]*$/.test(suffix)) {
      malformed += 1;
      continue;
    }
    const attempt = Number(suffix);
    if (!Number.isSafeInteger(attempt)) {
      malformed += 1;
      continue;
    }
    if (artifact.id <= 0 || artifact.size_in_bytes <= 0 || artifact.expired || !/^sha256:[0-9a-f]{64}$/.test(artifact.digest || '')) {
      invalid += 1;
      continue;
    }
    if (attempt > currentAttempt) {
      future += 1;
      continue;
    }
    parsed.push({artifact, attempt});
  }

  const byAttempt = new Map();
  for (const candidate of parsed) {
    const rows = byAttempt.get(candidate.attempt) || [];
    rows.push(candidate);
    byAttempt.set(candidate.attempt, rows);
  }
  const ambiguous = [...byAttempt.values()].filter(rows => rows.length > 1).reduce((total, rows) => total + rows.length, 0);
  const compatible = parsed.length;
  const current = (byAttempt.get(currentAttempt) || []).length;
  const selectedRow = ambiguous === 0 && parsed.length > 0
    ? parsed.slice().sort((left, right) => right.attempt - left.attempt || left.artifact.id - right.artifact.id)[0]
    : null;
  const selected = selectedRow ? {
    id: selectedRow.artifact.id,
    name: selectedRow.artifact.name,
    size_bytes: selectedRow.artifact.size_in_bytes,
    expired: selectedRow.artifact.expired,
    digest: selectedRow.artifact.digest,
    run_id: runId,
    run_attempt: selectedRow.attempt,
  } : null;
  const fallbackDistance = selected ? currentAttempt - selected.run_attempt : 0;

  let decision = EXACT;
  let reason = 'CURRENT_ATTEMPT_ARTIFACT_BOUND';
  let resolution = 'run_attempt';
  let nextOperation = 'consume-current-attempt-artifact';
  if (malformed > 0) {
    decision = FAIL_CLOSED;
    reason = 'ARTIFACT_LINEAGE_NAME_MALFORMED';
    resolution = 'none';
    nextOperation = 'repair-artifact-lineage';
  } else if (future > 0) {
    decision = FAIL_CLOSED;
    reason = 'ARTIFACT_LINEAGE_FUTURE_ATTEMPT';
    resolution = 'none';
    nextOperation = 'repair-artifact-lineage';
  } else if (ambiguous > 0) {
    decision = FAIL_CLOSED;
    reason = 'ARTIFACT_LINEAGE_AMBIGUOUS';
    resolution = 'none';
    nextOperation = 'repair-artifact-lineage';
  } else if (invalid > 0) {
    decision = FAIL_CLOSED;
    reason = 'ARTIFACT_LINEAGE_INVALID';
    resolution = 'none';
    nextOperation = 'repair-artifact-lineage';
  } else if (!selected) {
    decision = FAIL_CLOSED;
    reason = 'ARTIFACT_LINEAGE_NOT_FOUND';
    resolution = 'none';
    nextOperation = 'rerun-all-jobs';
  } else if (selected.run_attempt !== currentAttempt) {
    decision = LOWER_RESOLUTION;
    reason = 'CURRENT_ATTEMPT_ARTIFACT_UNAVAILABLE';
    resolution = 'run_lineage';
    nextOperation = 'rerun-all-jobs';
  }

  return finalizeLineageReport({
    schema: 'gooo/ci-artifact-lineage/v1',
    stem,
    run_id: runId,
    current_attempt: currentAttempt,
    decision,
    reason,
    resolution,
    next_operation: nextOperation,
    exact_consumption_authorized: decision === EXACT,
    selected: decision === FAIL_CLOSED ? null : selected,
    summary: {
      observed_candidates: observed.length,
      compatible_candidates: compatible,
      current_attempt_candidates: current,
      malformed_candidates: malformed,
      invalid_candidates: invalid,
      future_candidates: future,
      ambiguous_candidates: ambiguous,
      selected_attempt: decision === FAIL_CLOSED || !selected ? 0 : selected.run_attempt,
      fallback_distance_attempts: decision === FAIL_CLOSED ? 0 : fallbackDistance,
      repository_writes: 0,
    },
  });
}

function selectCurrentEvidenceArtifact(artifacts, runId, runAttempt) {
  const report = selectArtifactLineage(artifacts, 'ci-evidence', runId, runAttempt);
  if (report.decision !== EXACT) {
    throw new Error(`workflow artifact lineage ${report.decision}: ${report.reason}; next=${report.next_operation}`);
  }
  return report.selected;
}

module.exports = {listWorkflowArtifacts, selectArtifactLineage, selectCurrentEvidenceArtifact, selectUniqueArtifact};
