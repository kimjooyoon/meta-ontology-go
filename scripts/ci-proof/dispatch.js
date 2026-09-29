'use strict';

async function resolvePromotionPull(github, context) {
  if (context.eventName !== 'workflow_dispatch') return context.payload.pull_request || null;

  const inputs = context.payload.inputs || {};
  const numberText = String(inputs.promotion_pr_number || '');
  const headSHA = String(inputs.promotion_head_sha || '');
  const baseSHA = String(inputs.promotion_base_sha || '');
  const headRef = context.ref.startsWith('refs/heads/') ? context.ref.slice('refs/heads/'.length) : '';
  if (!/^[1-9][0-9]*$/.test(numberText) || !/^[0-9a-f]{40}$/.test(headSHA) ||
      !/^[0-9a-f]{40}$/.test(baseSHA) || !headRef || headSHA !== context.sha) {
    throw new Error('promotion dispatch inputs are malformed or do not bind the exact workflow ref');
  }

  const number = Number(numberText);
  if (!Number.isSafeInteger(number) || number <= 0) throw new Error('promotion PR number is invalid');
  const repository = context.payload.repository.full_name;
  const pull = (await github.rest.pulls.get({
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: number,
  })).data;
  if (!pull || pull.number !== number || pull.state !== 'open' || pull.draft || pull.merged ||
      pull.base?.ref !== 'main' || pull.base?.sha !== baseSHA || pull.head?.ref !== headRef ||
      pull.head?.sha !== headSHA || pull.base?.repo?.full_name !== repository ||
      pull.head?.repo?.full_name !== repository || context.ref !== `refs/heads/${pull.head.ref}`) {
    throw new Error('promotion dispatch no longer matches the exact open same-repository main PR');
  }
  return pull;
}

module.exports = {resolvePromotionPull};
