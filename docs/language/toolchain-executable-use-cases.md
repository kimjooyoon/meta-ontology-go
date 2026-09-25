# Toolchain executable use cases

## Meaning

`TOOLCHAIN-EXECUTABLE-USE-CASES` is satisfied only by a dynamic receipt. A
catalog row without that receipt remains `NOT_SATISFIED`. The receipt executes
one accepted path and two rejection paths against the exact language-concept
artifact, then CI repeats the evaluator and compares the bytes.

## Fixed measurements

| Class | Metric | Target |
| --- | --- | ---: |
| OUTCOME | executable-use-case readiness | 10000 BPS |
| DRIVER | executed cases | 3 |
| DRIVER | accepted paths | 1 |
| DRIVER | fail-closed paths | 2 |
| GUARDRAIL | unresolved cases | 0 |
| GUARDRAIL | repository writes | 0 |
| GUARDRAIL | mutation authority | 0 |
| GUARDRAIL | registry drift | 0 |

The readiness denominator remains 24. The only accepted transition for this
change is `12/24 -> 13/24`, completed delta `+1`, with zero regressions and zero
unresolved evidence. CI, not this document, determines the exact basis-point
delta.

## Munchhausen choices

`FOUNDATION` binds the registry, head SHA, and source artifact digests.
`COHERENCE` requires the canonical success and both rejections to replay.
`REGRESSION` rejects tampering, writes, authority, and registry drift.

No novelty claim is made. The positive property is the combination of
versioned scenarios, executable negative cases, exact-head evidence, and a
fixed readiness denominator.
