# Cross-platform release readiness

The versioned denominator is fixed before execution.

- targets: 3
- cases: 20
- indicators: 39
- outcome / driver / guardrail: 3 / 16 / 20
- proofs: FOUNDATION / COHERENCE / REGRESSION
- use cases: 3

Acceptance requires `20/20` cases, `39/39` indicators, and every guardrail at zero.
The report must be `PASS / EXACT` and bind the exact head SHA.

The readiness transition uses the unchanged 24-obligation registry:

- before: `23/24 = 9583`
- after: `24/24 = 10000`
- delta: `+1 / +417`
- regressions / unresolved / repository writes: `0 / 0 / 0`

The denominator does not infer support for ARM, mobile, WebAssembly, package
signing, or public release publication. Those require a future versioned corpus.
