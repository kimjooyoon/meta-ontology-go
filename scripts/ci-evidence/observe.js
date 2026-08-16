'use strict';

const canonicalJobNames = Object.freeze([
  'gofmt',
  'go vet',
  'go test',
  'go test -race',
  'Semantic conformance',
  'CI policy',
]);

const canonicalJobSet = new Set(canonicalJobNames);
const schedulerSources = Object.freeze({
  gofmt: 'needs.format.result',
  'go vet': 'needs.vet.result',
  'go test': 'needs.test.result',
  'go test -race': 'needs.race.result',
  'Semantic conformance': 'needs.semantic.result',
  'CI policy': 'needs.policy.result',
});
const schedulerNeedKeys = Object.freeze({
  gofmt: 'format',
  'go vet': 'vet',
  'go test': 'test',
  'go test -race': 'race',
  'Semantic conformance': 'semantic',
  'CI policy': 'policy',
});
const rawStatuses = new Set(['queued', 'in_progress', 'completed', 'waiting', 'requested', 'pending']);
const rawConclusions = new Set(['success', 'failure', 'neutral', 'cancelled', 'skipped', 'timed_out', 'action_required', 'stale', 'startup_failure']);
const defaultMaxObservations = 6;
const defaultDelayMs = 1000;
const maxPages = 1000;

function unknown(reason) {
  return new Error(`CI_EVIDENCE_UNKNOWN: ${reason}`);
}

function positiveInteger(value) {
  return Number.isInteger(value) && value > 0;
}

function stableReason(value) {
  return String(value || 'unknown observation error').replace(/[\r\n]+/g, ' ');
}

async function listWorkflowJobs(readPage, pageLimit = maxPages) {
  const jobs = [];
  for (let page = 1; page <= pageLimit; page += 1) {
    let response;
    try {
      response = await readPage(page);
    } catch (error) {
      throw unknown(`canonical job API read failed: ${stableReason(error && error.message)}`);
    }
    if (!response || !response.data || !Array.isArray(response.data.jobs)) {
      throw unknown('canonical job API returned malformed data');
    }
    jobs.push(...response.data.jobs);
    if (response.data.jobs.length < 100) return jobs;
  }
  throw unknown(`canonical job API pagination exceeded the fail-closed page limit of ${pageLimit}`);
}

function schedulerResultsFromNeeds(results, expected) {
  if (!results || typeof results !== 'object') throw unknown('same-run scheduler results are missing');
  return canonicalJobNames.map(name => {
    const result = results[schedulerNeedKeys[name]];
    return {name, source: schedulerSources[name], result, head_sha: expected.headSHA, run_id: expected.runID, run_attempt: expected.runAttempt};
  });
}

function validateSchedulerResults(results, expected) {
  if (!Array.isArray(results)) throw unknown('same-run scheduler results are missing');
  const byName = new Map();
  for (const binding of results) {
    if (!binding || typeof binding !== 'object' || typeof binding.name !== 'string' || !canonicalJobSet.has(binding.name)) {
      throw unknown('same-run scheduler results are malformed');
    }
    if (byName.has(binding.name)) throw unknown(`duplicate same-run scheduler result "${binding.name}"`);
    if (binding.source !== schedulerSources[binding.name] || binding.result !== 'success') {
      throw unknown(`same-run scheduler result for "${binding.name}" is not an exact success`);
    }
    if (binding.head_sha !== expected.headSHA || binding.run_id !== expected.runID || binding.run_attempt !== expected.runAttempt) {
      throw unknown(`same-run scheduler result for "${binding.name}" is bound to an unexpected head/run/attempt`);
    }
    byName.set(binding.name, binding);
  }
  for (const name of canonicalJobNames) {
    if (!byName.has(name)) throw unknown(`same-run scheduler result for "${name}" is missing`);
  }
  return byName;
}

function inspectSnapshot(jobs, expected, previousIDs, schedulerByName) {
  if (!Array.isArray(jobs)) return {kind: 'fatal', reason: 'canonical job API returned malformed data'};

  const byName = new Map();
  const seenIDs = new Set();
  for (const job of jobs) {
    if (!job || typeof job !== 'object' || typeof job.name !== 'string') {
      return {kind: 'fatal', reason: 'canonical job API returned a malformed job record'};
    }
    if (!canonicalJobSet.has(job.name)) continue;
    if (byName.has(job.name)) {
      return {kind: 'fatal', reason: `duplicate canonical CI job "${job.name}"`};
    }
    if (!positiveInteger(job.id) || seenIDs.has(job.id)) {
      return {kind: 'fatal', reason: `duplicate or invalid canonical CI job id for "${job.name}"`};
    }
    if (typeof job.head_sha !== 'string' || !positiveInteger(job.run_id) || !positiveInteger(job.run_attempt)) {
      return {kind: 'fatal', reason: `canonical CI job "${job.name}" has an incomplete exact tuple`};
    }
    if (job.head_sha !== expected.headSHA || job.run_id !== expected.runID || job.run_attempt !== expected.runAttempt) {
      return {kind: 'fatal', reason: `canonical CI job "${job.name}" is bound to an unexpected head/run/attempt`};
    }
    const previousID = previousIDs.get(job.name);
    if (previousID !== undefined && previousID !== job.id) {
      return {kind: 'fatal', reason: `canonical CI job "${job.name}" changed identity during observation`};
    }
    previousIDs.set(job.name, job.id);
    seenIDs.add(job.id);
    byName.set(job.name, job);
  }

  for (const name of canonicalJobNames) {
    if (!byName.has(name)) {
      return {kind: 'pending', reason: `canonical CI job "${name}" is not yet visible`};
    }
  }

  const states = [];
  for (const name of canonicalJobNames) {
    const job = byName.get(name);
    if (typeof job.status !== 'string' || !rawStatuses.has(job.status) || !Object.prototype.hasOwnProperty.call(job, 'conclusion')) {
      return {kind: 'fatal', reason: `canonical CI job "${name}" has a malformed raw status projection`};
    }
    if (job.conclusion !== null && (typeof job.conclusion !== 'string' || !rawConclusions.has(job.conclusion))) {
      return {kind: 'fatal', reason: `canonical CI job "${name}" has a malformed raw conclusion projection`};
    }
    if (job.status === 'completed' && job.conclusion === 'success') {
      states.push({name, state: 'api_terminal_success', scheduler_result: schedulerByName.get(name).result});
      continue;
    }
    if (job.conclusion !== null && job.status !== 'completed') {
      return {kind: 'fatal', reason: `canonical CI job "${name}" has a nonterminal status with a terminal conclusion`};
    }
    if (job.conclusion !== null && job.conclusion !== 'success') {
      return {kind: 'fatal', reason: `canonical CI job "${name}" contradicts same-run scheduler success with raw conclusion "${job.conclusion}"`};
    }
    states.push({name, state: 'observer_lag', scheduler_result: schedulerByName.get(name).result});
  }

  return {
    kind: 'ready',
    canonicalJobs: canonicalJobNames.map(name => byName.get(name)),
    states,
  };
}

async function observeCanonicalJobs({
  readPage,
  expectedHead,
  runID,
  runAttempt,
  schedulerResults,
  maxObservations: observationLimit = defaultMaxObservations,
  delayMs = defaultDelayMs,
  wait = ms => new Promise(resolve => setTimeout(resolve, ms)),
  pageLimit = maxPages,
  onObservation,
}) {
  if (typeof readPage !== 'function' || typeof expectedHead !== 'string' || expectedHead === '' || !positiveInteger(runID) || !positiveInteger(runAttempt)) {
    throw unknown('canonical job observer received an incomplete exact tuple');
  }
  if (!positiveInteger(observationLimit) || delayMs < 0 || !Number.isInteger(delayMs)) {
    throw unknown('canonical job observer received an invalid finite observation bound');
  }

  const expected = {headSHA: expectedHead, runID, runAttempt};
  const schedulerByName = validateSchedulerResults(schedulerResults, expected);
  const previousIDs = new Map();
  let lastReason = 'no observation completed';
  for (let observation = 1; observation <= observationLimit; observation += 1) {
    const jobs = await listWorkflowJobs(readPage, pageLimit);
    const result = inspectSnapshot(jobs, expected, previousIDs, schedulerByName);
    if (result.kind === 'fatal') throw unknown(result.reason);
    if (result.kind === 'ready') {
      return {jobs, canonicalJobs: result.canonicalJobs, states: result.states, schedulerResults, observations: observation, bound: observationLimit};
    }
    lastReason = result.reason;
    if (typeof onObservation === 'function') await onObservation({observation, bound: observationLimit, reason: lastReason});
    if (observation < observationLimit) {
      try {
        await wait(delayMs);
      } catch (error) {
        throw unknown(`canonical job observer wait failed: ${stableReason(error && error.message)}`);
      }
    }
  }
  throw unknown(`canonical job identities were not all observed after ${observationLimit} observations; last observation: ${lastReason}`);
}

module.exports = {canonicalJobNames, listWorkflowJobs, observeCanonicalJobs, schedulerResultsFromNeeds, schedulerSources};
