# Claim lifecycle calculus

This is an isolated Gooo meta-programming experiment. The six `activity`
signatures are the source relation: each ordered input/result relation and its
`computes` value program becomes one durable claim. The experiment does not
add syntax or a general-purpose evidence library.

The producer emits a receipt with six claims and twelve append-only events:

- every claim is registered as `UNRECORDED -> OPEN`;
- an observation equal to a proposition closes one claim as
  `OPEN -> DISCHARGED`;
- an observation with the same subject/predicate but a different object
  closes another as `OPEN -> REFUTED`;
- missing or insufficient observations keep claims `OPEN` with `UNKNOWN`;
- conflicting observations keep a claim `OPEN` and are `FAIL_CLOSED`;
- a dependency blocked by an upstream `OPEN` claim remains `OPEN` without
  propagating a terminal state.

Each `computes` value is a structured `claim-case/v1` source case containing
the claim ID, proposition tuple `(subject,predicate,object)`, zero or more
observation tuples `(subject,predicate,observed_object,provenance)`, dependency
claim ID, and observation stage/step. It contains no expected status, decision,
or relation label. The evaluator computes equality, contradiction, insufficient
observation, and conflict to derive `SUPPORTS`, `CONTRADICTS`, `UNAVAILABLE`, or
`AMBIGUOUS`. The producer and the independent judge each run
`syntax.ParseFile` followed by `bidir.Lower`, recover the six cases from those
value programs, and derive states from the tuples. No activity table or
post-hoc claim index is authoritative.

The separate judge recomputes the source relation, replays every digest chain,
and checks every fixed numerator/denominator. It does not import the
producer's evaluator. The workflow changes one source `observed_object` from
`down` to `healthy`; this must change the relation, subject counts, semantic
receipt, decision, and `OPEN -> REFUTED`/`OPEN -> DISCHARGED` transition. A
comment-only intervention may change only the raw source digest. A separately
resealed ledger tamper is rejected by source reconstruction.

The receipt separates `conformance_decision` from the subject decision and
`subject_counts`. Subject decisions are `MIXED` or `FAIL_CLOSED`, with
`resolution=LOWER_RESOLUTION` and a separate reason, so conformance `PASS`
does not hide the `UNKNOWN` cases or the `FAIL_CLOSED` case.

The default effect boundary is read-only: repository writes are derived from
before/after repository snapshots and are `0`, mutation authority is `false`,
and the artifact binds the actual `runtime.Version()`. The receipt is evidence
for this experiment, not authority to promote or mutate semantic state.
