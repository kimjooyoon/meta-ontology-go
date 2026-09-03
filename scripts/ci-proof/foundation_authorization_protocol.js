'use strict';

const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');

const PROTOCOL_SCHEMA = 'gooo/foundation-authorization-protocol/v1';
const ROOT_AXIOM_SCHEMA = 'gooo/foundation-authorization-root-axiom/v1';
const PROTOCOL_DEFINITION_PATH = '.github/foundation-authorization-protocol.json';
const REPOSITORY = 'kimjooyoon/meta-ontology-go';
const CI_APP_ID = 15368;
const UNKNOWN_FIELDS = Object.freeze(['stage', 'step', 'reason', 'unknown_class', 'next_operation', 'blocked_by']);
const DECISION_PRECEDENCE = Object.freeze(['REFUTED', 'UNKNOWN', 'CLOSED']);
const REQUIRED_STAGES = Object.freeze(['FOUNDATION', 'COHERENCE', 'REGRESSION']);
const REQUIRED_INDICATORS = Object.freeze(['DRIVER', 'OUTCOME', 'GUARDRAIL']);
const PROTECTED_PATH_PATTERNS = Object.freeze([
  '.github/agent-scope-table.md',
  '.github/branch-policy.md',
  '.github/ci-governance.json',
  '.github/conformance-plan.md',
  '.github/workflows/**',
  'go.mod',
  'go.sum',
  'internal/verify/**',
  'scripts/ci-evidence/**',
  'scripts/ci-proof/**',
  'scripts/verify/**',
]);

const definition = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', PROTOCOL_DEFINITION_PATH), 'utf8'));

function validSHA(value) {
  return typeof value === 'string' && /^[0-9a-f]{40}$/.test(value);
}

function validDigest(value) {
  return typeof value === 'string' && /^sha256:[0-9a-f]{64}$/.test(value);
}

function validPositiveInteger(value) {
  return Number.isInteger(value) && value > 0;
}

function sha256(value) {
  return `sha256:${crypto.createHash('sha256').update(value).digest('hex')}`;
}

function exactArray(left, right) {
  return JSON.stringify(left) === JSON.stringify(right);
}

function exactObject(left, right) {
  return JSON.stringify(left) === JSON.stringify(right);
}

function requireExact(condition, message) {
  if (!condition) throw new Error(message);
}

function isObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function canonicalPathNames(files) {
  const names = (Array.isArray(files) ? files : []).flatMap((file) => {
    if (typeof file === 'string') return [file];
    return [file && file.filename, file && file.previous_filename];
  }).filter((value) => typeof value === 'string' && value.length > 0);
  return [...new Set(names)].sort();
}

function canonicalPathList(paths) {
  return [...new Set((Array.isArray(paths) ? paths : []).filter((value) => typeof value === 'string' && value.length > 0))].sort();
}

function digestPathList(paths) {
  const canonical = canonicalPathList(paths);
  return sha256(canonical.length === 0 ? '' : `${canonical.join('\n')}\n`);
}

function digestChangedPaths(files) {
  return digestPathList(canonicalPathNames(files));
}

function protectedPathMatches(pathName) {
  return PROTECTED_PATH_PATTERNS.some((pattern) => {
    if (pattern.endsWith('/**')) return pathName === pattern.slice(0, -3) || pathName.startsWith(pattern.slice(0, -2));
    return pathName === pattern;
  });
}

function canonicalProtectedIntersection(files) {
  return canonicalPathNames(files).filter(protectedPathMatches);
}

function digestProtectedIntersection(files) {
  const paths = Array.isArray(files) && files.every((file) => typeof file === 'string')
    ? canonicalPathList(files)
    : canonicalProtectedIntersection(files);
  return digestPathList(paths);
}

function canonicalTreeEntries(entries, includedPaths = null) {
  const included = includedPaths === null ? null : new Set(canonicalPathList(includedPaths));
  return (Array.isArray(entries) ? entries : [])
    .filter((entry) => entry && entry.type === 'blob' && typeof entry.path === 'string' && (included === null || included.has(entry.path)))
    .map((entry) => ({path: entry.path, mode: entry.mode || null, sha: entry.sha}))
    .sort((left, right) => {
      const leftKey = `${left.path}\u0000${left.mode}\u0000${left.sha}`;
      const rightKey = `${right.path}\u0000${right.mode}\u0000${right.sha}`;
      return leftKey < rightKey ? -1 : leftKey > rightKey ? 1 : 0;
    });
}

function digestTreeEntries(entries, includedPaths = null) {
  const lines = canonicalTreeEntries(entries, includedPaths).map((entry) => `${entry.path}\t${entry.mode}\t${entry.sha}`);
  return sha256(lines.length === 0 ? '' : `${lines.join('\n')}\n`);
}

function digestKernelEntries(entries) {
  const kernelEntries = (Array.isArray(entries) ? entries : []).filter((entry) => entry && typeof entry.path === 'string' && protectedPathMatches(entry.path));
  return digestTreeEntries(kernelEntries);
}

function validateProtocolDefinition(candidate = definition) {
  requireExact(candidate && candidate.schema === PROTOCOL_SCHEMA, 'foundation authorization protocol schema is not exact');
  requireExact(exactArray(candidate.decision_precedence, DECISION_PRECEDENCE), 'foundation authorization decision precedence is not exact');
  requireExact(exactArray(candidate.unknown_fields, UNKNOWN_FIELDS), 'foundation authorization unknown fields are not exact');
  requireExact(!Object.prototype.hasOwnProperty.call(candidate, 'aggregate_score'), 'foundation authorization protocol must not define an aggregate score');
  requireExact(Array.isArray(candidate.cells) && candidate.cells.length === 12, 'foundation authorization denominator must contain twelve cells');
  const stageCounts = Object.fromEntries(REQUIRED_STAGES.map((stage) => [stage, 0]));
  const indicatorCounts = Object.fromEntries(REQUIRED_INDICATORS.map((indicator) => [indicator, 0]));
  const ids = new Set();
  for (const cell of candidate.cells) {
    requireExact(isObject(cell) && typeof cell.id === 'string' && cell.id.length > 0, 'foundation authorization cell identity is malformed');
    requireExact(!ids.has(cell.id), 'foundation authorization cell identity is duplicated');
    ids.add(cell.id);
    requireExact(REQUIRED_STAGES.includes(cell.stage), `foundation authorization cell stage is not supported: ${cell.id}`);
    requireExact(REQUIRED_INDICATORS.includes(cell.indicator), `foundation authorization cell indicator is not supported: ${cell.id}`);
    requireExact(cell.proof_choice === cell.stage && typeof cell.step === 'string' && cell.step.length > 0, `foundation authorization cell binding is malformed: ${cell.id}`);
    requireExact(cell.producer === 'foundation-authorization-protocol' && cell.consumer === 'ci-guardian', `foundation authorization cell producer/consumer is not exact: ${cell.id}`);
    stageCounts[cell.stage] += 1;
    indicatorCounts[cell.indicator] += 1;
  }
  requireExact(Object.values(stageCounts).every((count) => count === 4), 'foundation authorization stage denominator is not 4/4/4');
  requireExact(Object.values(indicatorCounts).every((count) => count === 4), 'foundation authorization indicator denominator is not 4/4/4');
  return candidate;
}

function unknownEvidence(stage, step, reason, unknownClass, nextOperation, blockedBy = []) {
  const evidence = {stage, step, reason, unknown_class: unknownClass, next_operation: nextOperation, blocked_by: [...blockedBy]};
  requireExact(exactArray(Object.keys(evidence), UNKNOWN_FIELDS), 'unknown evidence keys are not exact');
  return evidence;
}

function resolveDecision(states) {
  const present = new Set(Array.isArray(states) ? states : []);
  return DECISION_PRECEDENCE.find((decision) => present.has(decision)) || 'CLOSED';
}

function candidateDigest(candidate) {
  return sha256(JSON.stringify({
    pull_request: candidate && candidate.pull_request,
    base_repo: candidate && candidate.base_repo,
    base_branch: candidate && candidate.base_branch,
    base_sha: candidate && candidate.base_sha,
    head_repo: candidate && candidate.head_repo,
    head_branch: candidate && candidate.head_branch,
    head_sha: candidate && candidate.head_sha,
    merge_base_sha: candidate && candidate.merge_base_sha,
    changed_paths: candidate && candidate.changed_paths,
    protected_intersection: candidate && candidate.protected_intersection,
  }));
}

function consumptionReceiptDigest(consumption) {
  return sha256(JSON.stringify({
    nonce: consumption && consumption.nonce,
    merge_parent_sha: consumption && consumption.merge_parent_sha,
    consumed_at: consumption && consumption.consumed_at,
    replay_decision: consumption && consumption.replay_decision,
    post_adoption_verified: consumption && consumption.post_adoption_verified,
  }));
}

function validTimestamp(value) {
  return typeof value === 'string' && Number.isFinite(Date.parse(value));
}

function validIdentity(identity) {
  return isObject(identity)
    && typeof identity.login === 'string' && identity.login.length > 0
    && validPositiveInteger(identity.id)
    && typeof identity.type === 'string' && identity.type.length > 0;
}

function validPathBinding(binding, digestFunction) {
  if (!isObject(binding) || !Array.isArray(binding.paths) || !Number.isInteger(binding.count) || typeof binding.digest !== 'string') return 'incomplete';
  const paths = canonicalPathList(binding.paths);
  if (!exactArray(binding.paths, paths) || binding.count !== paths.length) return 'digest_mismatch';
  if (!validDigest(binding.digest) || binding.digest !== digestFunction(paths)) return 'digest_mismatch';
  return 'valid';
}

function evaluate(input, {now = new Date().toISOString()} = {}) {
  validateProtocolDefinition();
  const cells = definition.cells.map((cell) => ({...cell, state: 'CLOSED'}));
  const refuted = [];
  const unknown = [];
  const addRefuted = (reason, cellId, detail = null) => refuted.push({reason, cellId, detail});
  const addUnknown = (evidence, cellId) => unknown.push({evidence, cellId});
  const mark = (cellId, state) => {
    const cell = cells.find((candidate) => candidate.id === cellId);
    if (cell) cell.state = state;
  };

  if (!isObject(input)) {
    addUnknown(unknownEvidence('FOUNDATION', 'REPOSITORY_ACTOR_AUTHORITY', 'INCOMPLETE_PROTOCOL_INPUT', 'INCOMPLETE_EVIDENCE', 'PROVIDE_PROTOCOL_ENVELOPE', ['protocol-envelope']), 'REPOSITORY_ACTOR_AUTHORITY');
  }
  const repository = input && input.repository;
  const actor = input && input.actor;
  if (!isObject(repository) || typeof repository.full_name !== 'string' || typeof repository.owner_login !== 'string' || !validPositiveInteger(repository.owner_id) || typeof repository.owner_type !== 'string' || !isObject(actor) || typeof actor.login !== 'string' || !validPositiveInteger(actor.id) || typeof actor.type !== 'string') {
    addUnknown(unknownEvidence('FOUNDATION', 'REPOSITORY_ACTOR_AUTHORITY', 'INCOMPLETE_REPOSITORY_ACTOR_AUTHORITY', 'INCOMPLETE_EVIDENCE', 'PROVIDE_REPOSITORY_OWNER_AND_ACTOR_IDENTITY', ['repository-actor-authority']), 'REPOSITORY_ACTOR_AUTHORITY');
  } else if (actor.owner_match === false || actor.login !== repository.owner_login || actor.id !== repository.owner_id || actor.type !== repository.owner_type) {
    addRefuted('ACTOR_MISMATCH', 'REPOSITORY_ACTOR_AUTHORITY', 'actor is not the exact repository owner identity');
  }

  if (!input || typeof input.intent !== 'string' || input.intent.length === 0 || !REQUIRED_STAGES.includes(input.proof_choice)) {
    addUnknown(unknownEvidence('FOUNDATION', 'AUTHORIZATION_INTENT_PROOF_CHOICE', 'INCOMPLETE_AUTHORIZATION_INTENT', 'INCOMPLETE_EVIDENCE', 'PROVIDE_INTENT_AND_PROOF_CHOICE', ['authorization-intent']), 'AUTHORIZATION_INTENT_PROOF_CHOICE');
  }

  const candidate = input && input.candidate;
  const candidateComplete = isObject(candidate)
    && validPositiveInteger(candidate.pull_request)
    && typeof candidate.base_repo === 'string' && candidate.base_repo.length > 0
    && ['dev', 'main'].includes(candidate.base_branch) && validSHA(candidate.base_sha)
    && typeof candidate.head_repo === 'string' && candidate.head_repo.length > 0
    && typeof candidate.head_branch === 'string' && candidate.head_branch.length > 0 && validSHA(candidate.head_sha)
    && validSHA(candidate.merge_base_sha);
  if (!candidateComplete) {
    addUnknown(unknownEvidence('FOUNDATION', 'CANDIDATE_PR_REPOSITORY_IDENTITY', 'INCOMPLETE_CANDIDATE_TUPLE', 'INCOMPLETE_EVIDENCE', 'PROVIDE_EXACT_CANDIDATE_IDENTITY', ['candidate-tuple']), 'CANDIDATE_PR_REPOSITORY_IDENTITY');
  } else {
    if (repository && candidate.base_repo !== repository.full_name || candidate.head_repo !== candidate.base_repo) addRefuted('CANDIDATE_REPOSITORY_MISMATCH', 'CANDIDATE_PR_REPOSITORY_IDENTITY', 'candidate repository identity is not exact');
  }
  if (!candidate || !validSHA(candidate.base_sha) || !validSHA(candidate.head_sha) || !validSHA(candidate.merge_base_sha)) {
    addUnknown(unknownEvidence('FOUNDATION', 'EXACT_BASE_HEAD_MERGE_BASE_SHAS', 'INCOMPLETE_EXACT_SHA_TUPLE', 'INCOMPLETE_EVIDENCE', 'PROVIDE_BASE_HEAD_AND_MERGE_BASE_SHAS', ['sha-tuple']), 'EXACT_BASE_HEAD_MERGE_BASE_SHAS');
  }

  const changedStatus = validPathBinding(candidate && candidate.changed_paths, digestPathList);
  if (changedStatus === 'incomplete') addUnknown(unknownEvidence('COHERENCE', 'CANONICAL_CHANGED_PATH_BINDING', 'INCOMPLETE_CHANGED_PATH_BINDING', 'INCOMPLETE_EVIDENCE', 'PROVIDE_CANONICAL_CHANGED_PATHS', ['changed-path-binding']), 'CANONICAL_CHANGED_PATH_BINDING');
  if (changedStatus === 'digest_mismatch') addRefuted('DIGEST_MISMATCH', 'CANONICAL_CHANGED_PATH_BINDING', 'changed-path count or digest does not match canonical paths');
  const protectedStatus = validPathBinding(candidate && candidate.protected_intersection, digestPathList);
  if (protectedStatus === 'incomplete') addUnknown(unknownEvidence('COHERENCE', 'PROTECTED_INTERSECTION_BINDING', 'INCOMPLETE_PROTECTED_INTERSECTION_BINDING', 'INCOMPLETE_EVIDENCE', 'PROVIDE_PROTECTED_INTERSECTION_DIGEST', ['protected-intersection-binding']), 'PROTECTED_INTERSECTION_BINDING');
  if (protectedStatus === 'digest_mismatch') addRefuted('DIGEST_MISMATCH', 'PROTECTED_INTERSECTION_BINDING', 'protected intersection count or digest does not match canonical paths');

  const observed = input && input.observed;
  if (!isObject(observed) || !isObject(observed.candidate) || !isObject(observed.runtime) || !isObject(observed.kernel)) {
    addUnknown(unknownEvidence('COHERENCE', 'WORKFLOW_RUNTIME_CHECK_APP_IDENTITY', 'INCOMPLETE_OBSERVED_EVIDENCE', 'INCOMPLETE_EVIDENCE', 'PROVIDE_OBSERVED_RUNTIME_AND_KERNEL_EVIDENCE', ['observed-evidence']), 'WORKFLOW_RUNTIME_CHECK_APP_IDENTITY');
  }
  if (candidateComplete && observed && isObject(observed.candidate) && !exactObject(observed.candidate, candidate)) addRefuted('CANDIDATE_DRIFT', 'CANDIDATE_PR_REPOSITORY_IDENTITY', 'observed candidate tuple differs from authorized candidate tuple');
  const runtime = input && input.runtime;
  const runtimeComplete = isObject(runtime)
    && typeof runtime.workflow_ref === 'string' && runtime.workflow_ref.length > 0 && validSHA(runtime.workflow_sha)
    && typeof runtime.runtime_ref === 'string' && runtime.runtime_ref.length > 0 && validSHA(runtime.runtime_sha)
    && typeof runtime.check_name === 'string' && runtime.check_name.length > 0 && validPositiveInteger(runtime.check_run_id)
    && runtime.app_id === CI_APP_ID;
  if (!runtimeComplete) addUnknown(unknownEvidence('COHERENCE', 'WORKFLOW_RUNTIME_CHECK_APP_IDENTITY', 'INCOMPLETE_RUNTIME_CHECK_IDENTITY', 'INCOMPLETE_EVIDENCE', 'PROVIDE_EXACT_WORKFLOW_RUNTIME_CHECK_AND_APP', ['runtime-check-identity']), 'WORKFLOW_RUNTIME_CHECK_APP_IDENTITY');
  if (runtimeComplete && observed && isObject(observed.runtime) && !exactObject(observed.runtime, runtime)) addRefuted('RUNTIME_DRIFT', 'WORKFLOW_RUNTIME_CHECK_APP_IDENTITY', 'observed runtime identity differs from authorized runtime identity');
  const kernel = input && input.kernel;
  const kernelComplete = isObject(kernel) && validDigest(kernel.before_sha256) && validDigest(kernel.after_sha256);
  if (!kernelComplete) addUnknown(unknownEvidence('COHERENCE', 'KERNEL_BEFORE_AFTER_DIGEST', 'INCOMPLETE_KERNEL_DIGEST_BINDING', 'INCOMPLETE_EVIDENCE', 'PROVIDE_KERNEL_BEFORE_AND_AFTER_DIGESTS', ['kernel-digest-binding']), 'KERNEL_BEFORE_AFTER_DIGEST');
  if (kernelComplete && observed && isObject(observed.kernel) && !exactObject(observed.kernel, kernel)) addRefuted('KERNEL_DRIFT', 'KERNEL_BEFORE_AFTER_DIGEST', 'observed kernel digest differs from authorized kernel digest');

  const authorization = input && input.authorization;
  const authorizationState = authorization && authorization.state;
  if (!isObject(authorization) || !['PENDING', 'AUTHORIZED', 'CONSUMED'].includes(authorizationState)) {
    addUnknown(unknownEvidence('REGRESSION', 'ONE_USE_NONCE', 'INCOMPLETE_AUTHORIZATION_STATE', 'INCOMPLETE_EVIDENCE', 'PROVIDE_AUTHORIZATION_STATE', ['authorization-state']), 'ONE_USE_NONCE');
  }
  if (authorization && authorization.owner_selection === null) {
    addUnknown(unknownEvidence('FOUNDATION', 'REPOSITORY_ACTOR_AUTHORITY', 'OWNER_SELECTION_MISSING', 'DIRECT_MISSING', 'SELECT_REPOSITORY_OWNER', []), 'REPOSITORY_ACTOR_AUTHORITY');
  } else if (authorization && !validIdentity(authorization.owner_selection)) {
    addUnknown(unknownEvidence('FOUNDATION', 'REPOSITORY_ACTOR_AUTHORITY', 'OWNER_SELECTION_INCOMPLETE', 'INCOMPLETE_EVIDENCE', 'PROVIDE_REPOSITORY_OWNER_SELECTION', ['owner-selection']), 'REPOSITORY_ACTOR_AUTHORITY');
  } else if (authorization && authorization.owner_selection && repository && (authorization.owner_selection.login !== repository.owner_login || authorization.owner_selection.id !== repository.owner_id || authorization.owner_selection.type !== repository.owner_type)) {
    addRefuted('ACTOR_MISMATCH', 'REPOSITORY_ACTOR_AUTHORITY', 'selected owner is not the exact repository owner identity');
  }
  if (authorization && authorization.nonce === null && authorizationState === 'PENDING') {
    addUnknown(unknownEvidence('REGRESSION', 'ONE_USE_NONCE', 'NONCE_NOT_ISSUED', 'DIRECT_MISSING', 'ISSUE_ONE_USE_NONCE', ['one-use-authorization']), 'ONE_USE_NONCE');
  } else if (!authorization || typeof authorization.nonce !== 'string' || authorization.nonce.length < 16) {
    addUnknown(unknownEvidence('REGRESSION', 'ONE_USE_NONCE', 'INCOMPLETE_NONCE', 'INCOMPLETE_EVIDENCE', 'PROVIDE_ONE_USE_NONCE', ['one-use-authorization']), 'ONE_USE_NONCE');
  }
  if (!authorization || !validTimestamp(authorization.issued_at) || !validTimestamp(authorization.expires_at)) {
    addUnknown(unknownEvidence('REGRESSION', 'EXPIRY_STALENESS', 'INCOMPLETE_EXPIRY_BINDING', 'INCOMPLETE_EVIDENCE', 'PROVIDE_ISSUED_AND_EXPIRY_TIMESTAMPS', ['expiry-staleness']), 'EXPIRY_STALENESS');
  } else if (Date.parse(now) >= Date.parse(authorization.expires_at)) {
    addRefuted('AUTHORIZATION_EXPIRED', 'EXPIRY_STALENESS', 'one-use authorization is expired at evaluation time');
  }
  if (authorization && authorization.stale === true) addRefuted('AUTHORIZATION_STALE', 'EXPIRY_STALENESS', 'one-use authorization is explicitly stale');
  if (!authorization || ![0, 1].includes(authorization.use_count)) {
    addUnknown(unknownEvidence('REGRESSION', 'ONE_USE_NONCE', 'INCOMPLETE_NONCE_USE_COUNT', 'INCOMPLETE_EVIDENCE', 'PROVIDE_NONCE_USE_COUNT', ['one-use-authorization']), 'ONE_USE_NONCE');
  } else if (authorization.use_count > 1) {
    addRefuted('NONCE_REPLAY', 'ONE_USE_NONCE', 'one-use nonce was used more than once');
  }
  if (authorization && authorization.reuse_attempts !== 0) addRefuted('NONCE_REPLAY', 'REPLAY_REUSE_REVOCATION_POST_ADOPTION', 'nonce reuse attempt was observed');
  if (authorization && authorization.revoked === true) addRefuted('AUTHORIZATION_REVOKED', 'REPLAY_REUSE_REVOCATION_POST_ADOPTION', 'one-use authorization was revoked');
  if (authorization && typeof authorization.revoked !== 'boolean') addUnknown(unknownEvidence('REGRESSION', 'REPLAY_REUSE_REVOCATION_POST_ADOPTION', 'INCOMPLETE_REVOCATION_STATE', 'INCOMPLETE_EVIDENCE', 'PROVIDE_REVOCATION_STATE', ['revocation-state']), 'REPLAY_REUSE_REVOCATION_POST_ADOPTION');

  if (authorizationState === 'AUTHORIZED' || authorizationState === 'CONSUMED') {
    const receipt = authorization && authorization.authorization_receipt;
    if (!isObject(receipt) || receipt.nonce !== authorization.nonce || receipt.candidate_digest !== candidateDigest(candidate) || !validIdentity(receipt.owner) || !authorization.owner_selection || !exactObject(receipt.owner, authorization.owner_selection)) {
      addUnknown(unknownEvidence('REGRESSION', 'ONE_USE_NONCE', 'INCOMPLETE_AUTHORIZATION_RECEIPT', 'INCOMPLETE_EVIDENCE', 'PROVIDE_EXACT_AUTHORIZATION_RECEIPT', ['authorization-receipt']), 'ONE_USE_NONCE');
    }
  }
  if (authorizationState === 'CONSUMED') {
    const consumption = input && input.consumption;
    if (!isObject(consumption) || consumption.nonce !== authorization.nonce || consumption.merge_parent_sha !== (candidate && candidate.base_sha) || consumption.replay_decision !== 'REFUTED' || consumption.post_adoption_verified !== true || !validTimestamp(consumption.consumed_at) || !validDigest(consumption.receipt_digest)) {
      addUnknown(unknownEvidence('REGRESSION', 'MERGE_PARENT_CONSUMPTION_RECEIPT', 'INCOMPLETE_CONSUMPTION_RECEIPT', 'INCOMPLETE_EVIDENCE', 'PROVIDE_EXACT_MERGE_PARENT_AND_CONSUMPTION_RECEIPT', ['consumption-receipt']), 'MERGE_PARENT_CONSUMPTION_RECEIPT');
    } else if (consumption.receipt_digest !== consumptionReceiptDigest(consumption)) {
      addRefuted('DIGEST_MISMATCH', 'MERGE_PARENT_CONSUMPTION_RECEIPT', 'consumption receipt digest does not match the exact receipt');
    }
  }

  const decision = resolveDecision([...refuted.map((item) => 'REFUTED'), ...unknown.map((item) => 'UNKNOWN')]);
  for (const item of refuted) mark(item.cellId, 'REFUTED');
  if (decision === 'UNKNOWN') {
    const selected = unknown.find((item) => item.evidence);
    if (selected) mark(selected.cellId, 'UNKNOWN');
  }
  return {
    schema: PROTOCOL_SCHEMA,
    decision,
    state: authorizationState || null,
    reason: decision === 'REFUTED' ? refuted[0].reason : decision === 'UNKNOWN' ? unknown[0].evidence.reason : 'EXACT_ONE_USE_AUTHORIZATION_LIFECYCLE_CLOSED',
    precedence: DECISION_PRECEDENCE.join('>'),
    cells,
    unknown: decision === 'UNKNOWN' ? unknown[0].evidence : null,
    refuted: decision === 'REFUTED' ? refuted[0] : null,
    repository_writes: 0,
  };
}

function buildRootAxiomReceipt({repository, owner, actor, pullRequest, baseBranch = 'dev', baseSha, headSha, audit, requiredOwnerAction = 'OWNER_CREATE_ONE_USE_AUTHORIZATION_FOR_EXACT_CANDIDATE'} = {}) {
  requireExact(repository === REPOSITORY, 'root axiom repository is not exact');
  requireExact(validIdentity(owner) && validIdentity(actor), 'root axiom owner or actor identity is malformed');
  requireExact(validPositiveInteger(pullRequest) && validSHA(baseSha) && validSHA(headSha), 'root axiom candidate identity is malformed');
  return {
    schema: ROOT_AXIOM_SCHEMA,
    decision: 'UNKNOWN',
    reason: 'NO_SERVER_SIDE_OWNER_AUTHORIZATION_OPERATION',
    repository,
    owner,
    actor,
    candidate: {pull_request: pullRequest, base_branch: baseBranch, base_sha: baseSha, head_sha: headSha},
    bootstrap_audit: isObject(audit) ? audit : {},
    unknown: unknownEvidence('FOUNDATION', 'REPOSITORY_ACTOR_AUTHORITY', 'NO_SERVER_SIDE_OWNER_AUTHORIZATION_OPERATION', 'ROOT_AXIOM_BLOCKED', requiredOwnerAction, ['repository-owner']),
    required_owner_action: requiredOwnerAction,
    no_check_bypass: true,
    no_force_push: true,
    repository_writes: 0,
  };
}

function validateRootAxiomReceipt(receipt) {
  requireExact(receipt && receipt.schema === ROOT_AXIOM_SCHEMA && receipt.decision === 'UNKNOWN' && receipt.reason === 'NO_SERVER_SIDE_OWNER_AUTHORIZATION_OPERATION', 'root axiom receipt decision is not exact');
  requireExact(receipt.repository === REPOSITORY && validIdentity(receipt.owner) && validIdentity(receipt.actor), 'root axiom receipt repository identity is not exact');
  requireExact(receipt.candidate && validPositiveInteger(receipt.candidate.pull_request) && receipt.candidate.base_branch === 'dev' && validSHA(receipt.candidate.base_sha) && validSHA(receipt.candidate.head_sha), 'root axiom receipt candidate identity is not exact');
  requireExact(isObject(receipt.bootstrap_audit) && receipt.required_owner_action === 'OWNER_CREATE_ONE_USE_AUTHORIZATION_FOR_EXACT_CANDIDATE', 'root axiom receipt audit or owner action is not exact');
  requireExact(exactArray(Object.keys(receipt.unknown || {}), UNKNOWN_FIELDS), 'root axiom receipt unknown fields are not exact');
  requireExact(receipt.unknown.stage === 'FOUNDATION' && receipt.unknown.step === 'REPOSITORY_ACTOR_AUTHORITY' && receipt.unknown.unknown_class === 'ROOT_AXIOM_BLOCKED' && receipt.unknown.next_operation === receipt.required_owner_action && exactArray(receipt.unknown.blocked_by, ['repository-owner']), 'root axiom receipt unknown evidence is not exact');
  requireExact(receipt.no_check_bypass === true && receipt.no_force_push === true && receipt.repository_writes === 0, 'root axiom receipt mutation safeguards are not exact');
  return receipt;
}

validateProtocolDefinition();

module.exports = {
  CI_APP_ID,
  DECISION_PRECEDENCE,
  PROTECTED_PATH_PATTERNS,
  PROTOCOL_DEFINITION_PATH,
  PROTOCOL_SCHEMA,
  REPOSITORY,
  ROOT_AXIOM_SCHEMA,
  UNKNOWN_FIELDS,
  buildRootAxiomReceipt,
  candidateDigest,
  canonicalPathList,
  canonicalPathNames,
  canonicalProtectedIntersection,
  canonicalTreeEntries,
  consumptionReceiptDigest,
  digestChangedPaths,
  digestKernelEntries,
  digestPathList,
  digestProtectedIntersection,
  digestTreeEntries,
  evaluate,
  resolveDecision,
  sha256,
  validateProtocolDefinition,
  validateRootAxiomReceipt,
  validDigest,
  validSHA,
};
