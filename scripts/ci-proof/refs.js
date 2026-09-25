'use strict';

const ACTIVE_BASE_BRANCHES = new Set(['dev', 'main']);

function normalizeBranchName(value, label) {
  if (typeof value !== 'string' || value.length === 0 || value.length > 255 ||
      value.startsWith('/') || value.endsWith('/') || value.startsWith('.') ||
      value.endsWith('.') || value.includes('..') || value.includes('//') ||
      value.includes('@{') || /[\u0000-\u0020~^:?*\\[\\]\\\\]/.test(value)) {
    throw new Error(`${label || 'branch'} is missing or malformed`);
  }
  return value;
}

function normalizeProtectedBranchRef(ref) {
  if (typeof ref !== 'string' || !ref.startsWith('refs/heads/')) {
    throw new Error('protected push ref must be refs/heads/<branch>');
  }
  return normalizeActiveBaseBranch(ref.slice('refs/heads/'.length), 'protected push branch');
}

function normalizeActiveBaseBranch(value, label) {
  const branch = normalizeBranchName(value, label);
  if (!ACTIVE_BASE_BRANCHES.has(branch)) {
    throw new Error(`${label || 'base branch'} is retired or unsupported: ${branch}`);
  }
  return branch;
}

function normalizeBaseRef(event, ref, pullRequest) {
  if (event === 'pull_request') {
    return normalizeActiveBaseBranch(pullRequest && pullRequest.base && pullRequest.base.ref, 'pull request base branch');
  }
  if (event === 'push') {
    return normalizeProtectedBranchRef(ref);
  }
  throw new Error(`unsupported event for base branch normalization: ${event || '<missing>'}`);
}

function normalizeOwnerBranch(event, ref, pullRequest) {
  if (event === 'pull_request') {
    return normalizeBranchName(pullRequest && pullRequest.head && pullRequest.head.ref, 'pull request owner branch');
  }
  if (event === 'push') {
    return normalizeProtectedBranchRef(ref);
  }
  throw new Error(`unsupported event for owner branch normalization: ${event || '<missing>'}`);
}

module.exports = {normalizeBaseRef, normalizeOwnerBranch, normalizeProtectedBranchRef};
