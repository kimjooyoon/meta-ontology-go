'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {createRequire} = require('node:module');
const {test} = require('node:test');

const repository = process.env.GOOO_FAILURE_TEST_REPO || path.resolve(__dirname, '../..');
const workflow = process.env.GOOO_FAILURE_TEST_WORKFLOW || path.join(repository, '.github/workflows/ci.yml');
const source = fs.readFileSync(workflow, 'utf8');
const marker = '      - name: Capture exact failure tuple\n';
assert.equal(source.split(marker).length, 2, 'one real failure capture step');
const step = source.split(marker)[1].split('      - name: Generate versioned failure manifest and summary\n')[0];
assert.equal(step.split('          script: |\n').length, 2, 'one script body');
const script = step.split('          script: |\n')[1].split('\n').map(line => {
  assert.ok(line === '' || line.startsWith('            '), 'capture script indentation');
  return line.slice(12);
}).join('\n');
const capture = new (Object.getPrototypeOf(async function () {}).constructor)(
  'github', 'context', 'core', 'require', 'process', script);
const localRequire = createRequire(path.join(repository, 'failure-capture-test.cjs'));
const head = 'a'.repeat(40);
const base = 'b'.repeat(40);
const workflowSHA = 'c'.repeat(40);
const run = 9;
const attempt = 2;
const canonical = ['gofmt', 'go vet', 'go test', 'go test -race', 'Semantic conformance', 'CI policy'];
const context = {repo: {owner: 'owner', repo: 'repo'}, eventName: 'pull_request',
  ref: 'refs/pull/7/merge', sha: head, runId: run, payload: {pull_request: {
    number: 7, base: {ref: 'dev', sha: base}, head: {ref: 'agent/ci-workflow', sha: head},
  }}};
function failedJob(id, name, stepName = 'failed step') {
  return {id, name, status: 'completed', conclusion: 'failure', head_sha: head,
    run_id: run, run_attempt: attempt,
    steps: [{name: stepName, status: 'completed', conclusion: 'failure'}]};
}
function currentProof() {
  return {schema: 'gooo/ci-proof/v5', repository: 'owner/repo', event: 'pull_request',
    base_ref: 'dev', base_sha: base, head_sha: head, ref: context.ref, event_ref: context.ref,
    checkout_ref: head, pr_number: 7, run_id: run, run_attempt: attempt, workflow_sha: workflowSHA,
    jobs: canonical.map((name, i) => failedJob(100 + i, name)),
    artifacts: [{id: 201, name: 'ci-evidence-9-2', run_id: run, run_attempt: attempt,
      size_bytes: 1, expired: false, digest: `sha256:${'d'.repeat(64)}`}],
    rejections: ['scope_not_passed']};
}
async function observe(jobs, proof, proofRows) {
  const outputs = new Map();
  const fileAPI = {existsSync: () => proof !== undefined,
    readFileSync: () => typeof proof === 'string' ? proof : JSON.stringify(proof),
    writeFileSync: (name, value) => outputs.set(name, value)};
  const artifacts = proofRows || (proof === undefined ? [] : [{id: 202, name: 'ci-proof-9-2',
    size_in_bytes: 1, expired: false, digest: `sha256:${'e'.repeat(64)}`}]);
  const github = {rest: {actions: {
    listJobsForWorkflowRun: async () => ({data: {jobs}}),
    listWorkflowRunArtifacts: async () => ({data: {artifacts}}),
  }}};
  const allowed = new Set(['./scripts/ci-proof/dispatch', './scripts/ci-proof/refs', './scripts/ci-proof/artifacts']);
  await capture(github, context, {warning: () => {}}, name => {
    if (name === 'fs') return fileAPI;
    assert.ok(allowed.has(name), `unexpected module ${name}`);
    return localRequire(name);
  }, {env: {CI_RUN_ATTEMPT: String(attempt), CI_WORKFLOW_SHA: workflowSHA}});
  return {outputs, input: outputs.get('ci-failure-input.json') ? JSON.parse(outputs.get('ci-failure-input.json')) : null};
}

test('semantic, evidence, proof chain retains each cause and the missing proof', async () => {
  const jobs = [failedJob(11, 'Semantic conformance'),
    failedJob(12, 'CI evidence', 'Verify and write CI evidence bundle'),
    failedJob(13, 'CI proof bundle', 'Resolve canonical evidence lineage')];
  const {input} = await observe(jobs);
  assert.deepEqual(input.terminal_failure_codes, ['CI-TEST-001', 'CI-ARTIFACT-001', 'CI-ARTIFACT-001']);
  assert.deepEqual(input.failure_codes, ['CI-ARTIFACT-001', 'CI-TEST-001']);
  assert.equal(input.job.id, 11);
  assert.equal(input.artifact_status, 'missing');
  assert.equal(input.artifact_reason, 'artifact_missing');
  assert.deepEqual(input.rejections, ['proof_artifact_missing']);
  assert.equal(input.proof_artifact, null);
  assert.equal(input.terminal_failure_evidence[2].reason, 'proof_state:artifact_missing');
  const fixture = JSON.parse(fs.readFileSync(path.join(__dirname, 'testdata/failure-summary-chain.json'), 'utf8'));
  assert.deepEqual(input, fixture, 'Go consumer uses this exact producer output');
});

test('proof classification remains source-bound when primary failure order changes', async () => {
  for (const proofFirst of [false, true]) {
    const jobs = [failedJob(proofFirst ? 12 : 11, 'go test'), failedJob(proofFirst ? 11 : 12, 'CI proof bundle')];
    const {input} = await observe(jobs, currentProof());
    const proofIndex = input.terminal_failures.findIndex(job => job.name === 'CI proof bundle');
    assert.equal(input.terminal_failure_codes[proofIndex], 'CI-SCOPE-001');
    assert.equal(input.artifact_status, 'verified');
    assert.equal(input.terminal_failure_evidence[proofIndex].reason, 'proof_state:verified');
    assert.deepEqual(input.failure_codes, ['CI-SCOPE-001', 'CI-TEST-001']);
  }
});

test('older schema and stale source remain freshness failures', async () => {
  for (const mutate of [p => { p.schema = 'gooo/ci-proof/v4'; }, p => { p.head_sha = 'f'.repeat(40); }]) {
    const proof = currentProof();
    mutate(proof);
    const {input} = await observe([failedJob(11, 'go test'), failedJob(12, 'CI proof bundle')], proof);
    assert.equal(input.terminal_failure_codes[1], 'CI-FRESHNESS-001');
    assert.equal(input.artifact_status, 'not_applicable');
    assert.deepEqual(input.rejections, ['proof_artifact_tuple_mismatch']);
    assert.equal(input.artifacts.length, 0);
  }
});

test('invalid proof remains an artifact failure after an earlier semantic failure', async () => {
  const {input} = await observe([failedJob(11, 'Semantic conformance'), failedJob(12, 'CI proof bundle')], '{');
  assert.equal(input.terminal_failure_codes[1], 'CI-ARTIFACT-001');
  assert.equal(input.artifact_status, 'missing');
  assert.deepEqual(input.rejections, ['proof_artifact_invalid']);
});

test('known step mappings and future jobs retain distinct classifications', async () => {
  const rows = [
    ['CI evidence', 'Reproduce generated evidence input', 'CI-TEST-001'],
    ['CI evidence', 'Capture CLI domain evidence', 'CI-TEST-001'],
    ['CI evidence', 'Verify and write CI evidence bundle', 'CI-ARTIFACT-001'],
    ['CI policy', 'Validate source metric and meta evidence (caps are observations)', 'CI-CONTRACT-001'],
    ['CI policy', 'Validate source metric and meta evidence on protected push', 'CI-CONTRACT-001'],
    ['CI policy', 'Check source scope', 'CI-SCOPE-001'],
    ['future check', 'future step', 'CI-UNCLASSIFIED-001'],
  ];
  for (const [name, stepName, code] of rows) {
    const {input} = await observe([failedJob(11, name, stepName)], currentProof());
    assert.deepEqual(input.terminal_failure_codes, [code]);
    assert.deepEqual(input.failure_codes, [code]);
    assert.equal(input.terminal_failure_evidence[0].classification,
      name === 'future check' ? 'unclassified_job' : 'known_job');
  }
});

test('stale terminal job tuple stops capture before emitting a manifest', async () => {
  const job = failedJob(11, 'go test');
  job.run_attempt = 1;
  await assert.rejects(observe([job]), /exact head\/run binding/);
});

test('six successful jobs still produce only a no-failure closure', async () => {
  const jobs = canonical.map((name, i) => ({...failedJob(100 + i, name), conclusion: 'success'}));
  const {input, outputs} = await observe(jobs);
  assert.equal(input, null);
  assert.deepEqual(JSON.parse(outputs.get('ci-closure-input.json')).terminal_failures, []);
});

function failureSummaryAdmission() {
  const job = source.split('  failure-report:\n')[1].split('    runs-on:')[0];
  const dependencies = job.match(/needs: \[([^\]]+)\]/)[1].split(',').map(value => value.trim());
  const expression = job.match(/\$\{\{([\s\S]*?)\}\}/)[1];
  const evaluate = new Function('needs', 'github', 'cancelled', `return (${expression});`);
  return {dependencies, evaluate};
}

test('failure-summary admission uses the positive proof on successful runs', () => {
  const {dependencies, evaluate} = failureSummaryAdmission();
  assert.deepEqual(dependencies, ['format', 'vet', 'test', 'race', 'semantic', 'policy', 'proof']);
  const needs = Object.fromEntries(dependencies.map(name => [name, {result: 'success'}]));
  for (const github of [
    {event_name: 'pull_request', ref: 'refs/pull/1581/merge'},
    {event_name: 'workflow_dispatch', ref: 'refs/heads/agent/current-source'},
    {event_name: 'push', ref: 'refs/heads/dev'},
    {event_name: 'push', ref: 'refs/heads/main'},
  ]) {
    assert.equal(evaluate(needs, github, () => false), false);
  }
});

test('failure-summary admission retains every non-successful dependency', () => {
  const {dependencies, evaluate} = failureSummaryAdmission();
  const github = {event_name: 'pull_request', ref: 'refs/pull/1581/merge'};
  for (const name of dependencies) {
    for (const result of ['failure', 'skipped', 'cancelled', '']) {
      const needs = Object.fromEntries(dependencies.map(key => [key, {result: 'success'}]));
      needs[name].result = result;
      assert.equal(evaluate(needs, github, () => false), true, `${name}: ${result}`);
      assert.equal(evaluate(needs, github, () => true), false, 'workflow cancellation retains its existing guard');
      assert.equal(evaluate(needs, {event_name: 'push', ref: 'refs/heads/agent/current-source'}, () => false), false);
    }
  }
});
