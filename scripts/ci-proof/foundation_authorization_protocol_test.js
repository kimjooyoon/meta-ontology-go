'use strict';

const assert = require('node:assert/strict');
const protocol = require('./foundation_authorization_protocol');

const OWNER = Object.freeze({login: 'kimjooyoon', id: 115961382, type: 'User', permission: 'admin'});
const BASE_SHA = '1'.repeat(40);
const HEAD_SHA = '2'.repeat(40);
const MERGE_BASE_SHA = '3'.repeat(40);
const BEFORE_DIGEST = protocol.sha256('kernel-before');
const AFTER_DIGEST = protocol.sha256('kernel-after');
const NOW = '2026-09-03T00:00:00.000Z';
const EXPIRES = '2026-09-04T00:00:00.000Z';
const NONCE = 'nonce-20260903-exact';

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function buildCandidate() {
  const changedPaths = [
    '.github/foundation-authorization-protocol.json',
    '.github/workflows/ci.yml',
    'internal/verify/scope_foundation_authorization_protocol_20260903.go',
    'scripts/ci-proof/foundation_authorization_protocol.js',
    'scripts/ci-proof/foundation_authorization_protocol_test.js',
  ].sort();
  const protectedPaths = protocol.canonicalProtectedIntersection(changedPaths);
  return {
    pull_request: 700,
    base_repo: protocol.REPOSITORY,
    base_branch: 'dev',
    base_sha: BASE_SHA,
    head_repo: protocol.REPOSITORY,
    head_branch: 'agent/foundation-authorization-protocol-20260903',
    head_sha: HEAD_SHA,
    merge_base_sha: MERGE_BASE_SHA,
    changed_paths: {paths: changedPaths, count: changedPaths.length, digest: protocol.digestPathList(changedPaths)},
    protected_intersection: {paths: protectedPaths, count: protectedPaths.length, digest: protocol.digestPathList(protectedPaths)},
  };
}

function buildInput(state = 'PENDING') {
  const candidate = buildCandidate();
  const runtime = {
    workflow_ref: `${protocol.REPOSITORY}/.github/workflows/ci.yml@refs/heads/dev`,
    workflow_sha: '4'.repeat(40),
    runtime_ref: 'refs/heads/dev',
    runtime_sha: '5'.repeat(40),
    check_name: 'FOUNDATION authorization protocol',
    check_run_id: 700001,
    app_id: protocol.CI_APP_ID,
  };
  const input = {
    repository: {full_name: protocol.REPOSITORY, owner_login: OWNER.login, owner_id: OWNER.id, owner_type: OWNER.type},
    actor: {...OWNER, owner_match: true},
    intent: 'Replace the historical static one-tuple foundation authorization with a reusable one-use exact-candidate protocol.',
    proof_choice: 'FOUNDATION',
    candidate,
    runtime,
    kernel: {before_sha256: BEFORE_DIGEST, after_sha256: AFTER_DIGEST},
    authorization: {
      state,
      owner_selection: {...OWNER},
      nonce: NONCE,
      issued_at: NOW,
      expires_at: EXPIRES,
      stale: false,
      use_count: state === 'CONSUMED' ? 1 : 0,
      reuse_attempts: 0,
      revoked: false,
      authorization_receipt: null,
    },
    consumption: null,
    observed: {candidate: clone(candidate), runtime: clone(runtime), kernel: {before_sha256: BEFORE_DIGEST, after_sha256: AFTER_DIGEST}},
  };
  if (state === 'AUTHORIZED' || state === 'CONSUMED') {
    input.authorization.authorization_receipt = {
      nonce: NONCE,
      candidate_digest: protocol.candidateDigest(candidate),
      owner: {...OWNER},
    };
  }
  if (state === 'CONSUMED') {
    input.consumption = {
      nonce: NONCE,
      merge_parent_sha: BASE_SHA,
      consumed_at: '2026-09-03T00:10:00.000Z',
      replay_decision: 'REFUTED',
      post_adoption_verified: true,
    };
    input.consumption.receipt_digest = protocol.consumptionReceiptDigest(input.consumption);
  }
  return input;
}

protocol.validateProtocolDefinition();
assert.equal(protocol.DECISION_PRECEDENCE.join('>'), 'REFUTED>UNKNOWN>CLOSED');
assert.deepEqual(protocol.UNKNOWN_FIELDS, ['stage', 'step', 'reason', 'unknown_class', 'next_operation', 'blocked_by']);
assert.equal(protocol.validateProtocolDefinition().cells.length, 12);
assert.equal(new Set(protocol.validateProtocolDefinition().cells.map((cell) => cell.stage)).size, 3);
assert.ok(!Object.prototype.hasOwnProperty.call(protocol.validateProtocolDefinition(), 'aggregate_score'));

for (const state of ['PENDING', 'AUTHORIZED', 'CONSUMED']) {
  const result = protocol.evaluate(buildInput(state), {now: NOW});
  assert.equal(result.decision, 'CLOSED');
  assert.equal(result.state, state);
  assert.equal(result.unknown, null);
  assert.equal(result.refuted, null);
  assert.equal(result.repository_writes, 0);
  assert.equal(result.cells.length, 12);
  assert.ok(result.cells.every((cell) => cell.state === 'CLOSED'));
  assert.ok(!Object.prototype.hasOwnProperty.call(result, 'score'));
}

const missingOwner = buildInput('PENDING');
missingOwner.authorization.owner_selection = null;
let result = protocol.evaluate(missingOwner, {now: NOW});
assert.equal(result.decision, 'UNKNOWN');
assert.equal(result.reason, 'OWNER_SELECTION_MISSING');
assert.deepEqual(Object.keys(result.unknown), protocol.UNKNOWN_FIELDS);
assert.deepEqual(result.unknown, {
  stage: 'FOUNDATION',
  step: 'REPOSITORY_ACTOR_AUTHORITY',
  reason: 'OWNER_SELECTION_MISSING',
  unknown_class: 'DIRECT_MISSING',
  next_operation: 'SELECT_REPOSITORY_OWNER',
  blocked_by: [],
});

const incompleteTuple = buildInput('PENDING');
delete incompleteTuple.candidate.head_sha;
result = protocol.evaluate(incompleteTuple, {now: NOW});
assert.equal(result.decision, 'UNKNOWN');
assert.equal(result.reason, 'INCOMPLETE_CANDIDATE_TUPLE');
assert.deepEqual(Object.keys(result.unknown), protocol.UNKNOWN_FIELDS);

const candidateDrift = buildInput('PENDING');
candidateDrift.observed.candidate.head_sha = '6'.repeat(40);
result = protocol.evaluate(candidateDrift, {now: NOW});
assert.equal(result.decision, 'REFUTED');
assert.equal(result.reason, 'CANDIDATE_DRIFT');

const digestMismatch = buildInput('PENDING');
digestMismatch.candidate.changed_paths.digest = protocol.sha256('wrong');
result = protocol.evaluate(digestMismatch, {now: NOW});
assert.equal(result.decision, 'REFUTED');
assert.equal(result.reason, 'DIGEST_MISMATCH');

const expired = buildInput('AUTHORIZED');
result = protocol.evaluate(expired, {now: '2026-09-05T00:00:00.000Z'});
assert.equal(result.decision, 'REFUTED');
assert.equal(result.reason, 'AUTHORIZATION_EXPIRED');

const nonceReplay = buildInput('CONSUMED');
nonceReplay.authorization.reuse_attempts = 1;
result = protocol.evaluate(nonceReplay, {now: NOW});
assert.equal(result.decision, 'REFUTED');
assert.equal(result.reason, 'NONCE_REPLAY');

const actorMismatch = buildInput('PENDING');
actorMismatch.actor.id = OWNER.id + 1;
result = protocol.evaluate(actorMismatch, {now: NOW});
assert.equal(result.decision, 'REFUTED');
assert.equal(result.reason, 'ACTOR_MISMATCH');

const precedence = buildInput('PENDING');
precedence.authorization.owner_selection = null;
precedence.observed.candidate.head_sha = '7'.repeat(40);
result = protocol.evaluate(precedence, {now: NOW});
assert.equal(result.decision, 'REFUTED');
assert.equal(result.reason, 'CANDIDATE_DRIFT');
assert.equal(protocol.resolveDecision(['CLOSED']), 'CLOSED');
assert.equal(protocol.resolveDecision(['UNKNOWN', 'CLOSED']), 'UNKNOWN');
assert.equal(protocol.resolveDecision(['REFUTED', 'UNKNOWN', 'CLOSED']), 'REFUTED');

const deterministicInput = buildInput('AUTHORIZED');
assert.deepEqual(protocol.evaluate(deterministicInput, {now: NOW}), protocol.evaluate(deterministicInput, {now: NOW}));

const root = protocol.buildRootAxiomReceipt({
  repository: protocol.REPOSITORY,
  owner: {...OWNER},
  actor: {...OWNER, permission: 'admin'},
  pullRequest: 636,
  baseSha: '8'.repeat(40),
  headSha: '9'.repeat(40),
  audit: {
    installation: {read_status: 'unavailable', reason: 'A JSON web token could not be decoded'},
    rulesets: {read_status: 'unavailable', reason: 'administration permission is not exposed to this token'},
    environments: {read_status: 'unavailable', reason: 'administration permission is not exposed to this token'},
    owner_operation: {read_status: 'unavailable', reason: 'NO_SERVER_SIDE_OWNER_AUTHORIZATION_OPERATION'},
  },
});
protocol.validateRootAxiomReceipt(root);
assert.equal(root.schema, protocol.ROOT_AXIOM_SCHEMA);
assert.equal(root.decision, 'UNKNOWN');
assert.equal(root.candidate.pull_request, 636);
assert.equal(root.candidate.head_sha, '9'.repeat(40));
assert.equal(root.owner.login, OWNER.login);
assert.deepEqual(Object.keys(root.unknown), protocol.UNKNOWN_FIELDS);
assert.equal(root.no_check_bypass, true);
assert.equal(root.no_force_push, true);
assert.equal(root.repository_writes, 0);

console.log('foundation authorization protocol cases: PASS');
