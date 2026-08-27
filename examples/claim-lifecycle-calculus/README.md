# Claim lifecycle calculus

This is an isolated Gooo meta-programming experiment. The six `activity`
signatures are the source relation: each ordered input/result relation and its
`computes` value program becomes one durable claim. The experiment does not
add syntax or a general-purpose evidence library.

The producer emits a receipt with six claims and twelve append-only events:

- every claim is registered as `UNRECORDED -> OPEN`;
- supporting evidence closes one claim as `OPEN -> DISCHARGED`;
- contradicting evidence closes another as `OPEN -> REFUTED`;
- missing evidence keeps claims `OPEN` and records either a direct or
  dependency-blocked cause;
- ambiguous evidence remains `OPEN` and is `FAIL_CLOSED`.

Each `computes` value is a structured `claim-case/v1` source case containing
the claim ID, prior state, evidence kind and ID, dependency claim ID, observed
stage/step/reason, and an expected result used only for conformance comparison.
The producer and the independent judge each run `syntax.ParseFile` followed by
`bidir.Lower`, recover the six cases from those value programs, and derive
states from observed evidence. No activity table or post-hoc claim index is
authoritative.

The separate judge recomputes the source relation, replays every digest chain,
and checks every fixed numerator/denominator. It does not import the
producer's evaluator. The workflow also changes one source evidence kind and
adds one comment-only intervention: the former must change the semantic
receipt, decision, and transition, while the latter may change only the raw
source digest. A separately resealed ledger tamper is rejected by source
reconstruction.

The receipt separates `conformance_decision` from `subject_resolution` and
`subject_counts`, so a conformance `PASS` does not hide the three `UNKNOWN`
cases or the one `FAIL_CLOSED` case.

The default effect boundary is read-only: repository writes are derived from
before/after repository snapshots and are `0`, mutation authority is `false`,
and the artifact binds the actual `runtime.Version()`. The receipt is evidence
for this experiment, not authority to promote or mutate semantic state.
