# Meta-circular boundary

`main.gooo` declares a meta operation that can describe itself. The witness
keeps three claims separate:

1. `self-description`: the source declares the meta-operation vocabulary.
2. `authorization`: a capability is accepted only when an explicit external
   grant binds issuer, subject digest, operation, scope, and handle.
3. `execution`: a request executes only after authorization, and only at the
   `READ_ONLY` effect ceiling.

The fixed denominator has four cases: description only, explicit read-only
capability, forged capability, and a write capability outside the read-only
scope. The consumer-side `IndependentJudge` re-reads the source and derives
the expected result; it does not trust the producer's top-level decision.

## Evidence contract

Every case emits a `gooo/meta-circular-boundary-receipt/v1` receipt with
`producer`, `consumer`, `meta_operation`, `proof_choice`, `stage`, `step`,
`reason`, and a three-link claim transition from description to authorization
to execution. The complete report has ten fixed indicators, a SHA-256 digest,
and reports `0` repository writes and `false` mutation authority.

The experiment is inspired by the staged structure of MIT's
[Meta-Circular Evaluator](https://groups.csail.mit.edu/mac/classes/6.001/FT98/lectures/lec16/eval.pdf)
and the object-capability rule that authority is conveyed by explicit
references rather than ambient names in E's
[Capability-based Security](https://erights.org/elang/kernel/auditors/index.html)
notes. It does not claim to be a self-hosting evaluator or a cryptographically
unforgeable capability implementation: the handle is deterministic fixture
evidence, and the read-only boundary is a semantic contract exercised by CI.

## Meta value and falsifiers

The meta value is an observable non-equivalence:
`DESCRIPTION != AUTHORIZATION != EXECUTION`. A counterexample would be any
receipt where description-only input becomes authorized, a forged or write
grant reaches execution, or the source/receipt consumer can be changed without
the independent judge rejecting the digest-bound report. The experiment is
therefore falsifiable, not a self-hosting claim.
