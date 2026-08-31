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
const correctionWithOwnedOutcome = {
  ...correction,
  parent_outcome: 'REFUTED_INCOMPLETE_PROPAGATION',
  outcome: 'CLOSED',
  causal: authorization.SCHEMA_COHERENCE_MIGRATION_REASON,
};
assert.doesNotThrow(() => authorization.validateCorrectionChildReceipt(correctionWithOwnedOutcome));
const parentWithoutOutcome = JSON.parse(JSON.stringify(receipt));
delete parentWithoutOutcome.outcome;
delete parentWithoutOutcome.cells[0].outcome;
assert.doesNotThrow(() => authorization.validateRegressionRepairReceipt(parentWithoutOutcome));
const migration = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'governance-denominator-v4-schema-coherence.json'), 'utf8'));
assert.doesNotThrow(() => authorization.validateSchemaCoherenceMigrationReceipt(migration));
assert.deepEqual(authorization.schemaCoherenceInputFieldStates(authorization.SCHEMA_COHERENCE_MIGRATION_INPUT), Object.fromEntries(authorization.SCHEMA_COHERENCE_MIGRATION_INPUT_FIELDS.map((field) => [field, 'CLOSED'])));
assert.equal(authorization.classifySchemaCoherenceInput(authorization.SCHEMA_COHERENCE_MIGRATION_INPUT).decision, 'CLOSED');
assert.equal(authorization.classifySchemaCoherenceInput({schema: 'gooo/receipt-schema-migration/v0.1.2'}).unknown_count, 6);
const staleInput = {...authorization.SCHEMA_COHERENCE_MIGRATION_INPUT, target_commit: '0'.repeat(40)};
assert.equal(authorization.classifySchemaCoherenceInput(staleInput).decision, 'UNKNOWN');
assert.equal(authorization.resolveSchemaCoherenceDecision(['CLOSED', 'UNKNOWN', 'REFUTED']), 'REFUTED');
assert.equal(authorization.resolveSchemaCoherenceDecision(['CLOSED', 'UNKNOWN']), 'UNKNOWN');
assert.equal(authorization.resolveSchemaCoherenceDecision(['CLOSED']), 'CLOSED');
const executableGuardianScope = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', '.github', 'governance-denominator-v5-executable-guardian-scope.json'), 'utf8'));
assert.doesNotThrow(() => authorization.validateExecutableGuardianScopeReceipt(executableGuardianScope));
assert.deepEqual(executableGuardianScope.lineage, authorization.EXECUTABLE_GUARDIAN_SCOPE_LINEAGE);
assert.equal(executableGuardianScope.cells.length, 1);
assert.equal(executableGuardianScope.cells[0].id, 'EXECUTABLE_GUARDIAN_SCOPE_ADOPTION');
assert.equal(executableGuardianScope.cells[0].proof_choice, 'REGRESSION');
assert.equal(executableGuardianScope.cells[0].indicator, 'GUARDRAIL');
assert.equal(executableGuardianScope.cells[0].allowed, 1);
assert.equal(executableGuardianScope.cells[0].consumed, 1);
assert.equal(executableGuardianScope.cells[0].replay_decision, 'REFUTED');
assert.equal(executableGuardianScope.cells[0].parent_outcome, authorization.EXECUTABLE_GUARDIAN_SCOPE_PARENT_OUTCOME);
assert.equal(executableGuardianScope.cells[0].outcome, 'CLOSED');
assert.equal(executableGuardianScope.acceptance_ids.length, 8);
assert.deepEqual(authorization.executableGuardianScopeInputFieldStates(authorization.EXECUTABLE_GUARDIAN_SCOPE_INPUT), Object.fromEntries(authorization.EXECUTABLE_GUARDIAN_SCOPE_INPUT_FIELDS.map((field) => [field, 'CLOSED'])));
assert.equal(authorization.classifyExecutableGuardianScopeInput(authorization.EXECUTABLE_GUARDIAN_SCOPE_INPUT).decision, 'CLOSED');
assert.equal(authorization.classifyExecutableGuardianScopeInput({schema: 'gooo/receipt-schema-migration/v0.2.3'}).decision, 'UNKNOWN');
assert.equal(authorization.classifyExecutableGuardianScopeInput({schema: 'gooo/receipt-schema-migration/v0.2.3'}).unknown_count, 6);
assert.equal(authorization.classifyExecutableGuardianScopeInput({...authorization.EXECUTABLE_GUARDIAN_SCOPE_INPUT, target_commit: '0'.repeat(40)}).decision, 'REFUTED');
assert.equal(authorization.classifyExecutableGuardianScopeInput({...authorization.EXECUTABLE_GUARDIAN_SCOPE_INPUT, protocol: null}).decision, 'REFUTED');
assert.equal(authorization.resolveExecutableGuardianScopeDecision(['CLOSED', 'UNKNOWN', 'REFUTED']), 'REFUTED');
assert.equal(authorization.resolveExecutableGuardianScopeDecision(['CLOSED', 'UNKNOWN']), 'UNKNOWN');
assert.equal(authorization.resolveExecutableGuardianScopeDecision(['CLOSED']), 'CLOSED');
console.log('foundation authorization tests passed');
