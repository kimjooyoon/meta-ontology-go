# Proposal synthesis and continuity

## Separation

Metric strategy synthesis proves the current subject. Proposal continuity proves
that a prior synthesized subject can be selected. A missing predecessor must not
prevent the current proposal contract from being emitted, and it must not be
reported as a fixed point.

The `strategy` job therefore emits `metric-strategy-<subject-sha>` independently.
The `proposal-continuity` job may fail closed without deleting that immutable
artifact. Repository mutation and promotion authority remain false in both jobs.

## Job-level resolution

Selection schema `gooo/autonomous-change-proposal-predecessor-selection/v2`
accepts a completed push workflow only when exactly one `strategy` job completed
successfully and exactly one matching artifact contains an 8/8 proposal contract.
The containing workflow may have failed in the separate continuity job.

This is a lower observation resolution, not a weaker decision:

| Evidence | Required value |
| --- | ---: |
| Exact synthesis jobs | 1 |
| Valid proposal artifacts | 1 |
| Contract coordinates | 8/8 |
| Selection proofs | 5/5 |
| Unresolved candidates | 0 |
| Repository writes | 0 |
| Promotion authority | 0 |

## Use case

If an earlier merge lacks a proposal artifact, current synthesis still publishes
an exact read-only contract. Continuity remains `FAIL_CLOSED`. The next merged
subject can select that contract only through the successful synthesis job ID,
artifact checksum, report digest, commit SHA, and zero-write guardrails.

This repair does not change the fixed language-readiness numerator. It creates
the exact predecessor evidence required by a later quantified transition.
