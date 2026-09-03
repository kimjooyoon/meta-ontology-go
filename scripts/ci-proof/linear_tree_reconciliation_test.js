'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const reconciliation = require('./linear_tree_reconciliation');

const sha = (letter) => letter.repeat(40);
const blob = (name, letter) => ({path: name, type: 'blob', mode: '100644', sha: sha(letter)});
const tree = reconciliation.treeManifest([blob('README.md', 'a'), blob('internal/verify/example.go', 'b')]);
const pull = {number: 99, base: {ref: 'main', sha: sha('c'), repo: {full_name: 'kimjooyoon/meta-ontology-go'}}, head: {ref: 'agent/main-history-reconciliation-fixture-20260903', sha: sha('d'), repo: {full_name: 'kimjooyoon/meta-ontology-go'}}};
const workflow = {workflow_ref: 'kimjooyoon/meta-ontology-go/.github/workflows/ci-guardian.yml@refs/heads/dev', workflow_sha: sha('e'), runtime_ref: 'refs/heads/dev', runtime_sha: sha('e'), check_name: 'CI guardian', app_id: 15368};
const candidate = reconciliation.candidateFromPull({pull, changedFiles: [{filename: '.github/ci-governance.json', status: 'modified'}], sourceDevSHA: sha('f'), mergeBaseSHA: sha('1'), sourceDevTree: tree, candidateTree: tree, workflow});
const authorization = {schema: reconciliation.AUTHORIZATION_SCHEMA, state: 'AUTHORIZED', candidate_digest: reconciliation.candidateDigest(candidate), use_count: 0, reuse_attempts: 0, stale: false, revoked: false};
const ownerRecord = {
  schema: reconciliation.AUTHORIZATION_SCHEMA,
  binding: reconciliation.AUTHORIZATION_BINDING,
  state: 'AUTHORIZED',
  proof_choice: 'FOUNDATION',
  intent: 'AUTHORIZE_LINEAR_TREE_EQUIVALENT_RECONCILIATION_FOR_EXACT_CANDIDATE',
  required_operation: 'OWNER_CREATE_ONE_USE_LINEAR_TREE_RECONCILIATION_AUTHORIZATION',
  reusable: false,
  one_use: true,
  nonce: 'fixture-owner-authorization-nonce',
  issued_at: '2026-09-03T00:00:00.000Z',
  expires_at: '2026-09-04T00:00:00.000Z',
  use_count: 0,
  reuse_attempts: 0,
  stale: false,
  revoked: false,
  owner_selection: reconciliation.OWNER,
  actor: reconciliation.OWNER,
  candidate_binding: reconciliation.authorizationCandidateBinding(candidate),
  candidate_digest: reconciliation.candidateDigest(candidate),
  source_dev_sha_at_authorization: candidate.source_dev_sha,
  base_snapshot_sha: candidate.base_sha,
  candidate_tree_digest: candidate.candidate_tree.tree_digest,
  source_dev_tree_digest: candidate.source_dev_tree.tree_digest,
  authorization_receipt: {
    nonce: 'fixture-owner-authorization-nonce',
    candidate_digest: reconciliation.candidateDigest(candidate),
    owner: reconciliation.OWNER,
  },
  no_protection_mutation: true,
  no_check_bypass: true,
  no_force_push: true,
  repository_writes_before_authorization: 0,
};

test('compact owner authorization binds the complete candidate without embedding its full tree manifest', () => {
  assert.doesNotThrow(() => reconciliation.validateOwnerAuthorization(ownerRecord, {candidate, now: new Date('2026-09-03T12:00:00.000Z')}));
  assert.ok(JSON.stringify(ownerRecord).length < 65536);
});

test('NORMAL closes only after exact tree, current refs, authorization, and 7/7 checks', () => {
  const result = reconciliation.evaluate({route: reconciliation.ROUTE, candidate, live_main_sha: pull.base.sha, live_dev_sha: candidate.source_dev_sha, authorization, workflow, required_checks: reconciliation.REQUIRED_CHECKS.map((name) => ({name, status: 'completed', conclusion: 'success', head_sha: pull.head.sha, app_id: 15368}))});
  assert.equal(result.decision, 'CLOSED');
  assert.equal(result.unknown, null);
});

test('UNKNOWN preserves the exact six-field evidence shape when dev moves', () => {
  const result = reconciliation.evaluate({route: reconciliation.ROUTE, candidate, live_main_sha: pull.base.sha, live_dev_sha: sha('2'), authorization});
  assert.equal(result.decision, 'UNKNOWN');
  assert.deepEqual(Object.keys(result.unknown), reconciliation.UNKNOWN_FIELDS);
});

test('REFUTED wins for a tree mismatch and ordinary main route', () => {
  const mismatched = {...candidate, candidate_tree: {...tree, tree_digest: reconciliation.sha256('different')}};
  const result = reconciliation.evaluate({route: 'promotion_main', candidate: mismatched, live_main_sha: pull.base.sha, live_dev_sha: candidate.source_dev_sha, authorization});
  assert.equal(result.decision, 'REFUTED');
  assert.equal(result.refuted.reason, 'UNAUTHORIZED_ROUTE');
});

test('one-use replay is always REFUTED', () => {
  assert.deepEqual(reconciliation.replayReceipt({authorization: {use_count: 0, reuse_attempts: 0}}), {decision: 'REFUTED', reason: 'ONE_USE_AUTHORIZATION_REPLAY_ATTEMPT'});
  assert.deepEqual(reconciliation.replayReceipt({authorization: {use_count: 2, reuse_attempts: 1}}), {decision: 'REFUTED', reason: 'NONCE_REPLAY'});
});
