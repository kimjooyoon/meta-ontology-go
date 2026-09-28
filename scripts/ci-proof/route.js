'use strict';

const crypto = require('node:crypto');

const schema = 'gooo/ci-proof-route/v1';
const routes = Object.freeze({
  'pull_request:dev': 'feature_dev',
  'pull_request:main': 'promotion_main',
  'push:dev': 'protected_push_dev',
  'push:main': 'protected_push_main',
});

function classifyProofRoute(event, baseRef, input = {}) {
  if (event === 'pull_request' && baseRef === 'main' && input.headRef !== 'dev' && !isPromotionSnapshotHead(input.headRef)) {
    throw new Error('main promotion head must be dev or an exact dev-tree snapshot');
  }
  const route = routes[event + ':' + baseRef];
  if (!route) throw new Error('unsupported CI proof route tuple');
  return route;
}

function isPromotionSnapshotHead(headRef) {
  const match = typeof headRef === 'string' && headRef.match(/^agent\/main-promotion-snapshot-([0-9a-f]{40})$/);
  return Boolean(match && match[1] !== '0'.repeat(40));
}

function buildProofRouteEvidence(input) {
  const payload = {
    schema,
    event: input.event,
    event_ref: input.eventRef,
    base_ref: input.baseRef,
    head_sha: input.headSha,
    route: classifyProofRoute(input.event, input.baseRef, input),
  };
  const digest = crypto.createHash('sha256').update(JSON.stringify(payload)).digest('hex');
  return {...payload, digest: 'sha256:' + digest};
}

function validateProofRouteEvidence(evidence, input) {
  const expected = buildProofRouteEvidence(input);
  if (JSON.stringify(evidence) !== JSON.stringify(expected)) {
    throw new Error('proof route evidence is stale, malformed, or unbound');
  }
}

module.exports = {
  buildProofRouteEvidence,
  classifyProofRoute,
  validateProofRouteEvidence,
};
