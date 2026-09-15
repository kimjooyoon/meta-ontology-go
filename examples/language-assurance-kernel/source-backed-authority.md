# Source-backed authority contract

This example fixes the contract before adding an evaluator, adapter, CI gate, or
readiness promotion.

## Current quantitative state

- Language assurance denominator: 12
- Operating obligations before and after this contract: 6
- Readiness credit from this contract: 0
- Adoption state: CONTRACT_ONLY
- Contract rules: 7

The contract cannot certify its own implementation. A later evaluator must
consume this already-merged contract without changing it.

## Metric

The numerator is the number of accepted facts with all of these exact bindings:

1. a source reference,
2. a content-addressed source snapshot,
3. an exact byte span,
4. a span digest recomputed from source bytes,
5. an authority reference, and
6. an authority grant for that exact source scope.

The denominator is all accepted facts. The target is 10000 basis points. An
empty denominator is UNKNOWN, not perfect coverage.

## Resolution rule

Missing source bytes, an unpinned URL, an invalid span, or an unavailable
authority grant preserves observation UNKNOWN, lowers resolution to
INVARIANT_ONLY, and enforces BLOCK. Semantic interpretation remains Candidate
until a later contract defines an independently checkable transformation.

## Use cases

- A local Gooo declaration can be authoritative only inside its declared
  business scope and only at an exact source snapshot and byte span.
- A pinned gomacro repository snapshot can support a claim about gomacro's own
  documented behavior. It does not define Gooo semantics.
- A Hada News item is a discovery input. Its live URL alone remains Candidate;
  acceptance requires a pinned original source and an authority grant.

## Munchhausen choice

This epoch uses FOUNDATION to freeze the authority contract. COHERENCE belongs
to an independent validator epoch. REGRESSION belongs to a later external
corpus epoch. The three labels are not counted as three proofs in this PR.
