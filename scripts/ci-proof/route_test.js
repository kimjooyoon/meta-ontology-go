'use strict';

const assert = require('node:assert/strict');
const foundation = require('./foundation_bootstrap');
const route = require('./route');

const cases = [
  ['pull_request', 'dev', 'feature_dev', false],
  ['pull_request', 'main', 'promotion_main', true],
  ['push', 'dev', 'protected_push_dev', false],
  ['push', 'main', 'protected_push_main', false],
];

for (const [event, baseRef, expected, guardianRequired] of cases) {
  const input = {
    event,
    eventRef: event === 'push' ? 'refs/heads/' + baseRef : 'refs/pull/7/merge',
    baseRef,
    headSha: 'a'.repeat(40),
  };
  const evidence = route.buildProofRouteEvidence(input);
  assert.equal(evidence.route, expected);
  assert.equal(evidence.guardian_required, guardianRequired);
  assert.match(evidence.digest, /^sha256:[0-9a-f]{64}$/);
  assert.doesNotThrow(() => route.validateProofRouteEvidence(evidence, input));
}

assert.throws(
  () => route.classifyProofRoute('workflow_dispatch', 'main'),
  /unsupported CI proof route tuple/,
);

const foundationPull = () => ({
  number: foundation.FOUNDATION_BOOTSTRAP.pullRequest,
  state: 'open',
  merged: false,
  merged_at: null,
  base: {ref: foundation.FOUNDATION_BOOTSTRAP.baseRef, sha: 'c'.repeat(40), repo: {full_name: foundation.FOUNDATION_BOOTSTRAP.repository}},
  head: {ref: foundation.FOUNDATION_BOOTSTRAP.headRef, sha: foundation.FOUNDATION_BOOTSTRAP.headSha, repo: {full_name: foundation.FOUNDATION_BOOTSTRAP.repository}},
});

const foundationLive = () => ({
  refs: {dev_sha: 'c'.repeat(40), main_sha: foundation.FOUNDATION_BOOTSTRAP.sourceMainSha},
});

const foundationResult = (liveKernelPaths, baseChangedPaths = []) => ({
  decision: 'FAIL_CLOSED',
  code: 'CI-ROOT-OF-TRUST-001',
  reason: 'protected kernel path changed',
  kernelPaths: [...liveKernelPaths].sort(),
  files: [...liveKernelPaths, ...baseChangedPaths].map((filename) => ({filename, status: 'modified'})),
});

const foundationBlobs = () => ({...foundation.EXPECTED_BASE_DRIFT_BLOB_SHAS});
const foundationBaseCommit = () => ({sha: 'c'.repeat(40), parents: [{sha: foundation.PRE_CORRECTION_BASE_SNAPSHOT}]});

const foundationDecision = (liveKernelPaths, correctionChangedPaths = foundation.CORRECTION_CHANGED_KERNEL_PATHS, baseKernelBlobs = {[foundation.REMAINING_BASE_SATISFIED_KERNEL_PATHS[0]]: foundation.EXPECTED_BASE_DRIFT_BLOB_SHAS[foundation.REMAINING_BASE_SATISFIED_KERNEL_PATHS[0]]}, preCorrectionKernelBlobs = foundationBlobs(), baseCommit = foundationBaseCommit()) => foundation.foundationBootstrapDecision({
  pull: foundationPull(),
  result: foundationResult(liveKernelPaths),
  liveBefore: foundationLive(),
  liveAfter: foundationLive(),
  baseCommit,
  baseKernelBlobs,
  preCorrectionKernelBlobs,
  correctionChangedPaths,
});

const normalFoundation = foundationDecision(foundation.AUTHORIZED_KERNEL_PATHS);
assert.equal(normalFoundation.decision, 'FOUNDATION');
assert.equal(normalFoundation.reason, 'FOUNDATION_OVERRIDE_USED=2');
assert.deepEqual(normalFoundation.observedKernelPaths, foundation.EXPECTED_LIVE_KERNEL_PATHS);
assert.deepEqual(normalFoundation.pullKernelPaths, foundation.AUTHORIZED_KERNEL_PATHS);
assert.deepEqual(normalFoundation.alreadySatisfiedByBase, foundation.REMAINING_BASE_SATISFIED_KERNEL_PATHS);

const missingLive = foundationDecision(foundation.AUTHORIZED_KERNEL_PATHS.slice(1));
assert.equal(missingLive.decision, 'REFUTED');

const missingBase = foundationDecision(foundation.AUTHORIZED_KERNEL_PATHS, foundation.CORRECTION_CHANGED_KERNEL_PATHS.concat(foundation.REMAINING_BASE_SATISFIED_KERNEL_PATHS));
assert.equal(missingBase.decision, 'REFUTED');

const extraLive = foundationDecision([...foundation.AUTHORIZED_KERNEL_PATHS, 'scripts/ci-proof/not-authorized.js']);
assert.equal(extraLive.decision, 'REFUTED');

const wrongBlobDecision = foundationDecision(foundation.AUTHORIZED_KERNEL_PATHS, foundation.CORRECTION_CHANGED_KERNEL_PATHS, {[foundation.REMAINING_BASE_SATISFIED_KERNEL_PATHS[0]]: '0'.repeat(40)});
assert.equal(wrongBlobDecision.decision, 'REFUTED');

const wrongParentDecision = foundationDecision(foundation.AUTHORIZED_KERNEL_PATHS, foundation.CORRECTION_CHANGED_KERNEL_PATHS, undefined, undefined, {sha: 'c'.repeat(40), parents: [{sha: 'd'.repeat(40)}]});
assert.equal(wrongParentDecision.decision, 'REFUTED');

const closedPullDecision = foundation.foundationBootstrapDecision({
  pull: {...foundationPull(), state: 'closed'},
  result: foundationResult(foundation.AUTHORIZED_KERNEL_PATHS),
  liveBefore: foundationLive(),
  liveAfter: foundationLive(),
  baseCommit: foundationBaseCommit(),
  baseKernelBlobs: {[foundation.REMAINING_BASE_SATISFIED_KERNEL_PATHS[0]]: foundation.EXPECTED_BASE_DRIFT_BLOB_SHAS[foundation.REMAINING_BASE_SATISFIED_KERNEL_PATHS[0]]},
  preCorrectionKernelBlobs: foundationBlobs(),
  correctionChangedPaths: foundation.CORRECTION_CHANGED_KERNEL_PATHS,
});
assert.equal(closedPullDecision.decision, 'REFUTED');

const reusedPullDecision = foundation.foundationBootstrapDecision({
  pull: {...foundationPull(), merged: true, merged_at: '2026-08-31T00:00:00.000Z'},
  result: foundationResult(foundation.AUTHORIZED_KERNEL_PATHS),
  liveBefore: foundationLive(),
  liveAfter: foundationLive(),
  baseCommit: foundationBaseCommit(),
  baseKernelBlobs: {[foundation.REMAINING_BASE_SATISFIED_KERNEL_PATHS[0]]: foundation.EXPECTED_BASE_DRIFT_BLOB_SHAS[foundation.REMAINING_BASE_SATISFIED_KERNEL_PATHS[0]]},
  preCorrectionKernelBlobs: foundationBlobs(),
  correctionChangedPaths: foundation.CORRECTION_CHANGED_KERNEL_PATHS,
});
assert.equal(reusedPullDecision.decision, 'REFUTED');
