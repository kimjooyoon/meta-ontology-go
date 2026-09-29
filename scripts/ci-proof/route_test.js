'use strict';

const assert = require('node:assert/strict');
const route = require('./route');

const cases = [
  ['pull_request', 'dev', 'feature_dev'],
  ['pull_request', 'main', 'promotion_main', 'dev'],
  ['pull_request', 'main', 'promotion_main', 'agent/main-promotion-snapshot-0123456789abcdef0123456789abcdef01234567'],
  ['push', 'dev', 'protected_push_dev'],
  ['push', 'main', 'protected_push_main'],
];

for (const [event, baseRef, expected, headRef] of cases) {
  const input = {
    event,
    eventRef: event === 'push' ? 'refs/heads/' + baseRef : 'refs/pull/7/merge',
    baseRef,
    headSha: 'a'.repeat(40),
    ...(headRef ? {headRef} : {}),
  };
  const evidence = route.buildProofRouteEvidence(input);
  assert.equal(evidence.route, expected);
  assert.match(evidence.digest, /^sha256:[0-9a-f]{64}$/);
  assert.doesNotThrow(() => route.validateProofRouteEvidence(evidence, input));
}

assert.throws(() => route.classifyProofRoute('workflow_dispatch', 'main'), /main promotion head must be dev or an exact dev-tree snapshot/);

assert.throws(() => route.classifyProofRoute('pull_request', 'main'), /main promotion head must be dev or an exact dev-tree snapshot/);
assert.throws(() => route.classifyProofRoute('pull_request', 'main', {headRef: 'agent/unrelated-main-change'}), /main promotion head must be dev or an exact dev-tree snapshot/);
assert.throws(() => route.classifyProofRoute('pull_request', 'main', {headRef: 'agent/main-promotion-snapshot-' + '0'.repeat(40)}), /main promotion head must be dev or an exact dev-tree snapshot/);
