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

## Resolution decision protocol

Classify the observation at the smallest registered ownership scope before
deciding whether to wait. The repository is not the default unit of blockage.

| Decision | Applies to | Required action | Independent work |
| --- | --- | --- | --- |
| `STOP_AND_REPAIR` | `SEMANTIC_DEFECT` | Repair the compiler path and rerun its focused regression observation. | Continue only on disjoint scopes. |
| `HOLD_SEMANTIC_LANE` | `SEMANTIC_UNKNOWN` | Make the missing rule, input, or evidence boundary explicit before closing the lane. | Continue on disjoint scopes. |
| `HOLD_CONFLICTING_LANE` | `SCOPE_CONFLICT` | Resolve ownership or authority conflict without widening the proposed delta. | Continue on non-conflicting scopes. |
| `CONTINUE_INDEPENDENT` | `INFRASTRUCTURE_FAILURE` | Record the external cause and next operation; do not reinterpret it as a compiler result. | Yes. |
| `KEEP_PENDING_CONTINUE` | `QUEUE_OR_STALE_OBSERVATION` | Preserve the exact run, head, and missing terminal observation. | Yes. |
| `CONTINUE_ADVISORY` | `ADVISORY_FAILURE` | Keep the signal as follow-up evidence outside compiler closure. | Yes. |

The first three decisions constrain the affected lane. The last three do not
close the semantic question, but they also do not justify stopping unrelated
compiler work, documentation, provenance capture, or independent dogfood.
Continuing is not a retry disguised as success: the exact cause, next
operation, and evidence identity remain `UNKNOWN` until a terminal observation
exists.

### Merge authority is separate

This protocol answers “what work may continue?” It does not answer “may this
commit merge?”. A merge is accepted only through the repository's actual
protected-branch rules and required contexts. A focused compiler observation
can close the bounded semantic question without claiming that every advisory,
queued, or infrastructure observation is healthy; conversely, a protected
merge must never be inferred from a policy classification alone.

Every non-terminal record retains these fields:

- `source_head`
- `cause`
- `next_operation`
- `evidence_identity`

This keeps the continuation decision replayable and prevents a blocked CI
observation from becoming an implicit compiler verdict.

## Merge boundary

This policy does not weaken GitHub branch protection, required contexts, review,
or repository permissions. A normal protected merge still uses the server's
actual rules. The policy only determines which observations are sufficient to
close the bounded compiler question and which observations must remain separate
follow-up work.
