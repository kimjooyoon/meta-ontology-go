'use strict';

const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');

const SCHEMA = 'gooo/main-history-reconciliation/v2';
const PROTOCOL_DEFINITION_PATH = '.github/main-history-reconciliation-v2.json';
const REPOSITORY = 'kimjooyoon/meta-ontology-go';
const ROUTE = 'reconciliation_main';
const BRANCH_PREFIX = 'agent/main-history-reconciliation-';
const OWNER = Object.freeze({login: 'kimjooyoon', id: 115961382, type: 'User'});
const CI_APP_ID = 15368;
const REQUIRED_CHECKS = Object.freeze(['CI policy', 'Semantic conformance', 'go test', 'go test -race', 'go vet', 'gofmt', 'CI guardian']);
const UNKNOWN_FIELDS = Object.freeze(['stage', 'step', 'reason', 'unknown_class', 'next_operation', 'blocked_by']);
const DECISION_PRECEDENCE = Object.freeze(['REFUTED', 'UNKNOWN', 'CLOSED']);
const AUTHORIZATION_SCHEMA = 'gooo/linear-tree-reconciliation-owner-record/v2';
const AUTHORIZATION_BINDING = 'CANDIDATE_DIGEST_PLUS_EXACT_IDENTITY_AND_TREE_DIGESTS';
const RECONCILIATION_CODE = 'CI-GUARDIAN-RECONCILIATION-001';

const definition = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', PROTOCOL_DEFINITION_PATH), 'utf8'));

function sha256(value) {
  return `sha256:${crypto.createHash('sha256').update(value).digest('hex')}`;
}

function validSHA(value) {
  return typeof value === 'string' && /^[0-9a-f]{40}$/.test(value);
}

function validDigest(value) {
  return typeof value === 'string' && /^sha256:[0-9a-f]{64}$/.test(value);
}

function validPositiveInteger(value) {
  return Number.isInteger(value) && value > 0;
}

function exactObject(left, right) {
  return JSON.stringify(left) === JSON.stringify(right);
}

function exactArray(left, right) {
  return JSON.stringify(left) === JSON.stringify(right);
}

function codepointCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function canonicalPathList(paths) {
  return [...new Set((Array.isArray(paths) ? paths : []).filter((value) => typeof value === 'string' && value.length > 0))].sort();
}

function canonicalPathNames(files) {
  const paths = (Array.isArray(files) ? files : []).flatMap((file) => typeof file === 'string'
    ? [file]
    : [file && file.filename, file && file.previous_filename]).filter(Boolean);
  return canonicalPathList(paths);
}

function digestPathList(paths) {
  const canonical = canonicalPathList(paths);
  return sha256(canonical.length === 0 ? '' : `${canonical.join('\n')}\n`);
}

function isProtectedPath(value) {
  return value === '.github/agent-scope-table.md'
    || value === '.github/branch-policy.md'
    || value === '.github/ci-governance.json'
    || value === '.github/conformance-plan.md'
    || value === 'go.mod'
    || value === 'go.sum'
    || value.startsWith('.github/workflows/')
    || value.startsWith('internal/verify/')
    || value.startsWith('scripts/ci-evidence/')
    || value.startsWith('scripts/ci-proof/')
    || value.startsWith('scripts/verify/');
}

function canonicalProtectedIntersection(paths) {
  return canonicalPathList(paths).filter(isProtectedPath);
}

function canonicalTreeEntries(entries) {
  const result = [];
  const seen = new Set();
  for (const entry of Array.isArray(entries) ? entries : []) {
    if (!entry || entry.type !== 'blob' || typeof entry.path !== 'string' || !validSHA(entry.sha)) continue;
    if (seen.has(entry.path)) throw new Error(`tree manifest contains duplicate path: ${entry.path}`);
    seen.add(entry.path);
    result.push({path: entry.path, mode: entry.mode || null, sha: entry.sha});
  }
  return result.sort((left, right) => codepointCompare(left.path, right.path) || codepointCompare(String(left.mode), String(right.mode)) || codepointCompare(left.sha, right.sha));
}

function treeManifest(entries) {
  const canonical = canonicalTreeEntries(entries);
  const paths = canonical.map((entry) => entry.path);
  const lines = canonical.map((entry) => `${entry.path}\t${entry.mode}\t${entry.sha}`);
  return {
    paths,
    count: canonical.length,
    paths_digest: digestPathList(paths),
    tree_digest: sha256(lines.length === 0 ? '' : `${lines.join('\n')}\n`),
    manifest_digest: sha256(JSON.stringify(canonical)),
  };
}

function validateProtocolDefinition(candidate = definition) {
  if (!candidate || candidate.schema !== SCHEMA || !exactArray(candidate.decision_precedence, DECISION_PRECEDENCE) || !exactArray(candidate.unknown_fields, UNKNOWN_FIELDS) || !exactObject(candidate.owner_authorization, {
    schema: AUTHORIZATION_SCHEMA,
    binding: AUTHORIZATION_BINDING,
    one_use: true,
    full_candidate_evidence: 'ci-guardian-artifact',
  }) || !Array.isArray(candidate.cells) || candidate.cells.length !== 12) {
    throw new Error('linear-tree reconciliation protocol definition is not exact');
  }
  const ids = new Set();
  for (const cell of candidate.cells) {
    if (!cell || typeof cell.id !== 'string' || ids.has(cell.id) || !['FOUNDATION', 'COHERENCE', 'REGRESSION'].includes(cell.stage) || !['DRIVER', 'OUTCOME', 'GUARDRAIL'].includes(cell.indicator) || cell.proof_choice !== cell.stage || typeof cell.step !== 'string' || cell.producer !== 'linear-tree-reconciliation' || cell.consumer !== 'ci-guardian') {
      throw new Error('linear-tree reconciliation denominator cell is not exact');
    }
    ids.add(cell.id);
  }
  return candidate;
}

function unknownEvidence(stage, step, reason, unknownClass, nextOperation, blockedBy = []) {
  const evidence = {stage, step, reason, unknown_class: unknownClass, next_operation: nextOperation, blocked_by: [...blockedBy]};
  if (!exactArray(Object.keys(evidence), UNKNOWN_FIELDS)) throw new Error('linear-tree reconciliation unknown evidence keys are not exact');
  return evidence;
}

function candidateDigest(candidate) {
  return sha256(JSON.stringify({
    pull_request: candidate && candidate.pull_request,
    base_repo: candidate && candidate.base_repo,
    base_branch: candidate && candidate.base_branch,
    base_sha: candidate && candidate.base_sha,
    head_repo: candidate && candidate.head_repo,
    head_branch: candidate && candidate.head_branch,
    head_sha: candidate && candidate.head_sha,
    source_dev_sha: candidate && candidate.source_dev_sha,
    merge_base_sha: candidate && candidate.merge_base_sha,
    changed_paths: candidate && candidate.changed_paths,
    protected_intersection: candidate && candidate.protected_intersection,
    source_dev_tree: candidate && candidate.source_dev_tree,
    candidate_tree: candidate && candidate.candidate_tree,
    workflow: candidate && candidate.workflow,
  }));
}

function candidateFromPull({pull, changedFiles, sourceDevSHA, mergeBaseSHA, sourceDevTree, candidateTree, workflow}) {
  const paths = canonicalPathNames(changedFiles);
  const protectedPaths = canonicalProtectedIntersection(paths);
  return {
    pull_request: pull.number,
    base_repo: pull.base.repo.full_name,
    base_branch: pull.base.ref,
    base_sha: pull.base.sha,
    head_repo: pull.head.repo.full_name,
    head_branch: pull.head.ref,
    head_sha: pull.head.sha,
    source_dev_sha: sourceDevSHA,
    merge_base_sha: mergeBaseSHA,
    changed_paths: {paths, count: paths.length, digest: digestPathList(paths)},
    protected_intersection: {paths: protectedPaths, count: protectedPaths.length, digest: digestPathList(protectedPaths)},
    source_dev_tree: sourceDevTree,
    candidate_tree: candidateTree,
    workflow: workflow || null,
  };
}

function validateCandidate(candidate) {
  if (!candidate || !validPositiveInteger(candidate.pull_request) || candidate.base_branch !== 'main' || !validSHA(candidate.base_sha) || typeof candidate.base_repo !== 'string' || candidate.head_repo !== candidate.base_repo || !isReconciliationBranch(candidate.head_branch) || !validSHA(candidate.head_sha) || !validSHA(candidate.source_dev_sha) || !validSHA(candidate.merge_base_sha) || !candidate.changed_paths || !candidate.protected_intersection || !candidate.source_dev_tree || !candidate.candidate_tree || !candidate.workflow) throw new Error('linear-tree reconciliation candidate tuple is incomplete');
  const changedPaths = canonicalPathList(candidate.changed_paths.paths);
  const protectedPaths = canonicalProtectedIntersection(changedPaths);
  if (!exactArray(candidate.changed_paths.paths, changedPaths) || candidate.changed_paths.count !== changedPaths.length || candidate.changed_paths.digest !== digestPathList(changedPaths) || !exactArray(candidate.protected_intersection.paths, protectedPaths) || candidate.protected_intersection.count !== protectedPaths.length || candidate.protected_intersection.digest !== digestPathList(protectedPaths)) throw new Error('linear-tree reconciliation changed-path or protected-intersection binding is not exact');
  if (!validTreeBinding(candidate.source_dev_tree) || !validTreeBinding(candidate.candidate_tree)) throw new Error('linear-tree reconciliation candidate tree binding is incomplete');
  if (!candidate.workflow.workflow_ref || !validSHA(candidate.workflow.workflow_sha) || candidate.workflow.runtime_ref !== 'refs/heads/dev' || candidate.workflow.runtime_sha !== candidate.workflow.workflow_sha || candidate.workflow.check_name !== 'CI guardian' || candidate.workflow.app_id !== CI_APP_ID) throw new Error('linear-tree reconciliation workflow identity is incomplete');
  return candidate;
}

function validTreeBinding(binding) {
  return binding && Array.isArray(binding.paths) && Number.isInteger(binding.count) && binding.count === binding.paths.length && exactArray(binding.paths, canonicalPathList(binding.paths)) && validDigest(binding.paths_digest) && validDigest(binding.tree_digest) && validDigest(binding.manifest_digest) && binding.paths_digest === digestPathList(binding.paths);
}

function isReconciliationBranch(value) {
  return typeof value === 'string' && value.startsWith(BRANCH_PREFIX) && value.length > BRANCH_PREFIX.length;
}

function treeAuthorizationBinding(binding) {
  return {
    tree_sha: binding && binding.tree_sha,
    count: binding && binding.count,
    paths_digest: binding && binding.paths_digest,
    tree_digest: binding && binding.tree_digest,
    manifest_digest: binding && binding.manifest_digest,
  };
}

function authorizationCandidateBinding(candidate) {
  return {
    pull_request: candidate && candidate.pull_request,
    base_repo: candidate && candidate.base_repo,
    base_branch: candidate && candidate.base_branch,
    base_sha: candidate && candidate.base_sha,
    head_repo: candidate && candidate.head_repo,
    head_branch: candidate && candidate.head_branch,
    head_sha: candidate && candidate.head_sha,
    source_dev_sha: candidate && candidate.source_dev_sha,
    merge_base_sha: candidate && candidate.merge_base_sha,
    changed_paths: candidate && candidate.changed_paths,
    protected_intersection: candidate && candidate.protected_intersection,
    source_dev_tree: treeAuthorizationBinding(candidate && candidate.source_dev_tree),
    candidate_tree: treeAuthorizationBinding(candidate && candidate.candidate_tree),
    workflow: candidate && candidate.workflow,
  };
}

function validateOwnerAuthorization(record, {candidate, repository = REPOSITORY, now = new Date()} = {}) {
  const expired = !record || typeof record.expires_at !== 'string' || !Number.isFinite(Date.parse(record.expires_at)) || Date.parse(now) >= Date.parse(record.expires_at);
  if (!record || record.schema !== AUTHORIZATION_SCHEMA || record.binding !== AUTHORIZATION_BINDING || record.state !== 'AUTHORIZED' || record.proof_choice !== 'FOUNDATION' || record.intent !== 'AUTHORIZE_LINEAR_TREE_EQUIVALENT_RECONCILIATION_FOR_EXACT_CANDIDATE' || record.required_operation !== 'OWNER_CREATE_ONE_USE_LINEAR_TREE_RECONCILIATION_AUTHORIZATION' || record.reusable !== false || record.one_use !== true || typeof record.nonce !== 'string' || record.nonce.length < 16 || typeof record.issued_at !== 'string' || !Number.isFinite(Date.parse(record.issued_at)) || expired || record.use_count !== 0 || record.reuse_attempts !== 0 || record.stale !== false || record.revoked !== false || !exactObject(record.owner_selection, OWNER) || !exactObject(record.actor, OWNER) || !exactObject(record.candidate_binding, authorizationCandidateBinding(candidate)) || record.candidate_digest !== candidateDigest(candidate) || record.source_dev_sha_at_authorization !== candidate.source_dev_sha || record.base_snapshot_sha !== candidate.base_sha || record.candidate_tree_digest !== candidate.candidate_tree.tree_digest || record.source_dev_tree_digest !== candidate.source_dev_tree.tree_digest || !record.authorization_receipt || record.authorization_receipt.nonce !== record.nonce || record.authorization_receipt.candidate_digest !== record.candidate_digest || !exactObject(record.authorization_receipt.owner, OWNER) || record.no_protection_mutation !== true || record.no_check_bypass !== true || record.no_force_push !== true || record.repository_writes_before_authorization !== 0 || repository !== REPOSITORY) {
    throw new Error(expired || record?.stale === true ? 'linear-tree reconciliation authorization is stale or expired' : 'linear-tree reconciliation owner authorization is missing or not exact');
  }
  return {
    schema: record.schema,
    binding: record.binding,
    state: record.state,
    candidate_binding: record.candidate_binding,
    candidate_digest: record.candidate_digest,
    source_dev_sha_at_authorization: record.source_dev_sha_at_authorization,
    base_snapshot_sha: record.base_snapshot_sha,
    candidate_tree_digest: record.candidate_tree_digest,
    source_dev_tree_digest: record.source_dev_tree_digest,
    owner_selection: record.owner_selection,
    actor: record.actor,
    nonce: record.nonce,
    issued_at: record.issued_at,
    expires_at: record.expires_at,
    stale: record.stale,
    use_count: record.use_count,
    reuse_attempts: record.reuse_attempts,
    revoked: record.revoked,
    authorization_receipt: record.authorization_receipt,
  };
}

async function readOwnerAuthorization({listComments, owner, repo, pullNumber, candidate, now = new Date()}) {
  if (typeof listComments !== 'function') throw new Error('linear-tree reconciliation owner comment reader is unavailable');
  const comments = [];
  for (let page = 1; page <= 1000; page += 1) {
    const response = await listComments({owner, repo, issue_number: pullNumber, per_page: 100, page});
    if (!response || (response.status !== undefined && response.status !== 200) || !Array.isArray(response.data)) throw new Error('linear-tree reconciliation owner comment API returned malformed data');
    comments.push(...response.data);
    if (response.data.length < 100) break;
    if (page === 1000) throw new Error('linear-tree reconciliation owner comment pagination exceeded the limit');
  }
  const records = [];
  for (const comment of comments) {
    if (!comment || !comment.user || comment.user.login !== OWNER.login || comment.user.id !== OWNER.id) continue;
    try {
      const record = JSON.parse(comment.body);
      if (record && record.schema === AUTHORIZATION_SCHEMA) records.push({comment, record});
    } catch {
      // Malformed owner comments are evidence, not authorization.
    }
  }
  if (records.length === 0) throw new Error('linear-tree reconciliation owner authorization is unavailable');
  if (records.length !== 1) throw new Error('linear-tree reconciliation owner authorization is ambiguous');
  try {
    validateOwnerAuthorization(records[0].record, {candidate, now});
  } catch (error) {
    error.authorization = records[0].record;
    throw error;
  }
  return records[0];
}

function validateRequiredChecks(checks, headSHA) {
  if (!Array.isArray(checks) || !validSHA(headSHA)) return false;
  const rows = checks.map((check) => ({name: check && (check.name || check.context), status: check && check.status, conclusion: check && check.conclusion, head_sha: check && (check.head_sha || check.headSHA), app_id: check && (check.app_id === undefined ? check.app?.id : check.app_id)}));
  if (rows.length !== REQUIRED_CHECKS.length) return false;
  const names = rows.map((row) => row.name);
  if (!exactArray([...names].sort(), [...REQUIRED_CHECKS].sort())) return false;
  return rows.every((row) => row.status === 'completed' && row.conclusion === 'success' && row.head_sha === headSHA && row.app_id === CI_APP_ID);
}

function evaluate(input, {now = new Date()} = {}) {
  validateProtocolDefinition();
  const cells = definition.cells.map((cell) => ({...cell, state: 'CLOSED'}));
  const refuted = [];
  const unknown = [];
  const addRefuted = (reason, cellId, detail = null) => refuted.push({reason, cellId, detail});
  const addUnknown = (evidence, cellId) => unknown.push({evidence, cellId});
  const candidate = input && input.candidate;
  if (!candidate) {
    addUnknown(unknownEvidence('FOUNDATION', 'CANDIDATE_TUPLE', 'INCOMPLETE_CANDIDATE_TUPLE', 'INCOMPLETE_EVIDENCE', 'PROVIDE_EXACT_CANDIDATE_TUPLE', ['candidate-tuple']), 'CANDIDATE_TUPLE');
  } else {
    try {
      validateCandidate(candidate);
    } catch (error) {
      addRefuted('CANDIDATE_TUPLE_MISMATCH', 'CANDIDATE_TUPLE', error.message);
    }
  }
  if (candidate && input?.route !== ROUTE) addRefuted('UNAUTHORIZED_ROUTE', 'ROUTE_AUTHORITY', 'ordinary agent-to-main route is not a reconciliation route');
  if (!input || !validSHA(input.live_main_sha)) addUnknown(unknownEvidence('COHERENCE', 'CURRENT_MAIN_REF', 'INCOMPLETE_CURRENT_MAIN_REF', 'INCOMPLETE_EVIDENCE', 'READ_CURRENT_MAIN_REF', ['main-ref']), 'CURRENT_MAIN_REF');
  else if (candidate && candidate.base_sha !== input.live_main_sha) addUnknown(unknownEvidence('COHERENCE', 'CURRENT_MAIN_REF', 'MAIN_REF_MOVED', 'LIVE_REF_MOVED', 'REBASE_OR_RECREATE_RECONCILIATION_SNAPSHOT', ['main-ref']), 'CURRENT_MAIN_REF');
  if (!input || !validSHA(input.live_dev_sha)) addUnknown(unknownEvidence('COHERENCE', 'CURRENT_DEV_REF', 'INCOMPLETE_CURRENT_DEV_REF', 'INCOMPLETE_EVIDENCE', 'READ_CURRENT_DEV_REF', ['dev-ref']), 'CURRENT_DEV_REF');
  else if (candidate && candidate.source_dev_sha !== input.live_dev_sha) addUnknown(unknownEvidence('COHERENCE', 'CURRENT_DEV_REF', 'DEV_REF_MOVED', 'LIVE_REF_MOVED', 'RECREATE_CANDIDATE_FROM_CURRENT_DEV', ['dev-ref']), 'CURRENT_DEV_REF');
  if (candidate && candidate.source_dev_tree && candidate.candidate_tree && (candidate.source_dev_tree.tree_digest !== candidate.candidate_tree.tree_digest || candidate.source_dev_tree.manifest_digest !== candidate.candidate_tree.manifest_digest || !exactArray(candidate.source_dev_tree.paths, candidate.candidate_tree.paths))) {
    addRefuted('TREE_MISMATCH', 'TREE_EQUIVALENCE', 'candidate tree and pinned source-dev tree are not exactly equivalent');
  }
  if (candidate && input?.observed_candidate && !exactObject(candidate, input.observed_candidate)) addRefuted('CANDIDATE_DRIFT', 'CANDIDATE_TUPLE', 'observed candidate differs from the authorized candidate tuple');
  if (candidate && input?.required_checks !== undefined && !validateRequiredChecks(input.required_checks, candidate.head_sha)) addUnknown(unknownEvidence('COHERENCE', 'REQUIRED_CHECKS', 'INCOMPLETE_REQUIRED_CHECKS', 'INCOMPLETE_EVIDENCE', 'WAIT_FOR_EXACT_7_OF_7_REQUIRED_CHECKS', ['required-checks']), 'REQUIRED_CHECKS');
  if (candidate && input?.workflow && (!candidate.workflow || !exactObject(input.workflow, candidate.workflow))) addRefuted('WORKFLOW_IDENTITY_MISMATCH', 'WORKFLOW_IDENTITY', 'workflow identity differs from candidate authorization');
  if (!input || !input.authorization) addUnknown(unknownEvidence('REGRESSION', 'ONE_USE_AUTHORIZATION', 'INCOMPLETE_OWNER_AUTHORIZATION', 'INCOMPLETE_EVIDENCE', 'POST_ONE_FRESH_OWNER_AUTHORIZATION', ['owner-authorization']), 'ONE_USE_AUTHORIZATION');
  else {
    const authorization = input.authorization;
    if (authorization.reuse_attempts > 0 || authorization.use_count > 1) addRefuted('NONCE_REPLAY', 'ONE_USE_AUTHORIZATION', 'one-use authorization replay was observed');
    if (authorization.stale === true || authorization.revoked === true) addRefuted('AUTHORIZATION_STALE', 'ONE_USE_AUTHORIZATION', 'authorization is stale or revoked');
    if (candidate && authorization.candidate_digest && authorization.candidate_digest !== candidateDigest(candidate)) addRefuted('AUTHORIZATION_CANDIDATE_MISMATCH', 'ONE_USE_AUTHORIZATION', 'authorization is bound to another candidate');
    if (!['AUTHORIZED', 'CONSUMED'].includes(authorization.state) || authorization.use_count === undefined) addUnknown(unknownEvidence('REGRESSION', 'ONE_USE_AUTHORIZATION', 'INCOMPLETE_OWNER_AUTHORIZATION_STATE', 'INCOMPLETE_EVIDENCE', 'PROVIDE_EXACT_ONE_USE_AUTHORIZATION_STATE', ['owner-authorization']), 'ONE_USE_AUTHORIZATION');
  }
  const decision = refuted.length > 0 ? 'REFUTED' : unknown.length > 0 ? 'UNKNOWN' : 'CLOSED';
  for (const item of refuted) {
    const cell = cells.find((candidateCell) => candidateCell.id === item.cellId);
    if (cell) cell.state = 'REFUTED';
  }
  if (decision === 'UNKNOWN' && unknown[0]) {
    const cell = cells.find((candidateCell) => candidateCell.id === unknown[0].cellId);
    if (cell) cell.state = 'UNKNOWN';
  }
  return {
    schema: SCHEMA,
    decision,
    reason: decision === 'REFUTED' ? refuted[0].reason : decision === 'UNKNOWN' ? unknown[0].evidence.reason : 'EXACT_LINEAR_TREE_EQUIVALENT_RECONCILIATION_CLOSED',
    precedence: DECISION_PRECEDENCE.join('>'),
    cells,
    unknown: decision === 'UNKNOWN' ? unknown[0].evidence : null,
    refuted: decision === 'REFUTED' ? refuted[0] : null,
    repository_writes: 0,
    local_test_executions: 0,
  };
}

async function readTreeBinding({getCommit, getTree, owner, repo, ref}) {
  const commitResponse = await getCommit({owner, repo, ref});
  const treeSHA = commitResponse?.data?.commit?.tree?.sha;
  if (!validSHA(treeSHA)) throw new Error(`tree commit is unavailable for ${ref}`);
  const treeResponse = await getTree({owner, repo, tree_sha: treeSHA, recursive: '1'});
  if (!treeResponse?.data || treeResponse.data.truncated === true || !Array.isArray(treeResponse.data.tree)) throw new Error(`tree manifest is unavailable for ${ref}`);
  const binding = treeManifest(treeResponse.data.tree);
  return {...binding, tree_sha: treeSHA};
}

function protocolReceipt({candidate, authorization, ownerComment, evaluation, sourceDevTree, candidateTree, runtime}) {
  return {
    schema: SCHEMA,
    decision: 'PASS',
    protocol_decision: evaluation.decision,
    candidate,
    candidate_digest: candidateDigest(candidate),
    authorization: {...authorization, owner_comment_id: ownerComment?.id || null, owner_comment_url: ownerComment?.html_url || null},
    source_dev_tree: sourceDevTree,
    candidate_tree: candidateTree,
    runtime,
    evaluation,
    no_protection_mutation: true,
    no_check_bypass: true,
    no_force_push: true,
    repository_writes: 0,
  };
}

async function evaluatePromotion({pull, repository, workflowRef, workflowSha, runtimeRef, runtimeSha, runId, runAttempt, liveBefore, liveAfter, changedFiles, listComments, compareCommits, getCommit, getTree, now = new Date()}) {
  try {
    if (repository !== REPOSITORY || pull?.base?.ref !== 'main' || !isReconciliationBranch(pull?.head?.ref) || pull.base.repo.full_name !== repository || pull.head.repo.full_name !== repository || pull.state !== 'open' || pull.draft === true || pull.merged === true || pull.merged_at !== null || workflowRef !== `${repository}/.github/workflows/ci-guardian.yml@refs/heads/dev` || runtimeRef !== 'refs/heads/dev' || runtimeSha !== workflowSha || !validSHA(workflowSha) || !Number.isInteger(runId) || runId < 1 || !Number.isInteger(runAttempt) || runAttempt < 1 || !liveBefore?.refs || !liveAfter?.refs || liveBefore.refs.main_sha !== pull.base.sha || liveAfter.refs.main_sha !== pull.base.sha || liveBefore.refs.dev_sha !== liveAfter.refs.dev_sha || liveBefore.refs.dev_sha !== liveAfter.refs.dev_sha) throw new Error('linear-tree reconciliation identity or live references are not exact');
    const compareResponse = await compareCommits({owner: repository.split('/')[0], repo: repository.split('/')[1], base: pull.base.sha, head: pull.head.sha});
    const mergeBaseSHA = compareResponse?.data?.merge_base_commit?.sha;
    if (!validSHA(mergeBaseSHA)) throw new Error('linear-tree reconciliation candidate merge base is unavailable');
    const sourceDevTree = await readTreeBinding({getCommit, getTree, owner: repository.split('/')[0], repo: repository.split('/')[1], ref: liveBefore.refs.dev_sha});
    const candidateTree = await readTreeBinding({getCommit, getTree, owner: repository.split('/')[0], repo: repository.split('/')[1], ref: pull.head.sha});
    const runtime = {workflow_ref: workflowRef, workflow_sha: workflowSha, runtime_ref: runtimeRef, runtime_sha: runtimeSha, check_name: 'CI guardian', app_id: CI_APP_ID};
    const candidate = candidateFromPull({pull, changedFiles, sourceDevSHA: liveBefore.refs.dev_sha, mergeBaseSHA, sourceDevTree, candidateTree, workflow: runtime});
    let ownerAuthorization;
    let authorization;
    try {
      ownerAuthorization = await readOwnerAuthorization({listComments, owner: repository.split('/')[0], repo: repository.split('/')[1], pullNumber: pull.number, candidate, now});
      authorization = validateOwnerAuthorization(ownerAuthorization.record, {candidate, repository, now});
    } catch (error) {
      const evaluation = evaluate({route: ROUTE, candidate, live_main_sha: liveBefore.refs.main_sha, live_dev_sha: liveBefore.refs.dev_sha, authorization: error.authorization || null, workflow: runtime});
      return {schema: SCHEMA, decision: 'FAIL_CLOSED', code: RECONCILIATION_CODE, reason: evaluation.reason, protocol_decision: evaluation.decision, candidate, candidate_digest: candidateDigest(candidate), evaluation, repository_writes: 0};
    }
    const evaluation = evaluate({route: ROUTE, candidate, live_main_sha: liveBefore.refs.main_sha, live_dev_sha: liveBefore.refs.dev_sha, authorization, workflow: runtime});
    if (evaluation.decision !== 'CLOSED') return {schema: SCHEMA, decision: 'FAIL_CLOSED', code: RECONCILIATION_CODE, reason: evaluation.reason, protocol_decision: evaluation.decision, candidate, candidate_digest: candidateDigest(candidate), evaluation, repository_writes: 0};
    return protocolReceipt({candidate, authorization, ownerComment: ownerAuthorization.comment, evaluation, sourceDevTree, candidateTree, runtime});
  } catch (error) {
    return {schema: SCHEMA, decision: 'FAIL_CLOSED', code: RECONCILIATION_CODE, reason: error?.message || String(error), protocol_decision: 'REFUTED', repository_writes: 0, no_protection_mutation: true, no_check_bypass: true, no_force_push: true};
  }
}

function finalizePromotionReceipt(receipt, {requiredChecks, currentDevSHA, currentMainSHA, now = new Date()} = {}) {
  const baseResult = {schema: SCHEMA, precedence: DECISION_PRECEDENCE.join('>'), cells: [], checked_at: now.toISOString()};
  const refuted = (reason, detail = null) => ({...baseResult, decision: 'REFUTED', reason, unknown: null, refuted: {reason, detail}});
  const unknown = (step, reason, nextOperation, blockedBy) => {
    const evidence = unknownEvidence('COHERENCE', step, reason, 'INCOMPLETE_EVIDENCE', nextOperation, blockedBy);
    return {...baseResult, decision: 'UNKNOWN', reason: evidence.reason, unknown: evidence, refuted: null};
  };
  if (!receipt || receipt.schema !== SCHEMA || receipt.decision !== 'PASS' || receipt.protocol_decision !== 'CLOSED' || !receipt.candidate || receipt.candidate_digest !== candidateDigest(receipt.candidate)) {
    return refuted('MALFORMED_RECONCILIATION_RECEIPT');
  }
  try {
    validateCandidate(receipt.candidate);
  } catch (error) {
    return refuted('CANDIDATE_TUPLE_MISMATCH', error.message || String(error));
  }
  if (!receipt.authorization || receipt.authorization.use_count !== 0 || receipt.authorization.reuse_attempts !== 0) {
    return refuted('NONCE_REPLAY');
  }
  const sourceTree = receipt.source_dev_tree || receipt.candidate.source_dev_tree;
  const candidateTree = receipt.candidate_tree || receipt.candidate.candidate_tree;
  if (!validTreeBinding(sourceTree) || !validTreeBinding(candidateTree) || sourceTree.tree_digest !== candidateTree.tree_digest || sourceTree.manifest_digest !== candidateTree.manifest_digest || !exactArray(sourceTree.paths, candidateTree.paths)) {
    return refuted('TREE_MISMATCH', 'candidate tree and pinned source-dev tree are not exactly equivalent');
  }
  if (!validSHA(currentDevSHA) || !validSHA(currentMainSHA)) return unknown('CURRENT_REFS', 'INCOMPLETE_CURRENT_REFS', 'READ_CURRENT_MAIN_AND_DEV_REFS', ['main-ref', 'dev-ref']);
  if (receipt.candidate.source_dev_sha !== currentDevSHA) return unknown('CURRENT_DEV_REF', 'DEV_REF_MOVED', 'RETRY_WITH_CURRENT_EXACT_HEAD', ['dev-ref']);
  if (receipt.candidate.base_sha !== currentMainSHA) return unknown('CURRENT_MAIN_REF', 'MAIN_REF_MOVED', 'RETRY_WITH_CURRENT_EXACT_HEAD', ['main-ref']);
  if (!validateRequiredChecks(requiredChecks, receipt.candidate.head_sha)) return unknown('REQUIRED_CHECKS', 'INCOMPLETE_REQUIRED_CHECKS', 'WAIT_FOR_EXACT_7_OF_7_REQUIRED_CHECKS', ['required-checks']);
  return {schema: SCHEMA, decision: 'CLOSED', reason: 'EXACT_7_OF_7_REQUIRED_CHECKS_AND_TREE_EQUIVALENCE', precedence: DECISION_PRECEDENCE.join('>'), unknown: null, refuted: null, required_checks: REQUIRED_CHECKS, checked_at: now.toISOString()};
}

function replayReceipt(receipt) {
  if (!receipt || !receipt.authorization || receipt.authorization.use_count !== 0 || receipt.authorization.reuse_attempts !== 0) return {decision: 'REFUTED', reason: 'NONCE_REPLAY'};
  return {decision: 'REFUTED', reason: 'ONE_USE_AUTHORIZATION_REPLAY_ATTEMPT'};
}

validateProtocolDefinition();

module.exports = {
  AUTHORIZATION_BINDING,
  AUTHORIZATION_SCHEMA,
  BRANCH_PREFIX,
  CI_APP_ID,
  DECISION_PRECEDENCE,
  OWNER,
  RECONCILIATION_CODE,
  REQUIRED_CHECKS,
  ROUTE,
  SCHEMA,
  UNKNOWN_FIELDS,
  candidateDigest,
  candidateFromPull,
  authorizationCandidateBinding,
  canonicalPathNames,
  canonicalPathList,
  canonicalProtectedIntersection,
  canonicalTreeEntries,
  digestPathList,
  evaluate,
  evaluatePromotion,
  finalizePromotionReceipt,
  isReconciliationBranch,
  readOwnerAuthorization,
  readTreeBinding,
  replayReceipt,
  sha256,
  treeManifest,
  validateOwnerAuthorization,
  validateProtocolDefinition,
  validateRequiredChecks,
  validDigest,
  validSHA,
};
