'use strict';

const assert = require('assert');
const {canonicalJobNames, listWorkflowJobs, observeCanonicalJobs, schedulerResultsFromNeeds} = require('./observe');

const head = 'a'.repeat(40);
const expected = {expectedHead: head, runID: 31937165616, runAttempt: 1};
const schedulerResults = schedulerResultsFromNeeds({format: 'success', vet: 'success', test: 'success', race: 'success', semantic: 'success', policy: 'success'}, {headSHA: head, runID: expected.runID, runAttempt: expected.runAttempt});

function job(name, index, state = 'success') {
  if (state === 'success') {
    return {id: index, name, status: 'completed', conclusion: 'success', head_sha: head, run_id: expected.runID, run_attempt: 1, completed_at: '2026-08-16T08:44:18Z'};
  }
  return {id: index, name, status: 'in_progress', conclusion: null, head_sha: head, run_id: expected.runID, run_attempt: 1, completed_at: null};
}

function snapshot(state = 'success') {
  return canonicalJobNames.map((name, index) => job(name, index + 1, state));
}

async function testImmediateTerminalSuccess() {
  let calls = 0;
  const result = await observeCanonicalJobs({...expected, schedulerResults, readPage: async () => { calls += 1; return {data: {jobs: snapshot()}}; }, wait: async () => { throw new Error('wait should not be called'); }});
  assert.strictEqual(calls, 1);
  assert.strictEqual(result.observations, 1);
  assert.deepStrictEqual(result.canonicalJobs.map(item => item.name), canonicalJobNames);
}

async function testDelayedTerminalObservation() {
  const observations = [snapshot().slice(0, -1), snapshot()];
  let calls = 0;
  const result = await observeCanonicalJobs({...expected, schedulerResults, maxObservations: 3, readPage: async () => ({data: {jobs: observations[calls++]}}), wait: async () => {}});
  assert.strictEqual(result.observations, 2);
  assert.strictEqual(calls, 2);
}

async function testNeverTerminalTimesOut() {
  await assert.rejects(
    () => observeCanonicalJobs({...expected, schedulerResults, maxObservations: 2, readPage: async () => ({data: {jobs: snapshot().slice(0, -1)}}), wait: async () => {}}),
    /CI_EVIDENCE_UNKNOWN: canonical job identities were not all observed after 2 observations; last observation: canonical CI job "CI policy" is not yet visible/,
  );
}

async function testRawLagWithoutSchedulerRemainsUnknown() {
  await assert.rejects(
    () => observeCanonicalJobs({...expected, maxObservations: 2, readPage: async () => ({data: {jobs: snapshot('pending')}}), wait: async () => {}}),
    /CI_EVIDENCE_UNKNOWN: same-run scheduler results are missing/,
  );
}

async function testMalformedAndDuplicateJobsFailClosed() {
  await assert.rejects(
    () => observeCanonicalJobs({...expected, schedulerResults, readPage: async () => ({data: {}}), wait: async () => {}}),
    /CI_EVIDENCE_UNKNOWN: canonical job API returned malformed data/,
  );
  await assert.rejects(
    () => observeCanonicalJobs({...expected, schedulerResults, readPage: async () => ({data: {jobs: [...snapshot(), job('gofmt', 99)]}}), wait: async () => {}}),
    /CI_EVIDENCE_UNKNOWN: duplicate canonical CI job "gofmt"/,
  );
}

async function testWrongTupleFailsClosed() {
  const wrongHead = snapshot();
  wrongHead[0].head_sha = 'b'.repeat(40);
  await assert.rejects(
    () => observeCanonicalJobs({...expected, schedulerResults, readPage: async () => ({data: {jobs: wrongHead}}), wait: async () => {}}),
    /CI_EVIDENCE_UNKNOWN: canonical CI job "gofmt" is bound to an unexpected head\/run\/attempt/,
  );
  const wrongRun = snapshot();
  wrongRun[0].run_id += 1;
  await assert.rejects(
    () => observeCanonicalJobs({...expected, schedulerResults, readPage: async () => ({data: {jobs: wrongRun}}), wait: async () => {}}),
    /unexpected head\/run\/attempt/,
  );
  const wrongAttempt = snapshot();
  wrongAttempt[0].run_attempt = 2;
  await assert.rejects(
    () => observeCanonicalJobs({...expected, schedulerResults, readPage: async () => ({data: {jobs: wrongAttempt}}), wait: async () => {}}),
    /unexpected head\/run\/attempt/,
  );
}

async function testSchedulerEstablishesLagWithoutRewritingAPI() {
  const result = await observeCanonicalJobs({...expected, schedulerResults, readPage: async () => ({data: {jobs: snapshot('pending')}}), wait: async () => {}});
  assert.strictEqual(result.states[0].state, 'observer_lag');
  assert.strictEqual(result.canonicalJobs[0].status, 'in_progress');
  assert.strictEqual(result.canonicalJobs[0].conclusion, null);
}

async function testMalformedRawStatusFailsClosed() {
  const malformed = snapshot();
  malformed[0].status = 'finished';
  await assert.rejects(
    () => observeCanonicalJobs({...expected, schedulerResults, readPage: async () => ({data: {jobs: malformed}}), wait: async () => {}}),
    /CI_EVIDENCE_UNKNOWN: canonical CI job "gofmt" has a malformed raw status projection/,
  );
}

async function testNonterminalRawConclusionFailsClosed() {
  const malformed = snapshot('pending');
  malformed[0].conclusion = 'success';
  await assert.rejects(
    () => observeCanonicalJobs({...expected, schedulerResults, readPage: async () => ({data: {jobs: malformed}}), wait: async () => {}}),
    /CI_EVIDENCE_UNKNOWN: canonical CI job "gofmt" has a nonterminal status with a terminal conclusion/,
  );
}

async function testContradictorySourcesFailClosed() {
  const failed = snapshot();
  failed[0].status = 'completed';
  failed[0].conclusion = 'failure';
  await assert.rejects(
    () => observeCanonicalJobs({...expected, schedulerResults, readPage: async () => ({data: {jobs: failed}}), wait: async () => {}}),
    /CI_EVIDENCE_UNKNOWN: canonical CI job "gofmt" contradicts same-run scheduler success/,
  );
}

async function testPaginationRejectsMalformedPage() {
  const firstPage = Array.from({length: 100}, (_, index) => ({id: 1000 + index, name: `unrelated-${index}`}));
  await assert.rejects(
    () => listWorkflowJobs(async page => page === 1 ? {data: {jobs: firstPage}} : {data: {}}, 2),
    /CI_EVIDENCE_UNKNOWN: canonical job API returned malformed data/,
  );
}

Promise.all([
  testImmediateTerminalSuccess(),
  testDelayedTerminalObservation(),
  testNeverTerminalTimesOut(),
  testRawLagWithoutSchedulerRemainsUnknown(),
  testMalformedAndDuplicateJobsFailClosed(),
  testWrongTupleFailsClosed(),
  testSchedulerEstablishesLagWithoutRewritingAPI(),
  testMalformedRawStatusFailsClosed(),
  testNonterminalRawConclusionFailsClosed(),
  testContradictorySourcesFailClosed(),
  testPaginationRejectsMalformedPage(),
]).then(() => {
  console.log('canonical job observer tests passed');
}).catch(error => {
  console.error(error);
  process.exitCode = 1;
});
