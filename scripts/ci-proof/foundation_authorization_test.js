'use strict';

const assert = require('node:assert/strict');
const authorization = require('./foundation_authorization');

assert.equal(authorization.FOUNDATION_OVERRIDE_SUCCESS_COUNT, 3);
assert.equal(authorization.FOUNDATION_OVERRIDE_MARKER, 'FOUNDATION_OVERRIDE_SUCCESS_COUNT=3');
assert.deepEqual(authorization.canonicalPathNames([
  {filename: 'b', previous_filename: null},
  {filename: 'a', previous_filename: null},
  {filename: 'a', previous_filename: null},
]), ['a', 'b']);
assert.equal(authorization.digestChangedPaths(['a', 'b']), authorization.sha256('a\nb\n'));
assert.equal(authorization.digestTreeEntries([
  {path: 'z', type: 'blob', mode: '100644', sha: 'z'.repeat(40)},
  {path: 'a', type: 'blob', mode: '100644', sha: 'a'.repeat(40)},
  {path: 'excluded', type: 'blob', mode: '100644', sha: 'e'.repeat(40)},
], ['excluded']), authorization.sha256(JSON.stringify([
  {path: 'a', mode: '100644', sha: 'a'.repeat(40)},
  {path: 'z', mode: '100644', sha: 'z'.repeat(40)},
])));
console.log('foundation authorization tests passed');
