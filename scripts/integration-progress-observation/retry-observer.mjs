import {readFileSync} from 'node:fs';
import assert from 'node:assert/strict';

// Exercise the actual inline observer without GitHub requests or old cohorts.
const workflow = readFileSync(process.argv[2], 'utf8');
const begin = '// BEGIN GOOO_OBSERVATION_RETRY';
const end = '// END GOOO_OBSERVATION_RETRY';
assert.equal(workflow.split(begin).length, 2);
assert.equal(workflow.split(end).length, 2);
const source = workflow.slice(workflow.indexOf(begin), workflow.indexOf(end));
assert.ok(workflow.includes(`jq -e '.decision == "COMPLETE" or .decision == "PROGRESS_OBSERVED"'`));
const scenarios = [
  {statuses: [500, 502, 200], count: 3, success: true},
  {statuses: [503, 504, 500], count: 3, success: false},
  {statuses: [403, 200], count: 1, success: false},
  {statuses: [404, 200], count: 1, success: false},
  {statuses: [429, 200], count: 1, success: false},
  {statuses: [0, 200], count: 1, success: false},
  {statuses: [200], count: 1, success: true},
];
for (const scenario of scenarios) {
  const delays = [], logs = [];
  const observe = new Function('core', 'setTimeout', `${source}\nreturn observe;`)(
    {info: text => logs.push(text)}, (ready, delay) => {delays.push(delay); ready();});
  let calls = 0, succeeded = false;
  try {
    const actual = await observe(async () => {
      const status = scenario.statuses[calls++];
      if (status !== 200) throw Object.assign(new Error('synthetic read failure'), {status});
      return 'observed';
    }, 'synthetic-current-read');
    assert.equal(actual, 'observed');
    succeeded = true;
  } catch (error) {
    assert.equal(error.status, scenario.statuses[calls - 1]);
  }
  assert.equal(calls, scenario.count);
  assert.equal(succeeded, scenario.success);
  assert.deepEqual(delays, [250, 500].slice(0, scenario.count - 1));
  assert.equal(logs.length, delays.length);
}
console.log('7 synthetic read scenarios passed; max3 calls; permissions/quota/missing resources unchanged; network0/historical cohorts0');
