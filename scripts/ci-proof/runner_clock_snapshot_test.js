'use strict';

const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {spawnSync} = require('node:child_process');
const helper = path.resolve('.github/actions/work-domain/observe-clock.sh');
const head = 'a'.repeat(40);

function observe(t, sequence, apiFailure = false) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'gooo-clock-acquisition-'));
  t.after(() => fs.rmSync(root, {recursive: true, force: true}));
  const bin = path.join(root, 'bin');
  const origins = path.join(root, 'origins');
  fs.mkdirSync(bin); fs.mkdirSync(origins);
  const executable = (name, body) => {
    const file = path.join(bin, name);
    fs.writeFileSync(file, body, {mode: 0o700});
    return file;
  };
  executable('gh', `#!/usr/bin/env node
const fs = require('node:fs');
const root = process.env.TEST_ROOT;
const countPath = root + '/reads';
let count = fs.existsSync(countPath) ? Number(fs.readFileSync(countPath)) : 0;
if (process.argv.some(arg => arg.includes('/jobs?'))) {
  if (process.env.API_FAILURE === 'true') {
    process.stdout.write('{partial'); process.stderr.write('upstream read failed'); process.exit(1);
  }
  process.stdout.write(' [ {"total_count":1,"jobs":[{"observation":' + count + '}]} ]\\n');
} else {
  fs.writeFileSync(countPath, String(++count));
  process.stdout.write('  {"observation":' + count + '}\\n');
}
`);
  const tool = executable('domain', `#!/usr/bin/env node
const fs = require('node:fs');
const root = process.env.TEST_ROOT;
const file = root + '/native-calls.jsonl';
const count = fs.existsSync(file) ? fs.readFileSync(file, 'utf8').trim().split('\\n').length : 0;
fs.appendFileSync(file, JSON.stringify(process.argv.slice(2)) + '\\n');
const sequence = JSON.parse(process.env.NATIVE_SEQUENCE);
const code = sequence[Math.min(count, sequence.length - 1)];
if (code === 0) process.stdout.write('{"decision":"READY","effective_started_at":"2026-10-27T00:00:00Z"}\\n');
else process.stderr.write('native diagnostic ' + code + '\\n');
process.exit(code);
`);
  executable('sleep', '#!/bin/sh\nexit 0\n');
  const result = spawnSync('bash', [helper, tool, origins, 'CI evidence'], {
    encoding: 'utf8', timeout: 10000,
    env: {...process.env, PATH: `${bin}:${process.env.PATH}`, TEST_ROOT: root,
      NATIVE_SEQUENCE: JSON.stringify(sequence), API_FAILURE: String(apiFailure),
      EXPECTED_SOURCE_HEAD: head, GITHUB_REPOSITORY: 'owner/repo', GITHUB_RUN_ID: '17',
      GITHUB_RUN_ATTEMPT: '2', GOOO_CI_BUDGET_MS: '3000000'},
  });
  if (result.error) throw result.error;
  const parents = fs.readdirSync(origins).filter(name => name.startsWith('clock-observations.'));
  const frames = parents.flatMap(parent => fs.readdirSync(path.join(origins, parent))
    .sort().map(name => path.join(origins, parent, name)));
  const calls = fs.existsSync(path.join(root, 'native-calls.jsonl'))
    ? fs.readFileSync(path.join(root, 'native-calls.jsonl'), 'utf8').trim().split('\n').map(JSON.parse) : [];
  return {root, origins, result, frames, calls};
}

test('pending publication reobserves the same tuple and retains every original frame', t => {
  const {origins, result, frames, calls} = observe(t, [75, 0]);
  assert.equal(result.status, 0, result.stderr);
  assert.equal(frames.length, 2);
  for (let i = 0; i < frames.length; i++) {
    assert.equal(fs.readFileSync(path.join(frames[i], 'workflow-run.json'), 'utf8'),
      `  {"observation":${i + 1}}\n`);
    assert.equal(fs.readFileSync(path.join(frames[i], 'attempt-job-pages.json'), 'utf8'),
      ` [ {"total_count":1,"jobs":[{"observation":${i + 1}}]} ]\n`);
    assert.equal(fs.readFileSync(path.join(frames[i], 'runner-clock.exit'), 'utf8').trim(), String([75, 0][i]));
  }
  for (const args of calls) {
    for (const [flag, value] of [['--head', head], ['--run-id', '17'], ['--attempt', '2'],
      ['--budget-ms', '3000000'], ['--job', 'CI evidence'], ['--producer', 'Work domain preparation']]) {
      assert.equal(args[args.indexOf(flag) + 1], value);
    }
  }
  assert.equal(calls.length, 2);
  assert.equal(fs.readFileSync(path.join(origins, 'workflow-run.json'), 'utf8'),
    fs.readFileSync(path.join(frames[1], 'workflow-run.json'), 'utf8'));
  assert.equal(fs.readFileSync(path.join(origins, 'selected-clock-observation.txt'), 'utf8').trim(), frames[1]);
});

test('native contradictions never trigger another read or publish admission', t => {
  const {origins, result, frames, calls} = observe(t, [1, 0]);
  assert.equal(result.status, 1, result.stderr);
  assert.equal(frames.length, 1); assert.equal(calls.length, 1);
  assert.equal(fs.existsSync(path.join(origins, 'runner-clock.json')), false);
  assert.match(fs.readFileSync(path.join(frames[0], 'runner-clock.json.stderr'), 'utf8'), /native diagnostic 1/);
});

test('pending observations exhaust the acquisition count without a clock', t => {
  const {origins, result, frames, calls} = observe(t, [75]);
  assert.equal(result.status, 75, result.stderr);
  assert.equal(frames.length, 16); assert.equal(calls.length, 16);
  assert.equal(fs.existsSync(path.join(origins, 'runner-clock.json')), false);
});

test('failed API reads preserve partial bytes and never reach clock admission', t => {
  const {origins, result, frames, calls} = observe(t, [0], true);
  assert.equal(result.status, 1, result.stderr);
  assert.equal(frames.length, 1); assert.equal(calls.length, 0);
  assert.equal(fs.readFileSync(path.join(frames[0], 'attempt-job-pages.json'), 'utf8'), '{partial');
  assert.match(fs.readFileSync(path.join(frames[0], 'attempt-job-pages.json.stderr'), 'utf8'), /upstream read failed/);
  assert.equal(fs.existsSync(path.join(origins, 'runner-clock.json')), false);
});

test('composite exports the diagnostic root before its one bounded acquisition', () => {
  const action = fs.readFileSync('.github/actions/work-domain/action.yml', 'utf8');
  assert.match(action, /timeout --signal=KILL 30s bash "\$GITHUB_ACTION_PATH\/observe-clock.sh"/);
  const exportAt = action.indexOf("printf 'GOOO_CI_ORIGIN_ROOT=%s\\n'");
  assert.ok(exportAt >= 0 && exportAt < action.indexOf('observe-clock.sh'));
});
