'use strict';

const crypto = require('node:crypto');

const RECONCILIATION_ROUTE = 'reconciliation_main';
const RECONCILIATION_BRANCH_PREFIX = 'agent/main-history-reconciliation-';

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

function isReconciliationPromotion(input) {
  return input.event === 'pull_request' && input.baseRef === 'main' && typeof input.headRef === 'string' && input.headRef.startsWith(RECONCILIATION_BRANCH_PREFIX) && input.headRef.length > RECONCILIATION_BRANCH_PREFIX.length;
}

function classifyProofRoute(event, baseRef, input = {}) {
  if (isFoundationPromotion({...input, event, baseRef})) return 'foundation_promotion';
  if (isReconciliationPromotion({...input, event, baseRef})) return RECONCILIATION_ROUTE;
  if (event === 'pull_request' && baseRef === 'main' && typeof input.headRef === 'string' && input.headRef.startsWith('agent/') && input.headRef !== 'dev' && !input.headRef.startsWith('agent/foundation-discovery-recovery-')) {
    throw new Error('ordinary agent-to-main route is not authorized');
  }
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
  isReconciliationPromotion,
  RECONCILIATION_BRANCH_PREFIX,
  RECONCILIATION_ROUTE,
  validateProofRouteEvidence,
};
