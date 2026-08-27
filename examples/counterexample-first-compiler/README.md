# Counterexample-first meta compilation

This is a bounded, read-only vertical experiment over one real Gooo
transformation: `syntax.ParseFile` followed by `bidir.Lower`. The Gooo source
declares the candidate policy as `CanonicalEntityID(... ) computes
"identity:v1"`; the policy is then applied to the lowered entity nodes.

The corpus contains raw source inputs only. The producer executes every
observed input, compares lowered IDs with the policy predicate, discovers a
violation, and runs a deterministic shrinker. The shrinker executes each
immediate candidate and records the final local-minimality proof as a
numerator/denominator. A compile decision can be promoted only after the same
minimal source is re-executed through the resolution input and the observed
predicate passes.

The independent judge repeats ParseFile, Lower, predicate evaluation, shrinking,
and resolution from the raw source/corpus. It does not import the producer
package or a shared outcome table. Receipts preserve actual diagnostics,
lowering errors, semantic digests, producer/consumer/meta-operation/proof
choice, append-only claim transitions, and UNKNOWN stage/step/reason.

The CI artifact also contains a semantic intervention (`identity:v1` to
`identity:v2`) and a comment-only intervention. The former must change the
semantic digest, first minimal counterexample, and claim-transition digest; the
latter must preserve semantic digest and decision evidence.

This does not claim general compiler correctness, global minimality, theorem
proving, unbounded corpus coverage, or repository mutation authority.
