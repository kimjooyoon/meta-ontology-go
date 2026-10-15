'use strict';

const checks = [
  ['format', 'gofmt'], ['vet', 'go vet'], ['test', 'go test'],
  ['race', 'go test -race'], ['semantic', 'Semantic conformance'],
];

async function joinCanonicalChecks(github, context, env, options = {}) {
  const now = options.now || Date.now;
  const sleep = options.sleep || (ms => new Promise(resolve => setTimeout(resolve, ms)));
  const started = Date.parse(env.GOOO_CI_STARTED_AT);
  const budget = Number(env.GOOO_CI_BUDGET_MS);
  const attempt = Number(env.GITHUB_RUN_ATTEMPT);
  const head = env.GOOO_CI_SOURCE_HEAD;
  if (!Number.isFinite(started) || !Number.isSafeInteger(budget) || budget <= 0 ||
      !Number.isSafeInteger(attempt) || attempt <= 0 || !/^[0-9a-f]{40}$/.test(head || '')) {
    throw new Error('canonical join requires the original clock, head and attempt');
  }
  while (now() < started + budget) {
    const jobs = [];
    for (let page = 1; ; page++) {
      if (page > 1000) throw new Error('canonical join pagination limit');
      const response = await github.request(
        'GET /repos/{owner}/{repo}/actions/runs/{run_id}/attempts/{attempt_number}/jobs',
        {...context.repo, run_id: context.runId, attempt_number: attempt, per_page: 100, page},
      );
      if (!Array.isArray(response.data?.jobs)) throw new Error('canonical join jobs missing');
      jobs.push(...response.data.jobs);
      if (response.data.jobs.length < 100) break;
    }
    const retained = [];
    for (const [key, name] of checks) {
      const matches = jobs.filter(job => job.name === name);
      if (matches.length > 1) throw new Error(`canonical join duplicate ${name}`);
      if (matches.length === 0) continue;
      const job = matches[0];
      if (job.head_sha !== head || !Number.isSafeInteger(job.id) || job.id <= 0) {
        throw new Error(`canonical join source mismatch: ${name}`);
      }
      if (job.status !== 'completed') continue;
      if (job.conclusion !== 'success') throw new Error(`canonical check did not succeed: ${name}`);
      retained.push({key, name, id: job.id, conclusion: job.conclusion});
    }
    if (retained.length === checks.length) {
      return {schema: 'gooo/ci-canonical-join/v1', source_head: head, run_id: context.runId,
        run_attempt: attempt, jobs: retained, results: retained.map(job => `${job.key}=success`).join(',')};
    }
    const remaining = started + budget - now();
    if (remaining > 0) await sleep(Math.min(15000, remaining));
  }
  throw new Error('original shared work budget exhausted before canonical checks joined');
}

module.exports = {joinCanonicalChecks};
