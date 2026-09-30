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
    event: 'pull_request',
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

function prRun(overrides = {}) {
  return candidateRun({
    id: 601, event: 'pull_request', head_branch: branch, head_sha: sha,
    check_suite_id: 701, created_at: new Date(attemptTime + 1000).toISOString(),
    status: 'queued', conclusion: null,
    ...overrides,
  });
}

function prPollingAPI(listResponses, runResponses) {
  let listIndex = 0;
  let getIndex = 0;
  const calls = {jobs: 0, checks: 0};
  const run = prRun({status: 'completed', conclusion: 'success'});
  const jobs = ['gofmt', 'go vet', 'go test', 'go test -race', 'Semantic conformance', 'CI policy', 'CI proof bundle']
    .map((name, index) => ({
      name, status: 'completed', conclusion: 'success', head_sha: sha, run_attempt: 1,
      check_run_url: `https://api.github.com/repos/alice/meta-go/check-runs/${800 + index}`,
    }));
  const checks = jobs.map((job, index) => ({
    name: job.name, head_sha: sha, status: 'completed', conclusion: 'success',
    app: {id: 15368}, check_suite: {id: run.check_suite_id}, id: 800 + index,
  }));
  const github = {rest: {actions: {
    listWorkflowRuns() {}, getWorkflowRun() {}, listJobsForWorkflowRunAttempt() {},
  }, checks: {listForRef() {}}}};
  github.paginate = async (method, request) => {
    if (method === github.rest.actions.listWorkflowRuns) {
      const index = Math.min(listIndex++, listResponses.length - 1);
      return listResponses[index];
    }
    if (method === github.rest.actions.listJobsForWorkflowRunAttempt) {
      calls.jobs++;
      assert.equal(request.run_id, 601);
      assert.equal(request.attempt_number, 1);
      return jobs;
    }
    if (method === github.rest.checks.listForRef) {
      calls.checks++;
      return checks;
    }
    throw new Error('unexpected pagination target');
  };
  github.rest.actions.getWorkflowRun = async () => {
    const index = Math.min(getIndex++, runResponses.length - 1);
    return {data: runResponses[index]};
  };
  return {github, calls, listCalls: () => listIndex, getCalls: () => getIndex};
}

function prPollOptions(github, overrides = {}) {
  let now = attemptTime;
  return {
    github, owner: 'alice', repo: 'meta-go', repository, branch, sha,
    baseSHA: mainSHA, prNumber: 91, trigger: async () => ({number: 91}),
    clock: () => now, sleep: async ms => { now += ms; }, timeoutMs: 60000,
    ...overrides,
  };
}

test('PR polling waits through queued zero-job window then validates exact checks', async () => {
  const queued = prRun();
  const complete = {...queued, status: 'completed', conclusion: 'success'};
  const {github, calls, listCalls, getCalls} = prPollingAPI([[], [queued], [complete]], [queued, complete]);
  let now = attemptTime;
  const waits = [];
  const result = await promotion.pullRequestRunAndWait(prPollOptions(github, {
    prNumber: undefined,
    clock: () => now, sleep: async ms => { waits.push(ms); now += ms; },
  }));
  assert.deepEqual(result, {run_id: 601, attempt: 1, pr_number: 91});
  assert.deepEqual(waits, [5000]);
  assert.equal(listCalls(), 3);
  assert.equal(getCalls(), 2);
  assert.equal(calls.jobs, 1);
  assert.equal(calls.checks, 1);
});

test('multiple new PR runs are ambiguous and fail closed', async () => {
  const a = prRun();
  const b = prRun({id: 602});
  const {github} = prPollingAPI([[], [a, b]], []);
  await assert.rejects(promotion.pullRequestRunAndWait(prPollOptions(github)), /ambiguous/);
});

test('an old PR run replay is not attributed to the current PR update', async () => {
  const prior = prRun({id: 550, created_at: new Date(attemptTime - 60000).toISOString()});
  const {github} = prPollingAPI([[prior], [prior], [prior]], []);
  let now = attemptTime;
  await assert.rejects(promotion.pullRequestRunAndWait(prPollOptions(github, {
    timeoutMs: 10000, clock: () => now, sleep: async ms => { now += ms; },
  })), /timed out/);
});

test('a run linked to the wrong PR/base is rejected', async () => {
  const wrong = prRun({pull_requests: [{
    number: 92, head: {ref: branch, sha}, base: {ref: 'main', sha: mainSHA},
  }]});
  const {github} = prPollingAPI([[], [wrong]], []);
  await assert.rejects(promotion.pullRequestRunAndWait(prPollOptions(github)), /does not match/);
});

test('an authoritative PR run created before the current event is rejected', async () => {
  const old = prRun({created_at: new Date(attemptTime - 1000).toISOString()});
  const {github} = prPollingAPI([[], [old]], []);
  await assert.rejects(promotion.pullRequestRunAndWait(prPollOptions(github)), /predates/);
});

test('a rejected App PR event trigger starts no CI polling', async () => {
  const {github, listCalls} = prPollingAPI([[]], []);
  await assert.rejects(promotion.pullRequestRunAndWait(prPollOptions(github, {
    trigger: async () => { throw new Error('App API denied PR update'); },
  })), /App API denied/);
  assert.equal(listCalls(), 1);
});

test('missing App configuration fails before any pull-request API request', async () => {
  let requests = 0;
  await assert.rejects(promotion.appPullRequestRequest({
    fetchImpl: async () => { requests++; }, owner: 'alice', repo: 'meta-go',
    method: 'POST', body: {base: 'main', head: 'dev', title: 't', body: 'b'},
  }), error => error.code === 'PROMOTION_APP_CONFIGURATION_REQUIRED');
  assert.equal(requests, 0);
});

test('App PR API helper restricts calls to scoped create and body refresh', async () => {
  const seen = [];
  const fetchImpl = async (url, init) => {
    seen.push({url, init});
    return {ok: true, json: async () => ({number: 91, state: 'open'})};
  };
  await promotion.appPullRequestRequest({
    fetchImpl, apiURL: 'https://api.example.test', token: 'app-token',
    owner: 'alice', repo: 'meta-go', method: 'POST',
    body: {base: 'main', head: 'dev', title: 'promote', body: 'frozen'},
  });
  await promotion.appPullRequestRequest({
    fetchImpl, apiURL: 'https://api.example.test', token: 'app-token',
    owner: 'alice', repo: 'meta-go', method: 'PATCH', pullNumber: 91, body: {body: 'refreshed'},
  });
  assert.equal(seen[0].url, 'https://api.example.test/repos/alice/meta-go/pulls');
  assert.equal(seen[1].url, 'https://api.example.test/repos/alice/meta-go/pulls/91');
  assert.equal(seen[0].init.headers.authorization, 'Bearer app-token');
  assert.deepEqual(JSON.parse(seen[1].init.body), {body: 'refreshed'});
  await assert.rejects(promotion.appPullRequestRequest({
    fetchImpl, token: 'app-token', owner: 'alice', repo: 'meta-go',
    method: 'PATCH', pullNumber: 91, body: {state: 'closed'},
  }), /outside the promotion PR/);
  await assert.rejects(promotion.appPullRequestRequest({
    fetchImpl, token: 'app-token', owner: 'alice', repo: 'meta-go',
    method: 'DELETE', pullNumber: 91, body: {},
  }), /restricted to create or update/);
});

const requiredJobs = [
  'gofmt', 'go vet', 'go test', 'go test -race',
  'Semantic conformance', 'CI policy', 'CI proof bundle',
];

function validProof() {
  const digest = 'd'.repeat(64);
  return {
    decision: 'PASS', repository, event: 'pull_request',
    pr_number: 91, run_id: 501, run_attempt: 1,
    base_ref: 'main', base_sha: mainSHA,
    head_ref: 'dev', head_sha: sha,
    ref: 'refs/pull/91/merge', event_ref: 'refs/pull/91/merge',
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
  proof.ref = 'refs/pull/91/merge';
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
    listJobsForWorkflowRunAttempt() {},
  };
}

function pullMocks(overrides, calls, pull) {
  return {
    list() {},
    get: async () => ({data: overrides.latestPull || pull}),
    merge: async request => {
      calls.merge++;
      calls.events?.push('merge');
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
    checks: {listForRef() {}},
    pulls: pullMocks(overrides, calls, pull),
    git: gitMocks(overrides),
    repos: {compareCommits: overrides.compareCommits || (async () => ({data: {
      status: 'ahead', ahead_by: 1, behind_by: 0,
      merge_base_commit: {sha: mainSHA},
    }}))},
  }};
  github.paginate = async (method, request) => {
    if (method === github.rest.actions.listJobsForWorkflowRunAttempt) {
      assert.equal(request.run_id, run.id);
      assert.equal(request.attempt_number, run.run_attempt);
      return requiredJobs.map(name => ({
        name, status: 'completed', conclusion: 'success', head_sha: run.head_sha,
        run_attempt: run.run_attempt,
        check_run_url: `https://api.github.com/repos/alice/meta-go/check-runs/${900 + requiredJobs.indexOf(name)}`,
      }));
    }
    if (method === github.rest.checks.listForRef) {
      return requiredJobs.map((name, id) => ({
        id: 900 + id, name, status: 'completed', conclusion: 'success',
        head_sha: run.head_sha, app: {id: 15368}, check_suite: {id: run.check_suite_id},
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
    event: 'pull_request', check_suite_id: 701,
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

test('dispatch proof and wrong PR/ref proof cannot authorize a merge', async t => {
  await t.test('dispatch event', async () => {
    const api = mergeAPI();
    await assert.rejects(promotion.mergeAuthorizedPromotion(mergeOptions(api, {
      event: 'workflow_dispatch', ref: 'refs/heads/dev', event_ref: 'refs/heads/dev',
    })), /not digest-bound/);
    assert.equal(api.calls.merge, 0);
  });
  await t.test('wrong PR number', async () => {
    const api = mergeAPI();
    await assert.rejects(promotion.mergeAuthorizedPromotion(mergeOptions(api, {pr_number: 92})), /not digest-bound/);
    assert.equal(api.calls.merge, 0);
  });
  await t.test('wrong merge ref', async () => {
    const api = mergeAPI();
    await assert.rejects(promotion.mergeAuthorizedPromotion(mergeOptions(api, {
      ref: 'refs/pull/92/merge', event_ref: 'refs/pull/92/merge',
    })), /not digest-bound/);
    assert.equal(api.calls.merge, 0);
  });
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

test('latest exact attempt is verified before final live ref reads and merge', async () => {
  const events = [];
  const api = mergeAPI({getRef: async ({ref}) => {
    events.push(ref);
    return {data: {object: {sha: ref === 'heads/dev' ? sha : mainSHA}}};
  }});
  api.calls.events = events;
  await promotion.mergeAuthorizedPromotion(mergeOptions(api));
  assert.deepEqual(events.slice(-3), ['heads/dev', 'heads/main', 'merge']);
});

test('a rerun after dispatch completion fails the current-attempt recheck', async () => {
  const api = mergeAPI({run: calls => candidateRun({
    id: 501, event: 'pull_request', check_suite_id: 701,
    status: 'completed', conclusion: 'success', run_attempt: calls === 1 ? 1 : 2,
  })});
  await assert.rejects(promotion.mergeAuthorizedPromotion(mergeOptions(api)), /attempt changed/);
  assert.equal(api.calls.merge, 0);
});

test('historical or changed head SHA cannot authorize the current PR', async () => {
  const api = mergeAPI({run: () => candidateRun({
    id: 501, event: 'pull_request', check_suite_id: 701,
    status: 'completed', conclusion: 'success', head_sha: '8'.repeat(40),
  })});
  await assert.rejects(promotion.mergeAuthorizedPromotion(mergeOptions(api)), /run or current attempt changed/);
  assert.equal(api.calls.merge, 0);
});

test('missing one of the seven exact successful jobs fails closed', async () => {
  const api = mergeAPI();
  const originalPaginate = api.github.paginate;
  api.github.paginate = async (method, request) => {
    const result = await originalPaginate(method, request);
    if (method === api.github.rest.actions.listJobsForWorkflowRunAttempt) return result.slice(1);
    return result;
  };
  await assert.rejects(promotion.mergeAuthorizedPromotion(mergeOptions(api)), /required exact-attempt PR job/);
  assert.equal(api.calls.merge, 0);
});

test('a successful check from another app or check suite is not accepted', async () => {
  const api = mergeAPI();
  const originalPaginate = api.github.paginate;
  api.github.paginate = async (method, request) => {
    const result = await originalPaginate(method, request);
    if (method === api.github.rest.checks.listForRef) {
      return result.map((check, index) => index === 0 ? {...check, app: {id: 57789}} : check);
    }
    return result;
  };
  await assert.rejects(promotion.mergeAuthorizedPromotion(mergeOptions(api)), /GitHub Actions app 15368/);
  assert.equal(api.calls.merge, 0);
});

test('attempt-specific job check URLs ignore older same-name checks in the same suite', async () => {
  const api = mergeAPI();
  const originalPaginate = api.github.paginate;
  api.github.paginate = async (method, request) => {
    const result = await originalPaginate(method, request);
    if (method === api.github.rest.checks.listForRef) {
      return [
        {...result[0], id: 700, conclusion: 'failure'},
        ...result,
      ];
    }
    return result;
  };
  await promotion.mergeAuthorizedPromotion(mergeOptions(api));
  assert.equal(api.calls.merge, 1);
});

test('missing, wrong-repository, and malformed job check URLs fail closed', async t => {
  for (const [label, url] of [
    ['missing', undefined],
    ['wrong repository', 'https://api.github.com/repos/alice/other/check-runs/900'],
    ['wrong API host', 'https://example.test/repos/alice/meta-go/check-runs/900'],
    ['malformed ID', 'https://api.github.com/repos/alice/meta-go/check-runs/00900'],
  ]) {
    await t.test(label, async () => {
      const api = mergeAPI();
      const originalPaginate = api.github.paginate;
      api.github.paginate = async (method, request) => {
        const result = await originalPaginate(method, request);
        if (method === api.github.rest.actions.listJobsForWorkflowRunAttempt) {
          return result.map((job, index) => index === 0 ? {...job, check_run_url: url} : job);
        }
        return result;
      };
      await assert.rejects(promotion.mergeAuthorizedPromotion(mergeOptions(api)), /required PR check/);
      assert.equal(api.calls.merge, 0);
    });
  }
});

test('workflow helper is checked out from trusted workflow_run github.sha', () => {
  const workflowPath = path.join(__dirname, '../../.github/workflows/ci-automatic-promotion.yml');
  const workflow = fs.readFileSync(workflowPath, 'utf8');
  assert.equal((workflow.match(/ref: \$\{\{ github\.sha \}\}/g) || []).length, 2);
  assert.match(workflow, /github\.sha is the default-branch workflow revision/);
  assert.doesNotMatch(workflow, /ref: \$\{\{ github\.event\.workflow_run\.head_sha \}\}/);
  assert.match(workflow, /workflow_run\.event == 'push'/);
  assert.match(workflow, /ci-automatic-promotion-\$\{\{ github\.event\.workflow_run\.event \}\}/);
  assert.match(workflow, /successfulRequiredJobs\(\s*github, owner, repo, run, process\.env\.GITHUB_API_URL/);
  assert.doesNotMatch(workflow, /listJobsForWorkflowRun\(/);
  assert.match(workflow, /actions\/create-github-app-token\@bcd2ba49218906704ab6c1aa796996da409d3eb1/);
  assert.match(workflow, /permission-pull-requests: write/);
  assert.match(workflow, /PROMOTION_APP_TOKEN: \$\{\{ steps\.promotion-app-token\.outputs\.token \}\}/);
  assert.match(workflow, /github-token: \$\{\{ github\.token \}\}/);
  assert.doesNotMatch(workflow, /createWorkflowDispatch/);
  assert.match(workflow, /PROMOTION_APP_CONFIGURATION_REQUIRED/);
  assert.match(workflow, /app_config_action:/);
  assert.match(workflow, /pull-requests: read/);
  assert.match(workflow, /apiURL: process\.env\.GITHUB_API_URL/);
  const ciWorkflow = fs.readFileSync(path.join(__dirname, '../../.github/workflows/ci.yml'), 'utf8');
  assert.match(ciWorkflow, /types: \[opened, synchronize, reopened, ready_for_review, edited\]/);
});
