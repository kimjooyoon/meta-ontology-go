# Automatic promotion App

The executor uses a GitHub App only to open or edit a promotion PR. Its token
is restricted to `kimjooyoon/meta-ontology-go` and `pull_requests:write` (plus
GitHub's implicit metadata read permission). It receives no contents, Actions,
checks, administration, organization, or other repository grants. The workflow's
ordinary `GITHUB_TOKEN` reads runs/artifacts/checks, creates the candidate ref,
and performs the normal protected PR merge.

This solves two observed GitHub constraints: Actions checks produced by
`workflow_dispatch` are not eligible PR required checks, and PR events created
by `GITHUB_TOKEN` can require workflow execution approval. An App-created
`opened` event or App-generated body `edited` event starts authoritative PR CI.
See the official [required-check rules](https://docs.github.com/en/pull-requests/how-tos/merge-and-close-pull-requests/troubleshooting-required-status-checks)
and [token event rules](https://docs.github.com/en/actions/concepts/security/github_token).

## Configuration contract

The App is owned by the repository owner, installed on this one repository,
and requests only the permissions above. Webhooks and user authorization are
unneeded for this installation-token workflow. Repository configuration is:

| Key | Location | Meaning |
| --- | --- | --- |
| `PROMOTION_APP_CLIENT_ID` | Actions repository variable | The configured App's client ID |
| `PROMOTION_APP_PRIVATE_KEY` | Actions repository secret | The App's PEM private key |

The key must never enter Git, public logs, model data, or an artifact. The pinned
`actions/create-github-app-token` step requests a token for this repository and
only `permission-pull-requests: write`. The helper permits PR creation or a
single PR body update; other token uses are rejected locally. The action revokes
its generated token at the end of the job.

Missing either configuration value stops preparation with the code
`PROMOTION_APP_CONFIGURATION_REQUIRED` and an action description. A failed App
request stops the workflow with `PROMOTION_APP_PR_API_FAILED`. No alternate
token, direct protected-branch write, check replacement, or review requirement
is introduced.

## Evidence and limits

The parent run waits for a fresh exact `pull_request` workflow/head/branch/run
identity, then verifies successful current-attempt jobs and GitHub Actions
check-suite identities. Because GitHub can return an empty `pull_requests`
array on a genuine PR run, the downloaded proof must independently bind the PR
number, `refs/pull/N/merge`, source/target refs, head/base SHAs, and digest-bound
promotion authorization before any merge. Live dev/main/tree/parent/topology
checks remain mandatory.

The body refresh adds `edited` to CI's existing PR activity types, so other PR
title/body edits also run CI. The refresh marker changes once per source-run
identity and attempt. The protocol has mocked regression coverage; an installed
App and a real automatically merged promotion are required to demonstrate
operational autonomy. Publishing this source alone does not demonstrate it.
