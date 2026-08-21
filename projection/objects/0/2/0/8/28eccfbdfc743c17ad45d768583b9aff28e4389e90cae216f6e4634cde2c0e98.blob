'use strict';

const assert = require('assert');
const {listWorkflowArtifacts, selectCurrentEvidenceArtifact, selectUniqueArtifact} = require('./artifacts');

const digest = `sha256:${'a'.repeat(64)}`;
const evidence = (id = 1, name = 'ci-evidence-7-2') => ({id, name, size_in_bytes: 10, expired: false, digest});

async function testArtifactSelection() {
  const selected = selectCurrentEvidenceArtifact([evidence(), {id: 2, name: 'ci-proof-7-2', size_in_bytes: 10, expired: false, digest}], 7, 2);
  assert.deepStrictEqual(selected, {id: 1, name: 'ci-evidence-7-2', size_bytes: 10, expired: false, digest, run_id: 7, run_attempt: 2});
  assert.throws(() => selectCurrentEvidenceArtifact([], 7, 2), /missing/);
  assert.throws(() => selectCurrentEvidenceArtifact([evidence(), evidence(2)], 7, 2), /duplicate/);
  assert.strictEqual(selectUniqueArtifact([evidence(), {id: 3, name: 'other', size_in_bytes: 1}], 'other').id, 3);
  assert.throws(() => selectCurrentEvidenceArtifact([{...evidence(), digest: ''}], 7, 2), /invalid digest/);
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

Promise.all([testArtifactSelection(), testPagination()]).then(() => {
  console.log('artifact paginator tests passed');
}).catch(error => {
  console.error(error);
  process.exitCode = 1;
});
