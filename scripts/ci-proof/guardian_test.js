'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const foundationAuthorization = require('./foundation_authorization');
const {
  PROTECTED_FILES,
  PROTECTED_PREFIXES,
  DEFAULT_BRANCH_CODE,
  HEAD_BINDING_STATUS,
  HEAD_BINDING_VERIFIED,
  LIVE_REF_CODE,
  PROMOTION_TOPOLOGY_CODE,
  INSTALLATION_SCOPE_CODE,
  INSTALLATION_SCOPE_REPOSITORY,
  CHECK_IDENTITY_CODE,
  PROTECTION_CODE,
  OBSERVER_ENVIRONMENT,
  ROOT_FAILURE_CODE,
  buildGuardianArtifact,
  classifyGuardianDecision,
  digestGuardianArtifact,
  validateKernelDigestAttestation,
  inspectChangedFiles,
  readGitChangedPaths,
  observeGraphQLChangedFiles,
  kernelTreeDigest,
  revalidatePullRequest,
  validateGuardianArtifact,
  validateGuardianPullRequest,
  readLiveTopology,
  digestBranchProtection,
  observeBranchProtection,
  observeGuardianEnvironment,
  observeInstallationRepositoryScope,
  emptyInstallationRepositoryScope,
  digestInstallationRepositoryScope,
  OBSERVER_FRESHNESS_WINDOW_MS,
  validateBranchProtectionSnapshot,
  validatePublicBranchSummary,
  validateGuardianEnvironment,
  validateInstallationRepositoryScope,
  validObserverFreshness,
} = require('./guardian');

const sha = (letter) => letter.repeat(40);
const observerNow = new Date('2026-08-14T00:00:00.000Z');
const observerDateHeader = 'Thu, 13 Aug 2026 23:55:00 GMT';
const observedAt = '2026-08-13T23:55:00.000Z';
const validUntil = '2026-08-14T00:05:00.000Z';
const githubDateResponse = (data, status = 200) => ({status, headers: {date: observerDateHeader}, data});
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
    dismiss_stale_reviews: false, require_last_push_approval: false, linear_history: true, allow_force_pushes: false, allow_deletions: false, required_signatures: false, required_conversation_resolution: false, block_creations: false, lock_branch: false, allow_fork_syncing: false, restrictions: null, missing_reason: '', base_sha: sha('b'), head_sha: sha('d'), run_id: 108, run_attempt: 1, workflow_sha: sha('d'), observed_at: observedAt, valid_until: validUntil, digest_sha256: '',
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

async function testBoundChangedPathAuthority() {
  const exactFiles = readGitChangedPaths({
    baseSHA: sha('b'),
    headSHA: sha('a'),
    execute: (args) => {
      assert.deepEqual(args, ['diff', '--name-status', '-z', '--find-renames', '--find-copies', '--no-ext-diff', `${sha('b')}...${sha('a')}`]);
      return Buffer.from(`M\0docs/readme.md\0R100\0docs/old.md\0docs/new.md\0`);
    },
  });
  assert.deepEqual(exactFiles, [
    file('docs/new.md', 'renamed', 'docs/old.md'),
    file('docs/readme.md', 'modified'),
  ]);
  const result = await inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 2,
    exactFiles,
    listFiles: async () => ({status: 200, data: [file('docs/new.md', 'renamed', 'docs/old.md')]}),
  });
  assert.equal(result.decision, 'PASS');
  assert.equal(result.changedFilesSource, 'git-diff');
  assert.equal(result.apiChangedFilesCount, 1);
  assert.equal(result.apiChangedFilesExpectedCount, 2);
  assert.equal(result.apiChangedFilesComplete, false);
  await rejectsRoot(() => inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 1,
    exactFiles,
    listFiles: async () => ({status: 200, data: []}),
  }));
  assert.equal(await observeGraphQLChangedFiles({
    owner: 'owner', repo: 'repo', pullNumber: 108,
    graphql: async () => ({repository: {pullRequest: {changedFiles: 2}}}),
  }), 2);
  await rejectsRoot(() => observeGraphQLChangedFiles({
    owner: 'owner', repo: 'repo', pullNumber: 108,
    graphql: async () => ({repository: {pullRequest: {changedFiles: '2'}}}),
  }));
}

async function testKernelStatusesAndRenames() {
  for (const status of ['added', 'modified', 'removed']) {
    const result = await inspectChangedFiles({
      owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 1,
      listFiles: async () => ({status: 200, data: [file('.github/workflows/ci.yml', status)]}),
    });
    assert.equal(result.code, null);
    assert.equal(result.decision, 'AUTHORIZATION_REQUIRED');
    assert.equal(result.authorizationRequired, true);
  }
  const renamed = await inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 1,
    listFiles: async () => ({status: 200, data: [file('docs/new.md', 'renamed', '.github/ci-governance.json')]}),
  });
  assert.equal(renamed.code, null);
  assert.equal(renamed.decision, 'AUTHORIZATION_REQUIRED');
  assert.equal(renamed.authorizationRequired, true);
  const inert = await inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, expectedCount: 1,
    listFiles: async () => ({status: 200, data: [{...file('.github/workflows/ci.yml'), patch: '# name: CI guardian'}]}),
  });
  assert.equal(inert.code, null);
  assert.equal(inert.decision, 'AUTHORIZATION_REQUIRED');
  assert.equal(inert.authorizationRequired, true);
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
    result: {decision: 'PASS', code: null, reason: null, files: [file('docs/a.md')], kernelPaths: []}, branchProtection: mainProtectionFixture(), devBranchProtection: {...mainProtectionFixture(), branch: 'dev', required_checks: ['CI guardian shadow', 'CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'], required_check_bindings: ['CI guardian shadow', 'CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'].map((context) => ({context, app_id: 15368})).sort((left, right) => left.context < right.context ? -1 : left.context > right.context ? 1 : 0)}, observerEnvironment: OBSERVER_ENVIRONMENT, observerEnvironmentSnapshot: {repository: 'owner/repo', name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: ['branch_policy'], wait_timer: 0, reviewers: [], token_source: 'github.token', read_status: 'verified', missing_reason: '', run_id: 108, run_attempt: 1, workflow_sha: sha('d'), observed_at: observedAt, valid_until: validUntil, digest_sha256: ''}, installationRepositoryScope: {repository: 'owner/repo', installation_id: 42, token_source: 'github_app_installation', read_status: 'verified', repository_count: 1, repositories: ['owner/repo'], exact_match: true, missing_reason: '', run_id: 108, run_attempt: 1, workflow_sha: sha('d'), observed_at: observedAt, valid_until: validUntil, digest_sha256: ''},
  });
  promotionArtifact.dev_branch_protection.digest_sha256 = digestBranchProtection(promotionArtifact.dev_branch_protection);
  promotionArtifact.observer_environment_snapshot.digest_sha256 = require('./guardian').digestGuardianEnvironment(promotionArtifact.observer_environment_snapshot);
  promotionArtifact.observer_environment_digest = promotionArtifact.observer_environment_snapshot.digest_sha256;
  promotionArtifact.installation_repository_scope.digest_sha256 = digestInstallationRepositoryScope(promotionArtifact.installation_repository_scope);
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
  assert.throws(() => validateGuardianArtifact(promotionArtifact, promotionExpected, {now: observerNow}), (error) => error && error.code === CHECK_IDENTITY_CODE);
}

async function testPromotionPullRequestStateFailClosed() {
  const promotion = pull('main');
  promotion.head.ref = 'dev';
  promotion.state = 'open';
  promotion.draft = false;
  promotion.merged = false;
  promotion.merged_at = null;
  const liveResponse = (mutate = {}) => ({status: 200, data: {...promotion, ...mutate}});
  const revalidate = (mutate) => revalidatePullRequest({
    owner: 'owner', repo: 'repo', pullNumber: promotion.number, eventPull: promotion,
    getPull: async () => liveResponse(mutate),
  });
  const accepted = await revalidate();
  assert.equal(accepted.state, 'open');
  assert.equal(accepted.draft, false);
  for (const mutation of [
    {state: 'closed'},
    {draft: true},
    {merged: true},
    {merged_at: '2026-08-14T00:00:00Z'},
  ]) {
    await assert.rejects(() => revalidate(mutation), (error) => error && error.code === PROMOTION_TOPOLOGY_CODE);
  }
  const feature = pull('dev');
  const featureAccepted = await revalidatePullRequest({
    owner: 'owner', repo: 'repo', pullNumber: feature.number, eventPull: feature,
    getPull: async () => ({status: 200, data: feature}),
  });
  assert.equal(featureAccepted.head.ref, 'agent/ci-workflow');
}

async function testInstallationRepositoryScopeAttestation() {
  const valid = await observeInstallationRepositoryScope({
    repository: INSTALLATION_SCOPE_REPOSITORY, installationId: 42, tokenSource: 'github_app_installation', runId: 108, runAttempt: 1, workflowSHA: sha('d'), now: observerNow,
    listRepositories: async () => githubDateResponse({total_count: 1, repositories: [{full_name: INSTALLATION_SCOPE_REPOSITORY}]}),
  });
  assert.equal(valid.read_status, 'verified');
  assert.equal(valid.exact_match, true);
  assert.deepEqual(valid.repositories, [INSTALLATION_SCOPE_REPOSITORY]);
  assert.equal(valid.digest_sha256, digestInstallationRepositoryScope(valid));
  for (const data of [
    {total_count: 0, repositories: []},
    {total_count: 2, repositories: [{full_name: INSTALLATION_SCOPE_REPOSITORY}, {full_name: 'kimjooyoon/other'}]},
    {total_count: 1, repositories: [{full_name: 'kimjooyoon/other'}]},
  ]) {
    const unavailable = await observeInstallationRepositoryScope({
      repository: INSTALLATION_SCOPE_REPOSITORY, installationId: 42, tokenSource: 'github_app_installation', runId: 108, runAttempt: 1, workflowSHA: sha('d'), now: observerNow,
      listRepositories: async () => githubDateResponse(data),
    });
    assert.equal(unavailable.read_status, 'unavailable');
    assert.match(unavailable.missing_reason, /installation_repository_scope_api_mismatch/);
    assert.throws(() => validateInstallationRepositoryScope(unavailable, {requireVerified: true, expectedRepository: INSTALLATION_SCOPE_REPOSITORY, now: observerNow}), (error) => error && error.code === INSTALLATION_SCOPE_CODE);
  }
  const malformed = emptyInstallationRepositoryScope({repository: INSTALLATION_SCOPE_REPOSITORY, installationId: 0, tokenSource: 'github_app_installation', runId: 108, runAttempt: 1, workflowSHA: sha('d'), missingReason: 'test'});
  malformed.repositories = [INSTALLATION_SCOPE_REPOSITORY];
  malformed.repository_count = 1;
  malformed.digest_sha256 = digestInstallationRepositoryScope(malformed);
  assert.throws(() => validateInstallationRepositoryScope(malformed, {expectedRepository: INSTALLATION_SCOPE_REPOSITORY, now: observerNow}), (error) => error && error.code === INSTALLATION_SCOPE_CODE);
}

async function testProtectionObserverContracts() {
  assert.equal(OBSERVER_FRESHNESS_WINDOW_MS, 10 * 60 * 1000);
  assert.equal(validObserverFreshness(observedAt, validUntil, observerNow), true);
  assert.equal(validObserverFreshness(observedAt, validUntil, new Date('2026-08-14T00:06:00.000Z')), false);
  assert.equal(validObserverFreshness('2026-08-14T00:01:00.000Z', validUntil, observerNow), false);
  assert.equal(validObserverFreshness(observedAt, '2026-08-14T01:05:00.000Z', observerNow), false);
  assert.equal(validatePublicBranchSummary({protected: true, required_status_checks: {contexts: ['gofmt', 'CI guardian', 'go test', 'CI policy', 'go vet', 'go test -race', 'Semantic conformance'], checks: [{context: 'gofmt', app_id: 15368}, {context: 'CI guardian', app_id: 15368}, {context: 'go test', app_id: 15368}, {context: 'CI policy', app_id: 15368}, {context: 'go vet', app_id: 15368}, {context: 'go test -race', app_id: 15368}, {context: 'Semantic conformance', app_id: 15368}]}}), true);
  assert.equal(validatePublicBranchSummary({protected: true, required_status_checks: {contexts: ['gofmt'], checks: [{context: 'gofmt', app_id: 15368}]}}), false);
  const unavailable = await observeBranchProtection({repository: 'owner/repo', policySHA: '6'.repeat(64), eventRef: 'refs/heads/dev', checkoutRef: sha('d'), baseSHA: sha('b'), headSHA: sha('d'), runId: 108, runAttempt: 1, workflowSHA: sha('d'), tokenSource: 'github.token', getProtection: async () => ({status: 403})});
  assert.equal(unavailable.read_status, 'unavailable');
  assert.throws(() => validateBranchProtectionSnapshot(unavailable, {requireVerified: true}), (error) => error && error.code === PROTECTION_CODE);
  const protectionData = {required_status_checks: {strict: true, contexts: ['CI guardian', 'CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'], checks: [{context: 'CI guardian', app_id: 15368}, {context: 'CI policy', app_id: 15368}, {context: 'Semantic conformance', app_id: 15368}, {context: 'go test', app_id: 15368}, {context: 'go test -race', app_id: 15368}, {context: 'go vet', app_id: 15368}, {context: 'gofmt', app_id: 15368}]}, enforce_admins: {enabled: true}, required_linear_history: {enabled: true}, allow_force_pushes: {enabled: false}, allow_deletions: {enabled: false}, required_signatures: {enabled: false}, required_conversation_resolution: {enabled: false}, block_creations: {enabled: false}, lock_branch: {enabled: false}, allow_fork_syncing: {enabled: false}};
  const protectionArgs = {repository: 'owner/repo', branch: 'main', expectedContexts: ['CI guardian', 'CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'], policySHA: '6'.repeat(64), eventRef: 'refs/heads/dev', checkoutRef: sha('d'), baseSHA: sha('b'), headSHA: sha('d'), runId: 108, runAttempt: 1, workflowSHA: sha('d'), tokenSource: 'github_app_installation', appInstallationId: 42, appSlug: 'guardian'};
  const verified = await observeBranchProtection({...protectionArgs, now: observerNow, getProtection: async () => githubDateResponse(protectionData)});
  validateBranchProtectionSnapshot(verified, {requireVerified: true, now: observerNow});
  for (const freshness of [
    {observed_at: null, valid_until: validUntil},
    {observed_at: '2026-08-14T00:01:00.000Z', valid_until: '2026-08-14T00:11:00.000Z'},
    {observed_at: '2026-08-13T23:40:00.000Z', valid_until: '2026-08-13T23:50:00.000Z'},
  ]) {
    const stale = {...verified, ...freshness};
    stale.digest_sha256 = digestBranchProtection(stale);
    assert.throws(() => validateBranchProtectionSnapshot(stale, {requireVerified: true, now: observerNow}), (error) => error && error.code === PROTECTION_CODE);
  }
  const explicitNull = await observeBranchProtection({...protectionArgs, now: observerNow, getProtection: async () => githubDateResponse({...protectionData, required_pull_request_reviews: null, restrictions: null})});
  validateBranchProtectionSnapshot(explicitNull, {requireVerified: true, now: observerNow});
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
  const environment = await observeGuardianEnvironment({repository: 'owner/repo', tokenSource: 'github.token', runId: 108, runAttempt: 1, workflowSHA: sha('d'), now: observerNow, getEnvironment: async () => githubDateResponse({name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: [{type: 'branch_policy', id: 42, node_id: 'MDExOlByb3RlY3Rpb25SdWxlNDI='}]})});
  assert.equal(environment.read_status, 'verified');
  assert.equal(environment.observed_at, observedAt);
  assert.equal(environment.valid_until, validUntil);
  assert.deepEqual(environment.protection_rules, ['branch_policy']);
  assert.equal(environment.wait_timer, 0);
  assert.deepEqual(environment.reviewers, []);
  const environmentTamper = {...environment, deployment_branch_policy: {...environment.deployment_branch_policy, custom_branch_policies: true}};
  environmentTamper.digest_sha256 = require('./guardian').digestGuardianEnvironment(environmentTamper);
  assert.throws(() => validateGuardianEnvironment(environmentTamper, {requireVerified: true}), (error) => error && error.code === PROTECTION_CODE);
  const invalidEnvironmentResponses = [
    {name: 'missing policy', data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: []}},
    {name: 'required reviewers', data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: [{type: 'branch_policy'}, {type: 'required_reviewers'}]}},
    {name: 'wait timer', data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: [{type: 'wait_timer'}]}},
    {name: 'unknown rule', data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: [{type: 'deployment_branch_policy'}]}},
    {name: 'duplicate rule', data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: [{type: 'branch_policy'}, {type: 'branch_policy'}]}},
    {name: 'malformed rule', data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: [{id: 42}]}},
    {name: 'top-level wait timer', data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: [{type: 'branch_policy'}], wait_timer: 0}},
    {name: 'top-level reviewers', data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: true, custom_branch_policies: false}, protection_rules: [{type: 'branch_policy'}], reviewers: []}},
    {name: 'policy mismatch', data: {name: OBSERVER_ENVIRONMENT, deployment_branch_policy: {protected_branches: false, custom_branch_policies: false}, protection_rules: [{type: 'branch_policy'}]}},
  ];
  for (const invalid of invalidEnvironmentResponses) {
    const environmentMissing = await observeGuardianEnvironment({repository: 'owner/repo', tokenSource: 'github.token', runId: 108, runAttempt: 1, workflowSHA: sha('d'), getEnvironment: async () => githubDateResponse(invalid.data)});
    assert.equal(environmentMissing.read_status, 'unavailable', invalid.name);
    assert.equal(environmentMissing.missing_reason, 'guardian_environment_api_malformed', invalid.name);
  }
  const responseDateMissing = await observeBranchProtection({...protectionArgs, getProtection: async () => ({status: 200, data: protectionData})});
  assert.equal(responseDateMissing.read_status, 'unavailable');
  assert.match(responseDateMissing.missing_reason, /response_date/);
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
  const featurePull = pull('dev');
  featurePull.base.sha = sha('d');
  const identicalLive = {refs: {dev_sha: sha('d'), main_sha: sha('d')}, topology: {status: 'identical', ahead_by: 0, behind_by: 0, merge_base_sha: sha('d')}};
  const featureKernelBefore = 'sha256:' + '3'.repeat(64);
  const featureKernelAfter = 'sha256:' + '4'.repeat(64);
  const featureWithKernel = classifyGuardianDecision({
    pull: featurePull, repository: 'owner/repo', defaultBranch: 'dev', eventRef: 'refs/heads/dev', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'), runtimeSha: sha('d'), checkName: 'CI guardian shadow', liveBefore: identicalLive, liveAfter: identicalLive,
    kernelBeforeDigest: featureKernelBefore, kernelAfterDigest: featureKernelAfter,
    result: {decision: 'PASS', code: null, reason: 'authorized feature', foundationAuthorization: {decision: 'PASS'}, kernelPaths: ['.github/workflows/ci.yml']},
  });
  assert.equal(featureWithKernel.decision, 'PASS');
  assert.equal(featureWithKernel.kernelBeforeDigest, featureKernelBefore);
  assert.equal(featureWithKernel.kernelAfterDigest, featureKernelAfter);
  const featureKernelArtifact = buildGuardianArtifact({
    pull: featurePull, repository: 'owner/repo', action: 'reopened', defaultBranch: 'dev',
    workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'),
    runtimeRef: 'refs/heads/dev', runtimeSha: sha('d'), runId: 109, runAttempt: 1, eventRef: 'refs/heads/dev',
    liveBefore: identicalLive, liveAfter: identicalLive, checkName: 'CI guardian shadow', result: featureWithKernel,
  });
  assert.equal(featureKernelArtifact.kernel_before_sha256, featureKernelBefore);
  assert.equal(featureKernelArtifact.kernel_after_sha256, featureKernelAfter);
  assert.equal(validateKernelDigestAttestation({kernelPaths: featureKernelArtifact.kernel_paths, computedBeforeDigest: featureKernelBefore, computedAfterDigest: featureKernelAfter, artifactBeforeDigest: featureKernelArtifact.kernel_before_sha256, artifactAfterDigest: featureKernelArtifact.kernel_after_sha256}).decision, 'PASS');
  assert.equal(validateKernelDigestAttestation({kernelPaths: featureKernelArtifact.kernel_paths, computedBeforeDigest: featureKernelBefore, computedAfterDigest: featureKernelAfter, artifactBeforeDigest: featureKernelBefore, artifactAfterDigest: featureKernelBefore}).decision, 'REFUTED');
  assert.equal(validateKernelDigestAttestation({kernelPaths: featureKernelArtifact.kernel_paths, computedBeforeDigest: featureKernelBefore, computedAfterDigest: featureKernelAfter, artifactBeforeDigest: null, artifactAfterDigest: featureKernelAfter}).decision, 'REFUTED');
  assert.equal(DEFAULT_BRANCH_CODE, 'CI-GUARDIAN-DEFAULT-BRANCH-001');
  const missingLive = classifyGuardianDecision({
    pull: promotionPull, repository: 'owner/repo', defaultBranch: 'dev', eventRef: 'refs/heads/dev', workflowRef: 'owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev', workflowSha: sha('d'), runtimeSha: sha('d'), checkName: 'CI guardian',
    result: {decision: 'PASS', code: null, reason: null, kernelPaths: []},
  });
  assert.equal(missingLive.code, 'CI-GUARDIAN-LIVE-REF-001');
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
  assert.match(workflow, /client-id: \$\{\{ env\.GUARDIAN_APP_CLIENT_ID \}\}/);
  assert.doesNotMatch(workflow, /\bapp-id:/);
  assert.match(workflow, /permission-administration: read/);
  assert.match(workflow, /GUARDIAN_APP_PRIVATE_KEY/);
  const mintStart = workflow.indexOf('id: guardian-app-token');
  const mintEnd = workflow.indexOf('- name: Inspect changed paths from default authority', mintStart);
  assert(mintStart >= 0 && mintEnd > mintStart, 'Guardian App mint step is missing');
  const mintStep = workflow.slice(mintStart, mintEnd);
  assert.match(mintStep, /owner: \$\{\{ github\.repository_owner \}\}/);
  assert.match(mintStep, /repositories: \$\{\{ github\.event\.repository\.name \}\}/);
  assert.equal((mintStep.match(/^\s+permission-[^:]+:/gm) || []).length, 1, 'Guardian App token must request only one explicit permission');
  assert.equal((workflow.match(/\$\{\{ secrets\.[^}]+\}\}/g) || []).length, 1, 'workflow must contain exactly one secret reference');
  assert.equal((mintStep.match(/\$\{\{ secrets\.[^}]+\}\}/g) || []).join(''), '${{ secrets.GUARDIAN_APP_PRIVATE_KEY }}', 'only the Guardian private key may be read in the mint step');
  assert.match(workflow, /getBranchProtection/);
  assert.match(workflow, /GET \/installation\/repositories/);
  assert.match(workflow, /observeInstallationRepositoryScope/);
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
  assert.match(workflow, /const trustedPromotion = guardian\.trustedDevPromotion/);
  assert.match(workflow, /const authorizedFoundationFeature/);
  assert.match(workflow, /result\.foundationAuthorization\.decision === 'PASS'/);
  assert.match(workflow, /trustedPromotion \|\| authorizedFoundationFeature/);
  assert.match(workflow, /result\.authorizationRequired === true/);
  const dispatchIndex = workflow.indexOf('guardian.observeFoundationAuthorization');
  const kernelAssignmentIndex = workflow.indexOf('beforeDigest = await guardian.kernelTreeDigest');
  assert(dispatchIndex >= 0 && kernelAssignmentIndex >= 0 && dispatchIndex < kernelAssignmentIndex, 'Foundation authorization must be dispatched before kernel digest attestation');
  assert.match(workflow, /guardian\.validateKernelDigestAttestation/);
  assert.match(workflow, /computedBeforeDigest: beforeDigest/);
  assert.match(workflow, /artifactBeforeDigest: artifact\.kernel_before_sha256/);
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

function testExecutableGuardianScopeAcceptanceHarness() {
  const workflow = fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'workflows', 'ci-guardian.yml'), 'utf8');
  const receipt = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'governance-denominator-v5-executable-guardian-scope.json'), 'utf8'));
  const acceptance = new Map([
    ['SCOPE_INITIALIZED_BEFORE_POLICY_BRANCH', () => {
      const declaration = workflow.indexOf('let beforeDigest = null;');
      const policyBranch = workflow.indexOf('if (route === guardian.FOUNDATION_ROUTE)');
      assert(declaration >= 0 && declaration < policyBranch, 'digest scope is not initialized before the policy branch');
    }],
    ['SCOPE_REUSED_WITHOUT_REDECLARATION', () => {
      assert.equal((workflow.match(/let beforeDigest = null;/g) || []).length, 1);
      assert.equal((workflow.match(/let afterDigest = null;/g) || []).length, 1);
      assert.match(workflow, /beforeDigest = await guardian\.kernelTreeDigest/);
      assert.match(workflow, /afterDigest = await guardian\.kernelTreeDigest/);
    }],
    ['WORKFLOW_AUTHORITY_PINNED_TO_GITHUB_WORKFLOW_SHA', () => {
      assert.match(workflow, /ref: \$\{\{ github\.workflow_sha \}\}/);
      assert.doesNotMatch(workflow, /ref: \$\{\{ github\.event\.pull_request\.base\.sha \}\}/);
    }],
    ['LIVE_PR_CHANGED_PATHS_ATTESTED', () => {
      assert.match(workflow, /github\.rest\.pulls\.get/);
      assert.match(workflow, /listFiles/);
      assert.match(workflow, /expectedChangedFiles = observedPull\.changed_files/);
    }],
    ['PASS_KERNEL_DIGESTS_NON_NULL_EXACT', () => {
      const before = `sha256:${'a'.repeat(64)}`;
      const after = `sha256:${'b'.repeat(64)}`;
      assert.equal(validateKernelDigestAttestation({kernelPaths: ['go.mod'], computedBeforeDigest: before, computedAfterDigest: after, artifactBeforeDigest: before, artifactAfterDigest: after}).decision, 'PASS');
    }],
    ['NULL_STALE_MISMATCH_REFUTED', () => {
      const before = `sha256:${'a'.repeat(64)}`;
      const after = `sha256:${'b'.repeat(64)}`;
      assert.equal(validateKernelDigestAttestation({kernelPaths: ['go.mod'], computedBeforeDigest: null, computedAfterDigest: after, artifactBeforeDigest: before, artifactAfterDigest: after}).decision, 'REFUTED');
      assert.equal(validateKernelDigestAttestation({kernelPaths: ['go.mod'], computedBeforeDigest: before, computedAfterDigest: after, artifactBeforeDigest: before, artifactAfterDigest: before}).decision, 'REFUTED');
    }],
    ['FUTURE_SCHEMA_UNKNOWN_OVER_6_FIELDS', () => {
      const classified = foundationAuthorization.classifyExecutableGuardianScopeInput({schema: 'gooo/receipt-schema-migration/v0.2.3'});
      assert.equal(classified.decision, 'UNKNOWN');
      assert.equal(classified.unknown_count, 6);
    }],
    ['REFERENCE_ERROR_REFUTED_SCOPE_CLOSED', () => {
      assert.equal(receipt.prior_guardian_failure.message, 'ReferenceError: beforeDigest is not defined');
      assert.equal(receipt.cells[0].parent_outcome, foundationAuthorization.EXECUTABLE_GUARDIAN_SCOPE_PARENT_OUTCOME);
      assert.equal(receipt.cells[0].outcome, 'CLOSED');
    }],
  ]);
  assert.deepEqual([...acceptance.keys()], foundationAuthorization.EXECUTABLE_GUARDIAN_SCOPE_ACCEPTANCE_IDS);
  for (const [id, execute] of acceptance) {
    assert.doesNotThrow(execute, `acceptance case failed: ${id}`);
  }
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

function testRegressionRepairReceipt() {
  const receipt = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'governance-denominator-v2-migration.json'), 'utf8'));
  assert.doesNotThrow(() => foundationAuthorization.validateRegressionRepairReceipt(receipt));
  assert.doesNotThrow(() => foundationAuthorization.validateIncompletePropagationOutcome(receipt));
  assert.equal(receipt.foundation_override_success_count, foundationAuthorization.FOUNDATION_OVERRIDE_SUCCESS_COUNT);
  assert.equal(receipt.outcome, 'REFUTED_INCOMPLETE_PROPAGATION');
  const correction = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'governance-denominator-v3-correction.json'), 'utf8'));
  assert.doesNotThrow(() => foundationAuthorization.validateCorrectionChildReceipt(correction));
  assert.equal(correction.cells.length, 1);
  assert.equal(correction.cells[0].id, 'CORRECTION_CHILD');
  assert.equal(correction.cells[0].parent_repair_receipt, foundationAuthorization.CORRECTION_CHILD_PARENT_RECEIPT_SHA256);
  const migration = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'governance-denominator-v4-schema-coherence.json'), 'utf8'));
  assert.doesNotThrow(() => foundationAuthorization.validateSchemaCoherenceMigrationReceipt(migration));
  assert.equal(migration.cells.length, 1);
  assert.equal(migration.cells[0].id, 'SCHEMA_COHERENCE_MIGRATION_ADOPTION');
  assert.equal(migration.cells[0].proof_choice, 'COHERENCE');
  assert.equal(migration.cells[0].indicator, 'GUARDRAIL');
  assert.equal(migration.cells[0].allowed, 1);
  assert.equal(migration.cells[0].consumed, 1);
  assert.equal(migration.cells[0].replay_decision, 'REFUTED');
}

(async () => {
  await testPaginationAndNonKernelPass();
  await testBoundChangedPathAuthority();
  await testKernelStatusesAndRenames();
  await testForkAndMalformedAPI();
  await testStaleRaceAndArtifactDigest();
  await testPromotionPullRequestStateFailClosed();
  await testInstallationRepositoryScopeAttestation();
  await testCanonicalOrdering();
  await testPromotionAndKernelDigests();
  await testProtectionObserverContracts();
  await testPaginationLimit();
  testWorkflowIsReadOnlyAndBasePinned();
  testExecutableGuardianScopeAcceptanceHarness();
  testHeadBindingIsExplicitlyShadowOnly();
  testKernelSetIsMonotonic();
  testRegressionRepairReceipt();
  await testLiveRefsAndRouteIdentity();
  console.log('guardian tests passed');
})().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
