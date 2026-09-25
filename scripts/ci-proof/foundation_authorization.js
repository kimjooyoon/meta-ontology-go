'use strict';

const crypto = require('node:crypto');

const AUTHORIZATION_SCHEMA = 'gooo/meta-foundation-authorization/v1';
const FOUNDATION_OVERRIDE_SUCCESS_COUNT = 3;
const FOUNDATION_OVERRIDE_MARKER = `FOUNDATION_OVERRIDE_SUCCESS_COUNT=${FOUNDATION_OVERRIDE_SUCCESS_COUNT}`;
const AUTHORIZATION_BRANCH = 'agent/foundation-authorization-dev-sync-20260831';
const CANDIDATE_BRANCH = 'agent/dev-main-sync-20260831-rerun';
const CANDIDATE_PULL_REQUEST = 609;
const REPOSITORY = 'kimjooyoon/meta-ontology-go';
const REGRESSION_REPAIR_SCHEMA = 'gooo/ci-governance-denominator-migration/v2';
const REGRESSION_REPAIR_PATH = '.github/governance-denominator-v2-migration.json';
const REGRESSION_REPAIR_BRANCH = 'agent/foundation-regression-repair-20260831';
const REGRESSION_REPAIR_PULL_REQUEST = 611;
const REGRESSION_REPAIR_BASE_SHA = 'afa388a2d4f1a482c71f3dfdb1c5937a00d1eca3';
const REGRESSION_REPAIR_REASON = 'BASE_GUARDIAN_DIGEST_NULL_FOR_AUTHORIZED_FEATURE_PATH';
const REGRESSION_REPAIR_CHANGED_PATHS = Object.freeze([
  '.github/agent-scope-table.md',
  '.github/ci-governance.json',
  '.github/governance-denominator-v2-migration.json',
  '.github/workflows/ci-guardian.yml',
  'internal/verify/scope_foundation_regression_repair_20260831.go',
  'scripts/ci-proof/foundation_authorization.js',
  'scripts/ci-proof/foundation_authorization_test.js',
  'scripts/ci-proof/guardian_test.js',
  'scripts/ci-proof/guardian.js',
].sort());
const REGRESSION_REPAIR_EXCLUDED_PATHS = Object.freeze([REGRESSION_REPAIR_PATH]);
const REGRESSION_REPAIR_OLD_FAILURE = 'CI-ROOT-OF-TRUST-001: guardian artifact validation failed: CI-ROOT-OF-TRUST-001: guardian artifact kernel digest fields are inconsistent';
const CORRECTION_CHILD_SCHEMA = 'gooo/ci-governance-denominator-migration/v3';
const CORRECTION_CHILD_PATH = '.github/governance-denominator-v3-correction.json';
const CORRECTION_CHILD_BRANCH = 'agent/foundation-correction-child-20260831';
const CORRECTION_CHILD_PULL_REQUEST = 612;
const CORRECTION_CHILD_BASE_SHA = '0a640a796c7bb39d81f411ecb5bbb3f223b4ed1f';
const CORRECTION_CHILD_PARENT_RECEIPT_SHA256 = 'sha256:bb94d6ff8ef5467f44affad7fb45639239a8dde09020581d706f367645bc5335';
const CORRECTION_CHILD_REASON = 'COMPUTED_KERNEL_DIGEST_NOT_PROPAGATED_TO_PASS_ARTIFACT';
const CORRECTION_CHILD_CHANGED_PATHS = Object.freeze([
  '.github/agent-scope-table.md',
  '.github/ci-governance.json',
  '.github/governance-denominator-v2-migration.json',
  '.github/governance-denominator-v3-correction.json',
  '.github/workflows/ci-guardian.yml',
  'internal/verify/scope_foundation_correction_child_20260831.go',
  'scripts/ci-proof/foundation_authorization.js',
  'scripts/ci-proof/foundation_authorization_test.js',
  'scripts/ci-proof/guardian.js',
  'scripts/ci-proof/guardian_test.js',
].sort());
const CORRECTION_CHILD_EXCLUDED_PATHS = Object.freeze([CORRECTION_CHILD_PATH]);
const CORRECTION_CHILD_OLD_FAILURE = REGRESSION_REPAIR_OLD_FAILURE;
const CORRECTION_CHILD_OWNED_OUTCOME_FIELDS = Object.freeze(['parent_outcome', 'outcome', 'causal']);
const SCHEMA_COHERENCE_MIGRATION_SCHEMA = 'gooo/ci-governance-denominator-migration/v4';
const SCHEMA_COHERENCE_MIGRATION_PATH = '.github/governance-denominator-v4-schema-coherence.json';
const SCHEMA_COHERENCE_MIGRATION_BRANCH = 'agent/schema-coherence-migration-adoption-20260831';
const SCHEMA_COHERENCE_MIGRATION_PULL_REQUEST = 613;
const SCHEMA_COHERENCE_MIGRATION_BASE_SHA = '8bcb771b1d5cd84ee55900e42535d3c2d49545c4';
const SCHEMA_COHERENCE_MIGRATION_REASON = 'VERSIONED_CHILD_OWNS_NEW_OUTCOME_FIELDS';
const SCHEMA_COHERENCE_MIGRATION_PARENT_RECEIPTS = Object.freeze({
  v2: 'sha256:587f52d3917a2b0b54cbfb93859e7cbffa541b606ec792b6e68b1165f6cdf8db',
  v3: 'sha256:36bae817572ead0a449930f23e36ea9cb0aa4ab2ef61ce64d329b61841977571',
});
const SCHEMA_COHERENCE_MIGRATION_INPUT_FIELDS = Object.freeze(['target_commit', 'protocol', 'release_asset', 'main_run', 'adoption_proposal', 'proposal']);
const SCHEMA_COHERENCE_MIGRATION_INPUT = Object.freeze({
  schema: 'gooo/receipt-schema-migration/v0.1.1',
  target_commit: '69efc70d91dbd39420bbb214d4ae64887bf5e237',
  protocol: Object.freeze({rest: Object.freeze({immutable: true}), graphql: Object.freeze({immutable: true})}),
  release_asset: Object.freeze({name: 'gooo-receipt-schema-migration-v0.1.1-conformance.tar.gz', size_bytes: 18371, sha256: 'sha256:5bb52a3da96e8aabad76c27cbc900c71a4fbd3346c0b330804cb55fa2a9d3dd5'}),
  main_run: Object.freeze({run_id: 33353396559, conclusion: 'success', artifact: Object.freeze({artifact_id: 9744345755, size_bytes: 32262, sha256: 'sha256:eddada2591a81c8480e12f2d508ba89642df4852ad9b76a0de0666aa79063e7c'})}),
  adoption_proposal: Object.freeze({path: 'adoption-proposal.json', sha256: 'sha256:7f7edfd4e4e4a8622b20f15868e96d6707f3b44f1fab183f5c2e466f854be106'}),
  proposal: Object.freeze({sha256: 'sha256:182d99c99985ad49e5a7365e1f975b9662cf020a44bd335f45cdf6629b7578b0'}),
});
const SCHEMA_COHERENCE_MIGRATION_CHANGED_PATHS = Object.freeze([
  '.github/agent-scope-table.md',
  '.github/ci-governance.json',
  '.github/governance-denominator-v4-schema-coherence.json',
  '.github/workflows/ci.yml',
  'internal/verify/scope_schema_coherence_migration_adoption_20260831.go',
  'scripts/ci-proof/foundation_authorization.js',
  'scripts/ci-proof/foundation_authorization_test.js',
  'scripts/ci-proof/guardian.js',
  'scripts/ci-proof/guardian_test.js',
].sort());
const SCHEMA_COHERENCE_MIGRATION_EXCLUDED_PATHS = Object.freeze([SCHEMA_COHERENCE_MIGRATION_PATH]);
const EXECUTABLE_GUARDIAN_SCOPE_SCHEMA = 'gooo/ci-governance-denominator-migration/v5';
const EXECUTABLE_GUARDIAN_SCOPE_PATH = '.github/governance-denominator-v5-executable-guardian-scope.json';
const EXECUTABLE_GUARDIAN_SCOPE_BRANCH = 'agent/executable-guardian-scope-adoption-20260831';
const EXECUTABLE_GUARDIAN_SCOPE_PULL_REQUEST = 614;
const EXECUTABLE_GUARDIAN_SCOPE_BASE_SHA = '7f45792e3c23100cbb10cca8b229132060982a7b';
const EXECUTABLE_GUARDIAN_SCOPE_PARENT_V4_RECEIPT_SHA256 = 'sha256:13d14a8501938cc244b431d0c1a0b321cb610cfd14a09767f2e28d8c1652b370';
const EXECUTABLE_GUARDIAN_SCOPE_PARENT_OUTCOME = 'REFUTED_REFERENCE_ERROR_BEFORE_DIGEST_SCOPE';
const EXECUTABLE_GUARDIAN_SCOPE_REASON = 'INITIALIZE_DIGESTS_BEFORE_POLICY_BRANCH';
const EXECUTABLE_GUARDIAN_SCOPE_PROTECTED_FILES = Object.freeze([
  '.github/agent-scope-table.md',
  '.github/branch-policy.md',
  '.github/ci-governance.json',
  '.github/conformance-plan.md',
  '.github/foundation-authorization.json',
  'go.mod',
  'go.sum',
]);
const EXECUTABLE_GUARDIAN_SCOPE_ACCEPTANCE_IDS = Object.freeze([
  'SCOPE_INITIALIZED_BEFORE_POLICY_BRANCH',
  'SCOPE_REUSED_WITHOUT_REDECLARATION',
  'WORKFLOW_AUTHORITY_PINNED_TO_GITHUB_WORKFLOW_SHA',
  'LIVE_PR_CHANGED_PATHS_ATTESTED',
  'PASS_KERNEL_DIGESTS_NON_NULL_EXACT',
  'NULL_STALE_MISMATCH_REFUTED',
  'FUTURE_SCHEMA_UNKNOWN_OVER_6_FIELDS',
  'REFERENCE_ERROR_REFUTED_SCOPE_CLOSED',
]);
const EXECUTABLE_GUARDIAN_SCOPE_LINEAGE = Object.freeze([
  Object.freeze({version: 'v0.2.0', preserved: true, adopted: false}),
  Object.freeze({version: 'v0.2.1', preserved: true, adopted: false}),
]);
const EXECUTABLE_GUARDIAN_SCOPE_INPUT_FIELDS = Object.freeze(['target_commit', 'tag_object_sha', 'protocol', 'release_asset', 'main_run', 'proposal']);
const EXECUTABLE_GUARDIAN_SCOPE_INPUT = Object.freeze({
  schema: 'gooo/receipt-schema-migration/v0.2.2',
  target_commit: '977e622db99c16fbe37db5912b07f403cd09cdb2',
  tag_object_sha: 'c090f22191cd79a6d876ed26df084ac1d4720f3f',
  protocol: Object.freeze({
    rest: Object.freeze({immutable: true}),
    graphql: Object.freeze({immutable: true}),
    repository_setting: Object.freeze({enabled: true}),
  }),
  release_asset: Object.freeze({name: 'gooo-receipt-schema-migration-v0.2.2-conformance.tar.gz', size_bytes: 28877, sha256: 'sha256:e4a2cb8acd608141bdcdb66db6f6369a9480fc691d8a67b3572dd711d02dadf3'}),
  main_run: Object.freeze({run_id: 33357554531, conclusion: 'success', artifact: Object.freeze({artifact_id: 9745614408, size_bytes: 41821, sha256: 'sha256:6e3989aaf21e760f57868a756847dbc9b75824f3b0b8c47ed984e53f12f40146'})}),
  proposal: Object.freeze({file_sha256: 'sha256:f55a204da6e258f1345a52e5e9f164226eff4b6cafa8ba3a65daf97f2247e451', declared_sha256: 'sha256:54cde81ff704cc6afe10f3dfaf0d2dbca2bb29eb18d61a29efe9c9dc4d6d718e'}),
});
const EXECUTABLE_GUARDIAN_SCOPE_CHANGED_PATHS = Object.freeze([
  '.github/agent-scope-table.md',
  '.github/ci-governance.json',
  EXECUTABLE_GUARDIAN_SCOPE_PATH,
  '.github/workflows/ci-guardian.yml',
  'internal/verify/scope_executable_guardian_scope_adoption_20260831.go',
  'scripts/ci-proof/foundation_authorization.js',
  'scripts/ci-proof/foundation_authorization_test.js',
  'scripts/ci-proof/guardian.js',
  'scripts/ci-proof/guardian_test.js',
].sort());
const EXECUTABLE_GUARDIAN_SCOPE_EXCLUDED_PATHS = Object.freeze([EXECUTABLE_GUARDIAN_SCOPE_PATH]);
const EXECUTABLE_GUARDIAN_SCOPE_PRIOR_FAILURE = Object.freeze({
  run_id: 33355380192,
  job_id: 99376387819,
  code: 'CI-ROOT-OF-TRUST-001',
  message: 'ReferenceError: beforeDigest is not defined',
});
const GUARDIAN_DISPATCH_POLICY_SCHEMA = 'gooo/meta-foundation-authorization/v2';
const GUARDIAN_DISPATCH_SCHEMA = 'gooo/ci-governance-denominator-migration/v6';
const GUARDIAN_DISPATCH_PATH = '.github/governance-denominator-v6-foundation-authorization-dispatch.json';
const GUARDIAN_DISPATCH_POLICY_PATH = '.github/foundation-authorization.json';
const GUARDIAN_DISPATCH_BRANCH = 'agent/dev-main-sync-20260831-rerun';
const GUARDIAN_DISPATCH_PULL_REQUEST = 609;
const GUARDIAN_DISPATCH_BASE_SHA = 'e440cbc99f24ceb8385f1b89c70f8cdada10cdbb';
const GUARDIAN_DISPATCH_HEAD_SHA = '8b47db349315c02933296423b0ae7fa80ffeb1dc';
const GUARDIAN_DISPATCH_MERGE_BASE_SHA = 'bc5dc21788aa4c7d46d1f8ab516f8218bb423fdc';
const GUARDIAN_DISPATCH_CHANGED_FILE_COUNT = 92;
const GUARDIAN_DISPATCH_CHANGED_PATHS_SHA256 = 'sha256:7d50d27859d8e755edcce625c2cef34c5902528b19a3a4ebf73bd59dac296cac';
const GUARDIAN_DISPATCH_PROTECTED_INTERSECTION_COUNT = 26;
const GUARDIAN_DISPATCH_PROTECTED_INTERSECTION_SHA256 = 'sha256:204b3569310bffb0d3a6fafaddfec5930e99309281badd768b2c52dfbc07f3bf';
const GUARDIAN_DISPATCH_PROTECTED_PATHS = Object.freeze([
  '.github/agent-scope-table.md',
  '.github/ci-governance.json',
  '.github/workflows/ci-guardian.yml',
  '.github/workflows/ci.yml',
  '.github/workflows/self-improvement-minimal-loop.yml',
  '.github/workflows/transformation-effect.yml',
  'internal/verify/foundation_discovery_recovery_scope.go',
  'internal/verify/foundation_promotion.go',
  'internal/verify/scope_dev_main_sync_20260831_rerun.go',
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
].sort());
const GUARDIAN_DISPATCH_KERNEL_BEFORE_SHA256 = 'sha256:e3b0cd0fd2d95a0113b9a6755ebe318e856225630106fa00ccf24572fdfb69f6';
const GUARDIAN_DISPATCH_KERNEL_AFTER_SHA256 = 'sha256:b1e865e5c2f48735dc693b94807b8a8439819985353503ccffceecedac88d582';
const GUARDIAN_DISPATCH_PARENT_RECEIPT_SHA256 = 'sha256:ba2f25fd18cf161eaf973f2d570e87d161fb452686e3840f7a1b694e30f35f52';
const GUARDIAN_DISPATCH_SOURCE_DIGEST = 'sha256:ab3543bd805a53b6fa8a8e3fb6f10697e3b6f0db8658e471a14906718a137ed3';
const GUARDIAN_DISPATCH_CONTRACT_DIGEST = 'sha256:4384eb0f553e83b69c82ef25fa946f50a5b6c8782b400c7993173caaf876acb7';
const GUARDIAN_DISPATCH_SEMANTIC_IR_DIGEST = 'sha256:fc3e7442c4eb9c2b0298d7db0982ef8cb6ee26fe683f459783c0197466a62f5f';
const GUARDIAN_DISPATCH_VALIDATOR_DIGEST = 'sha256:f45978d7ffb5cbbb6a8fe9af5122ff10d989a8105219679078f1ad5acdf066a2';
const GUARDIAN_DISPATCH_HARNESS_DIGEST = 'sha256:d954abca510f9a7e99c70226d3673db89acb995dc40839973f6202447f7082f3';
const GUARDIAN_DISPATCH_PROPOSAL_FILE_DIGEST = 'sha256:578562ec738484c67d70581622981471f4a0eba86b35a726bc73714805fabb55';
const GUARDIAN_DISPATCH_PROPOSAL_DECLARED_DIGEST = 'sha256:42ae271ea27226bbbf0086ec1fed230e51787b8f44dc68468eafa5607c5a6c7e';
const GUARDIAN_DISPATCH_MARKER = 'FOUNDATION_AUTHORIZATION_DISPATCHED_BEFORE_KERNEL_ATTESTATION';
const GUARDIAN_DISPATCH_CELL_IDS = Object.freeze([
  'PROTECTED_PATH_AUTHORIZATION_DISPATCH',
  'FOUNDATION_RECEIPT_EVALUATION',
  'CHANGED_PATH_TUPLE_BINDING',
  'UNAUTHORIZED_PROTECTED_PATH_FAIL_CLOSED',
]);
const GUARDIAN_DISPATCH_LINEAGE = Object.freeze([
  Object.freeze({version: 'v0.1.0', preserved: true, immutable: false, decision: 'REFUTED'}),
  Object.freeze({version: 'v0.1.1', preserved: true, immutable: true, decision: 'CLOSED'}),
  Object.freeze({version: 'v0.2.0', preserved: true, immutable: true, decision: 'CLOSED'}),
  Object.freeze({version: 'v0.2.1', preserved: true, immutable: true, decision: 'CLOSED'}),
  Object.freeze({version: 'v0.2.2', preserved: true, immutable: true, decision: 'CLOSED'}),
  Object.freeze({version: 'v0.3.0', preserved: true, immutable: true, decision: 'CLOSED', release_id: 379538725, tag_object_sha: '4731fd84d3273ecb04e981066e9add9e6c6ff25b', target_commit: '41fd8cb827b88069b13caf271e2972a35d8ad1d4', asset: Object.freeze({id: 537535145, name: 'gooo-receipt-schema-migration-v0.3.0-conformance-proper.tar.gz', size_bytes: 35456, sha256: 'sha256:ca7c03961c74f10d16d6369e6900b7bc681bfd1cbb04523156d3f372bf2eb39a'})}),
]);
const AUTHORIZATION_PATHS = Object.freeze([
  '.github/agent-scope-table.md',
  '.github/ci-governance.json',
  '.github/foundation-authorization.json',
  '.github/workflows/ci-guardian.yml',
  '.github/workflows/ci.yml',
  'internal/verify/foundation_authorization.go',
  'internal/verify/scope_foundation_authorization_dev_sync_20260831.go',
  'scripts/ci-proof/foundation_authorization.js',
  'scripts/ci-proof/foundation_authorization_test.js',
  'scripts/ci-proof/guardian.js',
].sort());

function validSHA(value) {
  return typeof value === 'string' && /^[0-9a-f]{40}$/.test(value);
}

function validDigest(value) {
  return typeof value === 'string' && /^sha256:[0-9a-f]{64}$/.test(value);
}

function sha256(value) {
  return `sha256:${crypto.createHash('sha256').update(value).digest('hex')}`;
}

function exactArray(left, right) {
  return JSON.stringify(left) === JSON.stringify(right);
}

function canonicalPathNames(files) {
  const names = (Array.isArray(files) ? files : []).flatMap((file) => {
    if (typeof file === 'string') return [file];
    return [file && file.filename, file && file.previous_filename];
  }).filter((value) => typeof value === 'string' && value.length > 0);
  return [...new Set(names)].sort();
}

function digestChangedPaths(files) {
  const paths = canonicalPathNames(files);
  return sha256(paths.length === 0 ? '' : `${paths.join('\n')}\n`);
}

function canonicalProtectedIntersection(files) {
  const protectedPaths = new Set(GUARDIAN_DISPATCH_PROTECTED_PATHS);
  const names = (Array.isArray(files) ? files : []).flatMap((file) => {
    if (typeof file === 'string') return [file];
    return [file && file.filename, file && file.previous_filename];
  }).filter((value) => typeof value === 'string' && protectedPaths.has(value));
  return [...new Set(names)].sort();
}

function digestProtectedIntersection(files) {
  const paths = Array.isArray(files) && files.every((file) => typeof file === 'string')
    ? [...new Set(files)].sort()
    : canonicalProtectedIntersection(files);
  return sha256(paths.length === 0 ? '' : `${paths.join('\n')}\n`);
}

function canonicalTreeEntries(entries, excludedPaths = []) {
  const excluded = new Set(excludedPaths);
  return (Array.isArray(entries) ? entries : [])
    .filter((entry) => entry && entry.type === 'blob' && typeof entry.path === 'string' && !excluded.has(entry.path))
    .map((entry) => ({path: entry.path, mode: entry.mode || null, sha: entry.sha}))
    .sort((left, right) => {
      const leftKey = `${left.path}\u0000${left.mode}\u0000${left.sha}`;
      const rightKey = `${right.path}\u0000${right.mode}\u0000${right.sha}`;
      return leftKey < rightKey ? -1 : leftKey > rightKey ? 1 : 0;
    });
}

function digestTreeEntries(entries, excludedPaths = []) {
  const excluded = new Set(excludedPaths);
  const lines = (Array.isArray(entries) ? entries : [])
    .filter((entry) => entry && entry.type === 'blob' && typeof entry.path === 'string' && !excluded.has(entry.path))
    .map((entry) => `${entry.path}\t${entry.mode || null}\t${entry.sha}`)
    .sort();
  return sha256(lines.length === 0 ? '' : `${lines.join('\n')}\n`);
}

function requireExact(condition, message) {
  if (!condition) throw new Error(message);
}

function validateGuardianDispatchPolicy(policy) {
  requireExact(policy && policy.schema === GUARDIAN_DISPATCH_POLICY_SCHEMA, 'Guardian dispatch policy schema is not exact');
  requireExact(policy.decision === 'FOUNDATION' && policy.reason === 'NO_AUTHORIZED_MAIN_TO_DEV_PROTECTED_KERNEL_ROUTE', 'Guardian dispatch policy decision is not exact');
  requireExact(policy.repository === REPOSITORY, 'Guardian dispatch policy repository is not exact');
  requireExact(policy.foundation_override_success_count === FOUNDATION_OVERRIDE_SUCCESS_COUNT && policy.foundation_override_marker === FOUNDATION_OVERRIDE_MARKER, 'Guardian dispatch policy Foundation count is not exact');
  const candidate = policy.candidate;
  requireExact(candidate && candidate.pull_request === GUARDIAN_DISPATCH_PULL_REQUEST && candidate.branch === GUARDIAN_DISPATCH_BRANCH && candidate.base_branch === 'dev', 'Guardian dispatch candidate identity is not exact');
  requireExact(candidate.base_sha === GUARDIAN_DISPATCH_BASE_SHA && candidate.head_sha === GUARDIAN_DISPATCH_HEAD_SHA && candidate.merge_base_sha === GUARDIAN_DISPATCH_MERGE_BASE_SHA, 'Guardian dispatch candidate SHAs are not exact');
  requireExact(candidate.changed_file_count === GUARDIAN_DISPATCH_CHANGED_FILE_COUNT && candidate.changed_paths_sha256 === GUARDIAN_DISPATCH_CHANGED_PATHS_SHA256, 'Guardian dispatch changed-path binding is not exact');
  requireExact(candidate.protected_intersection_count === GUARDIAN_DISPATCH_PROTECTED_INTERSECTION_COUNT && candidate.protected_intersection_sha256 === GUARDIAN_DISPATCH_PROTECTED_INTERSECTION_SHA256, 'Guardian dispatch protected intersection binding is not exact');
  const dispatch = policy.dispatch;
  requireExact(dispatch && dispatch.receipt_path === GUARDIAN_DISPATCH_PATH && dispatch.receipt_sha256 && validDigest(dispatch.receipt_sha256), 'Guardian dispatch receipt binding is malformed');
  requireExact(dispatch.mode === 'single-use' && dispatch.consumed === false && dispatch.replay_decision === 'REFUTED', 'Guardian dispatch receipt consumption is not exact');
  return policy;
}

function validateGuardianDispatchReceipt(receipt, {candidatePull, changedFiles, kernelPaths, mergeBaseSHA} = {}) {
  requireExact(receipt && receipt.schema === GUARDIAN_DISPATCH_SCHEMA && receipt.migration_version === 'v3', 'Guardian dispatch receipt schema is not exact');
  requireExact(receipt.receipt_path === GUARDIAN_DISPATCH_PATH, 'Guardian dispatch receipt path is not exact');
  requireExact(receipt.parent_receipt && receipt.parent_receipt.schema === EXECUTABLE_GUARDIAN_SCOPE_SCHEMA && receipt.parent_receipt.path === EXECUTABLE_GUARDIAN_SCOPE_PATH && receipt.parent_receipt.sha256 === GUARDIAN_DISPATCH_PARENT_RECEIPT_SHA256, 'Guardian dispatch v5 parent lineage is not exact');
  requireExact(receipt.source_release && receipt.source_release.repository === 'kimjooyoon/gooo-receipt-schema-migration' && receipt.source_release.version === 'v0.3.1' && receipt.source_release.release_id === 379540253 && receipt.source_release.immutable === true && receipt.source_release.tag_object_sha === 'd98ea5f2a88bbc3e3bafa573a94a2058aba96e54' && receipt.source_release.target_commit === '7f651f45f548fb38357649578d9984fd77c5a451', 'Guardian dispatch source release is not exact');
  requireExact(receipt.source_release.asset && receipt.source_release.asset.name === 'gooo-receipt-schema-migration-v0.3.1-conformance.tar.gz' && receipt.source_release.asset.id === 537539263 && receipt.source_release.asset.size_bytes === 36114 && receipt.source_release.asset.sha256 === 'sha256:fe00e049fcf4c946fef1a6bc3a7c5295d3e3fa2aed630edcae0039dfcee81a76', 'Guardian dispatch source asset is not exact');
  requireExact(receipt.source_release.main_run && receipt.source_release.main_run.run_id === 33362658147 && receipt.source_release.main_run.conclusion === 'success', 'Guardian dispatch source run is not exact');
  requireExact(receipt.adoption_proposal && receipt.adoption_proposal.path === 'examples/receipt-schema-migration-v3/adoption-proposal.json' && receipt.adoption_proposal.file_sha256 === GUARDIAN_DISPATCH_PROPOSAL_FILE_DIGEST && receipt.adoption_proposal.declared_sha256 === GUARDIAN_DISPATCH_PROPOSAL_DECLARED_DIGEST, 'Guardian dispatch adoption proposal input is not exact');
  requireExact(receipt.semantic_ir && receipt.semantic_ir.path === 'examples/receipt-schema-migration-v3/semantic-ir.json' && receipt.semantic_ir.source_digest === GUARDIAN_DISPATCH_SOURCE_DIGEST && receipt.semantic_ir.contract_digest === GUARDIAN_DISPATCH_CONTRACT_DIGEST && receipt.semantic_ir.file_sha256 === GUARDIAN_DISPATCH_SEMANTIC_IR_DIGEST, 'Guardian dispatch semantic IR binding is not exact');
  requireExact(receipt.generated_evaluator && receipt.generated_evaluator.validator_path === 'examples/receipt-schema-migration-v3/generated/validator.json' && receipt.generated_evaluator.validator_sha256 === GUARDIAN_DISPATCH_VALIDATOR_DIGEST && receipt.generated_evaluator.harness_path === 'examples/receipt-schema-migration-v3/generated/guardian-harness-cases.json' && receipt.generated_evaluator.harness_sha256 === GUARDIAN_DISPATCH_HARNESS_DIGEST, 'Guardian dispatch generated evaluator binding is not exact');
  requireExact(receipt.guardian_harness && receipt.guardian_harness.schema === 'gooo/receipt-schema-migration/guardian-harness/v2' && receipt.guardian_harness.migration_version === 'v3' && receipt.guardian_harness.summary && receipt.guardian_harness.summary.closed_count === 2 && receipt.guardian_harness.summary.unknown_count === 0 && receipt.guardian_harness.summary.refuted_count === 9 && receipt.guardian_harness.summary.foundation_authorization_count === 8 && receipt.guardian_harness.summary.foundation_receipt_count === 8, 'Guardian dispatch harness summary is not exact');
  requireExact(receipt.denominator && receipt.denominator.previous_cell_count === 16 && receipt.denominator.cell_count === 20 && receipt.denominator.added === 4 && receipt.denominator.retired === 0 && receipt.denominator.split === 0 && exactArray(receipt.denominator.added_cell_ids, GUARDIAN_DISPATCH_CELL_IDS), 'Guardian dispatch denominator migration is not exact');
  requireExact(receipt.denominator.stage_counts && JSON.stringify(receipt.denominator.stage_counts) === JSON.stringify({FOUNDATION: 8, COHERENCE: 6, REGRESSION: 6}), 'Guardian dispatch stage counts are not exact');
  requireExact(receipt.denominator.role_counts && JSON.stringify(receipt.denominator.role_counts) === JSON.stringify({DRIVER: 6, OUTCOME: 6, GUARDRAIL: 8}), 'Guardian dispatch role counts are not exact');
  requireExact(Array.isArray(receipt.lineage) && exactArray(receipt.lineage, GUARDIAN_DISPATCH_LINEAGE), 'Guardian dispatch release lineage is not preserved');
  requireExact(receipt.fixture_v3 && receipt.fixture_v3.repository === REPOSITORY && receipt.fixture_v3.ref === 'dev' && receipt.fixture_v3.base_commit === GUARDIAN_DISPATCH_BASE_SHA && receipt.fixture_v3.head_commit === GUARDIAN_DISPATCH_HEAD_SHA && receipt.fixture_v3.merge_base === GUARDIAN_DISPATCH_MERGE_BASE_SHA, 'Guardian dispatch fixture identity is not exact');
  requireExact(receipt.fixture_v3.manifest_path === 'fixtures/guardian-v3-pr609.json' && receipt.fixture_v3.changed_files_count === GUARDIAN_DISPATCH_CHANGED_FILE_COUNT && receipt.fixture_v3.changed_paths_sha256 === GUARDIAN_DISPATCH_CHANGED_PATHS_SHA256 && receipt.fixture_v3.protected_intersection_count === GUARDIAN_DISPATCH_PROTECTED_INTERSECTION_COUNT && receipt.fixture_v3.protected_intersection_sha256 === GUARDIAN_DISPATCH_PROTECTED_INTERSECTION_SHA256, 'Guardian dispatch fixture path tuple evidence is not exact');
  requireExact(Array.isArray(receipt.fixture_v3.changed_path_tuples) && receipt.fixture_v3.changed_path_tuples.length === GUARDIAN_DISPATCH_CHANGED_FILE_COUNT && exactArray(receipt.fixture_v3.changed_path_tuples, [...receipt.fixture_v3.changed_path_tuples].sort()) && digestChangedPaths(receipt.fixture_v3.changed_path_tuples) === GUARDIAN_DISPATCH_CHANGED_PATHS_SHA256, 'Guardian dispatch changed-path tuple list is not exact');
  requireExact(exactArray(receipt.fixture_v3.protected_intersection_paths, GUARDIAN_DISPATCH_PROTECTED_PATHS), 'Guardian dispatch protected intersection paths are not exact');
  requireExact(receipt.fixture_v3.kernel_before_sha256 === GUARDIAN_DISPATCH_KERNEL_BEFORE_SHA256 && receipt.fixture_v3.kernel_after_sha256 === GUARDIAN_DISPATCH_KERNEL_AFTER_SHA256, 'Guardian dispatch kernel digest evidence is not exact');
  requireExact(receipt.fixture_v3.guardian_run_id === 33359548617 && receipt.fixture_v3.guardian_job_id === 99388126433 && receipt.fixture_v3.artifact_id === 9746232159 && receipt.fixture_v3.artifact_sha256 === 'sha256:41ae5d2398a001b16ecd72dba937924897234dd748fdcb5374caee7e70f026a8', 'Guardian dispatch live failure evidence is not exact');
  requireExact(receipt.provenance && receipt.provenance.development_local_go_commands && receipt.provenance.development_local_go_commands.build === 0 && receipt.provenance.development_local_go_commands.gofmt === 0 && receipt.provenance.development_local_go_commands.test === 0 && receipt.provenance.development_local_go_commands.vet === 0 && receipt.provenance.development_local_node_commands === 1 && receipt.provenance.local_guardian_harness_executions === 0 && receipt.provenance.local_conformance_executions === 0 && receipt.provenance.node_command_class === 'NODE_DIGEST_DERIVATION' && receipt.provenance.node_command_count === 1 && receipt.provenance.event_mutation_policy === 'APPEND_ONLY_NO_RESET_DELETE_REWRITE', 'Guardian dispatch current development provenance is not exact');
  const cells = receipt.cells;
  requireExact(Array.isArray(cells) && cells.length === GUARDIAN_DISPATCH_CELL_IDS.length && cells.every((cell, index) => cell && cell.id === GUARDIAN_DISPATCH_CELL_IDS[index] && cell.allowed === 1 && cell.consumed === 1 && cell.replay_decision === 'REFUTED' && cell.outcome === 'CLOSED'), 'Guardian dispatch cells are not exact');
  if (candidatePull !== undefined) {
    requireExact(candidatePull && candidatePull.number === GUARDIAN_DISPATCH_PULL_REQUEST && candidatePull.base && candidatePull.base.repo && candidatePull.base.repo.full_name === REPOSITORY && candidatePull.base.ref === 'dev' && candidatePull.base.sha === GUARDIAN_DISPATCH_BASE_SHA && candidatePull.head && candidatePull.head.repo && candidatePull.head.repo.full_name === REPOSITORY && candidatePull.head.ref === GUARDIAN_DISPATCH_BRANCH && candidatePull.head.sha === GUARDIAN_DISPATCH_HEAD_SHA, 'Guardian dispatch live candidate tuple is not exact');
  }
  if (mergeBaseSHA !== undefined) requireExact(mergeBaseSHA === GUARDIAN_DISPATCH_MERGE_BASE_SHA, 'Guardian dispatch live merge base is not exact');
  if (changedFiles !== undefined) requireExact(Array.isArray(changedFiles) && changedFiles.length === GUARDIAN_DISPATCH_CHANGED_FILE_COUNT && digestChangedPaths(changedFiles) === GUARDIAN_DISPATCH_CHANGED_PATHS_SHA256, 'Guardian dispatch live changed-path tuple digest is not exact');
  if (kernelPaths !== undefined) requireExact(Array.isArray(kernelPaths) && exactArray([...new Set(kernelPaths)].sort(), GUARDIAN_DISPATCH_PROTECTED_PATHS) && digestProtectedIntersection(kernelPaths) === GUARDIAN_DISPATCH_PROTECTED_INTERSECTION_SHA256, 'Guardian dispatch live protected intersection is not exact');
  return receipt;
}

function validateGuardianDispatchAuthorization(result) {
  validateGuardianDispatchReceipt(result);
  requireExact(result.decision === 'PASS' && result.code === null && result.reason === GUARDIAN_DISPATCH_MARKER && result.single_use === true && result.consumed === false && result.replay_decision === 'REFUTED', 'Guardian dispatch authorization result is not exact');
  return result;
}

function validatePolicy(policy) {
  if (policy && policy.schema === GUARDIAN_DISPATCH_POLICY_SCHEMA) return validateGuardianDispatchPolicy(policy);
  requireExact(policy && policy.schema === AUTHORIZATION_SCHEMA, 'foundation authorization schema is not exact');
  requireExact(policy.decision === 'FOUNDATION', 'foundation authorization decision is not FOUNDATION');
  requireExact(policy.reason === 'NO_AUTHORIZED_MAIN_TO_DEV_PROTECTED_KERNEL_ROUTE', 'foundation authorization reason is not exact');
  requireExact(policy.repository === REPOSITORY, 'foundation authorization repository is not exact');
  requireExact(policy.foundation_override_success_count === FOUNDATION_OVERRIDE_SUCCESS_COUNT, 'foundation override success count is not 3');
  requireExact(policy.foundation_override_marker === FOUNDATION_OVERRIDE_MARKER, 'foundation override success marker is not exact');
  const candidate = policy.candidate;
  requireExact(candidate && candidate.pull_request === CANDIDATE_PULL_REQUEST, 'candidate pull request is not exact');
  requireExact(candidate.branch === CANDIDATE_BRANCH, 'candidate branch is not exact');
  requireExact(candidate.base_branch === 'dev' && validSHA(candidate.base_sha), 'candidate base identity is malformed');
  requireExact(validSHA(candidate.head_sha), 'candidate head SHA is malformed');
  requireExact(typeof candidate.manifest_path === 'string' && validDigest(candidate.manifest_sha256), 'candidate manifest binding is malformed');
  requireExact(Number.isInteger(candidate.changed_path_count) && candidate.changed_path_count > 0 && validDigest(candidate.changed_paths_sha256), 'candidate changed-path binding is malformed');
  requireExact(validDigest(candidate.patch_sha256_excluding_authorization_paths), 'candidate patch binding is malformed');
  requireExact(validDigest(candidate.tree_sha256_excluding_authorization_paths), 'candidate tree binding is malformed');
  const authorization = policy.authorization;
  requireExact(authorization && Number.isInteger(authorization.pull_request) && authorization.pull_request > 0, 'authorization pull request is malformed');
  requireExact(authorization.branch === AUTHORIZATION_BRANCH && authorization.base_branch === 'dev', 'authorization branch identity is not exact');
  requireExact(exactArray(authorization.changed_paths, AUTHORIZATION_PATHS), 'authorization changed paths are not exact');
  requireExact(authorization.mode === 'single-use' && authorization.consumed === false && authorization.replay_decision === 'REFUTED', 'authorization consumption policy is not exact');
  requireExact(policy.auth_policy_paths && exactArray(policy.auth_policy_paths, AUTHORIZATION_PATHS), 'authorization policy path exclusion is not exact');
  const attestation = policy.base_attestation;
  requireExact(attestation && attestation.commit === candidate.base_sha && validSHA(attestation.tree_sha), 'candidate base tree attestation is malformed');
  requireExact(Array.isArray(attestation.parents) && attestation.parents.length > 0 && attestation.parents.every(validSHA), 'candidate base parent attestation is malformed');
  return policy;
}

function validateRegressionRepairReceipt(receipt) {
  requireExact(receipt && receipt.schema === REGRESSION_REPAIR_SCHEMA, 'regression repair receipt schema is not exact');
  requireExact(receipt.foundation_override_success_count === FOUNDATION_OVERRIDE_SUCCESS_COUNT, 'regression repair receipt changed FOUNDATION count');
  requireExact(Array.isArray(receipt.cells) && receipt.cells.length === 1, 'regression repair denominator must contain one cell');
  const cell = receipt.cells[0];
  requireExact(cell && cell.id === 'REGRESSION_REPAIR' && cell.meta_operation === 'RepairBaseGuardianDigestAttestation', 'regression repair cell identity is not exact');
  requireExact(cell.proof_choice === 'REGRESSION' && cell.indicator === 'GUARDRAIL', 'regression repair cell classification is not exact');
  requireExact(cell.allowed === 1 && cell.consumed === 1 && cell.replay_decision === 'REFUTED' && receipt.replay_second_use === 'REFUTED', 'regression repair consumption is not single-use');
  requireExact(cell.reason === REGRESSION_REPAIR_REASON, 'regression repair reason is not exact');
  requireExact(cell.pull_request === REGRESSION_REPAIR_PULL_REQUEST && cell.branch === REGRESSION_REPAIR_BRANCH && cell.base_sha === REGRESSION_REPAIR_BASE_SHA, 'regression repair identity is not exact');
  requireExact(Array.isArray(receipt.changed_paths) && exactArray(receipt.changed_paths, REGRESSION_REPAIR_CHANGED_PATHS), 'regression repair changed paths are not exact');
  requireExact(digestChangedPaths(receipt.changed_paths) === receipt.changed_paths_sha256 && validDigest(receipt.changed_paths_sha256), 'regression repair changed-path digest is not exact');
  requireExact(validDigest(receipt.patch_sha256_excluding_receipt) && validDigest(receipt.tree_sha256_excluding_receipt), 'regression repair content digests are malformed');
  requireExact(receipt.receipt_path === REGRESSION_REPAIR_PATH && exactArray(receipt.digest_exclusions, REGRESSION_REPAIR_EXCLUDED_PATHS), 'regression repair digest exclusions are not exact');
  requireExact(receipt.old_guardian_failure && receipt.old_guardian_failure.run_id === 33348091926 && receipt.old_guardian_failure.code === 'CI-ROOT-OF-TRUST-001' && receipt.old_guardian_failure.message === REGRESSION_REPAIR_OLD_FAILURE, 'regression repair old failure tuple is not exact');
  return receipt;
}

function validateIncompletePropagationOutcome(receipt) {
  validateRegressionRepairReceipt(receipt);
  requireExact(receipt.outcome === 'REFUTED_INCOMPLETE_PROPAGATION' && receipt.cells[0].outcome === 'REFUTED_INCOMPLETE_PROPAGATION', 'regression repair incomplete-propagation outcome is not recorded');
  return receipt;
}

function validateCorrectionChildReceipt(receipt) {
  requireExact(receipt && receipt.schema === CORRECTION_CHILD_SCHEMA, 'correction child receipt schema is not exact');
  requireExact(receipt.foundation_override_success_count === FOUNDATION_OVERRIDE_SUCCESS_COUNT, 'correction child receipt changed FOUNDATION count');
  requireExact(receipt.parent_repair_receipt === CORRECTION_CHILD_PARENT_RECEIPT_SHA256, 'correction child parent repair receipt is not exact');
  requireExact(receipt.receipt_path === CORRECTION_CHILD_PATH && exactArray(receipt.digest_exclusions, CORRECTION_CHILD_EXCLUDED_PATHS), 'correction child receipt exclusion is not exact');
  requireExact(Array.isArray(receipt.cells) && receipt.cells.length === 1, 'correction child denominator must contain one cell');
  const cell = receipt.cells[0];
  requireExact(cell && cell.id === 'CORRECTION_CHILD' && cell.meta_operation === 'PropagateKernelDigestToPassArtifact', 'correction child cell identity is not exact');
  requireExact(cell.proof_choice === 'REGRESSION' && cell.indicator === 'GUARDRAIL', 'correction child classification is not exact');
  requireExact(cell.parent_repair_receipt === CORRECTION_CHILD_PARENT_RECEIPT_SHA256, 'correction child cell parent repair receipt is not exact');
  requireExact(cell.reason === CORRECTION_CHILD_REASON, 'correction child reason is not exact');
  requireExact(cell.allowed === 1 && cell.consumed === 1 && cell.replay_decision === 'REFUTED' && receipt.replay_second_use === 'REFUTED', 'correction child consumption is not single-use');
  requireExact(cell.pull_request === CORRECTION_CHILD_PULL_REQUEST && cell.branch === CORRECTION_CHILD_BRANCH && cell.base_sha === CORRECTION_CHILD_BASE_SHA, 'correction child identity is not exact');
  requireExact(Array.isArray(receipt.changed_paths) && exactArray(receipt.changed_paths, CORRECTION_CHILD_CHANGED_PATHS), 'correction child changed paths are not exact');
  requireExact(digestChangedPaths(receipt.changed_paths) === receipt.changed_paths_sha256 && validDigest(receipt.changed_paths_sha256), 'correction child changed-path digest is not exact');
  requireExact(validDigest(receipt.patch_sha256_excluding_receipt) && validDigest(receipt.tree_sha256_excluding_receipt), 'correction child content digests are malformed');
  requireExact(receipt.prior_guardian_failure && receipt.prior_guardian_failure.run_id === 33349646371 && receipt.prior_guardian_failure.code === 'CI-ROOT-OF-TRUST-001' && receipt.prior_guardian_failure.message === CORRECTION_CHILD_OLD_FAILURE, 'correction child prior failure tuple is not exact');
  if (receipt.parent_outcome !== undefined || receipt.outcome !== undefined || receipt.causal !== undefined) {
    requireExact(receipt.parent_outcome === 'REFUTED_INCOMPLETE_PROPAGATION' && receipt.outcome === 'CLOSED' && receipt.causal === SCHEMA_COHERENCE_MIGRATION_REASON, 'correction child outcome ownership is not exact');
  }
  return receipt;
}

function schemaCoherenceInputFieldStates(input) {
  const fields = {};
  const expected = SCHEMA_COHERENCE_MIGRATION_INPUT;
  for (const field of SCHEMA_COHERENCE_MIGRATION_INPUT_FIELDS) {
    const present = input && Object.prototype.hasOwnProperty.call(input, field);
    const exact = present && JSON.stringify(input[field]) === JSON.stringify(expected[field]);
    fields[field] = exact ? 'CLOSED' : 'UNKNOWN';
  }
  return fields;
}

function resolveSchemaCoherenceDecision(decisions) {
  const values = Array.isArray(decisions) ? decisions : [];
  if (values.includes('REFUTED')) return 'REFUTED';
  if (values.includes('UNKNOWN')) return 'UNKNOWN';
  return values.length > 0 && values.every((value) => value === 'CLOSED') ? 'CLOSED' : 'UNKNOWN';
}

function classifySchemaCoherenceInput(input) {
  const fields = schemaCoherenceInputFieldStates(input);
  const schemaSupported = Boolean(input && input.schema === SCHEMA_COHERENCE_MIGRATION_INPUT.schema);
  if (!schemaSupported) {
    for (const field of SCHEMA_COHERENCE_MIGRATION_INPUT_FIELDS) fields[field] = 'UNKNOWN';
  }
  const decision = resolveSchemaCoherenceDecision(Object.values(fields));
  return {
    decision,
    schema: input && input.schema,
    fields,
    field_count: SCHEMA_COHERENCE_MIGRATION_INPUT_FIELDS.length,
    unknown_count: Object.values(fields).filter((value) => value === 'UNKNOWN').length,
  };
}

function validateImmutableReleaseInput(input) {
  const classification = classifySchemaCoherenceInput(input);
  requireExact(classification.decision === 'CLOSED', `immutable receipt-schema release input is ${classification.decision}`);
  requireExact(JSON.stringify(input) === JSON.stringify(SCHEMA_COHERENCE_MIGRATION_INPUT), 'immutable receipt-schema release input is not exact');
  return input;
}

function validateSchemaCoherenceMigrationReceipt(receipt) {
  requireExact(receipt && receipt.schema === SCHEMA_COHERENCE_MIGRATION_SCHEMA, 'schema coherence migration receipt schema is not exact');
  requireExact(receipt.foundation_override_success_count === FOUNDATION_OVERRIDE_SUCCESS_COUNT, 'schema coherence migration receipt changed FOUNDATION count');
  requireExact(receipt.receipt_path === SCHEMA_COHERENCE_MIGRATION_PATH && exactArray(receipt.digest_exclusions, SCHEMA_COHERENCE_MIGRATION_EXCLUDED_PATHS), 'schema coherence migration receipt exclusion is not exact');
  requireExact(receipt.base_sha === SCHEMA_COHERENCE_MIGRATION_BASE_SHA && receipt.merge_parent_sha === SCHEMA_COHERENCE_MIGRATION_BASE_SHA, 'schema coherence migration merge parent is not exact');
  requireExact(Array.isArray(receipt.parent_receipts) && receipt.parent_receipts.length === 2, 'schema coherence migration parent receipt cardinality is not exact');
  requireExact(receipt.parent_receipts[0].schema === REGRESSION_REPAIR_SCHEMA && receipt.parent_receipts[0].sha256 === SCHEMA_COHERENCE_MIGRATION_PARENT_RECEIPTS.v2, 'schema coherence migration v2 parent receipt is not exact');
  requireExact(receipt.parent_receipts[1].schema === CORRECTION_CHILD_SCHEMA && receipt.parent_receipts[1].sha256 === SCHEMA_COHERENCE_MIGRATION_PARENT_RECEIPTS.v3, 'schema coherence migration v3 parent receipt is not exact');
  requireExact(receipt.parent_field_policy && receipt.parent_field_policy.v2 && receipt.parent_field_policy.v2.schema_owned_only === true && receipt.parent_field_policy.v2.outcome_optional === true && receipt.parent_field_policy.v2.raw_json_rewrite === 'REFUTED' && receipt.parent_field_policy.v2.future_field_requirement === 'REFUTED', 'schema coherence v2 parent field policy is not exact');
  requireExact(receipt.parent_field_policy.v3 && receipt.parent_field_policy.v3.child_owned_fields && exactArray(receipt.parent_field_policy.v3.child_owned_fields, CORRECTION_CHILD_OWNED_OUTCOME_FIELDS), 'schema coherence v3 child-owned field policy is not exact');
  requireExact(receipt.previous_release && receipt.previous_release.version === 'v0.1.0' && receipt.previous_release.immutable === false && receipt.previous_release.decision === 'REFUTED', 'schema coherence v0.1.0 record is not exact');
  validateImmutableReleaseInput(receipt.immutable_release_input);
  requireExact(receipt.immutable_release_input.protocol.rest.immutable === true && receipt.immutable_release_input.protocol.graphql.immutable === true, 'schema coherence release protocol is not immutable');
  requireExact(receipt.input_field_count === SCHEMA_COHERENCE_MIGRATION_INPUT_FIELDS.length && receipt.missing_stale_future_schema_decision === 'UNKNOWN', 'schema coherence input denominator is not exact');
  requireExact(Array.isArray(receipt.cells) && receipt.cells.length === 1, 'schema coherence migration denominator must contain one cell');
  const cell = receipt.cells[0];
  requireExact(cell && cell.id === 'SCHEMA_COHERENCE_MIGRATION_ADOPTION' && cell.meta_operation === 'AdoptVersionedReceiptSchema', 'schema coherence migration cell identity is not exact');
  requireExact(cell.proof_choice === 'COHERENCE' && cell.indicator === 'GUARDRAIL', 'schema coherence migration cell classification is not exact');
  requireExact(cell.allowed === 1 && cell.consumed === 1 && cell.replay_decision === 'REFUTED' && receipt.replay_second_use === 'REFUTED', 'schema coherence migration consumption is not single-use');
  requireExact(cell.parent_outcome === 'REFUTED_INCOMPLETE_PROPAGATION' && cell.outcome === 'CLOSED' && cell.causal === SCHEMA_COHERENCE_MIGRATION_REASON, 'schema coherence migration child-owned outcome is not exact');
  requireExact(cell.reason === SCHEMA_COHERENCE_MIGRATION_REASON, 'schema coherence migration reason is not exact');
  requireExact(cell.pull_request === SCHEMA_COHERENCE_MIGRATION_PULL_REQUEST && cell.branch === SCHEMA_COHERENCE_MIGRATION_BRANCH && cell.base_sha === SCHEMA_COHERENCE_MIGRATION_BASE_SHA, 'schema coherence migration identity is not exact');
  requireExact(receipt.repository_gates === 0 && receipt.local_go_gates === 0 && receipt.cross_project_gates === 0, 'schema coherence migration gate counts are not zero');
  requireExact(Array.isArray(receipt.changed_paths) && exactArray(receipt.changed_paths, SCHEMA_COHERENCE_MIGRATION_CHANGED_PATHS), 'schema coherence migration changed paths are not exact');
  requireExact(digestChangedPaths(receipt.changed_paths) === receipt.changed_paths_sha256 && validDigest(receipt.changed_paths_sha256), 'schema coherence migration changed-path digest is not exact');
  requireExact(validDigest(receipt.patch_sha256_excluding_receipt) && validDigest(receipt.tree_sha256_excluding_receipt), 'schema coherence migration content digests are malformed');
  requireExact(receipt.input_release_digests && receipt.input_release_digests.asset === SCHEMA_COHERENCE_MIGRATION_INPUT.release_asset.sha256 && receipt.input_release_digests.main_artifact === SCHEMA_COHERENCE_MIGRATION_INPUT.main_run.artifact.sha256 && receipt.input_release_digests.adoption_proposal === SCHEMA_COHERENCE_MIGRATION_INPUT.adoption_proposal.sha256 && receipt.input_release_digests.proposal === SCHEMA_COHERENCE_MIGRATION_INPUT.proposal.sha256, 'schema coherence input release digests are not exact');
  return receipt;
}

function validateSchemaCoherenceMigration({receipt, parentReceiptBytes, correctionReceiptBytes, candidateBaseSHA, candidateBaseTreeEntries, migrationPull, migrationCommit, migrationCompare, migrationTreeEntries}) {
  validateSchemaCoherenceMigrationReceipt(receipt);
  requireExact(Buffer.isBuffer(parentReceiptBytes) && sha256(parentReceiptBytes) === SCHEMA_COHERENCE_MIGRATION_PARENT_RECEIPTS.v2, 'schema coherence v2 parent raw JSON was rewritten');
  requireExact(Buffer.isBuffer(correctionReceiptBytes) && sha256(correctionReceiptBytes) === SCHEMA_COHERENCE_MIGRATION_PARENT_RECEIPTS.v3, 'schema coherence v3 parent raw JSON was rewritten');
  requireExact(migrationPull && migrationPull.number === SCHEMA_COHERENCE_MIGRATION_PULL_REQUEST, 'schema coherence migration pull request number mismatch');
  requireExact(migrationPull.state === 'closed' && migrationPull.merged === true && validSHA(migrationPull.merge_commit_sha), 'schema coherence migration pull request is not merged exactly once');
  requireExact(migrationPull.head && migrationPull.head.ref === SCHEMA_COHERENCE_MIGRATION_BRANCH && migrationPull.head.repo && migrationPull.head.repo.full_name === REPOSITORY && validSHA(migrationPull.head.sha), 'schema coherence migration pull request head mismatch');
  requireExact(candidateBaseSHA === migrationPull.merge_commit_sha, 'candidate base is not the schema coherence migration merge commit');
  requireExact(migrationCommit && migrationCommit.sha === migrationPull.merge_commit_sha, 'schema coherence migration merge commit SHA mismatch');
  const parents = Array.isArray(migrationCommit.parents) ? migrationCommit.parents.map((parent) => parent && parent.sha) : [];
  requireExact(parents.length === 1 && parents[0] === SCHEMA_COHERENCE_MIGRATION_BASE_SHA, 'schema coherence migration merge parent is not dev@C');
  requireExact(migrationCompare && Array.isArray(migrationCompare.files) && exactArray(canonicalPathNames(migrationCompare.files), receipt.changed_paths), 'schema coherence migration changed paths do not match the receipt');
  requireExact(digestChangedPaths(migrationCompare.files) === receipt.changed_paths_sha256, 'schema coherence migration changed-path evidence does not match the receipt');
  requireExact(digestTreeEntries(migrationTreeEntries, SCHEMA_COHERENCE_MIGRATION_EXCLUDED_PATHS) === receipt.tree_sha256_excluding_receipt, 'schema coherence migration head tree does not match the receipt');
  requireExact(digestTreeEntries(candidateBaseTreeEntries, SCHEMA_COHERENCE_MIGRATION_EXCLUDED_PATHS) === receipt.tree_sha256_excluding_receipt, 'candidate base tree does not match the schema coherence migration receipt');
  return {
    schema: SCHEMA_COHERENCE_MIGRATION_SCHEMA,
    cell: 'SCHEMA_COHERENCE_MIGRATION_ADOPTION',
    proof_choice: 'COHERENCE',
    indicator: 'GUARDRAIL',
    pull_request: SCHEMA_COHERENCE_MIGRATION_PULL_REQUEST,
    branch: SCHEMA_COHERENCE_MIGRATION_BRANCH,
    base_sha: SCHEMA_COHERENCE_MIGRATION_BASE_SHA,
    merge_commit_sha: migrationPull.merge_commit_sha,
    merge_parent_sha: SCHEMA_COHERENCE_MIGRATION_BASE_SHA,
    parent_outcome: 'REFUTED_INCOMPLETE_PROPAGATION',
    outcome: 'CLOSED',
    causal: SCHEMA_COHERENCE_MIGRATION_REASON,
    reason: SCHEMA_COHERENCE_MIGRATION_REASON,
    allowed: 1,
    consumed: 1,
    replay_decision: 'REFUTED',
  };
}

function executableGuardianScopeInputFieldStates(input) {
  const fields = {};
  const expected = EXECUTABLE_GUARDIAN_SCOPE_INPUT;
  for (const field of EXECUTABLE_GUARDIAN_SCOPE_INPUT_FIELDS) {
    const present = input && Object.prototype.hasOwnProperty.call(input, field);
    const exact = present && JSON.stringify(input[field]) === JSON.stringify(expected[field]);
    fields[field] = exact ? 'CLOSED' : 'REFUTED';
  }
  return fields;
}

function resolveExecutableGuardianScopeDecision(decisions) {
  const values = Array.isArray(decisions) ? decisions : [];
  if (values.includes('REFUTED')) return 'REFUTED';
  if (values.includes('UNKNOWN')) return 'UNKNOWN';
  return values.length > 0 && values.every((value) => value === 'CLOSED') ? 'CLOSED' : 'UNKNOWN';
}

function classifyExecutableGuardianScopeInput(input) {
  const schemaSupported = Boolean(input && input.schema === EXECUTABLE_GUARDIAN_SCOPE_INPUT.schema);
  const fields = schemaSupported
    ? executableGuardianScopeInputFieldStates(input)
    : Object.fromEntries(EXECUTABLE_GUARDIAN_SCOPE_INPUT_FIELDS.map((field) => [field, 'UNKNOWN']));
  return {
    decision: resolveExecutableGuardianScopeDecision(Object.values(fields)),
    schema: input && input.schema,
    fields,
    field_count: EXECUTABLE_GUARDIAN_SCOPE_INPUT_FIELDS.length,
    unknown_count: Object.values(fields).filter((value) => value === 'UNKNOWN').length,
  };
}

function validateExecutableGuardianScopeInput(input) {
  const classification = classifyExecutableGuardianScopeInput(input);
  requireExact(classification.decision === 'CLOSED', `immutable receipt-schema release input is ${classification.decision}`);
  requireExact(JSON.stringify(input) === JSON.stringify(EXECUTABLE_GUARDIAN_SCOPE_INPUT), 'immutable receipt-schema release input is not exact');
  return input;
}

function validateExecutableGuardianScopeReceipt(receipt) {
  requireExact(receipt && receipt.schema === EXECUTABLE_GUARDIAN_SCOPE_SCHEMA, 'executable Guardian scope receipt schema is not exact');
  requireExact(receipt.foundation_override_success_count === FOUNDATION_OVERRIDE_SUCCESS_COUNT, 'executable Guardian scope receipt changed FOUNDATION count');
  requireExact(receipt.receipt_path === EXECUTABLE_GUARDIAN_SCOPE_PATH && exactArray(receipt.digest_exclusions, EXECUTABLE_GUARDIAN_SCOPE_EXCLUDED_PATHS), 'executable Guardian scope receipt exclusion is not exact');
  requireExact(receipt.base_sha === EXECUTABLE_GUARDIAN_SCOPE_BASE_SHA && receipt.merge_parent_sha === EXECUTABLE_GUARDIAN_SCOPE_BASE_SHA, 'executable Guardian scope merge parent is not exact');
  requireExact(receipt.parent_receipt && receipt.parent_receipt.schema === SCHEMA_COHERENCE_MIGRATION_SCHEMA && receipt.parent_receipt.path === SCHEMA_COHERENCE_MIGRATION_PATH && receipt.parent_receipt.sha256 === EXECUTABLE_GUARDIAN_SCOPE_PARENT_V4_RECEIPT_SHA256, 'executable Guardian scope v4 lineage is not exact');
  requireExact(Array.isArray(receipt.lineage) && exactArray(receipt.lineage, EXECUTABLE_GUARDIAN_SCOPE_LINEAGE), 'executable Guardian scope v0.2 lineage is not preserved');
  validateExecutableGuardianScopeInput(receipt.immutable_release_input);
  requireExact(receipt.immutable_release_input.protocol.rest.immutable === true && receipt.immutable_release_input.protocol.graphql.immutable === true && receipt.immutable_release_input.protocol.repository_setting.enabled === true, 'executable Guardian scope release protocol is not immutable or enabled');
  requireExact(receipt.input_field_count === EXECUTABLE_GUARDIAN_SCOPE_INPUT_FIELDS.length && receipt.missing_stale_future_schema_decision === 'UNKNOWN', 'executable Guardian scope input denominator is not exact');
  requireExact(Array.isArray(receipt.acceptance_ids) && exactArray(receipt.acceptance_ids, EXECUTABLE_GUARDIAN_SCOPE_ACCEPTANCE_IDS), 'executable Guardian scope acceptance IDs are not exact');
  requireExact(Array.isArray(receipt.protected_files) && exactArray(receipt.protected_files, EXECUTABLE_GUARDIAN_SCOPE_PROTECTED_FILES), 'executable Guardian scope protected-file lock is not exact');
  requireExact(receipt.prior_guardian_failure && JSON.stringify(receipt.prior_guardian_failure) === JSON.stringify(EXECUTABLE_GUARDIAN_SCOPE_PRIOR_FAILURE), 'executable Guardian scope prior ReferenceError snapshot is not exact');
  requireExact(Array.isArray(receipt.cells) && receipt.cells.length === 1, 'executable Guardian scope denominator must contain one cell');
  const cell = receipt.cells[0];
  requireExact(cell && cell.id === 'EXECUTABLE_GUARDIAN_SCOPE_ADOPTION' && cell.meta_operation === 'AdoptExecutableGuardianScope', 'executable Guardian scope cell identity is not exact');
  requireExact(cell.proof_choice === 'REGRESSION' && cell.indicator === 'GUARDRAIL', 'executable Guardian scope cell classification is not exact');
  requireExact(cell.allowed === 1 && cell.consumed === 1 && cell.replay_decision === 'REFUTED' && receipt.replay_second_use === 'REFUTED', 'executable Guardian scope consumption is not single-use');
  requireExact(cell.parent_outcome === EXECUTABLE_GUARDIAN_SCOPE_PARENT_OUTCOME && cell.outcome === 'CLOSED' && cell.causal === EXECUTABLE_GUARDIAN_SCOPE_REASON, 'executable Guardian scope child-owned outcome is not exact');
  requireExact(cell.reason === EXECUTABLE_GUARDIAN_SCOPE_REASON, 'executable Guardian scope reason is not exact');
  requireExact(cell.pull_request === EXECUTABLE_GUARDIAN_SCOPE_PULL_REQUEST && cell.branch === EXECUTABLE_GUARDIAN_SCOPE_BRANCH && cell.base_sha === EXECUTABLE_GUARDIAN_SCOPE_BASE_SHA, 'executable Guardian scope identity is not exact');
  requireExact(receipt.repository_gates === 0 && receipt.local_go_gates === 0 && receipt.cross_project_gates === 0, 'executable Guardian scope gate counts are not zero');
  requireExact(Array.isArray(receipt.changed_paths) && exactArray(receipt.changed_paths, EXECUTABLE_GUARDIAN_SCOPE_CHANGED_PATHS), 'executable Guardian scope changed paths are not exact');
  requireExact(digestChangedPaths(receipt.changed_paths) === receipt.changed_paths_sha256 && validDigest(receipt.changed_paths_sha256), 'executable Guardian scope changed-path digest is not exact');
  requireExact(validDigest(receipt.patch_sha256_excluding_receipt) && validDigest(receipt.tree_sha256_excluding_receipt), 'executable Guardian scope content digests are malformed');
  requireExact(receipt.input_release_digests && receipt.input_release_digests.release_asset === EXECUTABLE_GUARDIAN_SCOPE_INPUT.release_asset.sha256 && receipt.input_release_digests.main_artifact === EXECUTABLE_GUARDIAN_SCOPE_INPUT.main_run.artifact.sha256 && receipt.input_release_digests.proposal_file === EXECUTABLE_GUARDIAN_SCOPE_INPUT.proposal.file_sha256 && receipt.input_release_digests.declared_proposal === EXECUTABLE_GUARDIAN_SCOPE_INPUT.proposal.declared_sha256, 'executable Guardian scope input release digests are not exact');
  return receipt;
}

function validateExecutableGuardianScope({receipt, parentReceiptBytes, candidateBaseSHA, candidateBaseTreeEntries, migrationPull, migrationCommit, migrationCompare, migrationTreeEntries}) {
  validateExecutableGuardianScopeReceipt(receipt);
  requireExact(Buffer.isBuffer(parentReceiptBytes) && sha256(parentReceiptBytes) === EXECUTABLE_GUARDIAN_SCOPE_PARENT_V4_RECEIPT_SHA256, 'v4 parent receipt bytes were rewritten');
  requireExact(migrationPull && migrationPull.number === EXECUTABLE_GUARDIAN_SCOPE_PULL_REQUEST, 'executable Guardian scope pull request number mismatch');
  requireExact(migrationPull.state === 'closed' && migrationPull.merged === true && validSHA(migrationPull.merge_commit_sha), 'executable Guardian scope pull request is not merged exactly once');
  requireExact(migrationPull.head && migrationPull.head.ref === EXECUTABLE_GUARDIAN_SCOPE_BRANCH && migrationPull.head.repo && migrationPull.head.repo.full_name === REPOSITORY && validSHA(migrationPull.head.sha), 'executable Guardian scope pull request head mismatch');
  requireExact(candidateBaseSHA === migrationPull.merge_commit_sha, 'candidate base is not the executable Guardian scope merge commit');
  requireExact(migrationCommit && migrationCommit.sha === migrationPull.merge_commit_sha, 'executable Guardian scope merge commit SHA mismatch');
  const parents = Array.isArray(migrationCommit.parents) ? migrationCommit.parents.map((parent) => parent && parent.sha) : [];
  requireExact(parents.length === 1 && parents[0] === EXECUTABLE_GUARDIAN_SCOPE_BASE_SHA, 'executable Guardian scope merge parent is not dev@7f45792');
  requireExact(migrationCompare && Array.isArray(migrationCompare.files) && exactArray(canonicalPathNames(migrationCompare.files), receipt.changed_paths), 'executable Guardian scope changed paths do not match the receipt');
  requireExact(digestChangedPaths(migrationCompare.files) === receipt.changed_paths_sha256, 'executable Guardian scope changed-path evidence does not match the receipt');
  requireExact(digestTreeEntries(migrationTreeEntries, EXECUTABLE_GUARDIAN_SCOPE_EXCLUDED_PATHS) === receipt.tree_sha256_excluding_receipt, 'executable Guardian scope head tree does not match the receipt');
  requireExact(digestTreeEntries(candidateBaseTreeEntries, EXECUTABLE_GUARDIAN_SCOPE_EXCLUDED_PATHS) === receipt.tree_sha256_excluding_receipt, 'candidate base tree does not match executable Guardian scope receipt');
  return {
    schema: EXECUTABLE_GUARDIAN_SCOPE_SCHEMA,
    cell: 'EXECUTABLE_GUARDIAN_SCOPE_ADOPTION',
    proof_choice: 'REGRESSION',
    indicator: 'GUARDRAIL',
    pull_request: EXECUTABLE_GUARDIAN_SCOPE_PULL_REQUEST,
    branch: EXECUTABLE_GUARDIAN_SCOPE_BRANCH,
    base_sha: EXECUTABLE_GUARDIAN_SCOPE_BASE_SHA,
    merge_commit_sha: migrationPull.merge_commit_sha,
    merge_parent_sha: EXECUTABLE_GUARDIAN_SCOPE_BASE_SHA,
    parent_outcome: EXECUTABLE_GUARDIAN_SCOPE_PARENT_OUTCOME,
    outcome: 'CLOSED',
    causal: EXECUTABLE_GUARDIAN_SCOPE_REASON,
    reason: EXECUTABLE_GUARDIAN_SCOPE_REASON,
    allowed: 1,
    consumed: 1,
    replay_decision: 'REFUTED',
  };
}

function validateCorrectionChild({receipt, parentRepairReceipt, parentRepairReceiptBytes, parentRepairBaseCommit, parentRepairBaseTreeEntries, candidateBaseSHA, candidateBaseTreeEntries, correctionPull, correctionCommit, correctionCompare, correctionTreeEntries}) {
  validateCorrectionChildReceipt(receipt);
  validateIncompletePropagationOutcome(parentRepairReceipt);
  requireExact(Buffer.isBuffer(parentRepairReceiptBytes) && sha256(parentRepairReceiptBytes) === CORRECTION_CHILD_PARENT_RECEIPT_SHA256, 'correction child parent receipt bytes do not match');
  requireExact(parentRepairBaseCommit && parentRepairBaseCommit.sha === CORRECTION_CHILD_BASE_SHA, 'correction child parent repair base SHA mismatch');
  requireExact(digestTreeEntries(parentRepairBaseTreeEntries, REGRESSION_REPAIR_EXCLUDED_PATHS) === parentRepairReceipt.tree_sha256_excluding_receipt, 'correction child parent repair tree does not match');
  requireExact(correctionPull && correctionPull.number === CORRECTION_CHILD_PULL_REQUEST, 'correction child pull request number mismatch');
  requireExact(correctionPull.state === 'closed' && correctionPull.merged === true && validSHA(correctionPull.merge_commit_sha), 'correction child pull request is not merged exactly once');
  requireExact(correctionPull.head && correctionPull.head.ref === CORRECTION_CHILD_BRANCH && correctionPull.head.repo && correctionPull.head.repo.full_name === REPOSITORY && validSHA(correctionPull.head.sha), 'correction child pull request head mismatch');
  requireExact(candidateBaseSHA === correctionPull.merge_commit_sha, 'candidate base is not the correction child merge commit');
  requireExact(correctionCommit && correctionCommit.sha === correctionPull.merge_commit_sha, 'correction child merge commit SHA mismatch');
  const parents = Array.isArray(correctionCommit.parents) ? correctionCommit.parents.map((parent) => parent && parent.sha) : [];
  requireExact(parents.length === 1 && parents[0] === CORRECTION_CHILD_BASE_SHA, 'correction child merge is not the exact single-parent squash from dev@B');
  requireExact(correctionCompare && Array.isArray(correctionCompare.files) && exactArray(canonicalPathNames(correctionCompare.files), receipt.changed_paths), 'correction child changed paths do not match the receipt');
  requireExact(digestChangedPaths(correctionCompare.files) === receipt.changed_paths_sha256, 'correction child changed-path evidence does not match the receipt');
  requireExact(digestTreeEntries(correctionTreeEntries, CORRECTION_CHILD_EXCLUDED_PATHS) === receipt.tree_sha256_excluding_receipt, 'correction child head tree does not match the receipt');
  requireExact(digestTreeEntries(candidateBaseTreeEntries, CORRECTION_CHILD_EXCLUDED_PATHS) === receipt.tree_sha256_excluding_receipt, 'candidate base tree does not match the correction child receipt');
  return {
    schema: CORRECTION_CHILD_SCHEMA,
    cell: 'CORRECTION_CHILD',
    proof_choice: 'REGRESSION',
    indicator: 'GUARDRAIL',
    pull_request: CORRECTION_CHILD_PULL_REQUEST,
    branch: CORRECTION_CHILD_BRANCH,
    base_sha: CORRECTION_CHILD_BASE_SHA,
    merge_commit_sha: correctionPull.merge_commit_sha,
    parent_repair_receipt: CORRECTION_CHILD_PARENT_RECEIPT_SHA256,
    parent_repair_outcome: 'REFUTED_INCOMPLETE_PROPAGATION',
    reason: CORRECTION_CHILD_REASON,
    allowed: 1,
    consumed: 1,
    replay_decision: 'REFUTED',
  };
}

function validateRegressionRepair({receipt, candidateBaseSHA, candidateBaseCommit, candidateBaseTreeEntries, repairPull, repairCommit, repairCompare, repairTreeEntries}) {
  validateRegressionRepairReceipt(receipt);
  requireExact(repairPull && repairPull.number === REGRESSION_REPAIR_PULL_REQUEST, 'regression repair pull request number mismatch');
  requireExact(repairPull.state === 'closed' && repairPull.merged === true && validSHA(repairPull.merge_commit_sha), 'regression repair pull request is not merged exactly once');
  requireExact(repairPull.head && repairPull.head.ref === REGRESSION_REPAIR_BRANCH && repairPull.head.repo && repairPull.head.repo.full_name === REPOSITORY && validSHA(repairPull.head.sha), 'regression repair pull request head mismatch');
  requireExact(candidateBaseSHA === repairPull.merge_commit_sha, 'candidate base is not the regression repair merge commit');
  requireExact(repairCommit && repairCommit.sha === repairPull.merge_commit_sha, 'regression repair merge commit SHA mismatch');
  const parents = Array.isArray(repairCommit.parents) ? repairCommit.parents.map((parent) => parent && parent.sha) : [];
  requireExact(parents.length === 1 && parents[0] === REGRESSION_REPAIR_BASE_SHA, 'regression repair merge is not the exact single-parent squash from dev@A');
  requireExact(repairCompare && Array.isArray(repairCompare.files) && exactArray(canonicalPathNames(repairCompare.files), receipt.changed_paths), 'regression repair changed paths do not match the receipt');
  requireExact(digestChangedPaths(repairCompare.files) === receipt.changed_paths_sha256, 'regression repair changed-path evidence does not match the receipt');
  requireExact(digestTreeEntries(repairTreeEntries, REGRESSION_REPAIR_EXCLUDED_PATHS) === receipt.tree_sha256_excluding_receipt, 'regression repair head tree does not match the receipt');
  requireExact(digestTreeEntries(candidateBaseTreeEntries, REGRESSION_REPAIR_EXCLUDED_PATHS) === receipt.tree_sha256_excluding_receipt, 'candidate base tree does not match the regression repair receipt');
  return {
    schema: REGRESSION_REPAIR_SCHEMA,
    cell: 'REGRESSION_REPAIR',
    proof_choice: 'REGRESSION',
    indicator: 'GUARDRAIL',
    pull_request: REGRESSION_REPAIR_PULL_REQUEST,
    branch: REGRESSION_REPAIR_BRANCH,
    base_sha: REGRESSION_REPAIR_BASE_SHA,
    merge_commit_sha: repairPull.merge_commit_sha,
    reason: REGRESSION_REPAIR_REASON,
    allowed: 1,
    consumed: 1,
    replay_decision: 'REFUTED',
  };
}

function validateBaseAttestation(policy, commit) {
  const expected = policy.base_attestation;
  requireExact(commit && commit.sha === expected.commit, 'candidate base commit SHA attestation mismatch');
  requireExact(commit.commit && commit.commit.tree && commit.commit.tree.sha === expected.tree_sha, 'candidate base tree attestation mismatch');
  const parents = Array.isArray(commit.parents) ? commit.parents.map((parent) => parent && parent.sha) : [];
  requireExact(exactArray(parents, expected.parents), 'candidate base parent attestation mismatch');
}

function validateAuthorizationMerge(policy, authorizationPull, authorizationCommit) {
  requireExact(authorizationPull && authorizationPull.number === policy.authorization.pull_request, 'authorization pull request number mismatch');
  requireExact(authorizationPull.base && authorizationPull.base.ref === 'dev' && authorizationPull.base.repo && authorizationPull.base.repo.full_name === REPOSITORY, 'authorization pull request base mismatch');
  requireExact(authorizationPull.head && authorizationPull.head.ref === AUTHORIZATION_BRANCH && authorizationPull.head.repo && authorizationPull.head.repo.full_name === REPOSITORY, 'authorization pull request head mismatch');
  requireExact(authorizationPull.state === 'closed' && authorizationPull.merged === true && validSHA(authorizationPull.merge_commit_sha), 'authorization pull request is not a single merged authorization');
  requireExact(authorizationCommit && authorizationCommit.sha === authorizationPull.merge_commit_sha, 'authorization merge commit SHA mismatch');
  const parents = Array.isArray(authorizationCommit.parents) ? authorizationCommit.parents.map((parent) => parent && parent.sha) : [];
  requireExact(parents.length === 1 && parents[0] === policy.candidate.base_sha, 'authorization merge is not based on the attested dev snapshot');
  return authorizationPull.merge_commit_sha;
}

function validateCandidateIdentity(policy, candidatePull) {
  requireExact(candidatePull && candidatePull.number === policy.candidate.pull_request, 'candidate pull request number mismatch');
  requireExact(candidatePull.base && candidatePull.base.ref === 'dev' && candidatePull.base.repo && candidatePull.base.repo.full_name === REPOSITORY, 'candidate pull request base mismatch');
  requireExact(candidatePull.head && candidatePull.head.ref === CANDIDATE_BRANCH && candidatePull.head.repo && candidatePull.head.repo.full_name === REPOSITORY, 'candidate pull request head mismatch');
  requireExact(candidatePull.state === 'open' && candidatePull.merged !== true && candidatePull.merged_at === null, 'candidate pull request is already consumed or not open');
  requireExact(validSHA(candidatePull.base.sha) && validSHA(candidatePull.head.sha), 'candidate live base/head SHA is malformed');
}

function validateAncestry(compare, expectedAncestor, label) {
  requireExact(compare && compare.status === 'ahead' && compare.merge_base_commit && compare.merge_base_commit.sha === expectedAncestor, `${label} is not in candidate head ancestry`);
}

function validateCandidateEvidence({policy, candidatePull, authorizationPull, authorizationCommit, candidateCompare, authorizationCompare, baseCommit, manifestBytes, changedFiles, treeEntries, patchDigest, regressionRepair}) {
  validatePolicy(policy);
  validateCandidateIdentity(policy, candidatePull);
  validateBaseAttestation(policy, baseCommit);
  const authorizationMergeSHA = validateAuthorizationMerge(policy, authorizationPull, authorizationCommit);
  validateAncestry(candidateCompare, policy.candidate.head_sha, 'candidate S');
  validateAncestry(authorizationCompare, authorizationMergeSHA, 'authorization A');
  if (candidatePull.base.sha !== authorizationMergeSHA) {
    requireExact(regressionRepair && candidatePull.base.sha === regressionRepair.merge_commit_sha, 'candidate base is neither the authorization merge nor the exact regression repair merge');
  }
  requireExact(Buffer.isBuffer(manifestBytes) && sha256(manifestBytes) === policy.candidate.manifest_sha256, 'candidate manifest M does not match');
  requireExact(digestChangedPaths(changedFiles) === policy.candidate.changed_paths_sha256 && canonicalPathNames(changedFiles).length === policy.candidate.changed_path_count, 'candidate changed-path digest P does not match');
  requireExact(digestTreeEntries(treeEntries, policy.auth_policy_paths) === policy.candidate.tree_sha256_excluding_authorization_paths, 'candidate tree digest T does not match');
  if (patchDigest !== undefined) requireExact(patchDigest === policy.candidate.patch_sha256_excluding_authorization_paths, 'candidate patch digest T does not match');
  return {
    decision: 'PASS',
    reason: FOUNDATION_OVERRIDE_MARKER,
    candidate_pull_request: policy.candidate.pull_request,
    candidate_branch: policy.candidate.branch,
    candidate_head_sha: policy.candidate.head_sha,
    authorization_pull_request: policy.authorization.pull_request,
    authorization_merge_commit: authorizationMergeSHA,
    manifest_sha256: policy.candidate.manifest_sha256,
    changed_paths_sha256: policy.candidate.changed_paths_sha256,
    patch_sha256_excluding_authorization_paths: policy.candidate.patch_sha256_excluding_authorization_paths,
    tree_sha256_excluding_authorization_paths: policy.candidate.tree_sha256_excluding_authorization_paths,
    single_use: true,
    consumed: false,
    replay_decision: 'REFUTED',
    ...(regressionRepair ? {regression_repair: regressionRepair} : {}),
  };
}

module.exports = {
  AUTHORIZATION_BRANCH,
  AUTHORIZATION_PATHS,
  AUTHORIZATION_SCHEMA,
  CANDIDATE_BRANCH,
  CANDIDATE_PULL_REQUEST,
  FOUNDATION_OVERRIDE_MARKER,
  FOUNDATION_OVERRIDE_SUCCESS_COUNT,
  REPOSITORY,
  canonicalPathNames,
  canonicalProtectedIntersection,
  canonicalTreeEntries,
  digestChangedPaths,
  digestProtectedIntersection,
  digestTreeEntries,
  sha256,
  validDigest,
  validSHA,
  validateAuthorizationMerge,
  validateCandidateEvidence,
  validateCandidateIdentity,
  validateGuardianDispatchAuthorization,
  validateGuardianDispatchPolicy,
  validateGuardianDispatchReceipt,
  validatePolicy,
  validateCorrectionChild,
  validateCorrectionChildReceipt,
  validateIncompletePropagationOutcome,
  validateRegressionRepair,
  validateRegressionRepairReceipt,
  REGRESSION_REPAIR_BASE_SHA,
  REGRESSION_REPAIR_BRANCH,
  REGRESSION_REPAIR_CHANGED_PATHS,
  REGRESSION_REPAIR_EXCLUDED_PATHS,
  REGRESSION_REPAIR_PATH,
  REGRESSION_REPAIR_PULL_REQUEST,
  REGRESSION_REPAIR_REASON,
  REGRESSION_REPAIR_SCHEMA,
  CORRECTION_CHILD_BASE_SHA,
  CORRECTION_CHILD_BRANCH,
  CORRECTION_CHILD_CHANGED_PATHS,
  CORRECTION_CHILD_EXCLUDED_PATHS,
  CORRECTION_CHILD_PARENT_RECEIPT_SHA256,
  CORRECTION_CHILD_PATH,
  CORRECTION_CHILD_PULL_REQUEST,
  CORRECTION_CHILD_REASON,
  CORRECTION_CHILD_SCHEMA,
  CORRECTION_CHILD_OWNED_OUTCOME_FIELDS,
  classifySchemaCoherenceInput,
  resolveSchemaCoherenceDecision,
  schemaCoherenceInputFieldStates,
  validateImmutableReleaseInput,
  validateSchemaCoherenceMigration,
  validateSchemaCoherenceMigrationReceipt,
  classifyExecutableGuardianScopeInput,
  executableGuardianScopeInputFieldStates,
  resolveExecutableGuardianScopeDecision,
  validateExecutableGuardianScope,
  validateExecutableGuardianScopeInput,
  validateExecutableGuardianScopeReceipt,
  SCHEMA_COHERENCE_MIGRATION_BASE_SHA,
  SCHEMA_COHERENCE_MIGRATION_BRANCH,
  SCHEMA_COHERENCE_MIGRATION_CHANGED_PATHS,
  SCHEMA_COHERENCE_MIGRATION_EXCLUDED_PATHS,
  SCHEMA_COHERENCE_MIGRATION_INPUT,
  SCHEMA_COHERENCE_MIGRATION_INPUT_FIELDS,
  SCHEMA_COHERENCE_MIGRATION_PARENT_RECEIPTS,
  SCHEMA_COHERENCE_MIGRATION_PATH,
  SCHEMA_COHERENCE_MIGRATION_PULL_REQUEST,
  SCHEMA_COHERENCE_MIGRATION_REASON,
  SCHEMA_COHERENCE_MIGRATION_SCHEMA,
  EXECUTABLE_GUARDIAN_SCOPE_ACCEPTANCE_IDS,
  EXECUTABLE_GUARDIAN_SCOPE_BASE_SHA,
  EXECUTABLE_GUARDIAN_SCOPE_BRANCH,
  EXECUTABLE_GUARDIAN_SCOPE_CHANGED_PATHS,
  EXECUTABLE_GUARDIAN_SCOPE_EXCLUDED_PATHS,
  EXECUTABLE_GUARDIAN_SCOPE_INPUT,
  EXECUTABLE_GUARDIAN_SCOPE_INPUT_FIELDS,
  EXECUTABLE_GUARDIAN_SCOPE_PARENT_OUTCOME,
  EXECUTABLE_GUARDIAN_SCOPE_PARENT_V4_RECEIPT_SHA256,
  EXECUTABLE_GUARDIAN_SCOPE_PRIOR_FAILURE,
  EXECUTABLE_GUARDIAN_SCOPE_PATH,
  EXECUTABLE_GUARDIAN_SCOPE_PROTECTED_FILES,
  EXECUTABLE_GUARDIAN_SCOPE_PULL_REQUEST,
  EXECUTABLE_GUARDIAN_SCOPE_REASON,
  EXECUTABLE_GUARDIAN_SCOPE_SCHEMA,
  EXECUTABLE_GUARDIAN_SCOPE_LINEAGE,
  GUARDIAN_DISPATCH_BASE_SHA,
  GUARDIAN_DISPATCH_BRANCH,
  GUARDIAN_DISPATCH_CELL_IDS,
  GUARDIAN_DISPATCH_CHANGED_FILE_COUNT,
  GUARDIAN_DISPATCH_CHANGED_PATHS_SHA256,
  GUARDIAN_DISPATCH_CONTRACT_DIGEST,
  GUARDIAN_DISPATCH_MARKER,
  GUARDIAN_DISPATCH_HARNESS_DIGEST,
  GUARDIAN_DISPATCH_HEAD_SHA,
  GUARDIAN_DISPATCH_KERNEL_AFTER_SHA256,
  GUARDIAN_DISPATCH_KERNEL_BEFORE_SHA256,
  GUARDIAN_DISPATCH_LINEAGE,
  GUARDIAN_DISPATCH_MERGE_BASE_SHA,
  GUARDIAN_DISPATCH_PARENT_RECEIPT_SHA256,
  GUARDIAN_DISPATCH_PATH,
  GUARDIAN_DISPATCH_POLICY_PATH,
  GUARDIAN_DISPATCH_POLICY_SCHEMA,
  GUARDIAN_DISPATCH_PROTECTED_INTERSECTION_COUNT,
  GUARDIAN_DISPATCH_PROTECTED_INTERSECTION_SHA256,
  GUARDIAN_DISPATCH_PROTECTED_PATHS,
  GUARDIAN_DISPATCH_PROPOSAL_DECLARED_DIGEST,
  GUARDIAN_DISPATCH_PROPOSAL_FILE_DIGEST,
  GUARDIAN_DISPATCH_PULL_REQUEST,
  GUARDIAN_DISPATCH_SCHEMA,
  GUARDIAN_DISPATCH_SEMANTIC_IR_DIGEST,
  GUARDIAN_DISPATCH_SOURCE_DIGEST,
  GUARDIAN_DISPATCH_VALIDATOR_DIGEST,
};
