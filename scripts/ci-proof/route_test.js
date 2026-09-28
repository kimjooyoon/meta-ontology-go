'use strict';

const assert = require('node:assert/strict');
const route = require('./route');

const cases = [
  ['pull_request', 'dev', 'feature_dev'],
  ['pull_request', 'main', 'promotion_main'],
  ['push', 'dev', 'protected_push_dev'],
  ['push', 'main', 'protected_push_main'],
];

for (const [event, baseRef, expected] of cases) {
  const input = {
    event,
    eventRef: event === 'push' ? 'refs/heads/' + baseRef : 'refs/pull/7/merge',
    baseRef,
    headSha: 'a'.repeat(40),
  };
  const evidence = route.buildProofRouteEvidence(input);
  assert.equal(evidence.route, expected);
  assert.match(evidence.digest, /^sha256:[0-9a-f]{64}$/);
  assert.doesNotThrow(() => route.validateProofRouteEvidence(evidence, input));
}

assert.throws(
  () => route.classifyProofRoute('workflow_dispatch', 'main'),
  /unsupported CI proof route tuple/,
);

assert.throws(() => route.classifyProofRoute('pull_request', 'main', {headRef: 'agent/unrelated-main-change'}), /main promotion head must be dev/);
