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
accepts a completed push workflow only when the explicitly requested branch
route matches the run's `head_branch`, exactly one `strategy` job completed
successfully, and exactly one matching artifact contains an 8/8 proposal
contract. A same-SHA run on another route is a separate observation and is not
a candidate for the requested route. The containing workflow may have failed
in the separate continuity job.

Route identity is part of the evidence tuple: repository, predecessor SHA,
requested route, run ID/attempt, run head branch, workflow event/status/
conclusion, synthesis job identity, and artifact identity. Missing route data
or multiple exact candidates remain `UNKNOWN` with the six-field
`stage`/`step`/`reason`/`unknown_class`/`next_operation`/`blocked_by` receipt.
A contradictory success claim is `REFUTED` and takes precedence over
`UNKNOWN`; only one exact route-bound candidate is `CLOSED`.

The route-identity contract has a fixed five-case denominator:

| Case ID | Expected observation |
| --- | --- |
| `normal-route-bound-closed` | `CLOSED` |
| `other-route-only-unknown` | `UNKNOWN` |
| `missing-route-identity-unknown` | `UNKNOWN` |
| `duplicate-route-candidates-unknown` | `UNKNOWN` |
| `contradictory-success-refuted` | `REFUTED` |

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
