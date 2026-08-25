'use strict';

const assert = require('assert');
const fs = require('fs');
const path = require('path');
const {listWorkflowArtifacts, selectArtifactLineage, selectCurrentEvidenceArtifact, selectUniqueArtifact} = require('./artifacts');

const digest = `sha256:${'a'.repeat(64)}`;
const evidence = (id = 1, name = 'ci-evidence-7-2') => ({id, name, size_in_bytes: 10, expired: false, digest});

async function testArtifactSelection() {
  const selected = selectCurrentEvidenceArtifact([evidence(), {id: 2, name: 'ci-proof-7-2', size_in_bytes: 10, expired: false, digest}], 7, 2);
  assert.deepStrictEqual(selected, {id: 1, name: 'ci-evidence-7-2', size_bytes: 10, expired: false, digest, run_id: 7, run_attempt: 2});
  assert.throws(() => selectCurrentEvidenceArtifact([], 7, 2), /FAIL_CLOSED: ARTIFACT_LINEAGE_NOT_FOUND/);
  assert.throws(() => selectCurrentEvidenceArtifact([evidence(), evidence(2)], 7, 2), /FAIL_CLOSED: ARTIFACT_LINEAGE_AMBIGUOUS/);
  assert.strictEqual(selectUniqueArtifact([evidence(), {id: 3, name: 'other', size_in_bytes: 1}], 'other').id, 3);
  assert.throws(() => selectCurrentEvidenceArtifact([{...evidence(), digest: ''}], 7, 2), /FAIL_CLOSED: ARTIFACT_LINEAGE_INVALID/);
}

function testLineageUseCases() {
  const fixture = JSON.parse(fs.readFileSync(path.join(__dirname, '../../examples/ci-artifact-lineage/usecases.json'), 'utf8'));
  for (const usecase of fixture.cases) {
    const artifacts = usecase.attempts.map((attempt, index) => evidence(index + 1, `ci-evidence-7-${attempt}`));
    if (usecase.invalid_digest) artifacts[artifacts.length - 1].digest = '';
    if (usecase.malformed_name) artifacts.push(evidence(90, 'ci-evidence-7-unknown'));
    const first = selectArtifactLineage(artifacts, 'ci-evidence', 7, usecase.current_attempt);
    const replay = selectArtifactLineage(artifacts.slice().reverse(), 'ci-evidence', 7, usecase.current_attempt);
    assert.strictEqual(first.report_digest, replay.report_digest, `${usecase.id} replay digest`);
    assert.strictEqual(first.decision, usecase.expected_decision, `${usecase.id} decision`);
    assert.strictEqual(first.reason, usecase.expected_reason, `${usecase.id} reason`);
    assert.strictEqual(first.resolution, usecase.expected_resolution, `${usecase.id} resolution`);
    assert.strictEqual(first.exact_consumption_authorized, usecase.expected_consumption_authorized, `${usecase.id} consumption`);
    assert.strictEqual(first.summary.selected_attempt, usecase.expected_selected_attempt, `${usecase.id} selected attempt`);
    assert.strictEqual(first.summary.fallback_distance_attempts, usecase.expected_fallback_distance, `${usecase.id} fallback distance`);
    assert.strictEqual(first.summary.repository_writes, 0, `${usecase.id} writes`);
    assert.strictEqual(first.indicators.length, 5, `${usecase.id} indicator count`);
  }
}

async function testPagination() {
  const firstPage = Array.from({length: 100}, (_, index) => ({id: index + 1, name: `unrelated-${index}`}));
  const pages = [firstPage, [evidence(101)]];
  const seenPages = [];
  const rows = await listWorkflowArtifacts(async page => {
    seenPages.push(page);
    return {data: {artifacts: pages[page - 1]}};
  });
  assert.deepStrictEqual(seenPages, [1, 2]);
  assert.strictEqual(selectCurrentEvidenceArtifact(rows, 7, 2).id, 101);
  await assert.rejects(() => listWorkflowArtifacts(async () => ({data: {}})), /missing an artifacts array/);
  await assert.rejects(() => listWorkflowArtifacts(async () => { throw new Error('network'); }), /API read failed/);
}

Promise.all([testArtifactSelection(), testPagination(), testLineageUseCases()]).then(() => {
  console.log('artifact paginator tests passed');
}).catch(error => {
  console.error(error);
  process.exitCode = 1;
});
