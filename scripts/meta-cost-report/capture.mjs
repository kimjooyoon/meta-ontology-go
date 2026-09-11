import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';

const root = process.argv[2];
if (!root) throw new Error('An explicit caller-owned output directory is required');
const read = name => fs.readFileSync(path.join(root, name));
const digest = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const run = JSON.parse(read('source-run.json'));
const job = JSON.parse(read('source-job.json'));
const head = 'a8b3813c26252d523ae81ad8d81cfcdbf267788e';
if (run.id !== 34081014161 || run.head_sha !== head || run.run_attempt !== 1 ||
    run.repository?.full_name !== 'kimjooyoon/meta-ontology-go' ||
    run.status !== 'completed' || run.conclusion !== 'success' ||
    job.id !== 101617890564 || job.run_id !== run.id ||
    job.name !== 'CI policy' || job.status !== 'completed' || job.conclusion !== 'success') {
  throw new Error('Source run/job identity or terminal outcome does not match pinned baseline');
}
const log = read('source.log');
const lines = [];
for (const line of log.toString('utf8').split('\n')) {
  const match = line.match(/^\d{4}-\d\d-\d\dT\S+Z (\{"schema":"gooo\/meta-execution-driver-boundary\/v1".*\})\r?$/);
  if (!match) continue;
  const event = JSON.parse(match[1]);
  if (event.head_sha !== head) throw new Error('Mixed-head driver observation');
  lines.push(match[1]);
}
if (lines.length === 0) throw new Error('No driver observations; missing logs are not zero cost');
const events = Buffer.from(lines.join('\n') + '\n');
fs.writeFileSync(path.join(root, 'driver-events.ndjson'), events);
fs.writeFileSync(path.join(root, 'capture-receipt.json'), JSON.stringify({
  schema: 'gooo/meta-cost-capture/v1',
  source_repository: run.repository.full_name,
  source_head: head,
  source_run_id: run.id,
  source_run_attempt: run.run_attempt,
  source_job_id: job.id,
  consumer_workflow_sha: process.env.GITHUB_SHA,
  source_log_sha256: digest(log),
  events_sha256: digest(events),
  consumer_toolchain_observation_sha256: digest(read('toolchain.txt')),
  extracted_events: lines.length,
  scope: 'HISTORICAL_DRIVER_LOG_CAPTURE_ONLY',
  semantic_authority: 'NONE',
  improvement: 'UNKNOWN'
}, null, 2) + '\n');
