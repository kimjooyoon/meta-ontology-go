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

The first PR's full CI also found a missing syntax-registry entry for this new
policy source. The registration metaprogram generated all nine coordinated
members, including denominator v88. A fresh check confirms this source's five
preservation laws and an inventory with zero missing or unregistered files.
The [registration observation](../../docs/observations/transport-origin-registration/observations.json)
retains the exact request, generating activities and before/after member digests.
The new syntax case records an exercised use of existing language features.

GitHub documents the event-dependent commit semantics in
[workflow events](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#workflow_run)
and the workflow definition identity in
[contexts](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts#github-context).

## Carrying a candidate through later workflows

The next failure occurred after transport had correctly bound the origins.
Candidate run [38768968110](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/38768968110)
inspected subject `de821899…` using workflow `21725f7d…`. Two downstream readers
looked for a candidate named after the workflow commit. The attestation reader
failed to download it; the authorization bridge recorded `UNKNOWN`, then a later
job tried to continue without a verified authorization.

Delivery now carries a small `output/delivery.json` manifest alongside the
candidate and transport receipt. It declares the producer run, attempt, event,
workflow commit and logical subject, with hashes of the two payload files.
The shared Go reader checks the exact run and complete artifact population,
downloads one role by immutable ID, verifies its archive hash, and checks the
manifest against both payloads. A filename locates a role; its suffix supplies
no subject identity. The candidate's source-file hash binds the transport's
source-file hash. Its semantic observation hash remains a separate value.

The decisions are executable activities in the same registered Gooo source:

```gooo
activity BindCandidateSubject(candidate Text, transport Text, declared Text) -> Text
    id "gooo://self-improvement/transport/bind-candidate-subject" computes `
if candidate == "" || transport == "" || declared == "" {
    return "CANDIDATE_SUBJECT_UNKNOWN"
}
if candidate != transport || candidate != declared {
    return "CANDIDATE_SUBJECT_MISMATCH"
}
return "CANDIDATE_SUBJECT_BOUND"
`
```

The source also chooses delivery roles, continuation eligibility and which
origin names each continuation artifact. The Go projection executes those
choices; Go handles HTTP, bounded archives and typed receipt validation.

Authorization continuation requires an independently verified
`CLOSED/AUTHORIZED` result. An unavailable delivery keeps its causal `UNKNOWN`
and diagnostic artifact. It schedules no continuation. Continuation envelopes
carry `source_head_sha` for the workflow and `subject_sha` for the candidate.
Older envelopes with one origin retain their original interpretation.
The v25/v26 readers use their own workflow code while retaining the candidate's
subject in the payload. These records authorize scheduling; execution and grants
keep their existing separate contracts.

Fetches own a fresh directory, have a two-minute overall deadline and at most
eight metadata pages, and bound archives to 16 MiB, 128 members and 8 MiB per
selected member. Old candidate bundles without a delivery manifest remain
`UNKNOWN`; readers do not reconstruct missing provenance from a filename.
Fresh observations and their bounded case counts are recorded
[here](../../docs/observations/candidate-delivery-origins/observations.json).
