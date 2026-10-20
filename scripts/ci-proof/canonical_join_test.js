'use strict';

const assert = require('assert');
const fs = require('fs');
const {joinCanonicalChecks} = require('./canonical_join');

const head = 'a'.repeat(40);
const env = {GOOO_CI_STARTED_AT: '2026-10-15T00:00:00Z', GOOO_CI_BUDGET_MS: '60000',
  GITHUB_RUN_ATTEMPT: '1', GOOO_CI_SOURCE_HEAD: head};
const context = {repo: {owner: 'owner', repo: 'repo'}, runId: 17};
const names = ['gofmt', 'go vet', 'go test', 'go test -race', 'Semantic conformance'];
const jobs = names.map((name, i) => ({name, id: i + 1, head_sha: head, status: 'completed', conclusion: 'success'}));
const originTime = Date.parse(env.GOOO_CI_STARTED_AT);
const options = {now: () => originTime, sleep: async () => { throw new Error('unexpected wait'); }};
const api = rows => ({request: async (route, query) => {
  assert(route.includes('/attempts/{attempt_number}/jobs'));
  assert.strictEqual(query.run_id, 17);
  assert.strictEqual(query.attempt_number, 1);
  return {data: {jobs: rows}};
}});

async function main() {
  const receipt = await joinCanonicalChecks(api(jobs), context, env, options);
  assert.strictEqual(receipt.results, 'format=success,vet=success,test=success,race=success,semantic=success');
  assert.strictEqual(receipt.jobs.length, 5);
  for (const [key, value] of [['head_sha', 'b'.repeat(40)], ['conclusion', 'failure']]) {
    const rows = structuredClone(jobs); rows[0][key] = value;
    await assert.rejects(joinCanonicalChecks(api(rows), context, env, options));
  }
  await assert.rejects(joinCanonicalChecks(api([...jobs, jobs[0]]), context, env, options), /duplicate/);
  await assert.rejects(joinCanonicalChecks(api(jobs), context, {...env, GOOO_CI_STARTED_AT: ''}, options), /original clock/);
  let observed = 0;
  let waited = 0;
  const delayed = {request: async () => ({data: {jobs: ++observed === 1 ? jobs.slice(1) : jobs}})};
  await joinCanonicalChecks(delayed, context, env, {now: () => originTime + waited,
    sleep: async ms => { waited += ms; }});
  assert.strictEqual(observed, 2);
  assert.strictEqual(waited, 15000);
  let transientCalls = 0;
  let transientWaited = 0;
  const transient = {request: async () => {
    if (++transientCalls < 3) throw Object.assign(new Error('GitHub unavailable'), {status: 502});
    return {data: {jobs}};
  }};
  const recovered = await joinCanonicalChecks(transient, context, env,
    {now: () => originTime + transientWaited, sleep: async ms => { transientWaited += ms; }});
  assert.strictEqual(recovered.jobs.length, 5);
  assert.strictEqual(transientCalls, 3);
  assert.strictEqual(transientWaited, 30000);
  let permanentCalls = 0;
  const forbidden = {request: async () => {
    permanentCalls++;
    throw Object.assign(new Error('forbidden'), {status: 403});
  }};
  await assert.rejects(joinCanonicalChecks(forbidden, context, env, options), /forbidden/);
  assert.strictEqual(permanentCalls, 1);
  let unavailableWaited = 0;
  const unavailable = {request: async () => {
    throw Object.assign(new Error('GitHub unavailable'), {status: 503});
  }};
  await assert.rejects(joinCanonicalChecks(unavailable, context,
    {...env, GOOO_CI_BUDGET_MS: '30000'},
    {now: () => originTime + unavailableWaited, sleep: async ms => { unavailableWaited += ms; }}),
  /HTTP 503.*shared work budget exhausted/);
  assert.strictEqual(unavailableWaited, 30000);
  await assert.rejects(joinCanonicalChecks(api([]), context, env,
    {now: () => originTime + 60000}), /shared work budget exhausted/);

  const source = fs.readFileSync('.github/workflows/ci.yml', 'utf8');
  const policy = source.split('\n  policy:\n')[1].split('\n  evidence:\n')[0];
  assert(policy.includes('needs: [work-domain]'));
  assert(policy.indexOf('Execute declared meta-operation plan') < policy.indexOf('Join original canonical checks'));
  assert(policy.indexOf('Join original canonical checks') < policy.indexOf('Verify exact PR route'));
  assert(!policy.includes('needs.format.result'));
  const proof = source.split('\n  proof:\n')[1].split('\n  failure-report:\n')[0];
  assert(proof.includes('policy, evidence'));
  assert(proof.includes('node scripts/ci-proof/tool_origin.js verify'));
  assert(proof.includes('ci-execution-tools/gooo" generate examples/billing/main.gooo'));
  assert(!proof.includes('go run ./cmd/gooo') && !proof.includes('go run ./scripts/ci-proof'));
  for (const job of ['format', 'vet', 'test', 'race', 'semantic']) {
    const section = source.split(`\n  ${job}:\n`)[1].split(/\n  [a-z][a-z-]*:\n/)[0];
    assert(section.includes('needs: [work-domain]'));
  }
  console.log('canonical join: exact original five checks required; source observation work has no cycle with them');
}

main().catch(error => { console.error(error); process.exitCode = 1; });
