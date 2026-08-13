'use strict';

const ROOT_FAILURE_CODE = 'CI-ROOT-OF-TRUST-001';
const ALLOWED_BASES = new Set(['integration', 'dev', 'main']);
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
  const path = normalizePath(value);
  return PROTECTED_FILES.has(path) || PROTECTED_PREFIXES.some((prefix) => path.startsWith(prefix));
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
  if (!pull.head || typeof pull.head.ref !== 'string' || pull.head.ref.length === 0 || !validSHA(pull.head.sha)) {
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

async function inspectChangedFiles({listFiles, owner, repo, baseRepoFullName, pullNumber}) {
  if (typeof listFiles !== 'function' || typeof owner !== 'string' || typeof repo !== 'string' || baseRepoFullName !== `${owner}/${repo}` || !Number.isInteger(pullNumber) || pullNumber < 1) {
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

module.exports = {
  ALLOWED_BASES,
  PROTECTED_FILES,
  PROTECTED_PREFIXES,
  ROOT_FAILURE_CODE,
  guardianFailure,
  inspectChangedFiles,
  isProtectedKernelPath,
  normalizePath,
  validateGuardianPullRequest,
};
