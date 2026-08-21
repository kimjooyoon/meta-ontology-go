'use strict';

const DEFAULT_PAGE_SIZE = 100;
const DEFAULT_PAGE_LIMIT = 1000;

async function listWorkflowArtifacts(fetchPage, options = {}) {
  const pageSize = options.pageSize || DEFAULT_PAGE_SIZE;
  const pageLimit = options.pageLimit || DEFAULT_PAGE_LIMIT;
  const artifacts = [];
  for (let page = 1; page <= pageLimit; page += 1) {
    let response;
    try {
      response = await fetchPage(page);
    } catch (error) {
      throw new Error(`workflow artifact API read failed: ${error.message}`);
    }
    const rows = response && response.data && response.data.artifacts;
    if (!Array.isArray(rows)) {
      throw new Error('workflow artifact API response is missing an artifacts array');
    }
    artifacts.push(...rows);
    if (rows.length < pageSize) {
      return artifacts;
    }
  }
  throw new Error('workflow artifact API pagination exceeded the fail-closed page limit');
}

function selectUniqueArtifact(artifacts, name) {
  const matches = artifacts.filter(artifact => artifact && artifact.name === name);
  if (matches.length > 1) {
    throw new Error(`workflow artifact inventory contains duplicate ${name} artifacts`);
  }
  return matches[0] || null;
}

function selectCurrentEvidenceArtifact(artifacts, runId, runAttempt) {
  const name = `ci-evidence-${runId}-${runAttempt}`;
  const artifact = selectUniqueArtifact(artifacts, name);
  if (!artifact) {
    throw new Error(`workflow artifact inventory is missing ${name}`);
  }
  if (artifact.id <= 0 || artifact.size_in_bytes <= 0 || artifact.expired || !/^sha256:[0-9a-f]{64}$/.test(artifact.digest || '')) {
    throw new Error(`workflow artifact ${name} is zero-sized, expired, or has an invalid digest`);
  }
  return {
    id: artifact.id,
    name: artifact.name,
    size_bytes: artifact.size_in_bytes,
    expired: artifact.expired,
    digest: artifact.digest,
    run_id: runId,
    run_attempt: runAttempt,
  };
}

module.exports = {listWorkflowArtifacts, selectCurrentEvidenceArtifact, selectUniqueArtifact};
