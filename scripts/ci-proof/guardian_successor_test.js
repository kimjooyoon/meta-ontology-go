'use strict';

const assert = require('node:assert/strict');
const protocol = require('./foundation_authorization_protocol');
const successor = require('./guardian_successor');

const pull = {
  number: 640,
  base: {ref: 'main', sha: '8'.repeat(40), repo: {full_name: protocol.REPOSITORY}},
  head: {ref: 'dev', sha: '9'.repeat(40), repo: {full_name: protocol.REPOSITORY}},
};
const changedFiles = [
  {filename: '.github/workflows/ci-guardian.yml', status: 'modified'},
  {filename: 'scripts/ci-proof/foundation_authorization_protocol.js', status: 'modified'},
  {filename: 'scripts/ci-proof/guardian.js', status: 'modified'},
  {filename: 'scripts/ci-proof/guardian_successor.js', status: 'added'},
  {filename: 'scripts/ci-proof/guardian_successor_test.js', status: 'added'},
];
const mergeBaseSha = pull.base.sha;
const candidate = successor.candidateFromPull({pull, changedFiles, mergeBaseSha});
const runtime = {
  workflow_ref: `${protocol.REPOSITORY}/.github/workflows/ci-guardian.yml@refs/heads/dev`,
  workflow_sha: pull.head.sha,
  runtime_ref: 'refs/heads/dev',
  runtime_sha: pull.head.sha,
  check_name: 'CI guardian',
  check_run_id: 640001,
  app_id: protocol.CI_APP_ID,
};
const kernel = {before_sha256: protocol.sha256('before'), after_sha256: protocol.sha256('after')};
const live = {
  refs: {main_sha: pull.base.sha, dev_sha: pull.head.sha},
  topology: {status: 'ahead', ahead_by: 1, behind_by: 0, merge_base_sha: pull.base.sha},
};
const ownerRecord = {
  schema: successor.AUTHORIZATION_SCHEMA,
  authorization_route: 'NORMAL_ONE_USE_PROTOCOL',
  state: 'AUTHORIZED',
  proof_choice: 'FOUNDATION',
  intent: successor.AUTHORIZATION_INTENT,
  required_operation: successor.AUTHORIZATION_OPERATION,
  reason: successor.AUTHORIZATION_REASON,
  reusable: false,
  one_use: true,
  nonce: 'successor-protocol-nonce-20260903',
  issued_at: '2026-09-03T00:00:00.000Z',
  expires_at: '2026-09-04T00:00:00.000Z',
  use_count: 0,
  reuse_attempts: 0,
  stale: false,
  revoked: false,
  owner_selection: {...successor.OWNER},
  actor: {...successor.OWNER},
  candidate,
  current_dev_tip_at_authorization: pull.head.sha,
  base_snapshot_is_current_dev_tip: false,
  candidate_digest: protocol.candidateDigest(candidate),
  root_axiom_receipt_schema: null,
  root_axiom_receipt_digest: null,
  external_root_reuse: false,
  authorization_receipt: {
    nonce: 'successor-protocol-nonce-20260903',
    candidate_digest: protocol.candidateDigest(candidate),
    owner: {...successor.OWNER},
  },
  predecessor_root: {reused: false, source: 'NONE', reason: 'PR_640_SUCCESSOR_PROTOCOL'},
  no_protection_mutation: true,
  no_check_bypass: true,
  no_force_push: true,
  repository_writes_before_authorization: 0,
};

successor.validateOwnerRecord(ownerRecord, {candidate, now: new Date('2026-09-03T01:00:00.000Z')});
const input = {
  repository: {full_name: protocol.REPOSITORY, owner_login: successor.OWNER.login, owner_id: successor.OWNER.id, owner_type: successor.OWNER.type},
  actor: {...successor.OWNER, owner_match: true},
  intent: successor.AUTHORIZATION_INTENT,
  proof_choice: 'FOUNDATION',
  candidate,
  runtime,
  kernel,
  authorization: {
    state: ownerRecord.state,
    owner_selection: ownerRecord.owner_selection,
    nonce: ownerRecord.nonce,
    issued_at: ownerRecord.issued_at,
    expires_at: ownerRecord.expires_at,
    stale: ownerRecord.stale,
    use_count: ownerRecord.use_count,
    reuse_attempts: ownerRecord.reuse_attempts,
    revoked: ownerRecord.revoked,
    authorization_receipt: ownerRecord.authorization_receipt,
  },
  consumption: null,
  observed: {candidate: JSON.parse(JSON.stringify(candidate)), runtime: JSON.parse(JSON.stringify(runtime)), kernel: JSON.parse(JSON.stringify(kernel))},
};

let result = protocol.evaluate(input, {now: '2026-09-03T01:00:00.000Z'});
assert.equal(result.decision, 'CLOSED');
assert.equal(result.cells.length, 12);
assert.equal(result.unknown, null);
assert.equal(result.refuted, null);

const missingOwner = JSON.parse(JSON.stringify(input));
missingOwner.authorization.owner_selection = null;
result = protocol.evaluate(missingOwner, {now: '2026-09-03T01:00:00.000Z'});
assert.equal(result.decision, 'UNKNOWN');
assert.deepEqual(Object.keys(result.unknown), protocol.UNKNOWN_FIELDS);

const candidateDrift = JSON.parse(JSON.stringify(input));
candidateDrift.observed.candidate.head_sha = 'a'.repeat(40);
result = protocol.evaluate(candidateDrift, {now: '2026-09-03T01:00:00.000Z'});
assert.equal(result.decision, 'REFUTED');
assert.equal(result.reason, 'CANDIDATE_DRIFT');

const consumed = JSON.parse(JSON.stringify(input));
consumed.authorization.state = 'CONSUMED';
consumed.authorization.use_count = 1;
consumed.consumption = {
  nonce: consumed.authorization.nonce,
  merge_parent_sha: pull.base.sha,
  consumed_at: '2026-09-03T01:10:00.000Z',
  replay_decision: 'REFUTED',
  post_adoption_verified: true,
};
consumed.consumption.receipt_digest = protocol.consumptionReceiptDigest(consumed.consumption);
result = protocol.evaluate(consumed, {now: '2026-09-03T01:00:00.000Z'});
assert.equal(result.decision, 'CLOSED');

consumed.authorization.reuse_attempts = 1;
result = protocol.evaluate(consumed, {now: '2026-09-03T01:00:00.000Z'});
assert.equal(result.decision, 'REFUTED');
assert.equal(result.reason, 'NONCE_REPLAY');

assert.throws(() => successor.validateOwnerRecord({...ownerRecord, stale: true}, {candidate, now: new Date('2026-09-03T01:00:00.000Z')}));
assert.equal(protocol.resolveDecision(['REFUTED', 'UNKNOWN', 'CLOSED']), 'REFUTED');

console.log('guardian successor protocol cases: PASS');
