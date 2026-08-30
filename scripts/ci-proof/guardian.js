'use strict';

const crypto = require('node:crypto');
const foundationBootstrap = require('./foundation_bootstrap');

const ROOT_FAILURE_CODE = 'CI-ROOT-OF-TRUST-001';
const FOUNDATION_BOOTSTRAP_CODE = foundationBootstrap.FOUNDATION_BOOTSTRAP_CODE;
const HEAD_BINDING_STATUS = 'CI-GUARDIAN-HEAD-BINDING-UNVERIFIED';
const HEAD_BINDING_VERIFIED = 'verified';
const DEFAULT_BRANCH_CODE = 'CI-GUARDIAN-DEFAULT-BRANCH-001';
const LIVE_REF_CODE = 'CI-GUARDIAN-LIVE-REF-001';
const PROMOTION_TOPOLOGY_CODE = 'CI-GUARDIAN-PROMOTION-TOPOLOGY-001';
const CHECK_IDENTITY_CODE = 'CI-GUARDIAN-CHECK-IDENTITY-001';
const PROTECTION_CODE = 'CI-GUARDIAN-PROTECTION-001';
const INSTALLATION_SCOPE_CODE = 'CI-GUARDIAN-INSTALLATION-SCOPE-001';
const OBSERVER_ENVIRONMENT = 'guardian-observer';
const INSTALLATION_SCOPE_REPOSITORY = 'kimjooyoon/meta-ontology-go';
const OBSERVER_FRESHNESS_WINDOW_MS = 10 * 60 * 1000;
const GUARDIAN_SCHEMA = 'gooo/ci-guardian/v2';
const GUARDIAN_FAILURE_CODES = new Set([ROOT_FAILURE_CODE, FOUNDATION_BOOTSTRAP_CODE, DEFAULT_BRANCH_CODE, LIVE_REF_CODE, PROMOTION_TOPOLOGY_CODE, CHECK_IDENTITY_CODE, PROTECTION_CODE, INSTALLATION_SCOPE_CODE]);
const ALLOWED_BASES = new Set(['dev', 'main']);
const ALLOWED_ACTIONS = new Set(['opened', 'synchronize', 'reopened', 'ready_for_review']);
const PROOF_CONTEXTS = ['CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt'];
const DEV_PROTECTION_CONTEXTS = [...PROOF_CONTEXTS, 'CI guardian shadow'];
const MAIN_PROTECTION_CONTEXTS = [...PROOF_CONTEXTS, 'CI guardian'];
const VALID_STATUSES = new Set(['added', 'copied', 'changed', 'modified', 'removed', 'renamed']);
const PROTECTED_FILES = new Set([
  '.github/ci-governance.json',
  '.github/agent-scope-table.md',
  '.github/branch-policy.md',
  '.github/conformance-plan.md',
  'go.mod',
  'go.sum',
]);
const PROTECTED_PREFIXES = [
  '.github/workflows/',
  'scripts/ci-proof/',
  'scripts/ci-evidence/',
  'scripts/verify/',
  'internal/verify/',
];

function canonicalStringCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function guardianFailure(reason, code = ROOT_FAILURE_CODE) {
  const error = new Error(`${code}: ${reason}`);
  error.code = code;
  return error;
}

function validSHA(value) {
  return typeof value === 'string' && /^[0-9a-f]{40}$/i.test(value);
}

function validRepository(value) {
  return typeof value === 'string' && /^[^/\s]+\/[^/\s]+$/.test(value);
}

function validRef(value) {
  return typeof value === 'string' && value.length > 0 && !/[\s\\]/.test(value);
}

function expectedWorkflowRef(repository) {
  return `${repository}/.github/workflows/ci-guardian.yml@refs/heads/dev`;
}

function digestBranchProtection(snapshot) {
  const unsigned = {...snapshot, digest_sha256: ''};
  return crypto.createHash('sha256').update(JSON.stringify(unsigned)).digest('hex');
}

function digestGuardianEnvironment(snapshot) {
  const unsigned = {...snapshot, digest_sha256: ''};
  return crypto.createHash('sha256').update(JSON.stringify(unsigned)).digest('hex');
}

function observerFreshnessFromResponse(response) {
  const headers = response && response.headers;
  const rawDate = headers && typeof headers.get === 'function' ? headers.get('date') : headers && (headers.date || headers.Date);
  if (typeof rawDate !== 'string' || rawDate.trim() === '') return null;
  const observedMillis = Date.parse(rawDate);
  if (!Number.isFinite(observedMillis)) return null;
  const validUntilMillis = observedMillis + OBSERVER_FRESHNESS_WINDOW_MS;
  if (!Number.isFinite(validUntilMillis)) return null;
  return {observed_at: new Date(observedMillis).toISOString(), valid_until: new Date(validUntilMillis).toISOString()};
}

function validObserverFreshness(observedAt, validUntil, now = new Date()) {
  if (typeof observedAt !== 'string' || typeof validUntil !== 'string' || !(now instanceof Date) || !Number.isFinite(now.getTime())) return false;
  const observedMillis = Date.parse(observedAt);
  const validUntilMillis = Date.parse(validUntil);
  if (!Number.isFinite(observedMillis) || !Number.isFinite(validUntilMillis)) return false;
  return validUntilMillis - observedMillis === OBSERVER_FRESHNESS_WINDOW_MS && observedMillis <= now.getTime() && now.getTime() < validUntilMillis;
}

function emptyGuardianEnvironment({repository, tokenSource, runId, runAttempt, workflowSHA, missingReason}) {
  const snapshot = {
    repository: repository || null,
    name: OBSERVER_ENVIRONMENT,
    deployment_branch_policy: {protected_branches: false, custom_branch_policies: false},
    protection_rules: [],
    wait_timer: 0,
    reviewers: [],
    token_source: tokenSource || null,
    read_status: 'unavailable',
    missing_reason: missingReason || 'guardian_environment_api_unavailable',
    run_id: validPositiveInteger(runId) ? runId : null,
    run_attempt: validPositiveInteger(runAttempt) ? runAttempt : null,
    workflow_sha: workflowSHA || null,
    observed_at: null,
    valid_until: null,
    digest_sha256: '',
  };
  snapshot.digest_sha256 = digestGuardianEnvironment(snapshot);
  return snapshot;
}

function validateGuardianEnvironment(snapshot, {requireVerified = false, now = new Date()} = {}) {
  if (!snapshot || !validRepository(snapshot.repository) || snapshot.name !== OBSERVER_ENVIRONMENT || !snapshot.deployment_branch_policy || typeof snapshot.deployment_branch_policy.protected_branches !== 'boolean' || typeof snapshot.deployment_branch_policy.custom_branch_policies !== 'boolean' || !Array.isArray(snapshot.protection_rules) || !Number.isInteger(snapshot.wait_timer) || snapshot.wait_timer < 0 || !Array.isArray(snapshot.reviewers) || snapshot.reviewers.some((reviewer) => typeof reviewer !== 'string') || !validRef(snapshot.token_source) || !['verified', 'unavailable'].includes(snapshot.read_status) || typeof snapshot.missing_reason !== 'string' || !validPositiveInteger(snapshot.run_id) || !validPositiveInteger(snapshot.run_attempt) || !validSHA(snapshot.workflow_sha) || !/^[0-9a-f]{64}$/.test(snapshot.digest_sha256 || '')) {
    throw guardianFailure('guardian observer environment snapshot is malformed', PROTECTION_CODE);
  }
  if (snapshot.digest_sha256 !== digestGuardianEnvironment(snapshot)) throw guardianFailure('guardian observer environment digest mismatch', PROTECTION_CODE);
  if (snapshot.read_status === 'verified') {
    if (snapshot.token_source !== 'github.token' || snapshot.deployment_branch_policy.protected_branches !== true || snapshot.deployment_branch_policy.custom_branch_policies !== false || snapshot.protection_rules.length !== 1 || snapshot.protection_rules[0] !== 'branch_policy' || snapshot.wait_timer !== 0 || snapshot.reviewers.length !== 0 || snapshot.missing_reason !== '' || !validObserverFreshness(snapshot.observed_at, snapshot.valid_until, now)) {
      throw guardianFailure('guardian observer environment policy is not exact', PROTECTION_CODE);
    }
  } else if (snapshot.deployment_branch_policy.protected_branches || snapshot.deployment_branch_policy.custom_branch_policies || snapshot.protection_rules.length !== 0 || snapshot.wait_timer !== 0 || snapshot.reviewers.length !== 0 || snapshot.missing_reason === '' || snapshot.observed_at !== null || snapshot.valid_until !== null) {
    throw guardianFailure('guardian observer environment is not fail-closed', PROTECTION_CODE);
  }
  if (requireVerified && snapshot.read_status !== 'verified') throw guardianFailure('guardian observer environment evidence is unavailable', PROTECTION_CODE);
  return snapshot;
}

function digestInstallationRepositoryScope(snapshot) {
  const unsigned = {...snapshot, digest_sha256: ''};
  return crypto.createHash('sha256').update(JSON.stringify(unsigned)).digest('hex');
}

function emptyInstallationRepositoryScope({repository, installationId, tokenSource, runId, runAttempt, workflowSHA, missingReason}) {
  const snapshot = {
    repository: repository || null,
    installation_id: Number.isInteger(installationId) && installationId >= 0 ? installationId : 0,
    token_source: tokenSource || null,
    read_status: 'unavailable',
    repository_count: 0,
    repositories: [],
    exact_match: false,
    missing_reason: missingReason || 'installation_repository_scope_unavailable',
    run_id: validPositiveInteger(runId) ? runId : null,
    run_attempt: validPositiveInteger(runAttempt) ? runAttempt : null,
    workflow_sha: workflowSHA || null,
    observed_at: null,
    valid_until: null,
    digest_sha256: '',
  };
  snapshot.digest_sha256 = digestInstallationRepositoryScope(snapshot);
  return snapshot;
}

function validateInstallationRepositoryScope(snapshot, {requireVerified = false, expectedRepository, now = new Date()} = {}) {
  if (!snapshot || !validRepository(snapshot.repository) || !validRepository(expectedRepository) || snapshot.repository !== expectedRepository || !Number.isInteger(snapshot.installation_id) || snapshot.installation_id < 0 || snapshot.token_source !== 'github_app_installation' || !['verified', 'unavailable'].includes(snapshot.read_status) || !Number.isInteger(snapshot.repository_count) || snapshot.repository_count < 0 || !Array.isArray(snapshot.repositories) || snapshot.repository_count !== snapshot.repositories.length || snapshot.repositories.some((repository) => !validRepository(repository)) || typeof snapshot.exact_match !== 'boolean' || typeof snapshot.missing_reason !== 'string' || !validPositiveInteger(snapshot.run_id) || !validPositiveInteger(snapshot.run_attempt) || !validSHA(snapshot.workflow_sha) || !/^[0-9a-f]{64}$/.test(snapshot.digest_sha256 || '')) {
    throw guardianFailure('guardian installation repository scope attestation is malformed', INSTALLATION_SCOPE_CODE);
  }
  if (snapshot.digest_sha256 !== digestInstallationRepositoryScope(snapshot)) throw guardianFailure('guardian installation repository scope digest mismatch', INSTALLATION_SCOPE_CODE);
  if (snapshot.read_status === 'verified') {
    if (snapshot.installation_id < 1 || snapshot.repository_count !== 1 || JSON.stringify(snapshot.repositories) !== JSON.stringify([expectedRepository]) || snapshot.exact_match !== true || snapshot.missing_reason !== '' || !validObserverFreshness(snapshot.observed_at, snapshot.valid_until, now)) {
      throw guardianFailure('guardian installation repository scope is not exact', INSTALLATION_SCOPE_CODE);
    }
  } else if (snapshot.installation_id !== 0 || snapshot.repository_count !== 0 || snapshot.repositories.length !== 0 || snapshot.exact_match !== false || snapshot.missing_reason === '' || snapshot.observed_at !== null || snapshot.valid_until !== null) {
    throw guardianFailure('guardian installation repository scope is not fail-closed', INSTALLATION_SCOPE_CODE);
  }
  if (requireVerified && snapshot.read_status !== 'verified') throw guardianFailure('guardian installation repository scope is unavailable', INSTALLATION_SCOPE_CODE);
  return snapshot;
}

async function observeInstallationRepositoryScope({listRepositories, repository, installationId, tokenSource, runId, runAttempt, workflowSHA, now = new Date()}) {
  const unavailable = (reason) => emptyInstallationRepositoryScope({repository, installationId: 0, tokenSource, runId, runAttempt, workflowSHA, missingReason: reason});
  if (typeof listRepositories !== 'function' || repository !== INSTALLATION_SCOPE_REPOSITORY || !validPositiveInteger(installationId) || tokenSource !== 'github_app_installation' || !validPositiveInteger(runId) || !validPositiveInteger(runAttempt) || !validSHA(workflowSHA)) return unavailable('installation_repository_scope_observer_input_invalid');
  let response;
  try {
    response = await listRepositories({per_page: 100, page: 1});
  } catch (error) {
    return unavailable('installation_repository_scope_api_unavailable');
  }
  const freshness = observerFreshnessFromResponse(response);
  const data = response && response.data;
  if (!freshness || !response || (response.status !== undefined && response.status !== 200) || !data || !Number.isInteger(data.total_count) || data.total_count !== 1 || !Array.isArray(data.repositories) || data.repositories.length !== 1 || !data.repositories[0] || data.repositories[0].full_name !== INSTALLATION_SCOPE_REPOSITORY) return unavailable('installation_repository_scope_api_mismatch');
  const snapshot = {
    repository,
    installation_id: installationId,
    token_source: tokenSource,
    read_status: 'verified',
    repository_count: 1,
    repositories: [data.repositories[0].full_name],
    exact_match: true,
    missing_reason: '',
    run_id: runId,
    run_attempt: runAttempt,
    workflow_sha: workflowSHA,
    observed_at: freshness.observed_at,
    valid_until: freshness.valid_until,
    digest_sha256: '',
  };
  snapshot.digest_sha256 = digestInstallationRepositoryScope(snapshot);
  try {
    return validateInstallationRepositoryScope(snapshot, {requireVerified: true, expectedRepository: INSTALLATION_SCOPE_REPOSITORY, now});
  } catch (error) {
    return unavailable('installation_repository_scope_freshness_or_binding_mismatch');
  }
}

async function observeGuardianEnvironment({getEnvironment, repository, tokenSource, runId, runAttempt, workflowSHA, now = new Date()}) {
  const unavailable = (reason) => emptyGuardianEnvironment({repository, tokenSource, runId, runAttempt, workflowSHA, missingReason: reason});
  if (typeof getEnvironment !== 'function' || !validRepository(repository) || !validRef(tokenSource) || !validPositiveInteger(runId) || !validPositiveInteger(runAttempt) || !validSHA(workflowSHA)) return unavailable('guardian_environment_observer_input_invalid');
  let response;
  try {
    response = await getEnvironment({owner: repository.split('/')[0], repo: repository.split('/')[1], environment_name: OBSERVER_ENVIRONMENT});
  } catch (error) {
    return unavailable('guardian_environment_api_unavailable');
  }
  const freshness = observerFreshnessFromResponse(response);
  if (!freshness) return unavailable('guardian_environment_response_date_missing_or_malformed');
  const data = response && response.data;
  const policy = data && data.deployment_branch_policy;
  const rules = data && data.protection_rules;
  const hasOwn = (object, key) => Object.prototype.hasOwnProperty.call(object || {}, key);
  if (!response || (response.status !== undefined && response.status !== 200) || !data || data.name !== OBSERVER_ENVIRONMENT || !policy || !hasOwn(policy, 'protected_branches') || typeof policy.protected_branches !== 'boolean' || !hasOwn(policy, 'custom_branch_policies') || typeof policy.custom_branch_policies !== 'boolean' || policy.protected_branches !== true || policy.custom_branch_policies !== false || hasOwn(data, 'wait_timer') || hasOwn(data, 'reviewers') || !Array.isArray(rules) || rules.length === 0) return unavailable('guardian_environment_api_malformed');
  const seenRuleTypes = new Set();
  for (const rule of rules) {
    if (!rule || typeof rule !== 'object' || Array.isArray(rule) || !hasOwn(rule, 'type') || typeof rule.type !== 'string' || rule.type.length === 0 || seenRuleTypes.has(rule.type) || !['branch_policy', 'required_reviewers', 'wait_timer'].includes(rule.type)) return unavailable('guardian_environment_api_malformed');
    seenRuleTypes.add(rule.type);
    if (rule.type !== 'branch_policy') return unavailable('guardian_environment_api_malformed');
  }
  if (seenRuleTypes.size !== 1 || rules.length !== 1 || !seenRuleTypes.has('branch_policy')) return unavailable('guardian_environment_api_malformed');
  const snapshot = {
    repository,
    name: OBSERVER_ENVIRONMENT,
    deployment_branch_policy: {protected_branches: policy.protected_branches, custom_branch_policies: policy.custom_branch_policies},
    protection_rules: ['branch_policy'],
    wait_timer: 0,
    reviewers: [],
    token_source: tokenSource,
    read_status: 'verified',
    missing_reason: '',
    run_id: runId,
    run_attempt: runAttempt,
    workflow_sha: workflowSHA,
    observed_at: freshness.observed_at,
    valid_until: freshness.valid_until,
    digest_sha256: '',
  };
  snapshot.digest_sha256 = digestGuardianEnvironment(snapshot);
  try {
    validateGuardianEnvironment(snapshot, {requireVerified: true, now});
  } catch (error) {
    return emptyGuardianEnvironment({repository, tokenSource, runId, runAttempt, workflowSHA, missingReason: 'guardian_environment_freshness_or_policy_mismatch'});
  }
  return snapshot;
}

function protectionContextsForBranch(branch) {
  if (branch === 'dev') return DEV_PROTECTION_CONTEXTS;
  if (branch === 'main') return MAIN_PROTECTION_CONTEXTS;
  return [];
}

function emptyBranchProtection({branch, repository, policySHA, eventRef, checkoutRef, baseSHA, headSHA, runId, runAttempt, workflowSHA, tokenSource, missingReason, appInstallationId = 0, appSlug = ''}) {
  const snapshot = {
    repository: repository || null,
    branch: branch || null,
    policy_sha256: policySHA || null,
    event_ref: eventRef || null,
    checkout_ref: checkoutRef || null,
    token_source: tokenSource || null,
    app_installation_id: appInstallationId,
    app_slug: appSlug,
    read_status: 'unavailable',
    exists: false,
    strict: false,
    required_checks: [],
    required_check_bindings: [],
    enforce_admins: false,
    required_reviews: 0,
    dismiss_stale_reviews: false,
    require_last_push_approval: false,
    linear_history: false,
    allow_force_pushes: false,
    allow_deletions: false,
    required_signatures: false,
    required_conversation_resolution: false,
    block_creations: false,
    lock_branch: false,
    allow_fork_syncing: false,
    restrictions: null,
    missing_reason: missingReason || 'branch_protection_api_unavailable',
    base_sha: baseSHA || null,
    head_sha: headSHA || null,
    run_id: validPositiveInteger(runId) ? runId : null,
    run_attempt: validPositiveInteger(runAttempt) ? runAttempt : null,
    workflow_sha: workflowSHA || null,
    observed_at: null,
    valid_until: null,
    digest_sha256: '',
  };
  snapshot.digest_sha256 = digestBranchProtection(snapshot);
  return snapshot;
}

function sameSet(left, right) {
  if (!Array.isArray(left) || !Array.isArray(right) || left.length !== right.length) return false;
  const expected = [...right].sort(canonicalStringCompare);
  const actual = [...left].sort(canonicalStringCompare);
  return actual.every((value, index) => value === expected[index]);
}

function validRequiredCheckBindings(bindings, contexts) {
  if (!Array.isArray(bindings) || bindings.length !== contexts.length) return false;
  const actual = bindings.map((binding) => binding && `${binding.context}\u0000${binding.app_id}`).sort(canonicalStringCompare);
  const expected = contexts.map((context) => `${context}\u000015368`).sort(canonicalStringCompare);
  return actual.every((value, index) => value === expected[index]);
}

function validatePublicBranchSummary(data) {
  const checks = data && data.required_status_checks;
  if (!data || data.protected !== true || !checks || !Array.isArray(checks.contexts) || !Array.isArray(checks.checks)) return false;
  if (!sameSet(checks.contexts, MAIN_PROTECTION_CONTEXTS)) return false;
  return validRequiredCheckBindings(checks.checks.map((check) => ({context: check && check.context, app_id: Number(check && check.app_id)})), MAIN_PROTECTION_CONTEXTS);
}

function validateBranchProtectionSnapshot(snapshot, {requireVerified = false, expectedBranch = null, expectedContexts = null, now = new Date()} = {}) {
  const branch = expectedBranch || (snapshot && snapshot.branch);
  const contexts = expectedContexts || protectionContextsForBranch(branch);
  if (!snapshot || !['dev', 'main'].includes(branch) || snapshot.branch !== branch || contexts.length === 0 || !validRepository(snapshot.repository) || !/^[0-9a-f]{64}$/.test(snapshot.policy_sha256 || '') || !validRef(snapshot.event_ref) || !validRef(snapshot.checkout_ref) || !validSHA(snapshot.base_sha) || !validSHA(snapshot.head_sha) || !validPositiveInteger(snapshot.run_id) || !validPositiveInteger(snapshot.run_attempt) || !validSHA(snapshot.workflow_sha) || !validRef(snapshot.token_source) || !['verified', 'unavailable'].includes(snapshot.read_status) || !Array.isArray(snapshot.required_checks) || !Array.isArray(snapshot.required_check_bindings) || typeof snapshot.required_signatures !== 'boolean' || typeof snapshot.required_conversation_resolution !== 'boolean' || typeof snapshot.block_creations !== 'boolean' || typeof snapshot.lock_branch !== 'boolean' || typeof snapshot.allow_fork_syncing !== 'boolean' || (snapshot.restrictions !== undefined && snapshot.restrictions !== null) || !/^([0-9a-f]{64})$/.test(snapshot.digest_sha256 || '')) {
    throw guardianFailure('branch protection snapshot is malformed', PROTECTION_CODE);
  }
  if (snapshot.digest_sha256 !== digestBranchProtection(snapshot)) throw guardianFailure('branch protection snapshot digest mismatch', PROTECTION_CODE);
  if (snapshot.read_status === 'verified') {
    if (snapshot.token_source !== 'github_app_installation' || !validPositiveInteger(snapshot.app_installation_id) || typeof snapshot.app_slug !== 'string' || snapshot.app_slug.length === 0 || !snapshot.exists || snapshot.strict !== true || !sameSet(snapshot.required_checks, contexts) || !validRequiredCheckBindings(snapshot.required_check_bindings, contexts) || snapshot.enforce_admins !== true || snapshot.required_reviews !== 0 || snapshot.dismiss_stale_reviews !== false || snapshot.require_last_push_approval !== false || snapshot.linear_history !== true || snapshot.allow_force_pushes !== false || snapshot.allow_deletions !== false || snapshot.required_signatures !== false || snapshot.required_conversation_resolution !== false || snapshot.block_creations !== false || snapshot.lock_branch !== false || snapshot.allow_fork_syncing !== false || snapshot.restrictions !== null || snapshot.missing_reason !== '' || !validObserverFreshness(snapshot.observed_at, snapshot.valid_until, now)) {
      throw guardianFailure('full branch protection snapshot is incomplete or not exact', PROTECTION_CODE);
    }
  } else if (snapshot.exists || snapshot.strict || snapshot.required_checks.length !== 0 || snapshot.required_check_bindings.length !== 0 || snapshot.required_signatures || snapshot.required_conversation_resolution || snapshot.block_creations || snapshot.lock_branch || snapshot.allow_fork_syncing || snapshot.restrictions !== null || snapshot.missing_reason === '' || snapshot.observed_at !== null || snapshot.valid_until !== null) {
    throw guardianFailure('unavailable branch protection snapshot is not fail-closed', PROTECTION_CODE);
  }
  if (requireVerified && snapshot.read_status !== 'verified') throw guardianFailure('full branch protection observer evidence is unavailable', PROTECTION_CODE);
  return snapshot;
}

async function observeBranchProtection({getProtection, branch = 'main', expectedContexts = protectionContextsForBranch(branch), repository, policySHA, eventRef, checkoutRef, baseSHA, headSHA, runId, runAttempt, workflowSHA, tokenSource, appInstallationId = 0, appSlug = '', now = new Date()}) {
  const unavailable = (reason) => ({...emptyBranchProtection({branch, repository, policySHA, eventRef, checkoutRef, baseSHA, headSHA, runId, runAttempt, workflowSHA, tokenSource, missingReason: reason, appInstallationId, appSlug}), digest_sha256: ''});
  const finishUnavailable = (reason) => { const snapshot = unavailable(reason); snapshot.digest_sha256 = digestBranchProtection(snapshot); return snapshot; };
  if (typeof getProtection !== 'function' || !validRepository(repository) || !/^[0-9a-f]{64}$/.test(policySHA || '') || !validRef(eventRef) || !validRef(checkoutRef) || !validSHA(baseSHA) || !validSHA(headSHA) || !validPositiveInteger(runId) || !validPositiveInteger(runAttempt) || !validSHA(workflowSHA) || !validRef(tokenSource)) return finishUnavailable('branch_protection_observer_input_invalid');
  let response;
  try {
    response = await getProtection({owner: repository.split('/')[0], repo: repository.split('/')[1], branch});
  } catch (error) {
    return finishUnavailable('branch_protection_api_unavailable');
  }
  const freshness = observerFreshnessFromResponse(response);
  if (!freshness) return finishUnavailable('branch_protection_response_date_missing_or_malformed');
  const data = response && response.data;
  const checks = data && data.required_status_checks;
  const reviews = data && data.required_pull_request_reviews;
  const hasOwn = (object, key) => Object.prototype.hasOwnProperty.call(object || {}, key);
  const boolObject = (object, key) => Boolean(object && hasOwn(object, key) && typeof object[key] === 'boolean');
  if (!response || (response.status !== undefined && response.status !== 200) || !data || !hasOwn(data, 'required_status_checks') || !checks || !boolObject(checks, 'strict') || !Array.isArray(checks.contexts) || !Array.isArray(checks.checks) || !hasOwn(data, 'enforce_admins') || !boolObject(data.enforce_admins, 'enabled') || !hasOwn(data, 'required_linear_history') || !boolObject(data.required_linear_history, 'enabled') || !hasOwn(data, 'allow_force_pushes') || !boolObject(data.allow_force_pushes, 'enabled') || !hasOwn(data, 'allow_deletions') || !boolObject(data.allow_deletions, 'enabled') || (reviews !== null && reviews !== undefined) || !hasOwn(data, 'required_signatures') || !boolObject(data.required_signatures, 'enabled') || !hasOwn(data, 'required_conversation_resolution') || !boolObject(data.required_conversation_resolution, 'enabled') || !hasOwn(data, 'block_creations') || !boolObject(data.block_creations, 'enabled') || !hasOwn(data, 'lock_branch') || !boolObject(data.lock_branch, 'enabled') || !hasOwn(data, 'allow_fork_syncing') || !boolObject(data.allow_fork_syncing, 'enabled') || (hasOwn(data, 'restrictions') && data.restrictions !== null) || checks.contexts.some((context) => typeof context !== 'string') || checks.checks.some((check) => !check || typeof check !== 'object' || !hasOwn(check, 'context') || typeof check.context !== 'string' || !hasOwn(check, 'app_id') || !Number.isInteger(check.app_id))) return finishUnavailable('branch_protection_api_malformed');
  const bindings = checks.checks.map((check) => ({context: check && check.context, app_id: Number(check && check.app_id)}));
  const snapshot = {
    repository,
    branch,
    policy_sha256: policySHA,
    event_ref: eventRef,
    checkout_ref: checkoutRef,
    token_source: tokenSource,
    app_installation_id: appInstallationId,
    app_slug: appSlug,
    read_status: 'verified',
    exists: true,
    strict: Boolean(checks.strict),
    required_checks: [...checks.contexts].sort(canonicalStringCompare),
    required_check_bindings: bindings.sort((left, right) => canonicalStringCompare(`${left.context}\u0000${left.app_id}`, `${right.context}\u0000${right.app_id}`)),
    enforce_admins: Boolean(data.enforce_admins && data.enforce_admins.enabled),
    required_reviews: Number(reviews && reviews.required_approving_review_count || 0),
    dismiss_stale_reviews: Boolean(reviews && reviews.dismiss_stale_reviews),
    require_last_push_approval: Boolean(reviews && reviews.require_last_push_approval),
    linear_history: Boolean(data.required_linear_history && data.required_linear_history.enabled),
    allow_force_pushes: Boolean(data.allow_force_pushes && data.allow_force_pushes.enabled),
    allow_deletions: Boolean(data.allow_deletions && data.allow_deletions.enabled),
    required_signatures: Boolean(data.required_signatures && data.required_signatures.enabled),
    required_conversation_resolution: Boolean(data.required_conversation_resolution && data.required_conversation_resolution.enabled),
    block_creations: Boolean(data.block_creations && data.block_creations.enabled),
    lock_branch: Boolean(data.lock_branch && data.lock_branch.enabled),
    allow_fork_syncing: Boolean(data.allow_fork_syncing && data.allow_fork_syncing.enabled),
    restrictions: null,
    missing_reason: '',
    base_sha: baseSHA,
    head_sha: headSHA,
    run_id: runId,
    run_attempt: runAttempt,
    workflow_sha: workflowSHA,
    observed_at: freshness.observed_at,
    valid_until: freshness.valid_until,
    digest_sha256: '',
  };
  snapshot.digest_sha256 = digestBranchProtection(snapshot);
  try {
    validateBranchProtectionSnapshot(snapshot, {requireVerified: true, expectedBranch: branch, expectedContexts, now});
  } catch (error) {
    return finishUnavailable('branch_protection_policy_mismatch');
  }
  return snapshot;
}

function routeForPull(pull) {
  const identity = pullIdentity(pull);
  if (foundationBootstrap.exactIdentity(pull)) return foundationBootstrap.FOUNDATION_ROUTE;
  if (identity.base_ref === 'dev' && typeof identity.head_ref === 'string' && identity.head_ref.startsWith('agent/')) return 'feature_dev';
  if (identity.base_ref === 'main' && identity.head_ref === 'dev') return 'promotion_main';
  return null;
}

function checkNameForRoute(route) {
  if (route === foundationBootstrap.FOUNDATION_ROUTE) return 'CI guardian shadow';
  if (route === 'feature_dev') return 'CI guardian shadow';
  if (route === 'promotion_main') return 'CI guardian';
  return null;
}

function validateLiveRefResponse(response, ref) {
  if (!response || (response.status !== undefined && response.status !== 200) || !response.data || !response.data.object || !validSHA(response.data.object.sha)) {
    throw guardianFailure(`live ref ${ref} is unavailable or malformed`, LIVE_REF_CODE);
  }
  return response.data.object.sha;
}

async function readLiveTopology({getRef, compareCommits, owner, repo}) {
  if (typeof getRef !== 'function' || typeof compareCommits !== 'function' || typeof owner !== 'string' || typeof repo !== 'string') {
    throw guardianFailure('live ref API binding is missing or malformed', LIVE_REF_CODE);
  }
  let devResponse;
  let mainResponse;
  try {
    [devResponse, mainResponse] = await Promise.all([
      getRef({owner, repo, ref: 'heads/dev'}),
      getRef({owner, repo, ref: 'heads/main'}),
    ]);
  } catch (error) {
    throw guardianFailure(`live ref API failed: ${error.message || error}`, LIVE_REF_CODE);
  }
  const devSha = validateLiveRefResponse(devResponse, 'heads/dev');
  const mainSha = validateLiveRefResponse(mainResponse, 'heads/main');
  let compareResponse;
  try {
    compareResponse = await compareCommits({owner, repo, base: mainSha, head: devSha});
  } catch (error) {
    throw guardianFailure(`live topology compare API failed: ${error.message || error}`, PROMOTION_TOPOLOGY_CODE);
  }
  const data = compareResponse && compareResponse.data;
  if (!compareResponse || (compareResponse.status !== undefined && compareResponse.status !== 200) || !data || !['ahead', 'behind', 'identical', 'diverged'].includes(data.status) || !Number.isInteger(data.ahead_by) || !Number.isInteger(data.behind_by) || !data.merge_base_commit || !validSHA(data.merge_base_commit.sha)) {
    throw guardianFailure('live topology compare response is malformed', PROMOTION_TOPOLOGY_CODE);
  }
  return {
    refs: {dev_sha: devSha, main_sha: mainSha},
    topology: {status: data.status, ahead_by: data.ahead_by, behind_by: data.behind_by, merge_base_sha: data.merge_base_commit.sha},
  };
}

function validPositiveInteger(value) {
  return Number.isInteger(value) && value > 0;
}

function normalizePath(value) {
  if (typeof value !== 'string' || value.length === 0 || value.startsWith('/') || value.includes('\\')) {
    throw guardianFailure('changed file path is missing or malformed');
  }
  const parts = value.split('/');
  if (parts.some((part) => part.length === 0 || part === '.' || part === '..')) {
    throw guardianFailure(`changed file path is unsafe: ${value}`);
  }
  return value;
}

function isProtectedKernelPath(value) {
  return kernelPathKind(value) !== null;
}

function kernelPathKind(value) {
  const path = normalizePath(value);
  if (PROTECTED_FILES.has(path)) {
    return 'file';
  }
  if (PROTECTED_PREFIXES.some((prefix) => path === prefix.slice(0, -1) || path.startsWith(prefix))) {
    return 'prefix';
  }
  return null;
}

function validateGuardianPullRequest(pull) {
  if (!pull || !Number.isInteger(pull.number) || pull.number < 1) {
    throw guardianFailure('pull request identity is missing or malformed');
  }
  const base = pull.base;
  if (!base || !ALLOWED_BASES.has(base.ref) || !validSHA(base.sha)) {
    throw guardianFailure('pull request base branch or SHA is missing or unsupported');
  }
  if (!base.repo || typeof base.repo.full_name !== 'string' || !base.repo.full_name.includes('/')) {
    throw guardianFailure('pull request base repository is missing or malformed');
  }
  if (!pull.head || typeof pull.head.ref !== 'string' || pull.head.ref.length === 0 || !validSHA(pull.head.sha) || !pull.head.repo || typeof pull.head.repo.full_name !== 'string' || !pull.head.repo.full_name.includes('/')) {
    throw guardianFailure('pull request head identity is missing or malformed');
  }
  return pull;
}

function validatePromotionPullRequestState(pull) {
  const hasOwn = (key) => Object.prototype.hasOwnProperty.call(pull || {}, key);
  if (!pull || pull.state !== 'open' || pull.draft !== false || (hasOwn('merged') && pull.merged !== false) || (hasOwn('merged_at') && pull.merged_at !== null)) {
    throw guardianFailure('promotion pull request is not open, ready, and unmerged', PROMOTION_TOPOLOGY_CODE);
  }
  return pull;
}

function validateChangedFile(file) {
  if (!file || typeof file !== 'object' || !VALID_STATUSES.has(file.status)) {
    throw guardianFailure('changed-file response contains a missing or unsupported status');
  }
  const filename = normalizePath(file.filename);
  if (Object.prototype.hasOwnProperty.call(file, 'previous_filename') && file.previous_filename !== null) {
    normalizePath(file.previous_filename);
  }
  if (file.status === 'renamed' && typeof file.previous_filename !== 'string') {
    throw guardianFailure(`renamed file has no previous_filename: ${filename}`);
  }
  return {filename, previous_filename: file.previous_filename || null, status: file.status};
}

async function inspectChangedFiles({listFiles, owner, repo, baseRepoFullName, pullNumber, expectedCount}) {
  if (typeof listFiles !== 'function' || typeof owner !== 'string' || typeof repo !== 'string' || baseRepoFullName !== `${owner}/${repo}` || !Number.isInteger(pullNumber) || pullNumber < 1 || !Number.isInteger(expectedCount) || expectedCount < 0) {
    throw guardianFailure('changed-file API binding is missing or malformed');
  }
  const files = [];
  for (let page = 1; page <= 1000; page += 1) {
    let response;
    try {
      response = await listFiles({owner, repo, pull_number: pullNumber, per_page: 100, page});
    } catch (error) {
      throw guardianFailure(`changed-file API failed on page ${page}: ${error.message || error}`);
    }
    if (!response || (response.status !== undefined && response.status !== 200) || !Array.isArray(response.data)) {
      throw guardianFailure(`changed-file API returned malformed data on page ${page}`);
    }
    for (const file of response.data) {
      const validated = validateChangedFile(file);
      files.push(validated);
    }
    if (response.data.length < 100) {
      if (files.length !== expectedCount) {
        throw guardianFailure(`changed-file API count ${files.length} does not match live pull count ${expectedCount}`);
      }
      const kernelPaths = files
        .flatMap((file) => [file.filename, file.previous_filename])
        .filter((path) => path && isProtectedKernelPath(path));
      if (kernelPaths.length > 0) {
        return {decision: 'FAIL_CLOSED', code: ROOT_FAILURE_CODE, reason: `protected kernel path changed: ${kernelPaths.sort(canonicalStringCompare).join(', ')}`, files, kernelPaths: [...new Set(kernelPaths)].sort(canonicalStringCompare)};
      }
      return {decision: 'PASS', code: null, reason: null, files, kernelPaths: []};
    }
  }
  throw guardianFailure('changed-file pagination exceeded the fail-closed page limit');
}

function pullIdentity(pull) {
  return {
    number: pull && pull.number,
    base_repo: pull && pull.base && pull.base.repo && pull.base.repo.full_name,
    base_ref: pull && pull.base && pull.base.ref,
    base_sha: pull && pull.base && pull.base.sha,
    head_repo: pull && pull.head && pull.head.repo && pull.head.repo.full_name,
    head_ref: pull && pull.head && pull.head.ref,
    head_sha: pull && pull.head && pull.head.sha,
  };
}

function sameIdentity(left, right) {
  return JSON.stringify(pullIdentity(left)) === JSON.stringify(pullIdentity(right));
}

async function revalidatePullRequest({getPull, owner, repo, pullNumber, eventPull}) {
  if (typeof getPull !== 'function' || typeof owner !== 'string' || typeof repo !== 'string' || !validPositiveInteger(pullNumber)) {
    throw guardianFailure('live pull request API binding is missing or malformed');
  }
  let response;
  try {
    response = await getPull({owner, repo, pull_number: pullNumber});
  } catch (error) {
    throw guardianFailure(`live pull request API failed: ${error.message || error}`);
  }
  if (!response || (response.status !== undefined && response.status !== 200) || !response.data || !Number.isInteger(response.data.changed_files) || response.data.changed_files < 0) {
    throw guardianFailure('live pull request API returned malformed data');
  }
  const livePull = validateGuardianPullRequest(response.data);
  if (livePull.base.repo.full_name !== `${owner}/${repo}` || !sameIdentity(eventPull, livePull)) {
    throw guardianFailure('pull request base/head changed during guardian inspection');
  }
  if (routeForPull(livePull) === 'promotion_main') validatePromotionPullRequestState(livePull);
  return livePull;
}

async function kernelTreeDigest({getCommit, getTree, owner, repo, ref}) {
  if (typeof getCommit !== 'function' || typeof getTree !== 'function' || typeof owner !== 'string' || typeof repo !== 'string' || !validSHA(ref)) {
    throw guardianFailure('kernel digest API binding is missing or malformed');
  }
  let commitResponse;
  try {
    commitResponse = await getCommit({owner, repo, ref});
  } catch (error) {
    throw guardianFailure(`kernel commit API failed: ${error.message || error}`);
  }
  const treeSHA = commitResponse && commitResponse.data && commitResponse.data.commit && commitResponse.data.commit.tree && commitResponse.data.commit.tree.sha;
  if (!commitResponse || (commitResponse.status !== undefined && commitResponse.status !== 200) || !validSHA(treeSHA)) {
    throw guardianFailure('kernel commit response is missing an exact tree SHA');
  }
  let treeResponse;
  try {
    treeResponse = await getTree({owner, repo, tree_sha: treeSHA, recursive: '1'});
  } catch (error) {
    throw guardianFailure(`kernel tree API failed: ${error.message || error}`);
  }
  if (!treeResponse || (treeResponse.status !== undefined && treeResponse.status !== 200) || !treeResponse.data || treeResponse.data.sha !== treeSHA || treeResponse.data.truncated === true || !Array.isArray(treeResponse.data.tree)) {
    throw guardianFailure('kernel tree response is missing or truncated');
  }
  const seen = new Set();
  const entries = [];
  for (const entry of treeResponse.data.tree) {
    if (!entry || typeof entry.path !== 'string' || typeof entry.type !== 'string' || !validSHA(entry.sha)) {
      throw guardianFailure('kernel tree response contains a malformed entry');
    }
    const path = normalizePath(entry.path);
    if (seen.has(path)) {
      throw guardianFailure(`kernel tree response contains a duplicate path: ${path}`);
    }
    seen.add(path);
    const kind = kernelPathKind(path);
    if (kind === null) {
      continue;
    }
    if (!['blob', 'tree'].includes(entry.type) || (kind === 'file' && entry.type !== 'blob')) {
      throw guardianFailure(`kernel tree response contains an unsupported protected entry: ${path}`);
    }
    entries.push({path, type: entry.type, mode: entry.mode || null, sha: entry.sha});
  }
  if (entries.length === 0) {
    throw guardianFailure('kernel tree response contains no protected entries');
  }
  entries.sort((left, right) => canonicalStringCompare([left.path, left.type, left.sha].join('\u0000'), [right.path, right.type, right.sha].join('\u0000')));
  return `sha256:${crypto.createHash('sha256').update(JSON.stringify(entries)).digest('hex')}`;
}

function defaultBranchDecision(defaultBranch, eventRef) {
  if (defaultBranch !== 'dev' || eventRef !== `refs/heads/${defaultBranch}`) {
    return {decision: 'FAIL_CLOSED', code: DEFAULT_BRANCH_CODE, reason: 'guardian is not executing from the protected dev default branch'};
  }
  return {decision: 'PASS', code: null, reason: null};
}

function trustedDevPromotion({pull, repository, defaultBranch, workflowRef, workflowSha, runtimeSha, liveBefore, liveAfter, checkName}) {
  const identity = pullIdentity(pull);
  const liveMatches = validLiveSnapshot(liveBefore) && validLiveSnapshot(liveAfter) && liveBefore.refs.dev_sha === liveAfter.refs.dev_sha && liveBefore.refs.main_sha === liveAfter.refs.main_sha;
  const promotionTopology = validPromotionTopology(liveBefore) && validPromotionTopology(liveAfter);
  return defaultBranch === 'dev' && workflowRef === expectedWorkflowRef(repository) && runtimeSha === workflowSha && checkName === 'CI guardian' && identity.base_repo === repository && identity.head_repo === repository && identity.base_ref === 'main' && identity.head_ref === 'dev' && identity.head_sha === workflowSha && validSHA(workflowSha) && identity.base_sha === liveBefore?.refs.main_sha && identity.head_sha === liveBefore?.refs.dev_sha && identity.head_sha === liveAfter?.refs.dev_sha && identity.base_sha === liveAfter?.refs.main_sha && liveMatches && promotionTopology;
}

function validLiveSnapshot(snapshot) {
  return Boolean(snapshot && snapshot.refs && validSHA(snapshot.refs.dev_sha) && validSHA(snapshot.refs.main_sha) && snapshot.topology && ['ahead', 'behind', 'identical', 'diverged'].includes(snapshot.topology.status) && Number.isInteger(snapshot.topology.ahead_by) && snapshot.topology.ahead_by >= 0 && Number.isInteger(snapshot.topology.behind_by) && snapshot.topology.behind_by >= 0 && validSHA(snapshot.topology.merge_base_sha));
}

function validPromotionTopology(snapshot) {
  return validLiveSnapshot(snapshot) && snapshot.topology.status === 'ahead' && snapshot.topology.ahead_by > 0 && snapshot.topology.behind_by === 0 && snapshot.topology.merge_base_sha === snapshot.refs.main_sha;
}

function classifyGuardianDecision({pull, repository, defaultBranch, workflowRef, eventRef, workflowSha, runtimeSha, result, kernelBeforeDigest, kernelAfterDigest, liveBefore, liveAfter, checkName}) {
  const route = pullIdentity(pull);
  const routeName = routeForPull(pull);
  const expectedCheckName = checkNameForRoute(routeName);
  if (!routeName || !expectedCheckName || (checkName && checkName !== expectedCheckName)) {
    return {...result, decision: 'FAIL_CLOSED', code: CHECK_IDENTITY_CODE, reason: 'guardian route or check identity is not canonical'};
  }
  if (defaultBranch !== 'dev' || eventRef !== 'refs/heads/dev' || workflowRef !== expectedWorkflowRef(repository) || runtimeSha !== workflowSha || !validSHA(workflowSha) || route.base_repo !== repository || route.head_repo !== repository) {
    return {...result, decision: 'FAIL_CLOSED', code: DEFAULT_BRANCH_CODE, reason: 'guardian runtime is not the protected default-dev authority'};
  }
  if (!validLiveSnapshot(liveBefore) || !validLiveSnapshot(liveAfter) || liveBefore.refs.dev_sha !== liveAfter.refs.dev_sha || liveBefore.refs.main_sha !== liveAfter.refs.main_sha) {
    return {...result, decision: 'FAIL_CLOSED', code: LIVE_REF_CODE, reason: 'guardian live refs are missing, malformed, or drifted'};
  }
  const promotion = trustedDevPromotion({pull, repository, defaultBranch, workflowRef, workflowSha, runtimeSha, liveBefore, liveAfter, checkName: expectedCheckName});
  const featureRoute = routeName === 'feature_dev';
  const foundationRoute = routeName === foundationBootstrap.FOUNDATION_ROUTE;
  if (featureRoute && expectedCheckName !== 'CI guardian shadow') {
    return {...result, decision: 'FAIL_CLOSED', code: CHECK_IDENTITY_CODE, reason: 'feature route must emit the shadow guardian check'};
  }
  if (featureRoute && (route.base_sha !== workflowSha || route.base_sha !== liveBefore.refs.dev_sha || liveBefore.refs.dev_sha !== liveAfter.refs.dev_sha)) {
    return {...result, decision: 'FAIL_CLOSED', code: LIVE_REF_CODE, reason: 'feature base, workflow, and live dev SHAs are not identical'};
  }
  if (foundationRoute && (expectedCheckName !== 'CI guardian shadow' || route.base_ref !== 'dev' || !foundationBootstrap.exactIdentity(pull))) {
    return {...result, decision: 'FAIL_CLOSED', code: foundationBootstrap.FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap route identity is not exact'};
  }
  if (foundationRoute && result.decision !== 'PASS') {
    return {...result, decision: 'FAIL_CLOSED', code: foundationBootstrap.FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap route was not explicitly authorized'};
  }
  if (foundationRoute && result.reason !== 'FOUNDATION_OVERRIDE_USED=1') {
    return {...result, decision: 'FAIL_CLOSED', code: foundationBootstrap.FOUNDATION_BOOTSTRAP_CODE, reason: 'foundation bootstrap route lacks the explicit one-time override marker'};
  }
  if (result.decision === 'PASS' && featureRoute && route.base_sha !== workflowSha) {
    return {...result, decision: 'FAIL_CLOSED', code: LIVE_REF_CODE, reason: 'feature base SHA is not the exact workflow SHA'};
  }
  if (route.base_ref === 'main' && !promotion) {
    return {...result, decision: 'FAIL_CLOSED', code: PROMOTION_TOPOLOGY_CODE, reason: 'main promotion is not the exact same-repository dev workflow authority'};
  }
  if (result.decision === 'FAIL_CLOSED' && (result.kernelPaths || []).length > 0 && promotion) {
    if (!/^sha256:[0-9a-f]{64}$/.test(kernelBeforeDigest || '') || !/^sha256:[0-9a-f]{64}$/.test(kernelAfterDigest || '')) {
      return {...result, decision: 'FAIL_CLOSED', code: ROOT_FAILURE_CODE, reason: 'trusted kernel propagation is missing before/after digests'};
    }
    return {...result, decision: 'PASS', code: null, reason: 'exact dev-to-main kernel propagation', kernelBeforeDigest, kernelAfterDigest};
  }
  if (result.decision === 'PASS') {
    const activation = defaultBranchDecision(defaultBranch, eventRef);
    if (activation.decision !== 'PASS') {
      return {...result, ...activation};
    }
    if (workflowRef !== expectedWorkflowRef(repository) || runtimeSha !== workflowSha || !validSHA(workflowSha) || !validSHA(runtimeSha)) {
      return {...result, decision: 'FAIL_CLOSED', code: DEFAULT_BRANCH_CODE, reason: 'runtime and default workflow identities are not exactly bound'};
    }
  }
  return result;
}

function sortedChangedFiles(files) {
  return [...(Array.isArray(files) ? files : [])].sort((left, right) => {
    const leftKey = [left.filename, left.previous_filename || '', left.status].join('\u0000');
    const rightKey = [right.filename, right.previous_filename || '', right.status].join('\u0000');
    return canonicalStringCompare(leftKey, rightKey);
  });
}

function digestGuardianArtifact(manifest) {
  const unsigned = {...manifest, bundle_sha256: ''};
  return `sha256:${crypto.createHash('sha256').update(JSON.stringify(unsigned)).digest('hex')}`;
}

function buildGuardianArtifact({pull, repository, action, defaultBranch, workflowRef, workflowSha, runtimeRef, runtimeSha, runId, runAttempt, eventRef, result, liveBefore, liveAfter, checkName, branchProtection = null, devBranchProtection = null, observerEnvironment = null, observerEnvironmentSnapshot = null, installationRepositoryScope = null}) {
  const identity = pullIdentity(pull);
  const route = routeForPull(pull);
  const topology = liveAfter && liveAfter.topology ? liveAfter.topology : liveBefore && liveBefore.topology ? liveBefore.topology : null;
  const manifest = {
    schema: GUARDIAN_SCHEMA,
    repository: repository || null,
    pull_request_number: identity.number || null,
    action: action || null,
    base_repo: identity.base_repo || null,
    base_ref: identity.base_ref || null,
    base_sha: identity.base_sha || null,
    head_repo: identity.head_repo || null,
    head_ref: identity.head_ref || null,
    head_sha: identity.head_sha || null,
    workflow_ref: workflowRef || null,
    workflow_sha: workflowSha || null,
    runtime_ref: runtimeRef || null,
    runtime_sha: runtimeSha || null,
    run_id: validPositiveInteger(runId) ? runId : null,
    run_attempt: validPositiveInteger(runAttempt) ? runAttempt : null,
    event_ref: eventRef || null,
    default_branch: defaultBranch || null,
    observer_environment: observerEnvironment || null,
    observer_environment_snapshot: observerEnvironmentSnapshot,
    observer_environment_digest: observerEnvironmentSnapshot && observerEnvironmentSnapshot.digest_sha256 ? observerEnvironmentSnapshot.digest_sha256 : null,
    installation_repository_scope: installationRepositoryScope,
    foundation_bootstrap: result && result.foundationBootstrap ? result.foundationBootstrap : null,
    head_binding_status: result && result.decision === 'PASS' ? HEAD_BINDING_VERIFIED : HEAD_BINDING_STATUS,
    route,
    check_name: checkName || checkNameForRoute(route),
    live_refs_before: liveBefore ? liveBefore.refs : null,
    live_refs_after: liveAfter ? liveAfter.refs : null,
    topology,
    branch_protection: branchProtection,
    dev_branch_protection: devBranchProtection,
    kernel_before_sha256: result && result.kernelBeforeDigest ? result.kernelBeforeDigest : null,
    kernel_after_sha256: result && result.kernelAfterDigest ? result.kernelAfterDigest : null,
    changed_files: sortedChangedFiles(result && result.files),
    changed_files_count: Array.isArray(result && result.files) ? result.files.length : 0,
    kernel_paths: [...new Set((result && result.kernelPaths) || [])].sort(canonicalStringCompare),
    decision: result && result.decision ? result.decision : 'FAIL_CLOSED',
    code: result ? result.code : ROOT_FAILURE_CODE,
    reason: result && result.reason ? result.reason : result && result.decision === 'PASS' ? 'guardian observation passed' : 'guardian observation was incomplete',
    bundle_sha256: '',
  };
  manifest.bundle_sha256 = digestGuardianArtifact(manifest);
  return manifest;
}

function validateSortedArtifactFiles(files) {
  let previousKey = null;
  for (const file of files) {
    const validated = validateChangedFile(file);
    const key = [validated.filename, validated.previous_filename || '', validated.status].join('\u0000');
    if (previousKey !== null && canonicalStringCompare(key, previousKey) <= 0) {
      throw guardianFailure('guardian artifact changed files are not sorted and unique');
    }
    previousKey = key;
  }
}

function derivedKernelPaths(files) {
  return [...new Set(files.flatMap((file) => [file.filename, file.previous_filename || null]).filter((path) => path && isProtectedKernelPath(path)))].sort(canonicalStringCompare);
}

function validateSortedKernelPaths(paths) {
  let previous = null;
  for (const path of paths) {
    if (typeof path !== 'string' || !isProtectedKernelPath(path) || (previous !== null && canonicalStringCompare(path, previous) <= 0)) {
      throw guardianFailure('guardian artifact kernel paths are not protected, sorted, and unique');
    }
    previous = path;
  }
}

function validateExpectedArtifactTuple(manifest, expected) {
  const fields = ['repository', 'pull_request_number', 'action', 'base_repo', 'base_ref', 'base_sha', 'head_repo', 'head_ref', 'head_sha', 'default_branch', 'workflow_ref', 'workflow_sha', 'runtime_ref', 'runtime_sha', 'event_ref', 'run_id', 'run_attempt'];
  if (!expected || fields.some((field) => expected[field] === undefined || expected[field] === null)) {
    throw guardianFailure('guardian artifact external expected tuple is missing');
  }
  if (!validRepository(expected.repository) || !validPositiveInteger(expected.pull_request_number) || !validRef(expected.action) || !validRepository(expected.base_repo) || !ALLOWED_BASES.has(expected.base_ref) || !validSHA(expected.base_sha) || !validRepository(expected.head_repo) || !validRef(expected.head_ref) || !validSHA(expected.head_sha) || !validRef(expected.default_branch) || !validRef(expected.workflow_ref) || !validSHA(expected.workflow_sha) || !validRef(expected.runtime_ref) || !validSHA(expected.runtime_sha) || !validRef(expected.event_ref) || !validPositiveInteger(expected.run_id) || !validPositiveInteger(expected.run_attempt)) {
    throw guardianFailure('guardian artifact external expected tuple is malformed');
  }
  for (const field of fields) {
    if (manifest[field] !== expected[field]) {
      throw guardianFailure(`guardian artifact external tuple mismatch: ${field}`);
    }
  }
}

function validateGuardianArtifact(manifest, expected, {now = new Date()} = {}) {
  if (!manifest || manifest.schema !== GUARDIAN_SCHEMA || !validRepository(manifest.repository) || !validPositiveInteger(manifest.pull_request_number) || !validRef(manifest.action) || !validRepository(manifest.base_repo) || !ALLOWED_BASES.has(manifest.base_ref) || !validSHA(manifest.base_sha) || !validRepository(manifest.head_repo) || !validRef(manifest.head_ref) || !validSHA(manifest.head_sha) || !validRef(manifest.workflow_ref) || !validSHA(manifest.workflow_sha) || !validRef(manifest.runtime_ref) || !validSHA(manifest.runtime_sha) || !validPositiveInteger(manifest.run_id) || !validPositiveInteger(manifest.run_attempt) || !validRef(manifest.event_ref) || !validRef(manifest.default_branch) || ![HEAD_BINDING_STATUS, HEAD_BINDING_VERIFIED].includes(manifest.head_binding_status) || !Array.isArray(manifest.changed_files) || !Array.isArray(manifest.kernel_paths) || !['PASS', 'FAIL_CLOSED'].includes(manifest.decision) || !/^sha256:[0-9a-f]{64}$/.test(manifest.bundle_sha256 || '') || typeof manifest.reason !== 'string' || manifest.reason.length === 0 || !['feature_dev', 'promotion_main', foundationBootstrap.FOUNDATION_ROUTE].includes(manifest.route) || !['CI guardian shadow', 'CI guardian'].includes(manifest.check_name)) {
    throw guardianFailure('guardian artifact schema or identity is malformed');
  }
  validateExpectedArtifactTuple(manifest, expected);
  if (manifest.check_name !== checkNameForRoute(manifest.route)) {
    throw guardianFailure('guardian artifact check identity does not match route', CHECK_IDENTITY_CODE);
  }
  if (manifest.route !== foundationBootstrap.FOUNDATION_ROUTE && manifest.foundation_bootstrap !== null) {
    throw guardianFailure('non-FOUNDATION guardian artifact must not carry a FOUNDATION receipt', FOUNDATION_BOOTSTRAP_CODE);
  }
  if (manifest.route === 'promotion_main') {
    if (manifest.observer_environment !== OBSERVER_ENVIRONMENT) throw guardianFailure('guardian observer environment is not the protected environment', PROTECTION_CODE);
    validateBranchProtectionSnapshot(manifest.branch_protection, {requireVerified: manifest.decision === 'PASS', expectedBranch: 'main', expectedContexts: MAIN_PROTECTION_CONTEXTS, now});
    validateBranchProtectionSnapshot(manifest.dev_branch_protection, {requireVerified: manifest.decision === 'PASS', expectedBranch: 'dev', expectedContexts: DEV_PROTECTION_CONTEXTS, now});
    validateGuardianEnvironment(manifest.observer_environment_snapshot, {requireVerified: manifest.decision === 'PASS', now});
    if (manifest.observer_environment_digest !== manifest.observer_environment_snapshot.digest_sha256) throw guardianFailure('guardian observer environment digest is not bound', PROTECTION_CODE);
    validateInstallationRepositoryScope(manifest.installation_repository_scope, {requireVerified: manifest.decision === 'PASS', expectedRepository: manifest.repository, now});
  } else if (manifest.branch_protection !== null || manifest.dev_branch_protection !== null || manifest.observer_environment_snapshot !== null || manifest.observer_environment_digest !== null || manifest.installation_repository_scope !== null) {
    throw guardianFailure('feature guardian artifact must not carry privileged observer snapshots', PROTECTION_CODE);
  }
  const snapshotsPresent = manifest.live_refs_before !== null && manifest.live_refs_before !== undefined && manifest.live_refs_after !== null && manifest.live_refs_after !== undefined;
  const snapshotsValid = validLiveSnapshot({refs: manifest.live_refs_before, topology: manifest.topology}) && validLiveSnapshot({refs: manifest.live_refs_after, topology: manifest.topology}) && manifest.live_refs_before.dev_sha === manifest.live_refs_after.dev_sha && manifest.live_refs_before.main_sha === manifest.live_refs_after.main_sha;
  if ((manifest.decision === 'PASS' && !snapshotsValid) || (snapshotsPresent && !snapshotsValid)) {
    throw guardianFailure('guardian artifact live ref snapshots are missing, malformed, or drifted', LIVE_REF_CODE);
  }
  if (manifest.base_repo !== manifest.repository) {
    throw guardianFailure('guardian artifact base repository is not the event repository');
  }
  if (!ALLOWED_ACTIONS.has(manifest.action)) {
    throw guardianFailure('guardian artifact action is not a supported pull-request trigger');
  }
  const expectedRef = `${manifest.repository}/.github/workflows/ci-guardian.yml@refs/heads/${manifest.default_branch}`;
  if (manifest.workflow_ref !== expectedRef || (manifest.default_branch === 'dev' && manifest.workflow_ref !== expectedWorkflowRef(manifest.repository))) {
    throw guardianFailure('guardian artifact workflow source is not the exact default-branch workflow');
  }
  if (manifest.runtime_sha !== manifest.workflow_sha && manifest.decision === 'PASS') {
    throw guardianFailure('guardian artifact PASS runtime and workflow SHA differ');
  }
  if (manifest.decision === 'FAIL_CLOSED' && (!GUARDIAN_FAILURE_CODES.has(manifest.code))) {
    throw guardianFailure('guardian artifact failure code is missing');
  }
  if (manifest.decision === 'PASS' && manifest.code !== null) {
    throw guardianFailure('guardian artifact PASS code must be null');
  }
  if (manifest.head_binding_status !== HEAD_BINDING_VERIFIED && manifest.decision === 'PASS') {
    throw guardianFailure('guardian PASS head binding must be verified', CHECK_IDENTITY_CODE);
  }
  if (manifest.decision === 'FAIL_CLOSED' && manifest.head_binding_status !== HEAD_BINDING_STATUS) {
    throw guardianFailure('guardian FAIL_CLOSED head binding must remain unverified', CHECK_IDENTITY_CODE);
  }
  if (!Number.isInteger(manifest.changed_files_count) || manifest.changed_files_count !== manifest.changed_files.length) {
    throw guardianFailure('guardian artifact changed-file count does not match collected files');
  }
  validateSortedArtifactFiles(manifest.changed_files);
  validateSortedKernelPaths(manifest.kernel_paths);
  const expectedKernelPaths = derivedKernelPaths(manifest.changed_files);
  if (JSON.stringify(expectedKernelPaths) !== JSON.stringify(manifest.kernel_paths)) {
    throw guardianFailure('guardian artifact kernel paths do not match changed filenames and previous_filename values');
  }
  const beforeValid = manifest.kernel_before_sha256 === null || /^sha256:[0-9a-f]{64}$/.test(manifest.kernel_before_sha256);
  const afterValid = manifest.kernel_after_sha256 === null || /^sha256:[0-9a-f]{64}$/.test(manifest.kernel_after_sha256);
  if (!beforeValid || !afterValid || (manifest.kernel_paths.length === 0 && (manifest.kernel_before_sha256 !== null || manifest.kernel_after_sha256 !== null)) || (manifest.decision === 'PASS' && manifest.kernel_paths.length > 0 && (manifest.kernel_before_sha256 === null || manifest.kernel_after_sha256 === null))) {
    throw guardianFailure('guardian artifact kernel digest fields are inconsistent');
  }
  if (manifest.decision === 'PASS') {
    if (manifest.default_branch !== 'dev' || manifest.workflow_ref !== expectedWorkflowRef(manifest.repository) || manifest.runtime_ref !== 'refs/heads/dev' || manifest.event_ref !== 'refs/heads/dev' || manifest.runtime_sha !== manifest.workflow_sha || !validSHA(manifest.workflow_sha)) {
      throw guardianFailure('guardian artifact PASS is not bound to the exact default dev identity');
    }
    const trustedPromotion = manifest.base_repo === manifest.repository && manifest.head_repo === manifest.repository && manifest.base_ref === 'main' && manifest.head_ref === 'dev' && manifest.head_sha === manifest.workflow_sha;
    const featureRoute = manifest.base_ref === 'dev' && manifest.head_ref.startsWith('agent/');
    const foundationRoute = manifest.route === foundationBootstrap.FOUNDATION_ROUTE;
    if (!featureRoute && !trustedPromotion && !foundationRoute) {
      throw guardianFailure('guardian artifact PASS route is neither an agent feature nor exact dev-to-main promotion');
    }
    if (featureRoute && manifest.base_sha !== manifest.workflow_sha) {
      throw guardianFailure('guardian artifact PASS feature base SHA is not the exact workflow SHA');
    }
    if (foundationRoute && (!foundationBootstrap.foundationArtifactIdentity(manifest) || manifest.check_name !== 'CI guardian shadow' || manifest.reason !== 'FOUNDATION_OVERRIDE_USED=1' || !manifest.foundation_bootstrap || manifest.foundation_bootstrap.decision !== 'FOUNDATION' || manifest.foundation_bootstrap.consumed !== false || JSON.stringify(manifest.foundation_bootstrap.authorization) !== JSON.stringify(foundationBootstrap.FOUNDATION_BOOTSTRAP) || JSON.stringify(manifest.foundation_bootstrap.observedKernelPaths) !== JSON.stringify(manifest.kernel_paths) || JSON.stringify(manifest.kernel_paths) !== JSON.stringify(foundationBootstrap.ALLOWED_KERNEL_PATHS))) {
      throw guardianFailure('guardian artifact FOUNDATION identity or receipt is not exact', FOUNDATION_BOOTSTRAP_CODE);
    }
    if (manifest.kernel_paths.length > 0 && ((!trustedPromotion && !foundationRoute) || (!foundationRoute && (manifest.kernel_before_sha256 === null || manifest.kernel_after_sha256 === null)))) {
      throw guardianFailure('guardian artifact PASS kernel propagation is not exact dev-to-main authority');
    }
  }
  if (manifest.decision === 'PASS' && manifest.route === 'feature_dev' && (manifest.base_ref !== 'dev' || manifest.check_name !== 'CI guardian shadow' || manifest.base_sha !== manifest.workflow_sha || manifest.base_sha !== manifest.live_refs_before.dev_sha)) {
    throw guardianFailure('guardian feature route identity is not exact', LIVE_REF_CODE);
  }
  if (manifest.decision === 'PASS' && manifest.route === foundationBootstrap.FOUNDATION_ROUTE && (manifest.base_ref !== 'dev' || manifest.check_name !== 'CI guardian shadow' || !foundationBootstrap.foundationArtifactIdentity(manifest) || manifest.live_refs_before.main_sha !== foundationBootstrap.FOUNDATION_BOOTSTRAP.sourceMainSha || manifest.live_refs_after.main_sha !== foundationBootstrap.FOUNDATION_BOOTSTRAP.sourceMainSha || manifest.live_refs_before.dev_sha !== manifest.base_sha || manifest.live_refs_after.dev_sha !== manifest.base_sha)) {
    throw guardianFailure('guardian FOUNDATION live topology is not exact', FOUNDATION_BOOTSTRAP_CODE);
  }
  if (manifest.decision === 'PASS' && manifest.route === 'promotion_main') {
    if (manifest.check_name !== 'CI guardian' || manifest.base_ref !== 'main' || manifest.head_ref !== 'dev' || manifest.live_refs_before.main_sha !== manifest.base_sha || manifest.live_refs_after.main_sha !== manifest.base_sha || manifest.live_refs_before.dev_sha !== manifest.head_sha || manifest.live_refs_after.dev_sha !== manifest.head_sha || !validPromotionTopology({refs: manifest.live_refs_before, topology: manifest.topology}) || !validPromotionTopology({refs: manifest.live_refs_after, topology: manifest.topology}) || manifest.branch_protection.read_status !== 'verified' || manifest.branch_protection.branch !== 'main' || manifest.branch_protection.base_sha !== manifest.base_sha || manifest.branch_protection.head_sha !== manifest.head_sha || manifest.branch_protection.workflow_sha !== manifest.workflow_sha || manifest.branch_protection.run_id !== manifest.run_id || manifest.branch_protection.run_attempt !== manifest.run_attempt || manifest.dev_branch_protection.read_status !== 'verified' || manifest.dev_branch_protection.branch !== 'dev' || manifest.dev_branch_protection.base_sha !== manifest.base_sha || manifest.dev_branch_protection.head_sha !== manifest.head_sha || manifest.dev_branch_protection.workflow_sha !== manifest.workflow_sha || manifest.dev_branch_protection.run_id !== manifest.run_id || manifest.dev_branch_protection.run_attempt !== manifest.run_attempt || manifest.observer_environment_snapshot.read_status !== 'verified' || manifest.observer_environment_snapshot.run_id !== manifest.run_id || manifest.observer_environment_snapshot.run_attempt !== manifest.run_attempt || manifest.observer_environment_snapshot.workflow_sha !== manifest.workflow_sha) {
      throw guardianFailure('guardian promotion topology evidence is not exact', PROMOTION_TOPOLOGY_CODE);
    }
  }
  if (manifest.bundle_sha256 !== digestGuardianArtifact(manifest)) {
    throw guardianFailure('guardian artifact digest does not match canonical content');
  }
  return manifest;
}

module.exports = {
  ALLOWED_BASES,
  DEFAULT_BRANCH_CODE,
  OBSERVER_FRESHNESS_WINDOW_MS,
  LIVE_REF_CODE,
  PROMOTION_TOPOLOGY_CODE,
  CHECK_IDENTITY_CODE,
  PROTECTION_CODE,
  DEV_PROTECTION_CONTEXTS,
  MAIN_PROTECTION_CONTEXTS,
  protectionContextsForBranch,
  OBSERVER_ENVIRONMENT,
  INSTALLATION_SCOPE_CODE,
  INSTALLATION_SCOPE_REPOSITORY,
  GUARDIAN_SCHEMA,
  HEAD_BINDING_STATUS,
  HEAD_BINDING_VERIFIED,
  PROTECTED_FILES,
  PROTECTED_PREFIXES,
  ROOT_FAILURE_CODE,
  FOUNDATION_BOOTSTRAP_CODE,
  FOUNDATION_BOOTSTRAP: foundationBootstrap.FOUNDATION_BOOTSTRAP,
  FOUNDATION_ROUTE: foundationBootstrap.FOUNDATION_ROUTE,
  readLiveTopology,
  digestBranchProtection,
  digestGuardianEnvironment,
  observerFreshnessFromResponse,
  validObserverFreshness,
  emptyGuardianEnvironment,
  observeBranchProtection,
  observeGuardianEnvironment,
  digestInstallationRepositoryScope,
  emptyInstallationRepositoryScope,
  observeInstallationRepositoryScope,
  validateBranchProtectionSnapshot,
  validatePublicBranchSummary,
  validateGuardianEnvironment,
  validateInstallationRepositoryScope,
  validatePublicBranchSummary,
  routeForPull,
  checkNameForRoute,
  guardianFailure,
  buildGuardianArtifact,
  classifyGuardianDecision,
  defaultBranchDecision,
  digestGuardianArtifact,
  inspectChangedFiles,
  isProtectedKernelPath,
  kernelTreeDigest,
  normalizePath,
  revalidatePullRequest,
  validateGuardianArtifact,
  validateGuardianPullRequest,
  validatePromotionPullRequestState,
  trustedDevPromotion,
  foundationBootstrapDecision: foundationBootstrap.foundationBootstrapDecision,
};
