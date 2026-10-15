# Exact-head transport experiment

This experiment separates three identities that a chained workflow must not
collapse:

- `D`: the producer-declared logical subject and its canonical payload digest;
- `R`: the workflow, run, attempt, job, and orchestration head;
- `T`: the immutable artifact ID, producer run, and archive digest.

The fixed `EHT-8` denominator is declared by the eight `Observe...` activities
in `transport.gooo`. The current CI can verify seven obligations. Authenticated
producer attestation remains `UNKNOWN`, so the receipt is deliberately
`OBSERVED / LOWER_RESOLUTION` at `7/8 = 8750` basis points. This lower
resolution authorizes only the existing read-only, non-executing candidate.
Any known identity or digest mismatch is `FAIL_CLOSED`.

Artifact names are lookup labels only. The consumer selects one artifact from
the exact producer run and attempt, downloads it by immutable artifact ID,
recomputes the archive digest, and then asks the Gooo metaprogram to verify the
producer receipt and logical payload.

## Which commit belongs to which role?

A receipt has a subject commit, a checkout commit, and a workflow commit. Think
of a reader opening an older edition of a book: the book's edition and the
reader's current tools have separate origins.

The executable relation lives in
[`originpolicy/source.gooo`](../../internal/meta/selfimprovementtransport/originpolicy/source.gooo).
Its named inputs and conditional body are compiled by Gooo's existing body
generator. The consumer calls that generated function; its receipt retains the
policy source digest and the producer event returned by the GitHub run API.

| Producer event | Relation to the API run head |
| --- | --- |
| `pull_request` | subject commit; the workflow may use a merge commit |
| `push` | subject commit |
| `workflow_run` | workflow commit; the observed subject may be older or from another branch |
| absent or unmodeled event | `UNKNOWN`, with a next observation and blocking fields |

The subject must still match the producer checkout, payload, and archive index.
A contradictory role binding is `REFUTED`. A missing event keeps the run-identity
obligation open. Authenticated producer attestation remains a separate obligation.

This corrects the origin comparison diagnosed in
[original run 38751432958](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/38751432958).
Its producer inspected `4fdbd1d9…` while the observation workflow and API run
head were `814f4dd7…`. The prior consumer compared the subject to the workflow
head. The rejected consumer receipt was unavailable; the diagnosis used the
original producer archive, run metadata, and the source comparison.

Fresh local checks cover nine archive/lifecycle/consumer cases and five partial
policy inputs. They passed, along with a check that the committed projection
matches Gooo source and changes when its relation changes. These are bounded
cases, not a measurement of overall language completeness or CI speed. The
older resolution-policy case counters describe the declared scenario inventory.

The candidate workflow now preserves transport diagnostics even when consumption
fails. Future failures can be traced from the observed event and source relation.
Regenerate the projection from `internal/meta/selfimprovementtransport/originpolicy`
with `go generate`; generated regions are checked against their Gooo source.

GitHub documents the event-dependent commit semantics in
[workflow events](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#workflow_run)
and the workflow definition identity in
[contexts](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts#github-context).
