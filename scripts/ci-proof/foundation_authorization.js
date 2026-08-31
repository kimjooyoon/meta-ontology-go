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
};
