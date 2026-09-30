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
const REQUIRED_CHECKS = [...REQUIRED_JOBS, 'CI proof bundle'];
const GITHUB_ACTIONS_APP_ID = 15368;
const POLL_TIMEOUT_MS = 40 * 60 * 1000;
const REQUEST_TIMEOUT_MS = 15 * 1000;

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

function exactPullRequestRunIdentity(run, expected) {
  const base = run && run.workflow_id === CI_WORKFLOW_ID && run.event === 'pull_request' &&
    run.head_branch === expected.branch && run.head_sha === expected.sha &&
    run.repository?.full_name === expected.repository &&
    run.head_repository?.full_name === expected.repository;
  if (!base) return false;
  // Some Actions API responses omit PR associations for pull_request runs.
  // When GitHub returns them, bind the run to exactly the expected PR tuple.
  if (!Array.isArray(run.pull_requests) || run.pull_requests.length === 0) return true;
  if (!Number.isSafeInteger(expected.prNumber) || run.pull_requests.length !== 1) return false;
  const pr = run.pull_requests[0];
  return pr.number === expected.prNumber && pr.head?.ref === expected.branch &&
    pr.head?.sha === expected.sha && pr.base?.ref === 'main' &&
    pr.base?.sha === expected.baseSHA;
}

async function matchingPullRequestRuns(github, owner, repo, branch, sha) {
  const runs = await github.paginate(github.rest.actions.listWorkflowRuns, {
    owner, repo, workflow_id: CI_WORKFLOW_ID, branch, event: 'pull_request',
    head_sha: sha, per_page: 100, request: {timeout: REQUEST_TIMEOUT_MS},
  });
  if (runs.length >= 1000) {
    throw new Error('matching PR CI history reached the API result cap; refusing ambiguous attribution');
  }
  return runs;
}

function validatePullRequestRun(run, expected, eventStartedAt) {
  if (!exactPullRequestRunIdentity(run, expected) || !Number.isSafeInteger(run.id) || run.id <= 0 ||
      !Number.isSafeInteger(run.run_attempt) || run.run_attempt < 1 ||
      !Number.isSafeInteger(run.check_suite_id) || run.check_suite_id <= 0) {
    throw new Error('PR CI run has an unexpected workflow, event, repository, branch, SHA, or check suite');
  }
  const createdAt = Date.parse(run.created_at || '');
  const eventSecond = Math.floor(eventStartedAt / 1000) * 1000;
  if (!Number.isFinite(createdAt) || createdAt < eventSecond) {
    throw new Error('matching PR CI run predates the current PR event and cannot be attributed to it');
  }
}

async function readListedPullRequestRun(github, owner, repo, listed, expected, eventStartedAt) {
  validatePullRequestRun(listed, expected, eventStartedAt);
  const run = (await github.rest.actions.getWorkflowRun({
    owner, repo, run_id: listed.id, request: {timeout: REQUEST_TIMEOUT_MS},
  })).data;
  validatePullRequestRun(run, expected, eventStartedAt);
  if (run.id !== listed.id) throw new Error('PR CI run lookup returned a different run identity');
  return run;
}

async function successfulRequiredJobs(github, owner, repo, run, apiURL) {
  const jobs = await github.paginate(github.rest.actions.listJobsForWorkflowRunAttempt, {
    owner, repo, run_id: run.id, attempt_number: run.run_attempt, per_page: 100,
    request: {timeout: REQUEST_TIMEOUT_MS},
  });
  apiURL = apiURL || 'https://api.github.com';
  const checkRuns = await github.paginate(github.rest.checks.listForRef, {
    owner, repo, ref: run.head_sha, filter: 'all', per_page: 100,
    request: {timeout: REQUEST_TIMEOUT_MS},
  });
  for (const name of REQUIRED_CHECKS) {
    const matches = jobs.filter(job => job.name === name);
    if (matches.length !== 1 || matches[0].status !== 'completed' ||
        matches[0].conclusion !== 'success' || matches[0].head_sha !== run.head_sha ||
        (Number.isSafeInteger(matches[0].run_attempt) && matches[0].run_attempt !== run.run_attempt)) {
      throw new Error(`required exact-attempt PR job is not uniquely successful: ${name}`);
    }
    const checkRunID = exactCheckRunID(matches[0].check_run_url, apiURL, owner, repo);
    const checkMatches = checkRuns.filter(check => check.id === checkRunID);
    if (!checkRunID || checkMatches.length !== 1 || checkMatches[0].name !== name ||
        checkMatches[0].check_suite?.id !== run.check_suite_id ||
        checkMatches[0].head_sha !== run.head_sha || checkMatches[0].status !== 'completed' ||
        checkMatches[0].conclusion !== 'success' ||
        checkMatches[0].app?.id !== GITHUB_ACTIONS_APP_ID) {
      throw new Error(`required PR check is not uniquely successful from GitHub Actions app ${GITHUB_ACTIONS_APP_ID}: ${name}`);
    }
  }
}

function exactCheckRunID(checkRunURL, apiURL, owner, repo) {
  try {
    const api = new URL(apiURL || 'https://api.github.com');
    const checkRun = new URL(checkRunURL);
    const basePath = api.pathname.replace(/\/+$/, '');
    const prefix = `${basePath}/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/check-runs/`;
    if (checkRun.origin !== api.origin || checkRun.username || checkRun.password ||
        checkRun.search || checkRun.hash || !checkRun.pathname.startsWith(prefix)) return 0;
    const value = checkRun.pathname.slice(prefix.length);
    if (!/^[1-9][0-9]*$/.test(value)) return 0;
    const id = Number(value);
    return Number.isSafeInteger(id) ? id : 0;
  } catch {
    return 0;
  }
}

async function pollPullRequestRun(options) {
  const {github, owner, repo, expected, priorIDs, eventStartedAt, clock, sleep} = options;
  const deadline = clock() + Math.min(options.timeoutMs ?? POLL_TIMEOUT_MS, POLL_TIMEOUT_MS);
  let delay = 5000;
  let observedAttempt = 0;
  let pinnedRunID = 0;
  while (clock() < deadline) {
    const runs = await matchingPullRequestRuns(
      github, owner, repo, expected.branch, expected.sha,
    );
    uniqueRunIDs(runs);
    const fresh = runs.filter(run => !priorIDs.has(run.id));
    if (fresh.some(run => !exactPullRequestRunIdentity(run, expected))) {
      throw new Error('new PR CI run does not match the exact pull_request repository, branch, and SHA');
    }
    if (fresh.length > 1) throw new Error('multiple new PR CI runs make current-event identity ambiguous');
    if (pinnedRunID && (fresh.length !== 1 || fresh[0].id !== pinnedRunID)) {
      throw new Error('the exact PR CI run disappeared or was replaced during polling');
    }
    if (fresh.length === 1) {
      const listed = fresh[0];
      if (!pinnedRunID) pinnedRunID = listed.id;
      const run = await readListedPullRequestRun(
        github, owner, repo, listed, expected, eventStartedAt,
      );
      if (run.run_attempt < observedAttempt) throw new Error('PR CI attempt moved backwards');
      if (run.status === 'completed') {
        if (run.conclusion !== 'success') {
          throw new Error(`exact PR CI run completed with ${run.conclusion || 'no conclusion'}`);
        }
        await successfulRequiredJobs(github, owner, repo, run, options.apiURL);
        return {run_id: run.id, attempt: run.run_attempt};
      }
      if (!['queued', 'in_progress', 'waiting', 'pending', 'requested'].includes(run.status)) {
        throw new Error(`exact PR CI run has unsupported status ${run.status || 'missing'}`);
      }
      // An empty job list while the workflow is starting is not terminal.
      observedAttempt = run.run_attempt;
    }
    const remaining = deadline - clock();
    if (remaining <= 0) break;
    await sleep(Math.min(delay, remaining));
    delay = Math.min(delay * 2, 30000);
  }
  throw new Error('timed out waiting for the exact PR-authoritative CI run to complete');
}

async function pullRequestRunAndWait(options) {
  const {github, owner, repo, repository, branch, sha, prNumber, trigger} = options;
  if (!repository || !/^([0-9a-f]{40})$/.test(sha || '') ||
      (prNumber !== undefined && (!Number.isSafeInteger(prNumber) || prNumber <= 0)) ||
      !/^([0-9a-f]{40})$/.test(options.baseSHA || '') || typeof trigger !== 'function' ||
      (branch !== 'dev' && !/^agent\/main-promotion-snapshot-[0-9a-f]{40}$/.test(branch))) {
    throw new Error('refusing to await PR CI for an unrecognized promotion candidate');
  }
  const before = await matchingPullRequestRuns(github, owner, repo, branch, sha);
  const priorIDs = uniqueRunIDs(before);
  const clock = options.clock || Date.now;
  const eventStartedAt = clock();
  const triggerResult = await trigger();
  const triggeredPRNumber = prNumber ?? triggerResult?.number;
  if (!Number.isSafeInteger(triggeredPRNumber) || triggeredPRNumber <= 0) {
    throw new Error('PR event trigger did not identify its exact pull request');
  }
  const result = await pollPullRequestRun({
    ...options, expected: {
      repository, branch, sha, prNumber: triggeredPRNumber, baseSHA: options.baseSHA,
    }, priorIDs, eventStartedAt,
    clock, sleep: options.sleep || (ms => new Promise(resolve => setTimeout(resolve, ms))),
  });
  return {...result, pr_number: triggeredPRNumber};
}

async function appPullRequestRequest(options) {
  const {fetchImpl = fetch, apiURL, token, owner, repo, method, pullNumber, body} = options;
  if (typeof token !== 'string' || token.trim() === '') {
    const error = new Error('PROMOTION_APP_CONFIGURATION_REQUIRED: configure the scoped GitHub App token before opening or updating a promotion PR');
    error.code = 'PROMOTION_APP_CONFIGURATION_REQUIRED';
    throw error;
  }
  if (!['POST', 'PATCH'].includes(method) ||
      (method === 'POST' && pullNumber !== undefined) ||
      (method === 'PATCH' && (!Number.isSafeInteger(pullNumber) || pullNumber <= 0)) ||
      !/^[A-Za-z0-9_.-]+$/.test(owner || '') || !/^[A-Za-z0-9_.-]+$/.test(repo || '')) {
    throw new Error('App token is restricted to create or update a repository pull request');
  }
  const allowedFields = method === 'POST' ? ['base', 'body', 'head', 'title'] : ['body'];
  if (!body || typeof body !== 'object' || Array.isArray(body) ||
      Object.keys(body).some(key => !allowedFields.includes(key)) ||
      (method === 'POST' && (!body.base || !body.head || !body.title || !body.body)) ||
      (method === 'PATCH' && (Object.keys(body).length !== 1 || typeof body.body !== 'string'))) {
    throw new Error('App token request contains fields outside the promotion PR create/body-refresh contract');
  }
  const root = (apiURL || 'https://api.github.com').replace(/\/$/, '');
  const endpoint = method === 'POST' ? '/pulls' : `/pulls/${pullNumber}`;
  const response = await fetchImpl(`${root}/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}${endpoint}`, {
    method,
    headers: {
      accept: 'application/vnd.github+json',
      authorization: `Bearer ${token}`,
      'content-type': 'application/json',
      'x-github-api-version': '2022-11-28',
    },
    body: JSON.stringify(body),
    signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
  });
  let result;
  try { result = await response.json(); } catch { result = {}; }
  if (!response.ok) {
    throw new Error(`PROMOTION_APP_PR_API_FAILED: ${method} pull request returned ${response.status}: ${result.message || 'no API message'}`);
  }
  if (!result || typeof result !== 'object' || !Number.isSafeInteger(result.number)) {
    throw new Error(`PROMOTION_APP_PR_API_FAILED: ${method} pull request returned an invalid response`);
  }
  return result;
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

function exactSuccessfulAttempt(run, expected) {
  if (!exactPullRequestRunIdentity(run, expected) || run.id !== expected.runID ||
      run.run_attempt !== expected.attempt || run.status !== 'completed' ||
      run.conclusion !== 'success') {
    throw new Error('PR-authoritative CI run or current attempt changed after authorization');
  }
}

function verifyPromotionProof(proof, expected, pull, repository) {
  const authorization = proof.promotion_authorization;
  const observation = proof.promotion_observation;
  const proofDigest = proof.digests?.bundle_sha256;
  if (proof.decision !== 'PASS' || proof.repository !== repository ||
      proof.event !== 'pull_request' || proof.pr_number !== pull.number ||
      proof.run_id !== expected.runID || proof.run_attempt !== expected.attempt ||
      proof.base_ref !== 'main' || proof.base_sha !== expected.baseSHA ||
      pull.base.sha !== expected.baseSHA ||
      proof.head_ref !== pull.head.ref || proof.head_sha !== expected.sha ||
      proof.ref !== `refs/pull/${pull.number}/merge` || proof.event_ref !== proof.ref ||
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
  await successfulRequiredJobs(github, owner, repo, runResponse.data, options.apiURL);
  const open = await github.paginate(github.rest.pulls.list, {
    owner, repo, state: 'open', base: 'main', head: `${owner}:${expected.branch}`, per_page: 100,
  });
  if (open.length !== 1 || !exactPromotionPull(open[0], expected, repository)) {
    throw new Error('promotion head does not identify exactly one exact open same-repository PR');
  }
  const pull = open[0];
  const proof = JSON.parse(fs.readFileSync(proofPath, 'utf8'));
  verifyPromotionProof(proof, expected, pull, repository);
  const latestRun = await github.rest.actions.getWorkflowRun({
    owner, repo, run_id: expected.runID, request: {timeout: REQUEST_TIMEOUT_MS},
  });
  exactSuccessfulAttempt(latestRun.data, expected);
  await successfulRequiredJobs(github, owner, repo, latestRun.data, options.apiURL);
  if (!await finalLiveTuple(github, owner, repo, repository, expected, pull)) {
    core.info('dev, main, topology, or the promotion PR tuple changed; no merge performed.');
    return;
  }
  const merged = (await github.rest.pulls.merge({
    owner, repo, pull_number: pull.number, sha: expected.sha, merge_method: 'squash',
  })).data;
  if (!merged.merged) {
    throw new Error(`GitHub did not merge the authorized promotion: ${merged.message || 'unknown reason'}`);
  }
  core.info(`Merged CI-authorized promotion PR ${pull.number} as ${merged.sha}.`);
}

module.exports = {
  appPullRequestRequest,
  mergeAuthorizedPromotion,
  exactCheckRunID,
  pullRequestRunAndWait,
  pollPullRequestRun,
  successfulRequiredJobs,
};
