'use strict';

const assert = require('assert');
const {normalizeBaseRef, normalizeOwnerBranch, normalizeProtectedBranchRef} = require('./refs');

assert.strictEqual(normalizeProtectedBranchRef('refs/heads/dev'), 'dev');
assert.strictEqual(normalizeProtectedBranchRef('refs/heads/main'), 'main');
assert.throws(() => normalizeProtectedBranchRef('refs/heads/integration'), /retired|unsupported/);
assert.strictEqual(normalizeBaseRef('push', 'refs/heads/dev'), 'dev');
assert.strictEqual(normalizeBaseRef('pull_request', 'refs/pull/106/merge', {base: {ref: 'dev'}}), 'dev');
assert.throws(() => normalizeBaseRef('pull_request', 'refs/pull/106/merge', {base: {ref: 'integration'}}), /retired|unsupported/);
assert.strictEqual(normalizeOwnerBranch('pull_request', 'refs/pull/106/merge', {head: {ref: 'agent/ci-workflow'}}), 'agent/ci-workflow');

for (const value of [null, '', 'integration', 'refs/tags/integration', 'refs/pull/106/merge', 'refs/heads/', 'refs/heads/../main', 'refs/heads/main//dev']) {
  assert.throws(() => normalizeProtectedBranchRef(value), /malformed|refs\/heads/);
}
assert.throws(() => normalizeBaseRef('pull_request', 'refs/pull/106/merge', null), /malformed/);
assert.throws(() => normalizeBaseRef('workflow_dispatch', 'refs/heads/main'), /unsupported event/);
console.log('ref normalization tests passed');
