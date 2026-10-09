# Cross-platform release readiness

The versioned denominator is fixed before execution.

- targets: 4 (Linux amd64, Darwin amd64, Darwin arm64, Windows amd64)
- structural release cases: 26
- indicators: 39
- outcome / driver / guardrail: 3 / 16 / 20
- proofs: FOUNDATION / COHERENCE / REGRESSION
- use cases: 3

Acceptance requires `26/26` structural cases, `39/39` indicators, and every guardrail at zero.
The report must be `PASS / EXACT` and bind the exact head SHA.

The readiness transition uses the unchanged 24-obligation registry:

- before: `23/24 = 9583`
- after: `24/24 = 10000`
- delta: `+1 / +417`
- regressions / unresolved / repository writes: `0 / 0 / 0`

The native profile additionally executes language and package examples and
replays their saved construction. Package caller construction adds three raw
observations per platform: partial 1/4, complete 4/4 and saved replay 4/4, with
zero new inference during replay. Original source, cases and failure histories
are checked independently of the structural denominator.

Darwin arm64 is one of the four explicit targets. Other ARM ports, mobile,
WebAssembly and signing need their own evidence. Public release publication is
governed by the separate publication contract.
