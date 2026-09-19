# Problem-solving scope

This project does not treat every CI result as one undifferentiated compiler verdict.
The compiler lane closes a bounded semantic problem; other observations remain
valuable, but they do not silently redefine that problem.

## Primary lane: compiler definition and defect resolution

A compiler change is in scope when all of the following are true:

1. A reproducible `.gooo` or compiler-input case identifies the observed behavior.
2. The intended semantic rule is explicit, including the rejected or `UNKNOWN` cases.
3. The source delta is limited to the registered ownership paths for the lane.
4. A focused CI observation exercises the changed path and its regression case.
5. The result records the exact source head, input identity, toolchain identity, and outcome.

Candidate generation, deterministic replay, and provenance are evidence in this
lane. They are not adoption or self-improvement success until a subsequent run
uses the accepted semantic result.

## Failure classes

| Class | Meaning | Effect on the lane |
| --- | --- | --- |
| `SEMANTIC_DEFECT` | The changed compiler path violates the declared rule or regression case. | Stop this lane, repair, and rerun focused evidence. |
| `SEMANTIC_UNKNOWN` | The rule, input, or required evidence is insufficient to decide the changed path. | Stop this lane until the missing boundary is made explicit. |
| `SCOPE_CONFLICT` | The proposed delta crosses an owned path or changes an unrelated authority. | Stop only the conflicting lane; preserve independent work. |
| `INFRASTRUCTURE_FAILURE` | Runner, dependency service, cache, permission, or workflow setup failed outside the changed semantic path. | Record the exact cause and continue independent work. Do not call it a compiler pass or failure. |
| `QUEUE_OR_STALE_OBSERVATION` | A job is queued, stale, superseded, or has not produced a terminal observation. | Keep the observation pending; do not convert it to success or block unrelated work. |
| `ADVISORY_FAILURE` | A security, ecosystem, cost, or external-use observation is not required for this compiler delta. | Preserve it as a separate follow-up signal. |

`INFRASTRUCTURE_FAILURE`, `QUEUE_OR_STALE_OBSERVATION`, and
`ADVISORY_FAILURE` are not CI bypasses. They are classifications that prevent an
unrelated signal from being misreported as a semantic compiler defect.

## What is allowed while a lane is blocked

Work may continue in another registered ownership scope when it does not depend
on the unresolved semantic decision or protected operation. Examples include
documentation, provenance capture, independent examples, workflow diagnosis,
and another compiler primitive with a disjoint source surface.

The blocked lane must retain its exact head, cause, next operation, and evidence
identity. A timeout or missing artifact is never relabeled as a pass.

## Merge boundary

This policy does not weaken GitHub branch protection, required contexts, review,
or repository permissions. A normal protected merge still uses the server's
actual rules. The policy only determines which observations are sufficient to
close the bounded compiler question and which observations must remain separate
follow-up work.
