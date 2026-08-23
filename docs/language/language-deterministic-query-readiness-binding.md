# Deterministic query readiness binding

The binding joins one exact-head concept artifact, one deterministic query
report, and the fixed 24-obligation language-readiness registry.

It accepts only these exact current-state values:

- query cases: `32/32`
- query indicators: `18 = 1 OUTCOME + 8 DRIVER + 9 GUARDRAIL`
- readiness: `16/24 = 6666 BPS`
- bound coordinates: `12/12`
- unresolved, effects, repository writes, mutation authorities: `0`

The separate readiness transition ledger must prove `15/24 -> 16/24` and
`6250 -> 6666 BPS`; this binding does not infer a predecessor from current
state. An unknown or non-`PASS` upstream decision lowers the resolution and
returns `FAIL_CLOSED`; it is never interpreted as a fixed point.
