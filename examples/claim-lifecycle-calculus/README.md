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

The separate judge reparses this source, recomputes the source relation,
replays every digest chain, and checks every fixed numerator/denominator. It
does not import the producer's evaluator. The workflow uploads the receipt as
the result artifact and reports `PASS`, `FAIL_CLOSED`, and `UNKNOWN` case
counts.

The default effect boundary is read-only: repository writes are `0` and
mutation authority is `false`. The receipt is evidence for this experiment,
not authority to promote or mutate semantic state.
