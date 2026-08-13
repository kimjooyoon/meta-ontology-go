'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {
  PROTECTED_FILES,
  PROTECTED_PREFIXES,
  ROOT_FAILURE_CODE,
  inspectChangedFiles,
  validateGuardianPullRequest,
} = require('./guardian');

const sha = (letter) => letter.repeat(40);
const file = (filename, status = 'modified', previous_filename) => ({filename, status, ...(previous_filename ? {previous_filename} : {})});
const pull = (base = 'integration', fork = false) => ({
  number: 108,
  base: {ref: base, sha: sha('b'), repo: {full_name: 'owner/repo'}},
  head: {ref: 'agent/ci-workflow', sha: sha('a'), repo: {full_name: fork ? 'fork/repo' : 'owner/repo'}},
});

async function rejectsRoot(operation) {
  await assert.rejects(operation, (error) => error && error.code === ROOT_FAILURE_CODE);
}

async function testPaginationAndNonKernelPass() {
  const pages = [Array.from({length: 100}, (_, index) => file(index === 0 ? 'docs/readme.md' : `docs/page-${index}.md`, index === 0 ? 'added' : 'modified')), [file('docs/final.md', 'modified')]];
  const calls = [];
  const result = await inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108,
    listFiles: async (options) => {
      calls.push(options);
      return {status: 200, data: pages[options.page - 1] || []};
    },
  });
  assert.equal(result.decision, 'PASS');
  assert.equal(result.files.length, 101);
  assert.deepEqual(calls.map((call) => call.page), [1, 2]);
  assert.equal(calls[1].per_page, 100);
}

async function testKernelStatusesAndRenames() {
  for (const status of ['added', 'modified', 'removed']) {
    const result = await inspectChangedFiles({
      owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108,
      listFiles: async () => ({status: 200, data: [file('.github/workflows/ci.yml', status)]}),
    });
    assert.equal(result.code, ROOT_FAILURE_CODE);
    assert.equal(result.decision, 'FAIL_CLOSED');
  }
  const renamed = await inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108,
    listFiles: async () => ({status: 200, data: [file('docs/new.md', 'renamed', '.github/ci-governance.json')]}),
  });
  assert.equal(renamed.code, ROOT_FAILURE_CODE);
  const inert = await inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108,
    listFiles: async () => ({status: 200, data: [{...file('.github/workflows/ci.yml'), patch: '# name: CI guardian'}]}),
  });
  assert.equal(inert.code, ROOT_FAILURE_CODE);
}

async function testForkAndMalformedAPI() {
  validateGuardianPullRequest(pull('dev', true));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, listFiles: async () => ({status: 200, data: [{status: 'modified'}]})}));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, listFiles: async () => ({status: 200, data: null})}));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, listFiles: async () => undefined}));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, listFiles: async () => { throw new Error('forbidden'); }}));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108, listFiles: async () => ({status: 200, data: [file('.github/workflows/ci.yml', 'renamed')]})}));
  await rejectsRoot(() => inspectChangedFiles({owner: 'owner', repo: 'repo', baseRepoFullName: 'other/repo', pullNumber: 108, listFiles: async () => ({status: 200, data: []})}));
  for (const malformed of [pull('release'), {...pull(), base: {...pull().base, sha: 'short'}}, {...pull(), head: null}]) {
    assert.throws(() => validateGuardianPullRequest(malformed), (error) => error.code === ROOT_FAILURE_CODE);
  }
}

async function testPaginationLimit() {
  let calls = 0;
  await rejectsRoot(() => inspectChangedFiles({
    owner: 'owner', repo: 'repo', baseRepoFullName: 'owner/repo', pullNumber: 108,
    listFiles: async () => {
      calls += 1;
      return {status: 200, data: Array.from({length: 100}, (_, index) => file(`docs/page-${calls}-${index}.md`))};
    },
  }));
  assert.equal(calls, 1000);
}

function testWorkflowIsReadOnlyAndBasePinned() {
  const workflow = fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'workflows', 'ci-guardian.yml'), 'utf8');
  assert.match(workflow, /^name: CI guardian/m);
  assert.match(workflow, /pull_request_target:/);
  assert.match(workflow, /- integration\n      - dev\n      - main/);
  assert.match(workflow, /name: CI guardian/);
  assert.match(workflow, /ref: \$\{\{ github\.event\.pull_request\.base\.sha \}\}/);
  assert.match(workflow, /persist-credentials: false/);
  assert.match(workflow, /actions\/checkout@11d5960a326750d5838078e36cf38b85af677262/);
  assert.match(workflow, /actions\/github-script@f28e40c7f34bde8b3046d885e986cb6290c5673b/);
  assert.match(workflow, /listFiles/);
  assert.doesNotMatch(workflow, /github\.event\.pull_request\.head\.sha/);
  assert.doesNotMatch(workflow, /refs\/pull\//);
  assert.doesNotMatch(workflow, /secrets\./);
  assert.doesNotMatch(workflow, /contents: write|pull-requests: write/);
  assert.doesNotMatch(workflow, /^\s+run:/m);
  assert.doesNotMatch(workflow, /^\s+pull_request:/m);
  assert.doesNotMatch(workflow, /agent\/ci-workflow/);
}

function testKernelSetIsMonotonic() {
  for (const path of ['.github/ci-governance.json', '.github/agent-scope-table.md', '.github/branch-policy.md', '.github/conformance-plan.md', 'go.mod', 'go.sum']) {
    assert(PROTECTED_FILES.has(path), `missing protected file ${path}`);
  }
  for (const prefix of ['.github/workflows/', 'scripts/ci-proof/', 'scripts/ci-evidence/', 'scripts/verify/', 'internal/verify/']) {
    assert(PROTECTED_PREFIXES.includes(prefix), `missing protected prefix ${prefix}`);
  }
}

(async () => {
  await testPaginationAndNonKernelPass();
  await testKernelStatusesAndRenames();
  await testForkAndMalformedAPI();
  await testPaginationLimit();
  testWorkflowIsReadOnlyAndBasePinned();
  testKernelSetIsMonotonic();
  console.log('guardian tests passed');
})().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
