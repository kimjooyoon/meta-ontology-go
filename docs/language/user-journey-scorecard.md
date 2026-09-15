# User journey and resource scorecard

The scorecard answers human questions about the actual `gooo` executable: can a
user identify the version, check syntax in text and JSON, round-trip a source,
and run a semantic check; how large is the binary; and what wall time and peak
resident memory were observed on the CI runner?

The functional authority remains the existing 12/12 toolchain CLI receipt. The
scorecard adds six positive journeys, five samples per journey, and runner-scoped
resource observations. It does not create a second parser or CLI executor.

## Fixed indicators

The v1 denominator is 15: 3 outcomes, 7 drivers, and 5 guardrails. `USER` reads
6/6, `TOOL_AUTHOR` reads 10/10, and `GOVERNOR` reads 15/15. A lower-resolution
view cannot promote a higher one.

The v1 ceilings are 5000 ms per invocation, 262144 KiB maximum RSS, and
33554432 binary bytes. They are safety envelopes, not optimization targets.
Each command reports five raw observations and min/median/max reductions.

## Determinism boundary

CLI outputs and the reducer are replayed exactly. Wall time and RSS are marked
`RUNNER_SCOPED_NONDETERMINISTIC`; measurements are never byte-replay evidence.
A missing sample, unknown runner identity, unknown operation, or stale subject
lowers resolution and fails closed. A known ceiling violation fails at
`INVARIANT_ONLY` without being relabeled unknown.

All 18 functional and resource definitions name a meta operation and a
Munchausen proof choice. The profiled source is the unchanged
`examples/billing/main.gooo`, whose digest is joined to the binary and exact
head. PASS does not imply overall language completeness or cross-run resource
improvement.
