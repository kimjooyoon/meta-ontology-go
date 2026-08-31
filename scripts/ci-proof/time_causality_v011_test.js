'use strict';

const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');

const root = path.join(__dirname, '..', '..');
const read = (relative) => fs.readFileSync(path.join(root, relative), 'utf8');
const readJSON = (relative) => JSON.parse(read(relative));
const digest = (bytes) => `sha256:${crypto.createHash('sha256').update(bytes).digest('hex')}`;

const sourcePath = 'examples/ci-time-causality/main.gooo';
const source = read(sourcePath);
const manifest = readJSON('examples/ci-time-causality/time-manifest.json');
const duration = readJSON('examples/ci-time-causality/duration-receipt.json');
const clocks = readJSON('examples/ci-time-causality/clock-domains.json');
const replay = readJSON('examples/ci-time-causality/replay-receipt.json');
const generated = readJSON('examples/ci-time-causality/generated/evaluator-binding.json');
const dispatchReceipt = readJSON('.github/governance-denominator-v6-foundation-authorization-dispatch.json');
const assets = [
  ['clock-domains.json', 537649700, 770, 'sha256:4d858df4758881a329b6e4c52d76c2d4d2e54f114419b873ba90659fc5c9988b'],
  ['duration-receipt.json', 537649697, 9519, 'sha256:58b2236d930da8f6fd8724f766b3ca0a26a4d90de51b163194ecc8350d765a67'],
  ['operations.ndjson', 537649701, 8742, 'sha256:c5e704fa5542a69f9e950982f4fdd492e469c900d3ace3ce0408e05113aa5707'],
  ['replay-receipt.json', 537649699, 505, 'sha256:ee2b19165fb9d56d2ddbd3f96416ed80c4642508321cf24db36fe2b510733a71'],
  ['time-manifest.json', 537649698, 4877, 'sha256:67ec377f670f925dbb346bd83eee233ab53a01b05be4b56368517e2fef7554a3'],
  ['time-report.md', 537649709, 3414, 'sha256:b44c42f2188747fff4dee88458ea04a2e2b54bff32367f24fe2af19226cecf17'],
];

assert.equal(digest(source), 'sha256:a45675079c7e80c8324578ed999e73ccdb6f553f2c1ce35cac223e2809fb4fc1');
for (const [name, id, size, expected] of assets) {
  const bytes = read(`examples/ci-time-causality/${name}`);
  assert.equal(bytes.length, size, `${name} size`);
  assert.equal(digest(bytes), expected, `${name} digest`);
  assert.equal(id > 0, true);
}

assert.equal(manifest.schema, 'gooo/ci-time-causality/time-manifest/v1');
assert.equal(manifest.contract_id, 'ci-time-causality-v1');
assert.equal(manifest.source_path, 'examples/time-causality/main.gooo');
assert.equal(manifest.source_digest, digest(source));
assert.equal(manifest.immutable_fixture_digest, 'sha256:d41f4c6fdc33a5442339d83fdc1f4e68f8fcb317393188aa6b0fe67004d662cd');
assert.equal(manifest.ir_digest, 'sha256:1292a28d338696ff8245a309ca7abcf18ed01991b9eaebba0a4c1ba18c43e5f0');
assert.equal(manifest.generated_evaluator_digest, 'sha256:cd837bed10b30d0d42d5990d39535c168904f3ce261f03787574666d6be12d6f');
assert.equal(manifest.activity_count, 12);
assert.equal(manifest.cell_count, 12);
assert.equal(manifest.activity_cell_one_to_one, true);
assert.equal(manifest.artifact_count, 6);
assert.deepEqual(manifest.summary, {total: 12, closed: 3, unknown: 4, refuted: 5});
assert.deepEqual(manifest.retry_attempts, [1, 2]);
assert.equal(manifest.repository_writes, 0);
assert.equal(manifest.local_test_executions, 0);
assert.equal(manifest.cross_project_required_gates, 0);
assert.equal(manifest.verification_authority, 'github-actions');

assert.equal(duration.schema, 'gooo/ci-time-causality/duration-receipt/v1');
assert.equal(duration.contract_id, 'ci-time-causality-v1');
assert.deepEqual(duration.summary, {total: 12, closed: 3, unknown: 4, refuted: 5});
assert.equal(duration.results.length, 12);
assert.equal(duration.results.filter((item) => item.decision === 'CLOSED').length, 3);
assert.equal(duration.results.filter((item) => item.decision === 'UNKNOWN').length, 4);
assert.equal(duration.results.filter((item) => item.decision === 'REFUTED').length, 5);
assert.equal(duration.aggregation_rule, 'Only same operation_id, run_id, job_id, provider, and clock_domain may form one duration; source-ci and opentofu remain separate observations.');
assert.deepEqual(duration.retry_attempts, [1, 2]);
assert.equal(duration.results.find((item) => item.case_id === 'refuted-negative-duration').reason, 'REFUTED_CLOCK_ORDER');
assert.equal(duration.results.find((item) => item.case_id === 'refuted-clamp-to-zero-attempt').reason, 'CLAMP_TO_ZERO_FORBIDDEN');

assert.equal(clocks.schema, 'gooo/ci-time-causality/clock-domains/v1');
assert.equal(clocks.domains.length, 3);
assert.equal(clocks.domains.find((domain) => domain.id === 'github.actions.run.api.v1').resolution_ms, 1000);
assert.equal(clocks.domains.find((domain) => domain.id === 'github.actions.job.api.v1').resolution_ms, 1000);
assert.equal(replay.schema, 'gooo/ci-time-causality/replay/v1');
assert.equal(replay.deterministic, true);
assert.equal(replay.replay_count, 2);
assert.equal(replay.decision, 'CLOSED');

const operations = read('examples/ci-time-causality/operations.ndjson').trimEnd().split('\n').map((line) => JSON.parse(line));
assert.equal(operations.length, 16);
assert.equal(operations.filter((item) => item.record_type === 'observation').length, 12);
const retries = operations.filter((item) => item.record_type === 'retry_lineage').sort((left, right) => left.attempt - right.attempt);
assert.deepEqual(retries.map((item) => item.attempt), [1, 2]);
assert.deepEqual(retries.map((item) => item.job_id), ['99405870188', '99408612206']);
assert.deepEqual(retries.map((item) => item.artifact_id), ['9748462083', '9748520364']);
assert.deepEqual(retries.map((item) => item.decision), ['REFUTED', 'REFUTED']);
assert.deepEqual(retries.map((item) => item.artifact_digest), [
  'sha256:74b24a4c24acd853f5d661aa5ad88b64f56dd8186d98288f87981fd7b4bd3979',
  'sha256:1d0218ed0945ef76ea9cc396ad7a9e0916bc6243ebf6be0fe04efc8e9878b656',
]);

const audit = read('examples/ci-time-causality/release-history/post-release-audit.ndjson').trimEnd().split('\n').map((line) => JSON.parse(line));
assert.equal(audit.length, 6);
assert.deepEqual(audit.filter((item) => item.status === 'FAILED').map((item) => item.failure_code), ['UNDEFINED_SCHEMA_FIELD', 'SHELL_VARIABLE_PATH_COLLISION']);
assert.equal(audit.at(-1).failed_attempts_counted_as_success, false);
assert.equal(audit.at(-1).v0_1_0_deleted_or_recreated, false);
assert.equal(generated.schema, 'gooo/ci-time-causality/generated-evaluator-binding/v1');
assert.equal(generated.source_path, manifest.source_path);
assert.equal(generated.source_digest, manifest.source_digest);
assert.equal(generated.semantic_ir_digest, manifest.ir_digest);
assert.equal(generated.generated_evaluator_digest, manifest.generated_evaluator_digest);
assert.equal(generated.ci_implementation_path, 'scripts/ci-effort-observation/time_causality.go');
assert.equal(generated.negative_duration_reason, 'REFUTED_CLOCK_ORDER');
assert.equal(generated.clamp_to_zero_policy, 'FORBIDDEN');

const binding = dispatchReceipt.time_causality;
assert.equal(binding.schema, 'gooo/ci-time-causality/v0.1.1');
assert.equal(binding.contract_id, 'ci-time-causality-v1');
assert.equal(binding.release.release_id, 379586518);
assert.equal(binding.release.immutable, true);
assert.equal(binding.release.target_commit, '59b72a990b473199af81b8714b107798ab0533aa');
assert.equal(binding.source.sha256, manifest.source_digest);
assert.equal(binding.immutable_fixture_digest, manifest.immutable_fixture_digest);
assert.equal(binding.semantic_ir_digest, manifest.ir_digest);
assert.equal(binding.generated_evaluator.path, generated.ci_implementation_path);
assert.equal(binding.generated_evaluator.sha256, manifest.generated_evaluator_digest);
assert.equal(binding.generated_evaluator.binding_file_sha256, digest(read('examples/ci-time-causality/generated/evaluator-binding.json')));
assert.deepEqual(binding.output_assets, assets.map(([name, id, size, sha256]) => ({id, name, size_bytes: size, sha256})));
assert.deepEqual(binding.summary, {cells: 12, activities: 12, total: 12, closed: 3, unknown: 4, refuted: 5});
const migration = binding.scope_denominator_migration;
assert.equal(migration.schema, 'gooo/language-syntax-roundtrip/scope-denominator-migration/v1');
assert.equal(migration.reason, 'CI_TIME_CAUSALITY_SOURCE_AND_SCOPE_REGISTRATION');
assert.equal(migration.append_only, true);
assert.deepEqual(migration.previous, {
  total: 45,
  valid: 42,
  invalid: 3,
  capability: 44,
  governance: 1,
  registered_source_path: null,
});
assert.deepEqual(migration.current, {
  total: 46,
  valid: 43,
  invalid: 3,
  capability: 45,
  governance: 1,
  registered_source_path: 'examples/ci-time-causality/main.gooo',
});
assert.deepEqual(migration.delta, {add: 1, retire: 0, split: 0});
assert.deepEqual(migration.added_source_paths, ['examples/ci-time-causality/main.gooo']);
assert.deepEqual(migration.added_workflow_paths, ['.github/workflows/transformation-effect.yml']);
assert.deepEqual(migration.source_registration_paths, [
  'examples/language-syntax-roundtrip/corpus.json',
  'internal/meta/languagereadiness/languagesyntax/registry.go',
  'internal/meta/languagereadiness/languagesyntax/model.go',
  'internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go',
]);
assert.deepEqual(migration.workflow_path_ids, ['.github/workflows/transformation-effect.yml']);
assert.equal(migration.lowered, false);
assert.equal(binding.aggregation_rule, duration.aggregation_rule);
assert.equal(binding.negative_duration_reason, 'REFUTED_CLOCK_ORDER');
assert.equal(binding.clamp_to_zero_policy, 'FORBIDDEN');
assert.equal(binding.source_ci_opentofu_separate, true);
assert.deepEqual(binding.retry_attempts, [1, 2]);
assert.deepEqual(binding.ci_effort_retry_history.map((item) => ({attempt: item.run_attempt, job_id: item.job_id, artifact_id: item.artifact_id, artifact_sha256: item.artifact_sha256, decision: item.decision})), [
  {attempt: 1, job_id: 99405870188, artifact_id: 9748462083, artifact_sha256: 'sha256:74b24a4c24acd853f5d661aa5ad88b64f56dd8186d98288f87981fd7b4bd3979', decision: 'REFUTED'},
  {attempt: 2, job_id: 99408612206, artifact_id: 9748520364, artifact_sha256: 'sha256:1d0218ed0945ef76ea9cc396ad7a9e0916bc6243ebf6be0fe04efc8e9878b656', decision: 'REFUTED'},
]);
assert.deepEqual(binding.post_release_audit.failures.map((item) => item.failure_code), ['UNDEFINED_SCHEMA_FIELD', 'SHELL_VARIABLE_PATH_COLLISION']);
assert.equal(binding.post_release_audit.sha256, digest(read('examples/ci-time-causality/release-history/post-release-audit.ndjson')));
assert.equal(binding.post_release_audit.failed_attempts_counted_as_success, false);
assert.equal(binding.post_release_audit.v0_1_0_deleted_or_recreated, false);
assert.equal(binding.post_release_audit.reclassification, 'FORBIDDEN');
assert.equal(binding.append_only, true);

console.log('v0.1.1 immutable CI time-causality source, assets, evaluator binding, and append-only audit passed');
