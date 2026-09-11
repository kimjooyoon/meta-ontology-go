'use strict';

const protocol = require('./foundation_authorization_protocol');

const SUCCESSOR_SCHEMA = 'gooo/ci-guardian-successor-protocol/v1';
const SUCCESSOR_CODE = 'CI-GUARDIAN-SUCCESSOR-PROTOCOL-001';
const AUTHORIZATION_SCHEMA = 'gooo/foundation-authorization-owner-record/v1';
const AUTHORIZATION_INTENT = 'AUTHORIZE_NORMAL_ONE_USE_PROTOCOL_FOR_EXACT_CANDIDATE';
const AUTHORIZATION_OPERATION = 'OWNER_CREATE_ONE_USE_AUTHORIZATION_FOR_EXACT_CANDIDATE';
const AUTHORIZATION_REASON = 'NORMAL_ONE_USE_PROTOCOL_ACTIVE';
const OWNER = Object.freeze({login: 'kimjooyoon', id: 115961382, type: 'User'});

function exactObject(left, right) {
  return JSON.stringify(left) === JSON.stringify(right);
}

function validSHA(value) {
  return protocol.validSHA(value);
}

function validTimestamp(value) {
  return typeof value === 'string' && Number.isFinite(Date.parse(value));
}

function fail(reason, protocolDecision = 'REFUTED') {
  return {
    schema: SUCCESSOR_SCHEMA,
    decision: 'FAIL_CLOSED',
    code: SUCCESSOR_CODE,
    reason,
    protocol_decision: protocolDecision,
    repository_writes: 0,
    no_protection_mutation: true,
    no_check_bypass: true,
    no_force_push: true,
  };
}

function candidateFromPull({pull, changedFiles, mergeBaseSha}) {
  const paths = protocol.canonicalPathNames(changedFiles);
  const protectedPaths = protocol.canonicalProtectedIntersection(paths);
  return {
    pull_request: pull.number,
    base_repo: pull.base.repo.full_name,
    base_branch: pull.base.ref,
    base_sha: pull.base.sha,
    head_repo: pull.head.repo.full_name,
    head_branch: pull.head.ref,
    head_sha: pull.head.sha,
    merge_base_sha: mergeBaseSha,
    changed_paths: {paths, count: paths.length, digest: protocol.digestPathList(paths)},
    protected_intersection: {paths: protectedPaths, count: protectedPaths.length, digest: protocol.digestPathList(protectedPaths)},
  };
}

function validateCandidate(candidate, pull, mergeBaseSha, changedFiles) {
  if (!candidate || !pull || candidate.pull_request !== pull.number
    || candidate.base_repo !== pull.base.repo.full_name || candidate.base_branch !== pull.base.ref
    || candidate.base_sha !== pull.base.sha || candidate.head_repo !== pull.head.repo.full_name
    || candidate.head_branch !== pull.head.ref || candidate.head_sha !== pull.head.sha
    || candidate.merge_base_sha !== mergeBaseSha || !validSHA(candidate.merge_base_sha)) {
    throw new Error('successor protocol candidate tuple is not exact');
  }
  const expected = candidateFromPull({pull, changedFiles, mergeBaseSha});
  if (!exactObject(candidate, expected) || candidate.digest !== undefined) {
    throw new Error('successor protocol candidate path binding is not exact');
  }
  if (candidate.digest !== undefined || protocol.candidateDigest(candidate) !== protocol.candidateDigest(expected)) {
    throw new Error('successor protocol candidate digest is not exact');
  }
  return candidate;
}

function validateOwnerRecord(record, {candidate, repository = protocol.REPOSITORY, now = new Date()} = {}) {
  if (!record || record.schema !== AUTHORIZATION_SCHEMA || record.authorization_route !== 'NORMAL_ONE_USE_PROTOCOL'
    || record.state !== 'AUTHORIZED' || record.proof_choice !== 'FOUNDATION' || record.intent !== AUTHORIZATION_INTENT
    || record.required_operation !== AUTHORIZATION_OPERATION || record.reason !== AUTHORIZATION_REASON
    || record.reusable !== false || record.one_use !== true || typeof record.nonce !== 'string' || record.nonce.length < 16
    || !validTimestamp(record.issued_at) || !validTimestamp(record.expires_at) || Date.parse(now) >= Date.parse(record.expires_at)
    || record.use_count !== 0 || record.reuse_attempts !== 0 || record.stale !== false || record.revoked !== false
    || !exactObject(record.owner_selection, OWNER) || !exactObject(record.actor, OWNER)
    || !exactObject(record.candidate, candidate) || record.candidate_digest !== protocol.candidateDigest(candidate)
    || record.current_dev_tip_at_authorization !== candidate.head_sha || record.base_snapshot_is_current_dev_tip !== false
    || record.root_axiom_receipt_schema !== null || record.root_axiom_receipt_digest !== null || record.external_root_reuse !== false
    || !record.authorization_receipt || record.authorization_receipt.nonce !== record.nonce
    || record.authorization_receipt.candidate_digest !== record.candidate_digest
    || !exactObject(record.authorization_receipt.owner, OWNER)
    || record.predecessor_root?.reused !== false || record.no_protection_mutation !== true
    || record.no_check_bypass !== true || record.no_force_push !== true || record.repository_writes_before_authorization !== 0
    || repository !== protocol.REPOSITORY) {
    throw new Error('successor protocol owner authorization is missing, stale, or not exact');
  }
  return {
    state: record.state,
    owner_selection: record.owner_selection,
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
  if (typeof listComments !== 'function') throw new Error('successor protocol owner comment reader is unavailable');
  const comments = [];
  for (let page = 1; page <= 1000; page++) {
    const response = await listComments({owner, repo, issue_number: pullNumber, per_page: 100, page});
    if (!response || (response.status !== undefined && response.status !== 200) || !Array.isArray(response.data)) {
      throw new Error('successor protocol owner comment API returned malformed data');
    }
    comments.push(...response.data);
    if (response.data.length < 100) break;
    if (page === 1000) throw new Error('successor protocol owner comment pagination exceeded the limit');
  }
  const matches = [];
  for (const comment of comments) {
    if (!comment || !comment.user || comment.user.login !== OWNER.login || comment.user.id !== OWNER.id) continue;
    let record;
    try {
      record = JSON.parse(comment.body);
    } catch {
      continue;
    }
    try {
      validateOwnerRecord(record, {candidate, now});
      matches.push({comment, record});
    } catch {
      // Non-matching and stale owner records are evidence, not authorization.
    }
  }
  if (matches.length !== 1) throw new Error(matches.length === 0
    ? 'successor protocol exact owner authorization is unavailable'
    : 'successor protocol exact owner authorization is ambiguous');
  return matches[0];
}

async function readKernelDigest({getCommit, getTree, owner, repo, ref}) {
  const commitResponse = await getCommit({owner, repo, ref});
  const commit = commitResponse && commitResponse.data;
  const treeSHA = commit && commit.commit && commit.commit.tree && commit.commit.tree.sha;
  if (!validSHA(treeSHA)) throw new Error(`successor protocol kernel commit tree is unavailable for ${ref}`);
  const treeResponse = await getTree({owner, repo, tree_sha: treeSHA, recursive: '1'});
  const tree = treeResponse && treeResponse.data;
  if (!tree || tree.truncated === true || !Array.isArray(tree.tree)) throw new Error(`successor protocol kernel tree is unavailable for ${ref}`);
  return protocol.digestKernelEntries(tree.tree);
}

function validatePromotionIdentity({pull, repository, workflowRef, workflowSha, runtimeRef, runtimeSha, runId, runAttempt, liveBefore, liveAfter}) {
  if (!pull || pull.base?.ref !== 'main' || pull.head?.ref !== 'dev' || pull.base?.repo?.full_name !== repository
    || pull.head?.repo?.full_name !== repository || !validSHA(pull.base.sha) || !validSHA(pull.head.sha)
    || workflowRef !== `${repository}/.github/workflows/ci-guardian.yml@refs/heads/dev`
    || workflowSha !== runtimeSha || runtimeRef !== 'refs/heads/dev' || !validSHA(workflowSha)
    || !Number.isInteger(runId) || runId < 1 || !Number.isInteger(runAttempt) || runAttempt < 1
    || !liveBefore || !liveAfter || !liveBefore.refs || !liveAfter.refs || liveBefore.refs.main_sha !== pull.base.sha
    || liveBefore.refs.dev_sha !== pull.head.sha || liveAfter.refs.main_sha !== pull.base.sha
    || liveAfter.refs.dev_sha !== pull.head.sha || liveBefore.refs.main_sha !== liveAfter.refs.main_sha
    || liveBefore.refs.dev_sha !== liveAfter.refs.dev_sha || liveBefore.topology?.status !== 'ahead'
    || liveAfter.topology?.status !== 'ahead' || liveBefore.topology?.behind_by !== 0 || liveAfter.topology?.behind_by !== 0
    || liveBefore.topology?.merge_base_sha !== pull.base.sha || liveAfter.topology?.merge_base_sha !== pull.base.sha) {
    throw new Error('successor protocol promotion identity or live topology is not exact');
  }
  return pull.base.sha;
}

async function evaluatePromotion({pull, repository, workflowRef, workflowSha, runtimeRef, runtimeSha, runId, runAttempt, liveBefore, liveAfter, changedFiles, listComments, getCommit, getTree, now = new Date()}) {
  try {
    const mergeBaseSha = validatePromotionIdentity({pull, repository, workflowRef, workflowSha, runtimeRef, runtimeSha, runId, runAttempt, liveBefore, liveAfter});
    const candidate = candidateFromPull({pull, changedFiles, mergeBaseSha});
    const ownerAuthorization = await readOwnerAuthorization({listComments, owner: repository.split('/')[0], repo: repository.split('/')[1], pullNumber: pull.number, candidate, now});
    const authorization = validateOwnerRecord(ownerAuthorization.record, {candidate, repository, now});
    const kernel = {
      before_sha256: await readKernelDigest({getCommit, getTree, owner: repository.split('/')[0], repo: repository.split('/')[1], ref: pull.base.sha}),
      after_sha256: await readKernelDigest({getCommit, getTree, owner: repository.split('/')[0], repo: repository.split('/')[1], ref: pull.head.sha}),
    };
    const runtime = {
      workflow_ref: workflowRef,
      workflow_sha: workflowSha,
      runtime_ref: runtimeRef,
      runtime_sha: runtimeSha,
      check_name: 'CI guardian',
      check_run_id: runId,
      app_id: protocol.CI_APP_ID,
    };
    const input = {
      repository: {full_name: repository, owner_login: OWNER.login, owner_id: OWNER.id, owner_type: 'User'},
      actor: {...OWNER, owner_match: true},
      intent: AUTHORIZATION_INTENT,
      proof_choice: 'FOUNDATION',
      candidate,
      runtime,
      kernel,
      authorization,
      consumption: null,
      observed: {candidate: JSON.parse(JSON.stringify(candidate)), runtime: JSON.parse(JSON.stringify(runtime)), kernel: JSON.parse(JSON.stringify(kernel))},
    };
    const evaluation = protocol.evaluate(input, {now: now.toISOString()});
    if (evaluation.decision !== 'CLOSED') return fail(`successor protocol decision ${evaluation.decision}: ${evaluation.reason}`, evaluation.decision);
    return {
      schema: SUCCESSOR_SCHEMA,
      decision: 'PASS',
      reason: 'EXACT_SUCCESSOR_PROTOCOL',
      protocol_decision: evaluation.decision,
      candidate,
      candidate_digest: protocol.candidateDigest(candidate),
      authorization: {
        state: authorization.state,
        owner_selection: authorization.owner_selection,
        nonce: authorization.nonce,
        issued_at: authorization.issued_at,
        expires_at: authorization.expires_at,
        stale: authorization.stale,
        use_count: authorization.use_count,
        reuse_attempts: authorization.reuse_attempts,
        revoked: authorization.revoked,
        authorization_receipt: authorization.authorization_receipt,
        owner_comment_id: ownerAuthorization.comment.id,
        owner_comment_url: ownerAuthorization.comment.html_url,
      },
      runtime,
      kernel,
      evaluation: {decision: evaluation.decision, precedence: evaluation.precedence, cells: evaluation.cells, unknown: evaluation.unknown, refuted: evaluation.refuted},
      no_protection_mutation: true,
      no_check_bypass: true,
      no_force_push: true,
      repository_writes: 0,
    };
  } catch (error) {
    return fail(error && error.message ? error.message : String(error), 'REFUTED');
  }
}

function validateSuccessorProtocolReceipt(receipt, {pull, repository, workflowRef, workflowSha, runtimeRef, runtimeSha, runId, runAttempt, liveBefore, liveAfter, changedFiles, kernelBeforeDigest, kernelAfterDigest, now = new Date()}) {
  if (!receipt || receipt.schema !== SUCCESSOR_SCHEMA || receipt.decision !== 'PASS' || receipt.reason !== 'EXACT_SUCCESSOR_PROTOCOL'
    || receipt.protocol_decision !== 'CLOSED' || receipt.candidate_digest !== protocol.candidateDigest(receipt.candidate)
    || receipt.no_protection_mutation !== true || receipt.no_check_bypass !== true || receipt.no_force_push !== true || receipt.repository_writes !== 0) {
    throw new Error('successor protocol receipt is not an exact PASS');
  }
  const mergeBaseSha = validatePromotionIdentity({pull, repository, workflowRef, workflowSha, runtimeRef, runtimeSha, runId, runAttempt, liveBefore, liveAfter});
  const expectedCandidate = candidateFromPull({pull, changedFiles, mergeBaseSha});
  if (!exactObject(receipt.candidate, expectedCandidate)) throw new Error('successor protocol receipt candidate drifted');
  const expectedRuntime = {workflow_ref: workflowRef, workflow_sha: workflowSha, runtime_ref: runtimeRef, runtime_sha: runtimeSha, check_name: 'CI guardian', check_run_id: runId, app_id: protocol.CI_APP_ID};
  if (!exactObject(receipt.runtime, expectedRuntime)) throw new Error('successor protocol receipt runtime drifted');
  if (!receipt.kernel || receipt.kernel.before_sha256 !== kernelBeforeDigest || receipt.kernel.after_sha256 !== kernelAfterDigest) throw new Error('successor protocol receipt kernel digest drifted');
  const authorization = receipt.authorization;
  if (!authorization || authorization.state !== 'AUTHORIZED' || authorization.use_count !== 0 || authorization.reuse_attempts !== 0 || authorization.stale !== false || authorization.revoked !== false || typeof authorization.nonce !== 'string' || !validTimestamp(authorization.issued_at) || !validTimestamp(authorization.expires_at) || Date.parse(now) >= Date.parse(authorization.expires_at) || !exactObject(authorization.owner_selection, OWNER) || !authorization.authorization_receipt || authorization.authorization_receipt.nonce !== authorization.nonce || authorization.authorization_receipt.candidate_digest !== receipt.candidate_digest || !exactObject(authorization.authorization_receipt.owner, OWNER)) {
    throw new Error('successor protocol receipt one-use authorization is not exact');
  }
  if (!receipt.evaluation || receipt.evaluation.decision !== 'CLOSED' || !Array.isArray(receipt.evaluation.cells) || receipt.evaluation.cells.length !== 12 || receipt.evaluation.unknown !== null || receipt.evaluation.refuted !== null) throw new Error('successor protocol receipt evaluation is not exact');
  return receipt;
}

module.exports = {
  AUTHORIZATION_INTENT,
  AUTHORIZATION_OPERATION,
  AUTHORIZATION_REASON,
  AUTHORIZATION_SCHEMA,
  OWNER,
  SUCCESSOR_CODE,
  SUCCESSOR_SCHEMA,
  candidateFromPull,
  evaluatePromotion,
  readKernelDigest,
  validateOwnerRecord,
  validatePromotionIdentity,
  validateSuccessorProtocolReceipt,
};
