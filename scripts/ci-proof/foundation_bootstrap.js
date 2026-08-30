'use strict';

const FOUNDATION_BOOTSTRAP_CODE = 'CI-FOUNDATION-BOOTSTRAP-001';
const FOUNDATION_ROUTE = 'foundation_bootstrap_dev_sync';

const AUTHORIZED_KERNEL_PATHS = Object.freeze([
  '.github/agent-scope-table.md',
  '.github/ci-governance.json',
  '.github/workflows/ci.yml',
  '.github/workflows/self-improvement-minimal-loop.yml',
  'internal/verify/foundation_discovery_recovery_scope.go',
  'internal/verify/foundation_promotion.go',
  'internal/verify/governance_part01.go',
  'internal/verify/scope_dev_main_sync_20260831.go',
  'internal/verify/self_improvement_minimal_loop_scope.go',
  'scripts/ci-proof/build_part01.go',
  'scripts/ci-proof/build_part02.go',
  'scripts/ci-proof/build_part03.go',
  'scripts/ci-proof/foundation.go',
  'scripts/ci-proof/foundation_route_test.go',
  'scripts/ci-proof/foundation_types.go',
  'scripts/ci-proof/foundation_validate.go',
  'scripts/ci-proof/proof_types_part01.go',
  'scripts/ci-proof/route.go',
  'scripts/ci-proof/route.js',
  'scripts/ci-proof/route_test.js',
  'scripts/ci-proof/types_part02.go',
  'scripts/ci-proof/types_part05.go',
  'scripts/ci-proof/validate_part01.go',
  'scripts/verify/foundation_promotion.go',
  'scripts/verify/main_part02.go',
]);

const BASE_DRIFT_KERNEL_PATHS = Object.freeze([
  '.github/workflows/ci-guardian.yml',
  'internal/verify/scope_foundation_bootstrap_dev_sync_20260831.go',
  'scripts/ci-proof/foundation_bootstrap.js',
  'scripts/ci-proof/guardian.js',
]);

const ALLOWED_KERNEL_PATHS = Object.freeze([...AUTHORIZED_KERNEL_PATHS, ...BASE_DRIFT_KERNEL_PATHS].sort());

const FOUNDATION_BOOTSTRAP = Object.freeze({
  schema: 'gooo/meta-foundation-bootstrap/v1',
  decision: 'FOUNDATION',
  reason: 'NO_AUTHORIZED_MAIN_TO_DEV_PROTECTED_KERNEL_ROUTE',
  repository: 'kimjooyoon/meta-ontology-go',
  pullRequest: 606,
  headRef: 'agent/dev-main-sync-20260831',
  headSha: 'f681f555254785d1cba2b58d4185c04a6fcd895c',
  baseRef: 'dev',
  sourceMainSha: 'a0962dd1ac8376b9e88bb629c66e3a7f710b96a9',
  sourceDevSha: '7879cfea762d986a4ad9b0dc027b41593914388a',
  conflictManifestSHA256: '9d63ca1e020e21cf43ccb671c505264c5ddebffa0265cb6939f06d0f17e1da7d',
  mergeResolutionCommit: 'd357b7c5456805669235135a3ca79e145127bd15',
  proofChoice: 'FOUNDATION',
  allowedKernelPaths: ALLOWED_KERNEL_PATHS,
  authorizedKernelPaths: AUTHORIZED_KERNEL_PATHS,
  baseDriftKernelPaths: BASE_DRIFT_KERNEL_PATHS,
  consume: Object.freeze({
    mode: 'single-use',
    condition: 'one successful merge of exact PR #606 head',
    reuseDecision: 'REFUTED',
    consumedBy: 'pull_request:606',
  }),
  priorGuardianFailures: Object.freeze([
    Object.freeze({runId: 33336066336, headSha: 'd357b7c5456805669235135a3ca79e145127bd15', code: 'CI-ROOT-OF-TRUST-001'}),
    Object.freeze({runId: 33336219219, headSha: '195fe1b1', code: 'CI-ROOT-OF-TRUST-001'}),
    Object.freeze({runId: 33336436629, headSha: 'f681f555254785d1cba2b58d4185c04a6fcd895c', code: 'CI-ROOT-OF-TRUST-001'}),
  ]),
  goooMetaActivity: Object.freeze({
    source: 'examples/self-improvement-minimal-loop/main.gooo',
    activity: 'CAPTURE_RECEIPT',
    input: 'gooo://meta/self-improvement-loop/entity/receipt-input',
    output: 'gooo://meta/self-improvement-loop/entity/receipt',
  }),
});

function exactIdentity(pull) {
  return Boolean(
    pull
      && pull.number === FOUNDATION_BOOTSTRAP.pullRequest
      && pull.base
      && pull.base.ref === FOUNDATION_BOOTSTRAP.baseRef
      && pull.base.repo
      && pull.base.repo.full_name === FOUNDATION_BOOTSTRAP.repository
      && pull.head
      && pull.head.ref === FOUNDATION_BOOTSTRAP.headRef
      && pull.head.sha === FOUNDATION_BOOTSTRAP.headSha
      && pull.head.repo
      && pull.head.repo.full_name === FOUNDATION_BOOTSTRAP.repository,
  );
}

function exactPaths(paths, expected) {
  return Array.isArray(paths)
    && paths.length === expected.length
    && paths.every((path, index) => path === expected[index]);
}

function foundationBootstrapDecision({pull, result, liveBefore, liveAfter}) {
  if (!exactIdentity(pull)) {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap tuple mismatch'};
  }
  if (pull.state !== undefined && pull.state !== 'open') {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap was already consumed or is not open'};
  }
  if (pull.merged === true || (pull.merged_at !== undefined && pull.merged_at !== null)) {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap was already consumed'};
  }
  if (!liveBefore || !liveAfter || !liveBefore.refs || !liveAfter.refs
      || liveBefore.refs.main_sha !== FOUNDATION_BOOTSTRAP.sourceMainSha
      || liveAfter.refs.main_sha !== FOUNDATION_BOOTSTRAP.sourceMainSha
      || liveBefore.refs.dev_sha !== pull.base.sha
      || liveAfter.refs.dev_sha !== pull.base.sha) {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap live refs are not the exact pinned topology'};
  }
  if (!result || result.decision !== 'FAIL_CLOSED' || result.code !== 'CI-ROOT-OF-TRUST-001') {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap did not observe the exact root-of-trust deadlock'};
  }
  if (!exactPaths(result.kernelPaths, ALLOWED_KERNEL_PATHS)) {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap changed kernel paths outside the exact allowed scope'};
  }
  return {
    decision: 'FOUNDATION',
    code: null,
    reason: 'FOUNDATION_OVERRIDE_USED=1',
    schema: FOUNDATION_BOOTSTRAP.schema,
    authorization: FOUNDATION_BOOTSTRAP,
    consumed: false,
    observedKernelPaths: result.kernelPaths,
  };
}

function foundationArtifactIdentity(manifest) {
  return Boolean(
    manifest
      && manifest.pull_request_number === FOUNDATION_BOOTSTRAP.pullRequest
      && manifest.base_repo === FOUNDATION_BOOTSTRAP.repository
      && manifest.base_ref === FOUNDATION_BOOTSTRAP.baseRef
      && manifest.head_repo === FOUNDATION_BOOTSTRAP.repository
      && manifest.head_ref === FOUNDATION_BOOTSTRAP.headRef
      && manifest.head_sha === FOUNDATION_BOOTSTRAP.headSha,
  );
}

module.exports = {
  ALLOWED_KERNEL_PATHS,
  AUTHORIZED_KERNEL_PATHS,
  BASE_DRIFT_KERNEL_PATHS,
  FOUNDATION_BOOTSTRAP,
  FOUNDATION_BOOTSTRAP_CODE,
  FOUNDATION_ROUTE,
  exactIdentity,
  foundationArtifactIdentity,
  foundationBootstrapDecision,
};
