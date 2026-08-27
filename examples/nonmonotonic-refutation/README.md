# Non-monotonic refutation experiment

This experiment makes evidence-driven knowledge revision a meta-semantic
operation. It is deliberately narrower than an event-sourcing system: the
producer does not publish a replayable state mutation log. It publishes three
fixed claims and ordered evidence; an independent consumer derives the current
status from each evidence item and retains every prior status in the receipt.

The fixed denominator is three claims and the fixed transition total is six:

| Case | Status history | Purpose |
| --- | --- | --- |
| `support-only` | `OPEN -> DISCHARGED` | supporting evidence is accepted |
| `new-counterevidence` | `OPEN -> DISCHARGED -> REFUTED` | new evidence refutes a discharged claim |
| `re-evaluation` | `OPEN -> DISCHARGED -> REFUTED -> DISCHARGED` | later support re-evaluates a refuted claim |

The resulting counts are `OPEN->DISCHARGED=3`,
`DISCHARGED->REFUTED=2`, and `REFUTED->DISCHARGED=1`. Two claims are
currently `DISCHARGED`, one is currently `REFUTED`, and nine status values
are retained including the three initial states. Current discharge coverage is
`2/3 = 6666` basis points; it is not a claim that every current claim is true.

## Formal choices

The experiment adopts two principles from the primary literature:

1. Doyle's *A Truth Maintenance System* defines truth maintenance as recording
   justifications and updating beliefs when new information arrives. We adopt
   justification-bearing evidence, explicit dependency coordinates, and
   append-only historical annotations. Source: [Doyle, *A Truth Maintenance
   System*, Artificial Intelligence 12(3), 1979](https://doi.org/10.1016/0004-3702(79)90008-0).
2. Reiter's *A Logic for Default Reasoning* describes defaults as beliefs that
   may be modified or rejected by subsequent observations. We adopt
   non-monotonic transitions and reject the assumption that a prior
   `DISCHARGED` result remains valid after counterevidence. Source: [Reiter,
   *A Logic for Default Reasoning*, Artificial Intelligence 13(1-2),
   1980](https://doi.org/10.1016/0004-3702(80)90014-4).

The experiment rejects three tempting shortcuts:

- It does not delete or overwrite an old status; a transition stores `before`,
  `after`, evidence ID, `stage`, `step`, `reason`, and a previous digest.
- It does not treat the producer's final answer as authoritative; the
  independent consumer replays the evidence sequence and checks the source
  digest.
- It does not implement arbitrary default logic, belief ranking, probability,
  or global conflict resolution. The transition relation is intentionally
  finite and falsifiable.

Each evidence item binds `producer`, `consumer`, `meta_operation`, and
`proof_choice`. The `.gooo` file is the source vocabulary; Go code contains
the bounded experiment semantics and emits receipts outside the repository.
Both commands report `repository_writes=0` and `mutation_authority=false`.

## Replay and falsifiability

The consumer runs the independent adjudicator twice. A canonical run must
produce byte-identical receipts and a six-element digest chain. Reordering,
removing, or relabeling one evidence item changes the source-bound producer
receipt or violates the transition relation and must yield
`FAIL_CLOSED`. Adding more evidence is not silently accepted: the fixed
denominator and fixed transition total make the experiment fail closed until
the contract is intentionally revised.

This is a read-only semantic experiment, not a promotion gate and not a
general-purpose knowledge base.
