'use strict';

const crypto = require('node:crypto');

const schema = 'gooo/ci-proof-route/v1';
const foundationPromotion = Object.freeze({
  repository: 'kimjooyoon/meta-ontology-go',
  pullRequest: 602,
  headRef: 'agent/foundation-discovery-recovery-20260830',
  baseRef: 'main',
  baseSha: 'cd9727af80f5118405290d3be96890c18e1529c0',
});
const routes = Object.freeze({
  'pull_request:dev': 'feature_dev',
  'pull_request:main': 'promotion_main',
  'push:dev': 'protected_push_dev',
  'push:main': 'protected_push_main',
});

function isFoundationPromotion(input) {
  return input.event === 'pull_request' && input.repository === foundationPromotion.repository &&
    input.prNumber === foundationPromotion.pullRequest && input.headRef === foundationPromotion.headRef &&
    input.baseRef === foundationPromotion.baseRef && input.baseSha === foundationPromotion.baseSha;
}

function classifyProofRoute(event, baseRef, input = {}) {
  if (isFoundationPromotion({...input, event, baseRef})) return 'foundation_promotion';
  const route = routes[event + ':' + baseRef];
  if (!route) throw new Error('unsupported CI proof route tuple');
  return route;
}

function buildProofRouteEvidence(input) {
  const payload = {
    schema,
    event: input.event,
    event_ref: input.eventRef,
    base_ref: input.baseRef,
    head_sha: input.headSha,
    route: classifyProofRoute(input.event, input.baseRef, input),
    guardian_required: input.event === 'pull_request' && input.baseRef === 'main' && !isFoundationPromotion(input),
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
  foundationPromotion,
  isFoundationPromotion,
  validateProofRouteEvidence,
};
