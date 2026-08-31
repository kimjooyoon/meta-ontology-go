'use strict';

const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');

const root = path.join(__dirname, '..', '..');
const read = (relative) => fs.readFileSync(path.join(root, relative), 'utf8');
const readJSON = (relative) => JSON.parse(read(relative));
const digest = (bytes) => `sha256:${crypto.createHash('sha256').update(bytes).digest('hex')}`;

const sourcePath = 'examples/receipt-schema-migration-v3/migration.gooo';
const proposalPath = 'examples/receipt-schema-migration-v3/adoption-proposal.json';
const irPath = 'examples/receipt-schema-migration-v3/semantic-ir.json';
const validatorPath = 'examples/receipt-schema-migration-v3/generated/validator.json';
const harnessPath = 'examples/receipt-schema-migration-v3/generated/guardian-harness-cases.json';
const source = read(sourcePath);
const proposal = readJSON(proposalPath);
const ir = readJSON(irPath);
const validator = readJSON(validatorPath);
const harness = readJSON(harnessPath);

assert.equal(digest(source), 'sha256:ab3543bd805a53b6fa8a8e3fb6f10697e3b6f0db8658e471a14906718a137ed3');
assert.equal(digest(read(proposalPath)), 'sha256:578562ec738484c67d70581622981471f4a0eba86b35a726bc73714805fabb55');
assert.equal(digest(read(irPath)), 'sha256:fc3e7442c4eb9c2b0298d7db0982ef8cb6ee26fe683f459783c0197466a62f5f');
assert.equal(digest(read(validatorPath)), 'sha256:f45978d7ffb5cbbb6a8fe9af5122ff10d989a8105219679078f1ad5acdf066a2');
assert.equal(digest(read(harnessPath)), 'sha256:d954abca510f9a7e99c70226d3673db89acb995dc40839973f6202447f7082f3');
assert.equal(ir.schema, 'gooo/receipt-schema-migration/semantic-ir/v3');
assert.equal(ir.version, 'v3');
assert.equal(ir.source_digest, digest(source));
assert.equal(ir.denominator_id, 'receipt-schema-migration-v3');
assert.equal(ir.cell_count, 20);
assert.deepEqual(ir.stage_counts, {COHERENCE: 6, FOUNDATION: 8, REGRESSION: 6});
assert.deepEqual(ir.role_counts, {DRIVER: 6, GUARDRAIL: 8, OUTCOME: 6});
assert.equal(ir.authority.repository_writes, 0);
assert.equal(ir.authority.local_test_executions, 0);
assert.equal(ir.authority.cross_project_required_gates, 0);
assert.equal(ir.authority.product_generation_authorized, false);
assert.deepEqual(ir.precedence, ['REFUTED', 'UNKNOWN', 'CLOSED']);
assert.deepEqual(ir.unknown_fields, ['stage', 'step', 'reason', 'unknown_class', 'next_operation', 'blocked_by']);
assert.deepEqual(ir.migration, {
  from_version: 'v2',
  to_version: 'v3',
  added: 4,
  retired: 0,
  split: 0,
  added_cell_ids: [
    'PROTECTED_PATH_AUTHORIZATION_DISPATCH',
    'FOUNDATION_RECEIPT_EVALUATION',
    'CHANGED_PATH_TUPLE_BINDING',
    'UNAUTHORIZED_PROTECTED_PATH_FAIL_CLOSED',
  ],
  stage_counts_before: {COHERENCE: 5, FOUNDATION: 6, REGRESSION: 5},
  stage_counts_after: {COHERENCE: 6, FOUNDATION: 8, REGRESSION: 6},
  stage_delta: {COHERENCE: 1, FOUNDATION: 2, REGRESSION: 1},
  role_counts_before: {DRIVER: 5, GUARDRAIL: 6, OUTCOME: 5},
  role_counts_after: {DRIVER: 6, GUARDRAIL: 8, OUTCOME: 6},
  role_delta: {DRIVER: 1, GUARDRAIL: 2, OUTCOME: 1},
});
assert.equal(validator.schema, 'gooo/receipt-schema-migration/generated-validator/v1');
assert.equal(validator.ir_digest, harness.ir_digest);
assert.deepEqual(validator.precedence, ir.precedence);
assert.deepEqual(validator.unknown_fields, ir.unknown_fields);
assert.equal(harness.schema, 'gooo/receipt-schema-migration/guardian-harness/v2');
assert.equal(harness.migration_version, 'v3');
assert.deepEqual(harness.fixture_v3, ir.guardian_fixture_v3);
assert.equal(harness.cases.length, 11);
assert.equal(harness.cases.filter((testCase) => testCase.expected === 'CLOSED').length, 2);
assert.equal(harness.cases.filter((testCase) => testCase.expected === 'UNKNOWN').length, 0);
assert.equal(harness.cases.filter((testCase) => testCase.expected === 'REFUTED').length, 9);
assert.equal(proposal.schema, 'gooo/receipt-schema-migration/adoption-proposal/v3');
assert.equal(proposal.proposal_id, 'receipt-schema-migration-v3-adoption');
assert.equal(proposal.proposal_digest, 'sha256:42ae271ea27226bbbf0086ec1fed230e51787b8f44dc68468eafa5607c5a6c7e');
assert.deepEqual(proposal.migration, ir.migration);
assert.deepEqual(proposal.guardian_fixture_v3, ir.guardian_fixture_v3);
assert.deepEqual(proposal.expected_protected_paths, [
  '.github/ci-governance.json',
  '.github/agent-scope-table.md',
  '.github/branch-policy.md',
  '.github/conformance-plan.md',
  '.github/foundation-authorization.json',
  'go.mod',
  'go.sum',
]);
console.log('v0.3.1 source, semantic IR, and generated Guardian evaluator inputs passed');
