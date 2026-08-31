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
  '.github/workflows/ci-guardian.yml',
  '.github/workflows/ci.yml',
  'internal/verify/scope_schema_coherence_migration_adoption_20260831.go',
  'scripts/ci-proof/foundation_authorization.js',
  'scripts/ci-proof/foundation_authorization_test.js',
  'scripts/ci-proof/guardian.js',
  'scripts/ci-proof/guardian_test.js',
].sort());
const SCHEMA_COHERENCE_MIGRATION_EXCLUDED_PATHS = Object.freeze([SCHEMA_COHERENCE_MIGRATION_PATH]);
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

function validatePolicy(policy) {
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
  canonicalTreeEntries,
  digestChangedPaths,
  digestTreeEntries,
  sha256,
  validDigest,
  validSHA,
  validateAuthorizationMerge,
  validateCandidateEvidence,
  validateCandidateIdentity,
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
};
