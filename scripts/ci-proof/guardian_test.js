'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {
  PROTECTED_FILES,
  PROTECTED_PREFIXES,
  DEFAULT_BRANCH_CODE,
  HEAD_BINDING_STATUS,
  HEAD_BINDING_VERIFIED,
  LIVE_REF_CODE,
  PROMOTION_TOPOLOGY_CODE,
  CHECK_IDENTITY_CODE,
  PROTECTION_CODE,
  OBSERVER_ENVIRONMENT,
  ROOT_FAILURE_CODE,
  buildGuardianArtifact,
  classifyGuardianDecision,
  digestGuardianArtifact,
  inspectChangedFiles,
  kernelTreeDigest,
  revalidatePullRequest,
  validateGuardianArtifact,
  validateGuardianPullRequest,
  readLiveTopology,
  digestBranchProtection,
  observeBranchProtection,
  observeGuardianEnvironment,
  validateBranchProtectionSnapshot,
  validatePublicBranchSummary,
  validateGuardianEnvironment,
} = require('./guardian');

const sha = (letter) => letter.repeat(40);
const file = (filename, status = 'modified', previous_filename) => ({filename, status, ...(previous_filename ? {previous_filename} : {})});
const pull = (base = 'dev', fork = false) => ({
  number: 108,
  base: {ref: base, sha: sha('b'), repo: {full_name: 'owner/repo'}},
  head: {ref: 'agent/ci-workflow', sha: sha('a'), repo: {full_name: fork ? 'fork/repo' : 'owner/repo'}},
  changed_files: 1,
});

const expectedFixtureTuple = () => ({
  repository: 'owner/repo',
  pull_request_number: 108,
  action: 'synchronize',
  base_repo: 'owner/repo',
  base_ref: 'dev',
  base_sha: sha('d'),
  head_repo: 'owner/repo',
  head_ref: 'agent/ci-workflow',
  head_sha: sha('a'),
  default_branch: 'dev',
  workflow_ref: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev',
  workflow_sha: sha('d'),
  runtime_ref: 'refs/heads/dev',
  runtime_sha: sha('d'),
  event_ref: 'refs/heads/dev',
  run_id: 108,
  run_attempt: 1,
});

const liveFixture = () => ({
  refs: {dev_sha: sha('d'), main_sha: sha('b')},
  topology: {status: 'ahead', ahead_by: 1, behind_by: 0, merge_base_sha: sha('b')},
});

const mainProtectionFixture = () => {
  const contexts = ['CI guardian', 'CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'];
  const snapshot = {
    repository: 'owner/repo', branch: 'main', policy_sha256: '6'.repeat(64), event_ref: 'refs/heads/dev', checkout_ref: sha('d'), token_source: 'github_app_installation', app_installation_id: 42, app_slug: 'guardian', read_status: 'verified', exists: true, strict: true,
    required_checks: [...contexts].sort(), required_check_bindings: contexts.map((context) => ({context, app_id: 15368})).sort((left, right) => left.context < right.context ? -1 : left.context > right.context ? 1 : 0), enforce_admins: true, required_reviews: 0,
    dismiss_stale_reviews: false, require_last_push_approval: false, linear_history: true, allow_force_pushes: false, allow_deletions: false, required_signatures: false, required_conversation_resolution: false, block_creations: false, lock_branch: false, allow_fork_syncing: false, restrictions: null, missing_reason: '', base_sha: sha('b'), head_sha: sha('d'), run_id: 108, run_attempt: 1, workflow_sha: sha('d'), digest_sha256: '',
  };
  snapshot.digest_sha256 = digestBranchProtection(snapshot);
  return snapshot;
};

async function rejectsRoot(operation) {
  await assert.rejects(operation, (error) => error && error.code === ROOT_FAILURE_CODE);
}

async function testPaginationAndNonKernelPass() {
  const pages = [Array.from({length: 100}, (_, index) => file(index === 0 ? 'docs/readme.md' : `docs/page-${index}.md`, index === 0 ? 'added' : 'modified')), [file('docs/final.md', 'modified')]];
  const calls = [];
  const result = await inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 101,
    listFiles: async (options) => {
      calls.push(options);
      return {status: 200, data: pages[options.page - 1] || []};
    },
  });
  assert.equal(result.decision, 'PASS');
  assert.equal(result.files.length, 101);
  assert.deepEqual(calls.map((call) => call.page), [1, 2]);
  assert.equal(calls[1].per_page, 100);
}

async function testKernelStatusesAndRenames() {
  for (const status of ['added', 'modified', 'removed']) {
    const result = await inspectChangedFiles({
      owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 1,
      listFiles: async () => ({status: 200, data: [file('.github/workflows/ci.yml', status)]}),
    });
    assert.equal(result.code, ROOT_FAILURE_CODE);
    assert.equal(result.decision, 'FAIL_CLOSED');
  }
  const renamed = await inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 1,
    listFiles: async () => ({status: 200, data: [file('docs/new.md', 'renamed', '.github/ci-governance.json')]}),
  });
  assert.equal(renamed.code, ROOT_FAILURE_CODE);
  const inert = await inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 1,
    listFiles: async () => ({status: 200, data: [{...file('.github/workflows/ci.yml'), patch: '# name: CI guardian'}]}),
  });
  assert.equal(inert.code, ROOT_FAILURE_CODE);
}

async function testForkAndMalformedAPI() {
  validateGuardianPullRequest(pull('dev', true));
  assert.throws(() => validateGuardianPullRequest(pull('integration')), (error) => error && error.code === ROOT_FAILURE_CODE);
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 1, listFiles: async () => ({status: 200, data: [{status: 'modified'}]})}));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 0, listFiles: async () => ({status: 200, data: null})}));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 0, listFiles: async () => undefined}));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 0, listFiles: async () => { throw new Error('forbidden'); }}));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 1, listFiles: async () => ({status: 200, data: [file('.github/workflows/ci.yml', 'renamed')]})}));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'other/repo', pullNumber: 108, expectedCount: 0, listFiles: async () => ({status: 200, data: []})}));
  for (const malformed of [pull('release'), {...pull(), base: {...pull().base, sha: 'short'}}, {...pull(), head: null}]) {
    assert.throws(() => validateGuardianPullRequest(malformed), (error) => error.code === ROOT_FAILURE_CODE);
  }
}

async function testStaleRaceAndArtifactDigest() {
  const eventPull = pull('dev');
  eventPull.base.sha = sha('d');
  await rejectsRoot(() => revalidatePullRequest({
    owner: 'owner', repo: 'repo', pullNumber: 108, eventPull,
    getPull: async () => ({status: 200, data: {...eventPull, head: {...eventPull.head, sha: sha('c')}}}),
  }));
  await rejectsRoot(() => revalidatePullRequest({
    owner: 'owner', repo: 'repo', pullNumber: 108, eventPull,
    getPull: async () => ({status: 200, data: {...eventPull, base: {...eventPull.base, ref: 'main'}}}),
  }));
  const matching = await revalidatePullRequest({
    owner: 'owner', repo: 'repo', pullNumber: 108, eventPull,
    getPull: async () => ({status: 200, data: eventPull}),
  });
  assert.equal(matching.head.sha, eventPull.head.sha);
  const artifact = buildGuardianArtifact({
    pull: eventPull, repository: 'owner/repo', action: 'synchronize', defaultBranch: 'dev',
    workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'),
    runtimeRef: 'refs/heads/dev', runtimeSha: sha('d'), runId: 108, runAttempt: 1, eventRef: 'refs/heads/dev',
    liveBefore: liveFixture(), liveAfter: liveFixture(),
    result: {decision: 'PASS', code: null, reason: null, files: [file('docs/a.md', 'modified')], kernelPaths: []},
  });
  const expected = expectedFixtureTuple();
  assert.throws(() => validateGuardianArtifact(artifact), (error) => error && error.code === ROOT_FAILURE_CODE);
  validateGuardianArtifact(artifact, expected);
  artifact.changed_files[0].filename = 'docs/tampered.md';
  assert.throws(() => validateGuardianArtifact(artifact, expected), (error) => error && error.code === ROOT_FAILURE_CODE);
  const freshArtifact = () => buildGuardianArtifact({
    pull: eventPull, repository: 'owner/repo', action: 'synchronize', defaultBranch: 'dev',
    workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'),
    runtimeRef: 'refs/heads/dev', runtimeSha: sha('d'), runId: 108, runAttempt: 1, eventRef: 'refs/heads/dev',
    liveBefore: liveFixture(), liveAfter: liveFixture(),
    result: {decision: 'PASS', code: null, reason: null, files: [file('docs/a.md', 'modified')], kernelPaths: []},
  });
  for (const mutate of [
    (candidate) => { candidate.repository = ''; },
    (candidate) => { candidate.base_sha = 'short'; },
    (candidate) => { candidate.workflow_ref = 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/main'; },
    (candidate) => { candidate.runtime_sha = sha('e'); },
    (candidate) => { candidate.code = ROOT_FAILURE_CODE; },
    (candidate) => { candidate.action = 'closed'; },
    (candidate) => { candidate.changed_files.push(candidate.changed_files[0]); },
    (candidate) => { candidate.kernel_paths = ['scripts/ci-proof/a', 'scripts/ci-proof/a']; },
    (candidate) => { candidate.head_binding_status = HEAD_BINDING_STATUS; },
  ]) {
    const candidate = freshArtifact();
    const expected = expectedFixtureTuple();
    mutate(candidate);
    assert.throws(() => validateGuardianArtifact(candidate, expected), (error) => error && [ROOT_FAILURE_CODE, 'CI-GUARDIAN-CHECK-IDENTITY-001'].includes(error.code));
  }
  const rehash = (candidate) => {
    candidate.bundle_sha256 = digestGuardianArtifact(candidate);
    return candidate;
  };
  for (const mutate of [
    (candidate) => { candidate.default_branch = 'main'; },
    (candidate) => { candidate.event_ref = 'refs/heads/main'; },
    (candidate) => { candidate.runtime_ref = 'refs/heads/main'; },
    (candidate) => { candidate.base_ref = 'main'; candidate.head_ref = 'agent/ci-workflow'; },
    (candidate) => { candidate.pull_request_number = 999; },
    (candidate) => { candidate.head_sha = sha('f'); },
    (candidate) => { candidate.run_id = 999; },
    (candidate) => { candidate.run_attempt = 2; },
    (candidate) => { candidate.changed_files_count = 2; },
    (candidate) => { candidate.changed_files = [file('.github/workflows/ci.yml', 'modified')]; },
  ]) {
    const candidate = rehash(freshArtifact());
    const expected = expectedFixtureTuple();
    mutate(candidate);
    rehash(candidate);
    assert.throws(() => validateGuardianArtifact(candidate, expected), (error) => error && error.code === ROOT_FAILURE_CODE);
  }
  const featureBaseRace = rehash(freshArtifact());
  featureBaseRace.base_sha = sha('e');
  rehash(featureBaseRace);
  const featureBaseRaceExpected = expectedFixtureTuple();
  featureBaseRaceExpected.base_sha = sha('e');
  assert.throws(() => validateGuardianArtifact(featureBaseRace, featureBaseRaceExpected), (error) => error && error.code === ROOT_FAILURE_CODE);
  const omittedKernel = rehash(freshArtifact());
  const omittedKernelExpected = expectedFixtureTuple();
  omittedKernel.changed_files = [file('.github/workflows/ci.yml', 'modified')];
  rehash(omittedKernel);
  assert.throws(() => validateGuardianArtifact(omittedKernel, omittedKernelExpected), (error) => error && error.code === ROOT_FAILURE_CODE);
  const promotion = pull('main');
  promotion.head.ref = 'dev';
  promotion.head.sha = sha('d');
  const promotionArtifact = buildGuardianArtifact({
    pull: promotion, repository: 'owner/repo', action: 'synchronize', defaultBranch: 'dev',
    workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'), runtimeRef: 'refs/heads/dev', runtimeSha: sha('d'), runId: 108, runAttempt: 1, eventRef: 'refs/heads/dev', liveBefore: liveFixture(), liveAfter: liveFixture(), checkName: 'CI guardian',
    result: {decision: 'PASS', code: null, reason: null, files: [file('docs/a.md')], kernelPaths: []}, branchProtection: mainProtectionFixture(), devBranchProtection: {...mainProtectionFixture(), branch: 'dev', required_checks: ['CI guardian shadow', 'CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'], required_check_bindings: ['CI guardian shadow', 'CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'].map((context) => ({context, app_id: 15368})).sort((left, right) => left.context < right.context ? -1 : left.context > right.context ? 1 : 0)}, observerEnvironment: OBSERVER_ENVIRONMENT, observerEnvironmentSnapshot: {repository: 'owner/repo', name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: ['branch_policy'], wait_timer: 0, reviewers: [], token_source: 'github.token', read_status: 'verified', missing_reason: '', run_id: 108, run_attempt: 1, workflow_sha: sha('d'), digest_sha256: ''},
  });
  promotionArtifact.dev_branch_protection.digest_sha256 = digestBranchProtection(promotionArtifact.dev_branch_protection);
  promotionArtifact.observer_environment_snapshot.digest_sha256 = require('./guardian').digestGuardianEnvironment(promotionArtifact.observer_environment_snapshot);
  promotionArtifact.observer_environment_digest = promotionArtifact.observer_environment_snapshot.digest_sha256;
  promotionArtifact.bundle_sha256 = digestGuardianArtifact(promotionArtifact);
  const promotionExpected = expectedFixtureTuple();
  promotionExpected.base_ref = 'main';
  promotionExpected.base_sha = sha('b');
  promotionExpected.head_ref = 'dev';
  promotionExpected.head_sha = sha('d');
  promotionExpected.workflow_sha = sha('d');
  promotionExpected.runtime_sha = sha('d');
  promotionExpected.default_branch = 'dev';
  promotionExpected.workflow_ref = 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev';
  promotionArtifact.head_binding_status = HEAD_BINDING_STATUS;
  promotionArtifact.bundle_sha256 = digestGuardianArtifact(promotionArtifact);
  assert.throws(() => validateGuardianArtifact(promotionArtifact, promotionExpected), (error) => error && error.code === CHECK_IDENTITY_CODE);
}

async function testProtectionObserverContracts() {
  assert.equal(validatePublicBranchSummary({protected: true, required_status_checks: {contexts: ['gofmt', 'CI guardian', 'go test', 'CI policy', 'go vet', 'go test -race', 'Semantic conformance'], checks: [{context: 'gofmt', app_id: 15368}, {context: 'CI guardian', app_id: 15368}, {context: 'go test', app_id: 15368}, {context: 'CI policy', app_id: 15368}, {context: 'go vet', app_id: 15368}, {context: 'go test -race', app_id: 15368}, {context: 'Semantic conformance', app_id: 15368}]}}), true);
  assert.equal(validatePublicBranchSummary({protected: true, required_status_checks: {contexts: ['gofmt'], checks: [{context: 'gofmt', app_id: 15368}]}}), false);
  const unavailable = await observeBranchProtection({repository: 'owner/repo', policySHA: '6'.repeat(64), eventRef: 'refs/heads/dev', checkoutRef: sha('d'), baseSHA: sha('b'), headSHA: sha('d'), runId: 108, runAttempt: 1, workflowSHA: sha('d'), tokenSource: 'github.token', getProtection: async () => ({status: 403})});
  assert.equal(unavailable.read_status, 'unavailable');
  assert.throws(() => validateBranchProtectionSnapshot(unavailable, {requireVerified: true}), (error) => error && error.code === PROTECTION_CODE);
  const protectionData = {required_status_checks: {strict: true, contexts: ['CI guardian', 'CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'], checks: [{context: 'CI guardian', app_id: 15368}, {context: 'CI policy', app_id: 15368}, {context: 'Semantic conformance', app_id: 15368}, {context: 'go test', app_id: 15368}, {context: 'go test -race', app_id: 15368}, {context: 'go vet', app_id: 15368}, {context: 'gofmt', app_id: 15368}]}, enforce_admins: {enabled: true}, required_linear_history: {enabled: true}, allow_force_pushes: {enabled: false}, allow_deletions: {enabled: false}, required_signatures: {enabled: false}, required_conversation_resolution: {enabled: false}, block_creations: {enabled: false}, lock_branch: {enabled: false}, allow_fork_syncing: {enabled: false}};
  const protectionArgs = {repository: 'owner/repo', branch: 'main', expectedContexts: ['CI guardian', 'CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'], policySHA: '6'.repeat(64), eventRef: 'refs/heads/dev', checkoutRef: sha('d'), baseSHA: sha('b'), headSHA: sha('d'), runId: 108, runAttempt: 1, workflowSHA: sha('d'), tokenSource: 'github_app_installation', appInstallationId: 42, appSlug: 'guardian'};
  const verified = await observeBranchProtection({...protectionArgs, getProtection: async () => ({status: 200, data: protectionData})});
  validateBranchProtectionSnapshot(verified, {requireVerified: true});
  const explicitNull = await observeBranchProtection({...protectionArgs, getProtection: async () => ({status: 200, data: {...protectionData, required_pull_request_reviews: null, restrictions: null}})});
  validateBranchProtectionSnapshot(explicitNull, {requireVerified: true});
  const wrongApp = {...verified, required_check_bindings: verified.required_check_bindings.map((binding, index) => index === 0 ? {...binding, app_id: 1} : binding)};
  wrongApp.digest_sha256 = digestBranchProtection(wrongApp);
  assert.throws(() => validateBranchProtectionSnapshot(wrongApp, {requireVerified: true}), (error) => error && error.code === PROTECTION_CODE);
  for (const field of ['required_signatures', 'required_conversation_resolution', 'block_creations', 'lock_branch', 'allow_fork_syncing']) {
    const omitted = {...verified};
    delete omitted[field];
    omitted.digest_sha256 = digestBranchProtection(omitted);
    assert.throws(() => validateBranchProtectionSnapshot(omitted, {requireVerified: true}), (error) => error && error.code === PROTECTION_CODE);
  }
  const sixOnlyDev = {...verified, branch: 'dev', required_checks: ['CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'], required_check_bindings: verified.required_check_bindings.filter((binding) => binding.context !== 'CI guardian')};
  sixOnlyDev.digest_sha256 = digestBranchProtection(sixOnlyDev);
  assert.throws(() => validateBranchProtectionSnapshot(sixOnlyDev, {requireVerified: true, expectedBranch: 'dev'}), (error) => error && error.code === PROTECTION_CODE);
  const environment = await observeGuardianEnvironment({repository: 'owner/repo', tokenSource: 'github.token', runId: 108, runAttempt: 1, workflowSHA: sha('d'), getEnvironment: async () => ({status: 200, data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, wait_timer: 0, reviewers: [], protection_rules: [{type: 'branch_policy'}]}})});
  assert.equal(environment.read_status, 'verified');
  const environmentTamper = {...environment, deployment_branch_policy: {...environment.deployment_branch_policy, custom_branch_policies: true}};
  environmentTamper.digest_sha256 = require('./guardian').digestGuardianEnvironment(environmentTamper);
  assert.throws(() => validateGuardianEnvironment(environmentTamper, {requireVerified: true}), (error) => error && error.code === PROTECTION_CODE);
  const environmentMissing = await observeGuardianEnvironment({repository: 'owner/repo', tokenSource: 'github.token', runId: 108, runAttempt: 1, workflowSHA: sha('d'), getEnvironment: async () => ({status: 200, data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}}})});
  assert.equal(environmentMissing.read_status, 'unavailable');
}

async function testCanonicalOrdering() {
  const eventPull = pull('dev');
  eventPull.base.sha = sha('d');
  const orderingArtifact = buildGuardianArtifact({
    pull: eventPull, repository: 'owner/repo', action: 'synchronize', defaultBranch: 'dev',
    workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'),
    runtimeRef: 'refs/heads/dev', runtimeSha: sha('d'), runId: 108, runAttempt: 1, eventRef: 'refs/heads/dev',
    liveBefore: liveFixture(), liveAfter: liveFixture(),
    result: {
      decision: 'PASS', code: null, reason: null,
      files: [file('cmd/gooo/analyze_test.go'), file('cmd/gooo/analyze.go')], kernelPaths: [],
    },
  });
  assert.deepEqual(orderingArtifact.changed_files.map((candidate) => candidate.filename), [
    'cmd/gooo/analyze.go', 'cmd/gooo/analyze_test.go',
  ]);
  validateGuardianArtifact(orderingArtifact, expectedFixtureTuple());

  const unsortedArtifact = buildGuardianArtifact({
    pull: eventPull, repository: 'owner/repo', action: 'synchronize', defaultBranch: 'dev',
    workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'),
    runtimeRef: 'refs/heads/dev', runtimeSha: sha('d'), runId: 108, runAttempt: 1, eventRef: 'refs/heads/dev',
    liveBefore: liveFixture(), liveAfter: liveFixture(),
    result: {decision: 'PASS', code: null, reason: null, files: [file('docs/a.md')], kernelPaths: []},
  });
  unsortedArtifact.changed_files = [file('cmd/gooo/analyze_test.go'), file('cmd/gooo/analyze.go')];
  unsortedArtifact.changed_files_count = 2;
  unsortedArtifact.bundle_sha256 = digestGuardianArtifact(unsortedArtifact);
  assert.throws(() => validateGuardianArtifact(unsortedArtifact, expectedFixtureTuple()), (error) => error && error.code === ROOT_FAILURE_CODE);

  const treeEntries = [
    {path: 'scripts/ci-proof/analyze_test.go', type: 'blob', mode: '100644', sha: sha('1')},
    {path: 'scripts/ci-proof/analyze.go', type: 'blob', mode: '100644', sha: sha('2')},
  ];
  const treeDigest = async (entries) => kernelTreeDigest({
    owner: 'owner', repo: 'repo', ref: sha('d'),
    getCommit: async () => ({status: 200, data: {commit: {tree: {sha: sha('a')}}}}),
    getTree: async () => ({status: 200, data: {sha: sha('a'), truncated: false, tree: entries}}),
  });
  assert.equal(await treeDigest(treeEntries), await treeDigest([...treeEntries].reverse()));
}

async function testPromotionAndKernelDigests() {
  const promotionPull = pull('main');
  promotionPull.head.ref = 'dev';
  promotionPull.head.repo.full_name = 'owner/repo';
  promotionPull.head.sha = sha('d');
  const exact = classifyGuardianDecision({
    pull: promotionPull, repository: 'owner/repo', defaultBranch: 'dev', eventRef: 'refs/heads/dev', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'), runtimeSha: sha('d'),
    liveBefore: liveFixture(), liveAfter: liveFixture(), checkName: 'CI guardian',
    result: {decision: 'PASS', code: null, reason: null, kernelPaths: []},
  });
  assert.equal(exact.decision, 'PASS');
  const stale = classifyGuardianDecision({
    pull: promotionPull, repository: 'owner/repo', defaultBranch: 'dev', eventRef: 'refs/heads/dev', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('e'), runtimeSha: sha('e'),
    liveBefore: liveFixture(), liveAfter: liveFixture(), checkName: 'CI guardian',
    result: {decision: 'PASS', code: null, reason: null, kernelPaths: []},
  });
  assert.equal(stale.code, 'CI-GUARDIAN-PROMOTION-TOPOLOGY-001');
  const wrongDefault = classifyGuardianDecision({
    pull: promotionPull, repository: 'owner/repo', defaultBranch: 'main', eventRef: 'refs/heads/main', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/main', workflowSha: sha('d'), runtimeSha: sha('d'),
    liveBefore: liveFixture(), liveAfter: liveFixture(), checkName: 'CI guardian',
    result: {decision: 'PASS', code: null, reason: null, kernelPaths: []},
  });
  assert.equal(wrongDefault.code, DEFAULT_BRANCH_CODE);
  const bootstrapPull = pull('dev');
  bootstrapPull.base.sha = sha('d');
  const bootstrap = classifyGuardianDecision({
    pull: bootstrapPull, repository: 'owner/repo', defaultBranch: 'main', eventRef: 'refs/heads/main', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/main', workflowSha: sha('d'), runtimeSha: sha('d'),
    liveBefore: liveFixture(), liveAfter: liveFixture(), checkName: 'CI guardian shadow',
    result: {decision: 'PASS', code: null, reason: null, kernelPaths: []},
  });
  assert.equal(bootstrap.code, DEFAULT_BRANCH_CODE);
  const kernel = classifyGuardianDecision({
    pull: promotionPull, repository: 'owner/repo', defaultBranch: 'dev', eventRef: 'refs/heads/dev', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'), runtimeSha: sha('d'),
    liveBefore: liveFixture(), liveAfter: liveFixture(), checkName: 'CI guardian',
    kernelBeforeDigest: 'sha256:' + '1'.repeat(64), kernelAfterDigest: 'sha256:' + '2'.repeat(64),
    result: {decision: 'FAIL_CLOSED', code: ROOT_FAILURE_CODE, reason: 'kernel changed', kernelPaths: ['.github/workflows/ci.yml']},
  });
  assert.equal(kernel.decision, 'PASS');
  assert.equal(kernel.kernelBeforeDigest, 'sha256:' + '1'.repeat(64));
  assert.equal(DEFAULT_BRANCH_CODE, 'CI-GUARDIAN-DEFAULT-BRANCH-001');
  const missingLive = classifyGuardianDecision({
    pull: promotionPull, repository: 'owner/repo', defaultBranch: 'dev', eventRef: 'refs/heads/dev', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'), runtimeSha: sha('d'), checkName: 'CI guardian',
    result: {decision: 'PASS', code: null, reason: null, kernelPaths: []},
  });
  assert.equal(missingLive.code, 'CI-GUARDIAN-LIVE-REF-001');
  const featurePull = pull('dev');
  featurePull.base.sha = sha('d');
  const identicalLive = {refs: {dev_sha: sha('d'), main_sha: sha('d')}, topology: {status: 'identical', ahead_by: 0, behind_by: 0, merge_base_sha: sha('d')}};
  const stableFeature = classifyGuardianDecision({
    pull: featurePull, repository: 'owner/repo', defaultBranch: 'dev', eventRef: 'refs/heads/dev', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'), runtimeSha: sha('d'), checkName: 'CI guardian shadow', liveBefore: identicalLive, liveAfter: identicalLive,
    result: {decision: 'PASS', code: null, reason: null, kernelPaths: []},
  });
  assert.equal(stableFeature.decision, 'PASS');
  const identicalPromotion = classifyGuardianDecision({
    pull: promotionPull, repository: 'owner/repo', defaultBranch: 'dev', eventRef: 'refs/heads/dev', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'), runtimeSha: sha('d'), checkName: 'CI guardian', liveBefore: identicalLive, liveAfter: identicalLive,
    result: {decision: 'PASS', code: null, reason: null, kernelPaths: []},
  });
  assert.equal(identicalPromotion.code, 'CI-GUARDIAN-PROMOTION-TOPOLOGY-001');
  const tree = await kernelTreeDigest({
    owner: 'owner', repo: 'repo', ref: sha('d'),
    getCommit: async () => ({status: 200, data: {commit: {tree: {sha: sha('a')}}}}),
    getTree: async () => ({status: 200, data: {sha: sha('a'), truncated: false, tree: [
      {path: '.github/workflows/ci.yml', type: 'blob', mode: '100644', sha: sha('1')},
      {path: 'docs/readme.md', type: 'blob', mode: '100644', sha: sha('2')},
    ]}}),
  });
  assert.match(tree, /^sha256:[0-9a-f]{64}$/);
  const treeResponse = () => ({status: 200, data: {sha: sha('a'), truncated: false, tree: [{path: '.github/workflows/ci.yml', type: 'blob', mode: '100644', sha: sha('1')}]}});
  for (const mutate of [
    (response) => { response.data.sha = sha('b'); },
    (response) => { response.data.tree[0].sha = 'short'; },
    (response) => { response.data.tree[0].type = 'commit'; },
    (response) => { response.data.tree.push({...response.data.tree[0]}); },
  ]) {
    const response = treeResponse();
    mutate(response);
    await rejectsRoot(() => kernelTreeDigest({
      owner: 'owner', repo: 'repo', ref: sha('d'),
      getCommit: async () => ({status: 200, data: {commit: {tree: {sha: sha('a')}}}}),
      getTree: async () => response,
    }));
  }
}

async function testPaginationLimit() {
  let calls = 0;
  await rejectsRoot(() => inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 100000,
    listFiles: async () => {
      calls += 1;
      return {status: 200, data: Array.from({length: 100}, (_, index) => file(`docs/page-${calls}-${index}.md`))};
    },
  }));
  assert.equal(calls, 1000);
}

async function testLiveRefsAndRouteIdentity() {
  const topology = {status: 'ahead', ahead_by: 2, behind_by: 0, merge_base_commit: {sha: sha('b')}};
  const observed = await readLiveTopology({
    owner: 'owner', repo: 'repo',
    getRef: async ({ref}) => ({status: 200, data: {object: {sha: ref.endsWith('dev') ? sha('d') : sha('b')}}}),
    compareCommits: async () => ({status: 200, data: topology}),
  });
  assert.equal(observed.refs.dev_sha, sha('d'));
  await assert.rejects(() => readLiveTopology({owner: 'owner', repo: 'repo', getRef: async () => ({status: 403}), compareCommits: async () => ({status: 200, data: topology})}), (error) => error && error.code === LIVE_REF_CODE);
  const feature = pull('dev');
  feature.base.sha = sha('d');
  const withoutSnapshots = classifyGuardianDecision({pull: feature, repository: 'owner/repo', defaultBranch: 'dev', eventRef: 'refs/heads/dev', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'), runtimeSha: sha('d'), checkName: 'CI guardian shadow', result: {decision: 'PASS', code: null, reason: null, kernelPaths: []}});
  assert.equal(withoutSnapshots.code, LIVE_REF_CODE);
}

function testWorkflowIsReadOnlyAndBasePinned() {
  const workflow = fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'workflows', 'ci-guardian.yml'), 'utf8');
  const ciWorkflow = fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'workflows', 'ci.yml'), 'utf8');
  assert.match(workflow, /^name: CI guardian/m);
  assert.match(workflow, /pull_request_target:/);
  assert.match(workflow, /environment: \$\{\{ github\.base_ref == 'main' && 'guardian-observer'/);
  assert.match(workflow, /actions\/create-github-app-token@bcd2ba49218906704ab6c1aa796996da409d3eb1/);
  assert.match(workflow, /permission-administration: read/);
  assert.match(workflow, /GUARDIAN_APP_PRIVATE_KEY/);
  assert.match(workflow, /getBranchProtection/);
  assert.match(ciWorkflow, /token_source: 'not_observed',\s+app_installation_id: 0,\s+app_slug: '',\s+read_status: 'unavailable'/);
  assert.match(workflow, /- dev\n      - main/);
  assert.doesNotMatch(workflow, /- integration/);
  assert.match(workflow, /name: CI guardian/);
  assert.match(workflow, /ref: \$\{\{ github\.workflow_sha \}\}/);
  assert.doesNotMatch(workflow, /ref: \$\{\{ github\.event\.pull_request\.base\.sha \}\}/);
  assert.match(workflow, /persist-credentials: false/);
  assert.match(workflow, /actions\/checkout@11d5960a326750d5838078e36cf38b85af677262/);
  assert.match(workflow, /actions\/github-script@f28e40c7f34bde8b3046d885e986cb6290c5673b/);
  assert.match(workflow, /listFiles/);
  assert.match(workflow, /github\.rest\.pulls\.get/);
  assert.match(workflow, /github\.workflow_ref/);
  assert.match(workflow, /github\.workflow_sha/);
  assert.match(workflow, /github\.sha/);
  assert.match(workflow, /actions\/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02/);
  assert.match(workflow, /if: \$\{\{ always\(\) \}\}/);
  assert.match(workflow, /ci-guardian\.json/);
  assert.match(workflow, /pull_request_number: observedPull && observedPull\.number/);
  assert.match(workflow, /runtime_sha: runtimeSha/);
  const writeIndex = workflow.indexOf('writeFileSync');
  const validateIndex = workflow.indexOf('guardian.validateGuardianArtifact(artifact,');
  const setFailedIndex = workflow.indexOf('core.setFailed');
  assert(writeIndex >= 0 && writeIndex < validateIndex && validateIndex < setFailedIndex, 'guardian validates the written artifact before reporting failure');
  assert.match(workflow, /head_binding=\$\{artifact\.head_binding_status\}/);
  assert.doesNotMatch(workflow, /github\.event\.pull_request\.head\.sha/);
  assert.doesNotMatch(workflow, /refs\/pull\//);
  assert.doesNotMatch(workflow, /BRANCH_PROTECTION_TOKEN: \$\{\{ secrets\.BRANCH_PROTECTION_TOKEN \}\}/);
  assert.doesNotMatch(workflow, /contents: write|pull-requests: write/);
  assert.doesNotMatch(workflow, /^\s+run:/m);
  assert.doesNotMatch(workflow, /^\s+pull_request:/m);
  assert.doesNotMatch(workflow, /agent\/ci-workflow/);
}

function testHeadBindingIsExplicitlyShadowOnly() {
  assert.equal(HEAD_BINDING_STATUS, 'CI-GUARDIAN-HEAD-BINDING-UNVERIFIED');
  assert.equal(HEAD_BINDING_VERIFIED, 'verified');
}

function testKernelSetIsMonotonic() {
  for (const path of ['.github/ci-governance.json', '.github/agent-scope-table.md', '.github/branch-policy.md', '.github/conformance-plan.md', 'go.mod', 'go.sum']) {
    assert(PROTECTED_FILES.has(path), `missing protected file ${path}`);
  }
  for (const prefix of ['.github/workflows/', 'scripts/ci-proof/', 'scripts/ci-evidence/', 'scripts/verify/', 'internal/verify/']) {
    assert(PROTECTED_PREFIXES.includes(prefix), `missing protected prefix ${prefix}`);
  }
}

(async () => {
  await testPaginationAndNonKernelPass();
  await testKernelStatusesAndRenames();
  await testForkAndMalformedAPI();
  await testStaleRaceAndArtifactDigest();
  await testCanonicalOrdering();
  await testPromotionAndKernelDigests();
  await testProtectionObserverContracts();
  await testPaginationLimit();
  testWorkflowIsReadOnlyAndBasePinned();
  testHeadBindingIsExplicitlyShadowOnly();
  testKernelSetIsMonotonic();
  await testLiveRefsAndRouteIdentity();
  console.log('guardian tests passed');
})().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
