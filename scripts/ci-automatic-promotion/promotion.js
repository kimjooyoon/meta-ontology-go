'use strict';

const CI_WORKFLOW_ID = 332218049;
const REQUIRED_JOBS = [
  'gofmt',
  'go vet',
  'go test',
  'go test -race',
  'Semantic conformance',
  'CI policy',
];
const POLL_TIMEOUT_MS = 40 * 60 * 1000;
const REQUEST_TIMEOUT_MS = 15 * 1000;

function exactRunIdentity(run, expected) {
  return run && run.workflow_id === CI_WORKFLOW_ID &&
    run.event === 'workflow_dispatch' && run.head_branch === expected.branch &&
    run.head_sha === expected.sha &&
    run.repository?.full_name === expected.repository &&
    run.head_repository?.full_name === expected.repository;
}

function validateRunIdentity(run, expected) {
  if (!exactRunIdentity(run, expected) || !Number.isSafeInteger(run.id) || run.id <= 0 ||
      !Number.isSafeInteger(run.run_attempt) || run.run_attempt < 1) {
    throw new Error('dispatched CI run has an unexpected workflow, repository, branch, or SHA');
  }
  const createdAt = Date.parse(run.created_at || '');
  const dispatchSecond = Math.floor(expected.dispatchStartedAt / 1000) * 1000;
  if (!Number.isFinite(createdAt) || createdAt < dispatchSecond) {
    throw new Error('matching CI run predates this dispatch and cannot be attributed to it');
  }
}

async function matchingRuns(github, owner, repo, branch, sha) {
  const runs = await github.paginate(github.rest.actions.listWorkflowRuns, {
    owner, repo, workflow_id: CI_WORKFLOW_ID, branch,
    event: 'workflow_dispatch', head_sha: sha, per_page: 100,
    request: {timeout: REQUEST_TIMEOUT_MS},
  });
  if (runs.length >= 1000) {
    throw new Error('matching CI history reached the API result cap; refusing ambiguous attribution');
  }
  return runs;
}

function uniqueRunIDs(runs) {
  const ids = new Set();
  for (const run of runs) {
    if (!Number.isSafeInteger(run.id) || run.id <= 0 || ids.has(run.id)) {
      throw new Error('CI run listing contains an invalid or duplicate run identity');
    }
    ids.add(run.id);
  }
  return ids;
}

async function readListedRun(github, owner, repo, listed, expected, dispatchStartedAt) {
  validateRunIdentity(listed, {...expected, dispatchStartedAt});
  const run = (await github.rest.actions.getWorkflowRun({
    owner, repo, run_id: listed.id, request: {timeout: REQUEST_TIMEOUT_MS},
  })).data;
  validateRunIdentity(run, {...expected, dispatchStartedAt});
  if (run.id !== listed.id) throw new Error('CI run lookup returned a different run identity');
  return run;
}

function checkRunProgress(run, priorAttempt) {
  if (run.run_attempt < priorAttempt) throw new Error('CI run attempt moved backwards');
  if (run.status === 'completed') {
    if (run.conclusion !== 'success') {
      throw new Error(`exact dispatched CI run completed with ${run.conclusion || 'no conclusion'}`);
    }
    return true;
  }
  if (!['queued', 'in_progress', 'waiting', 'pending', 'requested'].includes(run.status)) {
    throw new Error(`exact dispatched CI run has unsupported status ${run.status || 'missing'}`);
  }
  return false;
}

async function pollDispatchedRun(options) {
  const {github, owner, repo, expected, priorIDs, dispatchStartedAt, clock, sleep} = options;
  const deadline = clock() + Math.min(options.timeoutMs ?? POLL_TIMEOUT_MS, POLL_TIMEOUT_MS);
  let delay = 5000;
  let observedAttempt = 0;
  let pinnedRunID = 0;
  while (clock() < deadline) {
    const runs = await matchingRuns(github, owner, repo, expected.branch, expected.sha);
    uniqueRunIDs(runs);
    const fresh = runs.filter(run => !priorIDs.has(run.id));
    if (fresh.some(run => !exactRunIdentity(run, expected))) {
      throw new Error('new CI run does not match the exact dispatched repository, branch, and SHA');
    }
    if (fresh.length > 1) throw new Error('multiple new matching CI runs make dispatch identity ambiguous');
    if (pinnedRunID && (fresh.length !== 1 || fresh[0].id !== pinnedRunID)) {
      throw new Error('the exact dispatched CI run disappeared or was replaced during polling');
    }
    if (fresh.length === 1) {
      const listed = fresh[0];
      if (!pinnedRunID) pinnedRunID = listed.id;
      const run = await readListedRun(github, owner, repo, listed, expected, dispatchStartedAt);
      if (checkRunProgress(run, observedAttempt)) return {run_id: run.id, attempt: run.run_attempt};
      observedAttempt = run.run_attempt;
    }
    const remaining = deadline - clock();
    if (remaining <= 0) break;
    await sleep(Math.min(delay, remaining));
    delay = Math.min(delay * 2, 30000);
  }
  throw new Error('timed out waiting for the exact dispatched CI run to complete');
}

async function dispatchAndWait(options) {
  const {github, owner, repo, repository, branch, sha, dispatch} = options;
  if (!repository || !/^([0-9a-f]{40})$/.test(sha || '') ||
      (branch !== 'dev' && !/^agent\/main-promotion-snapshot-[0-9a-f]{40}$/.test(branch))) {
    throw new Error('refusing to dispatch promotion CI for an unrecognized candidate');
  }
  const before = await matchingRuns(github, owner, repo, branch, sha);
  const priorIDs = uniqueRunIDs(before);
  const clock = options.clock || Date.now;
  const dispatchStartedAt = clock();
  await dispatch();
  return pollDispatchedRun({
    ...options, expected: {repository, branch, sha}, priorIDs, dispatchStartedAt,
    clock,
    sleep: options.sleep || (ms => new Promise(resolve => setTimeout(resolve, ms))),
  });
}

async function liveRef(github, owner, repo, name) {
  const response = await github.rest.git.getRef({owner, repo, ref: `heads/${name}`});
  const sha = response.data?.object?.sha;
  if (!/^[0-9a-f]{40}$/.test(sha || '')) throw new Error(`live ${name} ref has no exact commit SHA`);
  return sha;
}

async function exactCommit(github, owner, repo, sha) {
  return (await github.rest.git.getCommit({owner, repo, commit_sha: sha})).data;
}

async function successfulRequiredJobs(github, owner, repo, runID, sha) {
  const jobs = await github.paginate(github.rest.actions.listJobsForWorkflowRun, {
    owner, repo, run_id: runID, per_page: 100,
    request: {timeout: REQUEST_TIMEOUT_MS},
  });
  for (const name of [...REQUIRED_JOBS, 'CI proof bundle']) {
    const matches = jobs.filter(job => job.name === name);
    if (matches.length !== 1 || matches[0].status !== 'completed' ||
        matches[0].conclusion !== 'success' || matches[0].head_sha !== sha) {
      throw new Error(`required exact-head job is not uniquely successful: ${name}`);
    }
  }
}

function exactSuccessfulAttempt(run, expected) {
  if (!exactRunIdentity(run, expected) || run.id !== expected.runID ||
      run.run_attempt !== expected.attempt || run.status !== 'completed' ||
      run.conclusion !== 'success') {
    throw new Error('promotion CI run or current attempt changed after authorization');
  }
}

function verifyPromotionProof(proof, expected, pull, repository) {
  const authorization = proof.promotion_authorization;
  const observation = proof.promotion_observation;
  const proofDigest = proof.digests?.bundle_sha256;
  if (proof.decision !== 'PASS' || proof.repository !== repository ||
      proof.event !== 'workflow_dispatch' || proof.pr_number !== pull.number ||
      proof.run_id !== expected.runID || proof.run_attempt !== expected.attempt ||
      proof.base_ref !== 'main' || proof.base_sha !== expected.baseSHA ||
      pull.base.sha !== expected.baseSHA ||
      proof.head_ref !== pull.head.ref || proof.head_sha !== expected.sha ||
      proof.ref !== `refs/heads/${pull.head.ref}` || proof.event_ref !== proof.ref ||
      !/^[0-9a-f]{64}$/.test(proofDigest || '') ||
      authorization?.decision !== 'PASS' || authorization.operation !== 'fast_forward' ||
      authorization.source !== 'dev' || authorization.target !== 'main' ||
      authorization.base_sha !== pull.base.sha || authorization.head_sha !== expected.sha ||
      authorization.proof_digest !== proofDigest ||
      observation?.pr_number !== pull.number || observation.base_ref !== 'main' ||
      observation.head_ref !== pull.head.ref || observation.head_sha !== expected.sha ||
      observation.base_sha !== pull.base.sha || observation.head_repo !== repository ||
      observation.base_repo !== repository || observation.draft !== false ||
      observation.merged !== false || observation.topology?.status !== 'ahead' ||
      observation.topology?.ahead_by < 1 || observation.topology?.behind_by !== 0 ||
      observation.topology?.merge_base_sha !== pull.base.sha) {
    throw new Error('promotion authorization artifact is missing, stale, or not digest-bound to this PR');
  }
}

function exactPromotionPull(pull, expected, repository) {
  return !pull.draft && pull.state === 'open' && !pull.merged &&
    pull.base.repo?.full_name === repository && pull.head.repo?.full_name === repository &&
    pull.base.ref === 'main' && pull.base.sha === expected.baseSHA && pull.head.sha === expected.sha &&
    pull.head.ref === expected.branch && pull.number === expected.prNumber;
}

async function finalLiveTuple(github, owner, repo, repository, expected, pull) {
  const [devSHA, mainSHA] = await Promise.all([
    liveRef(github, owner, repo, 'dev'), liveRef(github, owner, repo, 'main'),
  ]);
  if (mainSHA !== pull.base.sha) return false;
  const devCommit = await exactCommit(github, owner, repo, devSHA);
  const candidate = await exactCommit(github, owner, repo, expected.sha);
  const isDirect = expected.branch === 'dev' && expected.sha === devSHA;
  const isSnapshot = expected.branch === `agent/main-promotion-snapshot-${devSHA}` &&
    candidate.parents?.length === 1 && candidate.parents[0]?.sha === mainSHA &&
    candidate.tree?.sha === devCommit.tree?.sha;
  if (!isDirect && !isSnapshot) return false;
  const topology = (await github.rest.repos.compareCommits({
    owner, repo, base: mainSHA, head: candidate.sha,
  })).data;
  if (topology.status !== 'ahead' || topology.ahead_by < 1 || topology.behind_by !== 0 ||
      topology.merge_base_commit?.sha !== mainSHA) return false;
  const latest = (await github.rest.pulls.get({owner, repo, pull_number: pull.number})).data;
  const latestMatches = latest.state === 'open' && !latest.draft && !latest.merged &&
    latest.base.ref === 'main' && latest.base.sha === mainSHA &&
    latest.head.ref === expected.branch && latest.head.sha === expected.sha &&
    latest.base.repo?.full_name === repository && latest.head.repo?.full_name === repository;
  if (!latestMatches) return false;
  const [finalDevSHA, finalMainSHA] = await Promise.all([
    liveRef(github, owner, repo, 'dev'), liveRef(github, owner, repo, 'main'),
  ]);
  return finalDevSHA === devSHA && finalMainSHA === mainSHA;
}

async function mergeAuthorizedPromotion(options) {
  const {github, core, fs, owner, repo, repository, proofPath} = options;
  const expected = {
    branch: options.branch, sha: options.sha, runID: options.runID,
    attempt: options.attempt, prNumber: options.prNumber,
    repository, baseSHA: options.baseSHA,
  };
  const runResponse = await github.rest.actions.getWorkflowRun({
    owner, repo, run_id: expected.runID, request: {timeout: REQUEST_TIMEOUT_MS},
  });
  exactSuccessfulAttempt(runResponse.data, expected);
  await successfulRequiredJobs(github, owner, repo, expected.runID, expected.sha);
  const open = await github.paginate(github.rest.pulls.list, {
    owner, repo, state: 'open', base: 'main', head: `${owner}:${expected.branch}`, per_page: 100,
  });
  if (open.length !== 1 || !exactPromotionPull(open[0], expected, repository)) {
    throw new Error('promotion head does not identify exactly one exact open same-repository PR');
  }
  const pull = open[0];
  const proof = JSON.parse(fs.readFileSync(proofPath, 'utf8'));
  verifyPromotionProof(proof, expected, pull, repository);
  if (!await finalLiveTuple(github, owner, repo, repository, expected, pull)) {
    core.info('dev, main, topology, or the promotion PR tuple changed; no merge performed.');
    return;
  }
  const latestRun = await github.rest.actions.getWorkflowRun({
    owner, repo, run_id: expected.runID, request: {timeout: REQUEST_TIMEOUT_MS},
  });
  exactSuccessfulAttempt(latestRun.data, expected);
  const merged = (await github.rest.pulls.merge({
    owner, repo, pull_number: pull.number, sha: expected.sha, merge_method: 'squash',
  })).data;
  if (!merged.merged) {
    throw new Error(`GitHub did not merge the authorized promotion: ${merged.message || 'unknown reason'}`);
  }
  core.info(`Merged CI-authorized promotion PR ${pull.number} as ${merged.sha}.`);
}

module.exports = {
  dispatchAndWait,
  mergeAuthorizedPromotion,
  pollDispatchedRun,
};
