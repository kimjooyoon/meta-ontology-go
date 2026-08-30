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
const PRE_CORRECTION_BASE_SNAPSHOT = '2b1d331c560a58c30def1bb8a5be3f66740bf954';
const CORRECTION_CHANGED_KERNEL_PATHS = Object.freeze([
  '.github/workflows/ci-guardian.yml',
  'scripts/ci-proof/foundation_bootstrap.js',
  'scripts/ci-proof/guardian.js',
  'scripts/ci-proof/route_test.js',
].sort());
const CORRECTION_CHANGED_BASE_DRIFT_KERNEL_PATHS = Object.freeze(CORRECTION_CHANGED_KERNEL_PATHS.filter((path) => BASE_DRIFT_KERNEL_PATHS.includes(path)));
const REMAINING_BASE_SATISFIED_KERNEL_PATHS = Object.freeze(BASE_DRIFT_KERNEL_PATHS.filter((path) => !CORRECTION_CHANGED_BASE_DRIFT_KERNEL_PATHS.includes(path)));
const EXPECTED_LIVE_KERNEL_PATHS = Object.freeze([...new Set([...AUTHORIZED_KERNEL_PATHS, ...CORRECTION_CHANGED_BASE_DRIFT_KERNEL_PATHS])].sort());
const EXPECTED_BASE_DRIFT_BLOB_SHAS = Object.freeze({
  '.github/workflows/ci-guardian.yml': 'd402078c1bb0c97c08755b7c65c814e15e34d567',
  'internal/verify/scope_foundation_bootstrap_dev_sync_20260831.go': '4b31c2e52d2d3e6d4c22881c520db08fc05d1431',
  'scripts/ci-proof/foundation_bootstrap.js': 'c2afa1e03e0c6e3098ddb09971d0c1384b47607a',
  'scripts/ci-proof/guardian.js': '6e4a4a8189971ad0409dab5f8dec869fd1022763',
});
const FOUNDATION_OVERRIDE_COUNT = 2;
const FOUNDATION_OVERRIDE_MARKER = `FOUNDATION_OVERRIDE_USED=${FOUNDATION_OVERRIDE_COUNT}`;

const FOUNDATION_BOOTSTRAP = Object.freeze({
  schema: 'gooo/meta-foundation-bootstrap/v1',
  decision: 'FOUNDATION',
  reason: 'NO_AUTHORIZED_MAIN_TO_DEV_PROTECTED_KERNEL_ROUTE',
  repository: 'kimjooyoon/meta-ontology-go',
  pullRequest: 606,
  headRef: 'agent/dev-main-sync-20260831',
  headSha: 'f681f555254785d1cba2b58d4185c04a6fcd895c',
  baseRef: 'dev',
  preCorrectionBaseSnapshot: PRE_CORRECTION_BASE_SNAPSHOT,
  sourceMainSha: 'a0962dd1ac8376b9e88bb629c66e3a7f710b96a9',
  sourceDevSha: '7879cfea762d986a4ad9b0dc027b41593914388a',
  conflictManifestSHA256: '9d63ca1e020e21cf43ccb671c505264c5ddebffa0265cb6939f06d0f17e1da7d',
  mergeResolutionCommit: 'd357b7c5456805669235135a3ca79e145127bd15',
  proofChoice: 'FOUNDATION',
  allowedKernelPaths: ALLOWED_KERNEL_PATHS,
  authorizedKernelPaths: AUTHORIZED_KERNEL_PATHS,
  baseDriftKernelPaths: BASE_DRIFT_KERNEL_PATHS,
  correctionChangedKernelPaths: CORRECTION_CHANGED_KERNEL_PATHS,
  correctionChangedBaseDriftKernelPaths: CORRECTION_CHANGED_BASE_DRIFT_KERNEL_PATHS,
  remainingBaseSatisfiedKernelPaths: REMAINING_BASE_SATISFIED_KERNEL_PATHS,
  expectedLiveKernelPaths: EXPECTED_LIVE_KERNEL_PATHS,
  expectedBaseDriftBlobSHAs: EXPECTED_BASE_DRIFT_BLOB_SHAS,
  foundationOverrideCount: FOUNDATION_OVERRIDE_COUNT,
  foundationOverrideMarker: FOUNDATION_OVERRIDE_MARKER,
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

function exactBaseDriftBlobSHAs(baseKernelBlobs, expectedPaths = BASE_DRIFT_KERNEL_PATHS) {
  return Boolean(
    baseKernelBlobs
      && typeof baseKernelBlobs === 'object'
      && !Array.isArray(baseKernelBlobs)
      && JSON.stringify(Object.keys(baseKernelBlobs).sort()) === JSON.stringify(expectedPaths.slice().sort())
      && expectedPaths.every((path) => baseKernelBlobs[path] === EXPECTED_BASE_DRIFT_BLOB_SHAS[path]),
  );
}

function exactCorrectionBaseCommit(baseCommit, baseSHA) {
  return Boolean(
    baseCommit
      && baseCommit.sha === baseSHA
      && Array.isArray(baseCommit.parents)
      && baseCommit.parents.length === 1
      && baseCommit.parents[0]
      && baseCommit.parents[0].sha === PRE_CORRECTION_BASE_SNAPSHOT,
  );
}

function foundationBootstrapDecision({pull, result, liveBefore, liveAfter, baseCommit, baseKernelBlobs, preCorrectionKernelBlobs, correctionChangedPaths}) {
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
      || liveAfter.refs.dev_sha !== pull.base.sha
      || !exactCorrectionBaseCommit(baseCommit, pull.base.sha)) {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap live refs are not the exact pinned topology'};
  }
  if (!result || result.decision !== 'FAIL_CLOSED' || result.code !== 'CI-ROOT-OF-TRUST-001') {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap did not observe the exact root-of-trust deadlock'};
  }
  if (!exactPaths(Array.isArray(correctionChangedPaths) ? correctionChangedPaths.slice().sort() : [], CORRECTION_CHANGED_KERNEL_PATHS)) {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation correction changed paths are not exact'};
  }
  const pullKernelPaths = Array.isArray(result.kernelPaths) ? [...new Set(result.kernelPaths)].sort() : [];
  const correctionChangedBaseDriftPaths = CORRECTION_CHANGED_BASE_DRIFT_KERNEL_PATHS.filter((path) => correctionChangedPaths.includes(path));
  const liveKernelPaths = [...new Set([...pullKernelPaths, ...correctionChangedBaseDriftPaths])].sort();
  const alreadySatisfiedByBase = BASE_DRIFT_KERNEL_PATHS.filter((path) => !correctionChangedBaseDriftPaths.includes(path));
  const disjoint = liveKernelPaths.every((path) => !alreadySatisfiedByBase.includes(path));
  const exactUnion = [...liveKernelPaths, ...alreadySatisfiedByBase].sort();
  if (!exactPaths(pullKernelPaths, AUTHORIZED_KERNEL_PATHS)
      || !exactPaths(liveKernelPaths, EXPECTED_LIVE_KERNEL_PATHS)
      || !disjoint
      || !exactPaths(alreadySatisfiedByBase, REMAINING_BASE_SATISFIED_KERNEL_PATHS)
      || !exactPaths(exactUnion, ALLOWED_KERNEL_PATHS)) {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap live scope and base-satisfied scope are not exact'};
  }
  if (!exactBaseDriftBlobSHAs(preCorrectionKernelBlobs)
      || !exactBaseDriftBlobSHAs(baseKernelBlobs, REMAINING_BASE_SATISFIED_KERNEL_PATHS)) {
    return {decision: 'REFUTED', code: FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap remaining base-satisfied blob attestation is not exact'};
  }
  return {
    decision: 'FOUNDATION',
    code: null,
    reason: FOUNDATION_OVERRIDE_MARKER,
    schema: FOUNDATION_BOOTSTRAP.schema,
    authorization: FOUNDATION_BOOTSTRAP,
    consumed: false,
    observedKernelPaths: liveKernelPaths,
    pullKernelPaths,
    alreadySatisfiedByBase,
    correctionChangedPaths: correctionChangedPaths.slice().sort(),
    baseCommitParentSha: baseCommit.parents[0].sha,
    preCorrectionBaseSnapshot: PRE_CORRECTION_BASE_SNAPSHOT,
    baseKernelBlobs,
    preCorrectionKernelBlobs,
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
  CORRECTION_CHANGED_BASE_DRIFT_KERNEL_PATHS,
  CORRECTION_CHANGED_KERNEL_PATHS,
  EXPECTED_BASE_DRIFT_BLOB_SHAS,
  EXPECTED_LIVE_KERNEL_PATHS,
  FOUNDATION_BOOTSTRAP,
  FOUNDATION_BOOTSTRAP_CODE,
  FOUNDATION_OVERRIDE_COUNT,
  FOUNDATION_OVERRIDE_MARKER,
  FOUNDATION_ROUTE,
  PRE_CORRECTION_BASE_SNAPSHOT,
  REMAINING_BASE_SATISFIED_KERNEL_PATHS,
  exactBaseDriftBlobSHAs,
  exactCorrectionBaseCommit,
  exactIdentity,
  foundationArtifactIdentity,
  foundationBootstrapDecision,
};
