'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
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
], ['excluded']), authorization.sha256(`a\t100644\t${'a'.repeat(40)}\nz\t100644\t${'z'.repeat(40)}\n`));
const receipt = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'governance-denominator-v2-migration.json'), 'utf8'));
assert.doesNotThrow(() => authorization.validateRegressionRepairReceipt(receipt));
assert.doesNotThrow(() => authorization.validateIncompletePropagationOutcome(receipt));
assert.equal(receipt.outcome, 'REFUTED_INCOMPLETE_PROPAGATION');
assert.equal(receipt.cells.length, 1);
assert.equal(receipt.cells[0].id, 'REGRESSION_REPAIR');
assert.equal(receipt.cells[0].proof_choice, 'REGRESSION');
assert.equal(receipt.cells[0].indicator, 'GUARDRAIL');
assert.equal(receipt.cells[0].allowed, 1);
assert.equal(receipt.cells[0].consumed, 1);
assert.equal(receipt.cells[0].replay_decision, 'REFUTED');
const correction = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'governance-denominator-v3-correction.json'), 'utf8'));
assert.doesNotThrow(() => authorization.validateCorrectionChildReceipt(correction));
assert.equal(correction.foundation_override_success_count, 3);
assert.equal(correction.cells.length, 1);
assert.equal(correction.cells[0].id, 'CORRECTION_CHILD');
assert.equal(correction.cells[0].proof_choice, 'REGRESSION');
assert.equal(correction.cells[0].indicator, 'GUARDRAIL');
assert.equal(correction.cells[0].allowed, 1);
assert.equal(correction.cells[0].consumed, 1);
assert.equal(correction.cells[0].replay_decision, 'REFUTED');
assert.equal(correction.parent_repair_receipt, authorization.CORRECTION_CHILD_PARENT_RECEIPT_SHA256);
console.log('foundation authorization tests passed');
