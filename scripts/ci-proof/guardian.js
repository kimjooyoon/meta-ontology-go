'use strict';

const crypto = require('node:crypto');

const ROOT_FAILURE_CODE = 'CI-ROOT-OF-TRUST-001';
const HEAD_BINDING_STATUS = 'CI-GUARDIAN-HEAD-BINDING-UNVERIFIED';
const DEFAULT_BRANCH_CODE = 'CI-GUARDIAN-DEFAULT-BRANCH-001';
const GUARDIAN_SCHEMA = 'gooo/ci-guardian/v1';
const GUARDIAN_FAILURE_CODES = new Set([ROOT_FAILURE_CODE, DEFAULT_BRANCH_CODE]);
const ALLOWED_BASES = new Set(['dev', 'main']);
const ALLOWED_ACTIONS = new Set(['opened', 'synchronize', 'reopened', 'ready_for_review']);
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

function guardianFailure(reason) {
  const error = new Error(`${ROOT_FAILURE_CODE}: ${reason}`);
  error.code = ROOT_FAILURE_CODE;
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
        return {decision: 'FAIL_CLOSED', code: ROOT_FAILURE_CODE, reason: `protected kernel path changed: ${kernelPaths.sort().join(', ')}`, files, kernelPaths: [...new Set(kernelPaths)].sort()};
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
  entries.sort((left, right) => [left.path, left.type, left.sha].join('\u0000').localeCompare([right.path, right.type, right.sha].join('\u0000')));
  return `sha256:${crypto.createHash('sha256').update(JSON.stringify(entries)).digest('hex')}`;
}

function defaultBranchDecision(defaultBranch, eventRef) {
  if (defaultBranch !== 'dev' || eventRef !== `refs/heads/${defaultBranch}`) {
    return {decision: 'FAIL_CLOSED', code: DEFAULT_BRANCH_CODE, reason: 'guardian is not executing from the protected dev default branch'};
  }
  return {decision: 'PASS', code: null, reason: null};
}

function trustedDevPromotion({pull, repository, defaultBranch, workflowRef, workflowSha, runtimeSha}) {
  const identity = pullIdentity(pull);
  return defaultBranch === 'dev' && workflowRef === expectedWorkflowRef(repository) && runtimeSha === workflowSha && identity.base_repo === repository && identity.head_repo === repository && identity.base_ref === 'main' && identity.head_ref === 'dev' && identity.head_sha === workflowSha && validSHA(workflowSha);
}

function classifyGuardianDecision({pull, repository, defaultBranch, workflowRef, eventRef, workflowSha, runtimeSha, result, kernelBeforeDigest, kernelAfterDigest}) {
  const route = pullIdentity(pull);
  const promotion = trustedDevPromotion({pull, repository, defaultBranch, workflowRef, workflowSha, runtimeSha});
  const featureRoute = route.base_ref === 'dev' && route.head_ref && route.head_ref.startsWith('agent/');
  if (result.decision === 'PASS' && featureRoute && route.base_sha !== workflowSha) {
    return {...result, decision: 'FAIL_CLOSED', code: ROOT_FAILURE_CODE, reason: 'feature base SHA is not the exact workflow SHA'};
  }
  if (route.base_ref === 'main' && !promotion) {
    return {...result, decision: 'FAIL_CLOSED', code: ROOT_FAILURE_CODE, reason: 'main promotion is not the exact same-repository dev workflow authority'};
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
      return {...result, decision: 'FAIL_CLOSED', code: ROOT_FAILURE_CODE, reason: 'runtime and default workflow identities are not exactly bound'};
    }
  }
  return result;
}

function sortedChangedFiles(files) {
  return [...(Array.isArray(files) ? files : [])].sort((left, right) => {
    const leftKey = [left.filename, left.previous_filename || '', left.status].join('\u0000');
    const rightKey = [right.filename, right.previous_filename || '', right.status].join('\u0000');
    return leftKey.localeCompare(rightKey);
  });
}

function digestGuardianArtifact(manifest) {
  const unsigned = {...manifest, bundle_sha256: ''};
  return `sha256:${crypto.createHash('sha256').update(JSON.stringify(unsigned)).digest('hex')}`;
}

function buildGuardianArtifact({pull, repository, action, defaultBranch, workflowRef, workflowSha, runtimeRef, runtimeSha, runId, runAttempt, eventRef, result}) {
  const identity = pullIdentity(pull);
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
    head_binding_status: HEAD_BINDING_STATUS,
    kernel_before_sha256: result && result.kernelBeforeDigest ? result.kernelBeforeDigest : null,
    kernel_after_sha256: result && result.kernelAfterDigest ? result.kernelAfterDigest : null,
    changed_files: sortedChangedFiles(result && result.files),
    changed_files_count: Array.isArray(result && result.files) ? result.files.length : 0,
    kernel_paths: [...new Set((result && result.kernelPaths) || [])].sort(),
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
    if (previousKey !== null && key <= previousKey) {
      throw guardianFailure('guardian artifact changed files are not sorted and unique');
    }
    previousKey = key;
  }
}

function derivedKernelPaths(files) {
  return [...new Set(files.flatMap((file) => [file.filename, file.previous_filename || null]).filter((path) => path && isProtectedKernelPath(path)))].sort();
}

function validateSortedKernelPaths(paths) {
  let previous = null;
  for (const path of paths) {
    if (typeof path !== 'string' || !isProtectedKernelPath(path) || (previous !== null && path <= previous)) {
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

function validateGuardianArtifact(manifest, expected) {
  if (!manifest || manifest.schema !== GUARDIAN_SCHEMA || !validRepository(manifest.repository) || !validPositiveInteger(manifest.pull_request_number) || !validRef(manifest.action) || !validRepository(manifest.base_repo) || !ALLOWED_BASES.has(manifest.base_ref) || !validSHA(manifest.base_sha) || !validRepository(manifest.head_repo) || !validRef(manifest.head_ref) || !validSHA(manifest.head_sha) || !validRef(manifest.workflow_ref) || !validSHA(manifest.workflow_sha) || !validRef(manifest.runtime_ref) || !validSHA(manifest.runtime_sha) || !validPositiveInteger(manifest.run_id) || !validPositiveInteger(manifest.run_attempt) || !validRef(manifest.event_ref) || !validRef(manifest.default_branch) || manifest.head_binding_status !== HEAD_BINDING_STATUS || !Array.isArray(manifest.changed_files) || !Array.isArray(manifest.kernel_paths) || !['PASS', 'FAIL_CLOSED'].includes(manifest.decision) || !/^sha256:[0-9a-f]{64}$/.test(manifest.bundle_sha256 || '') || typeof manifest.reason !== 'string' || manifest.reason.length === 0) {
    throw guardianFailure('guardian artifact schema or identity is malformed');
  }
  validateExpectedArtifactTuple(manifest, expected);
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
    if (!featureRoute && !trustedPromotion) {
      throw guardianFailure('guardian artifact PASS route is neither an agent feature nor exact dev-to-main promotion');
    }
    if (featureRoute && manifest.base_sha !== manifest.workflow_sha) {
      throw guardianFailure('guardian artifact PASS feature base SHA is not the exact workflow SHA');
    }
    if (manifest.kernel_paths.length > 0 && (!trustedPromotion || manifest.kernel_before_sha256 === null || manifest.kernel_after_sha256 === null)) {
      throw guardianFailure('guardian artifact PASS kernel propagation is not exact dev-to-main authority');
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
  GUARDIAN_SCHEMA,
  HEAD_BINDING_STATUS,
  PROTECTED_FILES,
  PROTECTED_PREFIXES,
  ROOT_FAILURE_CODE,
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
  trustedDevPromotion,
};
