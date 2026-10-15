'use strict';

const crypto = require('crypto');
const fs = require('fs');
const path = require('path');
const {execFileSync} = require('child_process');

const names = ['gooo', 'ci-proof'];
const sha256 = bytes => crypto.createHash('sha256').update(bytes).digest('hex');

function currentContext(env = process.env) {
  const context = {
    repository: env.GITHUB_REPOSITORY, source_head: env.GOOO_CI_SOURCE_HEAD,
    run_id: env.GITHUB_RUN_ID, run_attempt: env.GITHUB_RUN_ATTEMPT,
    platform: process.platform, arch: process.arch, toolchain: 'go1.27.2',
  };
  if (!context.repository || !/^[0-9a-f]{40}$/.test(context.source_head || '') ||
      !/^[1-9][0-9]*$/.test(context.run_id || '') || !/^[1-9][0-9]*$/.test(context.run_attempt || '')) {
    throw new Error('execution tool origin needs the current source/run/attempt');
  }
  return context;
}

function writeOrigin(root, context) {
  const actualHead = execFileSync('git', ['rev-parse', 'HEAD'], {encoding: 'utf8'}).trim();
  if (actualHead !== context.source_head) throw new Error('execution tool source head mismatch');
  const tools = names.map(name => {
    const file = path.join(root, name);
    const info = execFileSync('go', ['version', '-m', file], {encoding: 'utf8'});
    if (!info.split('\n')[0].endsWith(`: ${context.toolchain}`)) throw new Error('execution tool Go version mismatch');
    const bytes = fs.readFileSync(file);
    return {name, sha256: sha256(bytes), bytes: bytes.length, build_info: info};
  });
  const origin = {schema: 'gooo/ci-execution-tools/v1', ...context, producer_job: 'evidence', tools};
  fs.writeFileSync(path.join(root, 'origin.json'), JSON.stringify(origin, null, 2) + '\n');
  return origin;
}

function verifyOrigin(root, origin, context) {
  if (origin.schema !== 'gooo/ci-execution-tools/v1' || origin.producer_job !== 'evidence') {
    throw new Error('execution tool producer mismatch');
  }
  for (const [key, value] of Object.entries(context)) {
    if (origin[key] !== value) throw new Error(`execution tool ${key} mismatch`);
  }
  if (!Array.isArray(origin.tools) || origin.tools.length !== names.length) throw new Error('execution tools missing');
  for (const [index, name] of names.entries()) {
    const tool = origin.tools[index];
    if (tool.name !== name) throw new Error('execution tool member mismatch');
    const bytes = fs.readFileSync(path.join(root, name));
    if (tool.bytes !== bytes.length || tool.sha256 !== sha256(bytes)) throw new Error(`execution tool ${name} bytes mismatch`);
    const info = execFileSync('go', ['version', '-m', path.join(root, name)], {encoding: 'utf8'});
    // The download directory changes; retain and compare the embedded build data.
    const build = text => text.slice(text.indexOf(': ') + 2);
    if (!info.split('\n')[0].endsWith(`: ${context.toolchain}`) || build(info) !== build(tool.build_info)) {
      throw new Error(`execution tool ${name} build info mismatch`);
    }
  }
  return origin;
}

if (require.main === module) {
  const [mode, root] = process.argv.slice(2);
  if (!root || !['write', 'verify'].includes(mode)) throw new Error('usage: tool_origin.js write|verify DIRECTORY');
  const context = currentContext();
  if (mode === 'write') writeOrigin(root, context);
  else verifyOrigin(root, JSON.parse(fs.readFileSync(path.join(root, 'origin.json'), 'utf8')), context);
}

module.exports = {currentContext, writeOrigin, verifyOrigin};
