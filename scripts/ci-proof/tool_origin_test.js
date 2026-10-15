'use strict';

const assert = require('assert');
const fs = require('fs');
const os = require('os');
const path = require('path');
const {execFileSync} = require('child_process');
const {currentContext, writeOrigin, verifyOrigin} = require('./tool_origin');

const root = fs.mkdtempSync(path.join(os.tmpdir(), 'gooo-execution-tool-contract-'));
try {
  const context = currentContext({
    GITHUB_REPOSITORY: 'kimjooyoon/meta-ontology-go',
    GOOO_CI_SOURCE_HEAD: execFileSync('git', ['rev-parse', 'HEAD'], {encoding: 'utf8'}).trim(),
    GITHUB_RUN_ID: '17', GITHUB_RUN_ATTEMPT: '1',
  });
  assert.throws(() => currentContext({}), /current source\/run\/attempt/);
  const fixture = path.join(root, 'fixture.go');
  fs.writeFileSync(fixture, 'package main\nfunc main() {}\n');
  execFileSync('go', ['build', '-buildvcs=false', '-o', path.join(root, 'gooo'), fixture]);
  fs.copyFileSync(path.join(root, 'gooo'), path.join(root, 'ci-proof'));
  const origin = writeOrigin(root, context);
  assert.deepStrictEqual(verifyOrigin(root, origin, context), origin);
  for (const key of Object.keys(context)) {
    assert.throws(() => verifyOrigin(root, {...origin, [key]: 'other'}, context), /mismatch/);
  }
  assert.throws(() => verifyOrigin(root, {...origin, producer_job: 'other'}, context), /producer mismatch/);
  assert.throws(() => verifyOrigin(root, {...origin, tools: []}, context), /tools missing/);
  assert.throws(() => verifyOrigin(root, {...origin, tools: origin.tools.slice().reverse()}, context), /member mismatch/);
  const wrongBuild = structuredClone(origin);
  wrongBuild.tools[0].build_info = wrongBuild.tools[0].build_info.replace('go1.27.2', 'go1.27.1');
  assert.throws(() => verifyOrigin(root, wrongBuild, context), /build info mismatch/);
  fs.appendFileSync(path.join(root, 'gooo'), 'changed');
  assert.throws(() => verifyOrigin(root, origin, context), /bytes mismatch/);
  console.log('execution tool origin: same-source tools accepted; cross-source/run/attempt and changed bytes rejected');
} finally {
  fs.rmSync(root, {recursive: true, force: true});
}
