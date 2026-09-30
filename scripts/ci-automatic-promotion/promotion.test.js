'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const promotion = require('./promotion');

const repository = 'alice/meta-go';
const workflowID = 332218049;
const branch = 'dev';
const sha = 'a'.repeat(40);
const mainSHA = 'b'.repeat(40);
const attemptTime = Date.parse('2026-09-30T00:00:00.000Z');

function candidateRun(overrides = {}) {
  return {
    id: 501,
    workflow_id: workflowID,
    event: 'workflow_dispatch',
    head_branch: branch,
    head_sha: sha,
    repository: {full_name: repository},
    head_repository: {full_name: repository},
    created_at: new Date(attemptTime).toISOString(),
    run_attempt: 1,
    status: 'queued',
    conclusion: null,
    ...overrides,
  };
}

function pollingAPI(listResponses, runResponses) {
  let listIndex = 0;
  let getIndex = 0;
  const github = {
    rest: {actions: {
      listWorkflowRuns() {},
      getWorkflowRun() {},
    }},
    paginate(method) {
      assert.equal(method, github.rest.actions.listWorkflowRuns);
      const index = Math.min(listIndex++, listResponses.length - 1);
      return Promise.resolve(listResponses[index]);
    },
  };
  github.rest.actions.getWorkflowRun = async () => {
    const index = Math.min(getIndex++, runResponses.length - 1);
    return {data: runResponses[index]};
  };
  return {github, listCalls: () => listIndex, getCalls: () => getIndex};
}

function pollOptions(github, overrides = {}) {
  let now = attemptTime;
  return {
    github, owner: 'alice', repo: 'meta-go', repository,
    branch, sha, dispatch: async () => {},
    clock: () => now,
    sleep: async ms => { now += ms; },
    ...overrides,
  };
}

test('dispatch waits through delayed visibility and returns the exact successful attempt', async () => {
  const run = candidateRun({status: 'in_progress'});
  run.created_at = new Date(attemptTime + 1000).toISOString();
  const completed = {...run, status: 'completed', conclusion: 'success'};
  const {github} = pollingAPI([[], [], [run], [completed]], [run, completed]);
  const waits = [];
  let now = attemptTime + 1234;
  const result = await promotion.dispatchAndWait({
    github, owner: 'alice', repo: 'meta-go', repository, branch, sha,
    clock: () => now,
    sleep: async ms => { waits.push(ms); now += ms; },
    dispatch: async () => {},
    timeoutMs: 60000,
  });
  assert.deepEqual(result, {run_id: 501, attempt: 1});
  assert.deepEqual(waits, [5000, 10000]);
});

test('dispatch polling times out within its bounded deadline', async () => {
  const {github} = pollingAPI([[], []], []);
  let now = attemptTime;
  const options = pollOptions(github, {
    timeoutMs: 10000,
    clock: () => now,
    sleep: async ms => { now += ms; },
  });
  await assert.rejects(promotion.dispatchAndWait(options), /timed out/);
  assert.equal(now - attemptTime, 10000);
});

test('multiple new sibling runs fail closed as ambiguous', async () => {
  const first = candidateRun();
  const second = candidateRun({id: 502});
  const {github} = pollingAPI([[], [first, second]], []);
  await assert.rejects(promotion.dispatchAndWait(pollOptions(github)), /multiple new matching/);
});

test('a pinned child run cannot disappear or be replaced during polling', async t => {
  await t.test('disappearing run', async () => {
    const run = candidateRun({status: 'in_progress'});
    const {github} = pollingAPI([[], [run], []], [run]);
    await assert.rejects(promotion.dispatchAndWait(pollOptions(github)), /disappeared or was replaced/);
  });
  await t.test('replacement run', async () => {
    const run = candidateRun({status: 'in_progress'});
    const replacement = candidateRun({id: 502, status: 'in_progress'});
    const {github} = pollingAPI([[], [run], [replacement]], [run]);
    await assert.rejects(promotion.dispatchAndWait(pollOptions(github)), /disappeared or was replaced/);
  });
});

test('a newly listed wrong-SHA run is rejected', async () => {
  const wrong = candidateRun({head_sha: 'c'.repeat(40)});
  const {github} = pollingAPI([[], [wrong]], []);
  await assert.rejects(promotion.dispatchAndWait(pollOptions(github)), /does not match/);
});

test('a failed exact child attempt cannot authorize promotion', async () => {
  const failed = candidateRun({status: 'completed', conclusion: 'failure'});
  const {github} = pollingAPI([[], [failed]], [failed]);
  await assert.rejects(promotion.dispatchAndWait(pollOptions(github)), /completed with failure/);
});

test('an already existing matching workflow run is not mistaken for this dispatch', async () => {
  const prior = candidateRun({id: 450, created_at: new Date(attemptTime - 60000).toISOString()});
  const {github} = pollingAPI([[prior], [prior], [prior]], []);
  let now = attemptTime;
  const options = pollOptions(github, {
    timeoutMs: 10000,
    clock: () => now,
    sleep: async ms => { now += ms; },
  });
  await assert.rejects(promotion.dispatchAndWait(options), /timed out/);
});

test('poll follows the current attempt of the exact child run', async () => {
  const first = candidateRun({status: 'in_progress'});
  const second = {...first, run_attempt: 2, status: 'completed', conclusion: 'success'};
  const {github} = pollingAPI([[], [first], [second]], [first, second]);
  let now = attemptTime;
  const options = pollOptions(github, {clock: () => now, sleep: async ms => { now += ms; }});
  assert.deepEqual(await promotion.dispatchAndWait(options), {run_id: 501, attempt: 2});
});

const requiredJobs = [
  'gofmt', 'go vet', 'go test', 'go test -race',
  'Semantic conformance', 'CI policy', 'CI proof bundle',
];

function validProof() {
  const digest = 'd'.repeat(64);
  return {
    decision: 'PASS', repository, event: 'workflow_dispatch',
    pr_number: 91, run_id: 501, run_attempt: 1,
    base_ref: 'main', base_sha: mainSHA,
    head_ref: 'dev', head_sha: sha,
    ref: 'refs/heads/dev', event_ref: 'refs/heads/dev',
    digests: {bundle_sha256: digest},
    promotion_authorization: {
      decision: 'PASS', operation: 'fast_forward', source: 'dev', target: 'main',
      base_sha: mainSHA, head_sha: sha, proof_digest: digest,
    },
    promotion_observation: {
      pr_number: 91, base_ref: 'main', base_sha: mainSHA,
      head_ref: 'dev', head_sha: sha, head_repo: repository, base_repo: repository,
      draft: false, merged: false,
      topology: {status: 'ahead', ahead_by: 1, behind_by: 0, merge_base_sha: mainSHA},
    },
  };
}

function snapshotProof(snapshotBranch, snapshotSHA) {
  const proof = validProof();
  proof.head_ref = snapshotBranch;
  proof.head_sha = snapshotSHA;
  proof.ref = `refs/heads/${snapshotBranch}`;
  proof.event_ref = proof.ref;
  proof.promotion_authorization.head_sha = snapshotSHA;
  proof.promotion_observation.head_ref = snapshotBranch;
  proof.promotion_observation.head_sha = snapshotSHA;
  return proof;
}

function validPull(overrides = {}) {
  return {
    number: 91, state: 'open', draft: false, merged: false,
    base: {ref: 'main', sha: mainSHA, repo: {full_name: repository}},
    head: {ref: 'dev', sha, repo: {full_name: repository}},
    ...overrides,
  };
}

function actionMocks(overrides, calls, run) {
  return {
    getWorkflowRun: async () => ({data: overrides.run ? overrides.run(++calls.run) : run}),
    listJobsForWorkflowRun() {},
  };
}

function pullMocks(overrides, calls, pull) {
  return {
    list() {},
    get: async () => ({data: overrides.latestPull || pull}),
    merge: async request => {
      calls.merge++;
      calls.mergeRequest = request;
      return {data: {merged: true, sha: 'e'.repeat(40)}};
    },
  };
}

function gitMocks(overrides = {}) {
  return {
    getRef: overrides.getRef || (async ({ref}) => ({data: {object: {sha: ref === 'heads/dev' ? sha : mainSHA}}})),
    getCommit: overrides.getCommit || (async ({commit_sha}) => ({data: {
      sha: commit_sha,
      tree: {sha: commit_sha === mainSHA ? 'f'.repeat(40) : 'c'.repeat(40)},
      parents: [{sha: mainSHA}],
    }})),
  };
}

function mergeGithub(overrides, calls, run, pull) {
  const github = {rest: {
    actions: actionMocks(overrides, calls, run),
    pulls: pullMocks(overrides, calls, pull),
    git: gitMocks(overrides),
    repos: {compareCommits: overrides.compareCommits || (async () => ({data: {
      status: 'ahead', ahead_by: 1, behind_by: 0,
      merge_base_commit: {sha: mainSHA},
    }}))},
  }};
  github.paginate = async (method, request) => {
    if (method === github.rest.actions.listJobsForWorkflowRun) {
      return requiredJobs.map(name => ({
        name, status: 'completed', conclusion: 'success', head_sha: run.head_sha,
      }));
    }
    if (method === github.rest.pulls.list) return overrides.openPRs || [pull];
    throw new Error('unexpected pagination target');
  };
  return github;
}

function mergeAPI(overrides = {}) {
  const calls = {merge: 0, run: 0};
  const expected = overrides.expected || {
    branch: 'dev',
    sha,
    baseSHA: mainSHA,
    prNumber: 91,
    runID: 501, attempt: 1,
  };
  const run = candidateRun({
    id: expected.runID, status: 'completed', conclusion: 'success',
    run_attempt: expected.attempt, head_branch: expected.branch, head_sha: expected.sha,
  });
  const pull = validPull({number: expected.prNumber, base: {
    ref: 'main', sha: expected.baseSHA, repo: {full_name: repository},
  }, head: {ref: expected.branch, sha: expected.sha, repo: {full_name: repository}}});
  return {github: mergeGithub(overrides, calls, run, pull), calls, run, pull, expected};
}

function mergeOptions(api, proofOverrides = {}) {
  const proof = {...validProof(), ...proofOverrides};
  const expected = api.expected;
  return {
    github: api.github, core: {info() {}}, fs: {readFileSync: () => JSON.stringify(proof)},
    owner: 'alice', repo: 'meta-go', repository, proofPath: '/tmp/ci-proof.json',
    runID: expected.runID, attempt: expected.attempt,
    branch: expected.branch, sha: expected.sha,
    baseSHA: expected.baseSHA, prNumber: expected.prNumber,
  };
}

test('valid proof, exact six checks, live tuple, and normal PR API merge succeed', async () => {
  const api = mergeAPI();
  await promotion.mergeAuthorizedPromotion(mergeOptions(api));
  assert.equal(api.calls.merge, 1);
  assert.equal(api.calls.mergeRequest.sha, sha);
  assert.equal(api.calls.mergeRequest.merge_method, 'squash');
});

test('production-style snapshot with the live dev tree and exact main parent merges', async () => {
  const snapshotSHA = '9'.repeat(40);
  const snapshotBranch = `agent/main-promotion-snapshot-${sha}`;
  const expected = {
    branch: snapshotBranch, sha: snapshotSHA, baseSHA: mainSHA,
    prNumber: 91, runID: 501, attempt: 1,
  };
  const pull = validPull({head: {ref: snapshotBranch, sha: snapshotSHA, repo: {full_name: repository}}});
  const api = mergeAPI({expected, openPRs: [pull], latestPull: pull});
  await promotion.mergeAuthorizedPromotion(mergeOptions(api, snapshotProof(snapshotBranch, snapshotSHA)));
  assert.equal(api.calls.merge, 1);
  assert.equal(api.calls.mergeRequest.sha, snapshotSHA);
});

test('snapshot candidates with the wrong parent, tree, or live dev source do not merge', async () => {
  const snapshotSHA = '9'.repeat(40);
  const snapshotBranch = `agent/main-promotion-snapshot-${sha}`;
  const expected = {
    branch: snapshotBranch, sha: snapshotSHA, baseSHA: mainSHA,
    prNumber: 91, runID: 501, attempt: 1,
  };
  for (const change of ['parent', 'tree', 'dev']) {
    const pull = validPull({head: {ref: snapshotBranch, sha: snapshotSHA, repo: {full_name: repository}}});
    const invalid = change === 'dev' ? '8'.repeat(40) : null;
    const api = mergeAPI({expected, openPRs: [pull], latestPull: pull,
      getRef: async ({ref}) => {
        const value = ref === 'heads/dev' ? invalid || sha : mainSHA;
        return {data: {object: {sha: value}}};
      },
      getCommit: async ({commit_sha}) => ({data: {
        sha: commit_sha,
        tree: {sha: change === 'tree' && commit_sha === snapshotSHA ? '7'.repeat(40) :
          commit_sha === mainSHA ? 'f'.repeat(40) : 'c'.repeat(40)},
        parents: [{sha: change === 'parent' && commit_sha === snapshotSHA ?
          '6'.repeat(40) : mainSHA}],
      }}),
    });
    await promotion.mergeAuthorizedPromotion(mergeOptions(api, snapshotProof(snapshotBranch, snapshotSHA)));
    assert.equal(api.calls.merge, 0, `${change} should fail closed`);
  }
});

test('stale proof digest binding fails before the PR merge API', async () => {
  const api = mergeAPI();
  const options = mergeOptions(api, {
    promotion_authorization: {...validProof().promotion_authorization, proof_digest: '0'.repeat(64)},
  });
  await assert.rejects(promotion.mergeAuthorizedPromotion(options), /not digest-bound/);
  assert.equal(api.calls.merge, 0);
});

test('changed live PR tuple does not merge', async () => {
  const api = mergeAPI({latestPull: validPull({head: {
    ref: 'dev', sha: '9'.repeat(40), repo: {full_name: repository},
  }})});
  await promotion.mergeAuthorizedPromotion(mergeOptions(api));
  assert.equal(api.calls.merge, 0);
});

test('refs changing during final PR verification do not merge', async () => {
  let devReads = 0;
  const api = mergeAPI({getRef: async ({ref}) => {
    if (ref === 'heads/dev' && ++devReads > 1) return {data: {object: {sha: '8'.repeat(40)}}};
    return {data: {object: {sha: ref === 'heads/dev' ? sha : mainSHA}}};
  }});
  await promotion.mergeAuthorizedPromotion(mergeOptions(api));
  assert.equal(api.calls.merge, 0);
});

test('a rerun after dispatch completion fails the current-attempt recheck', async () => {
  const api = mergeAPI({run: calls => candidateRun({
    id: 501, status: 'completed', conclusion: 'success', run_attempt: calls === 1 ? 1 : 2,
  })});
  await assert.rejects(promotion.mergeAuthorizedPromotion(mergeOptions(api)), /attempt changed/);
  assert.equal(api.calls.merge, 0);
});

test('missing one of the seven exact successful jobs fails closed', async () => {
  const api = mergeAPI();
  const originalPaginate = api.github.paginate;
  api.github.paginate = async (method, request) => {
    const result = await originalPaginate(method, request);
    if (method === api.github.rest.actions.listJobsForWorkflowRun) return result.slice(1);
    return result;
  };
  await assert.rejects(promotion.mergeAuthorizedPromotion(mergeOptions(api)), /required exact-head job/);
  assert.equal(api.calls.merge, 0);
});

test('workflow helper is checked out from trusted workflow_run github.sha', () => {
  const workflowPath = path.join(__dirname, '../../.github/workflows/ci-automatic-promotion.yml');
  const workflow = fs.readFileSync(workflowPath, 'utf8');
  assert.equal((workflow.match(/ref: \$\{\{ github\.sha \}\}/g) || []).length, 2);
  assert.match(workflow, /github\.sha is the default-branch workflow revision/);
  assert.doesNotMatch(workflow, /ref: \$\{\{ github\.event\.workflow_run\.head_sha \}\}/);
  assert.match(workflow, /workflow_run\.event == 'push'/);
  assert.match(workflow, /ci-automatic-promotion-\$\{\{ github\.event\.workflow_run\.event \}\}/);
});
