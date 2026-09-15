# External ecosystem execution

This use case turns a merged reference into CI-only execution evidence without
promoting that evidence into an official language decision.

## Fixed measurement

The versioned denominator contains exactly 8 indicators: reference binding,
pinned commit, pinned tree, Go 1.27.0, two uncached external executions,
normalized replay equality, and repository write boundaries. Success remains
exactly `8/8 = 10000` basis points. The regression denominator is exactly
`10/10`, including 3 missing-evidence cases and 6 known invariant failures.

The measured Go 1.27.0 baseline is `6/8 = 7500` basis points and therefore
`FAIL_CLOSED / EXACT`, not success. Each run produced the same 981 normalized
terminal outcomes: 953 pass, 20 skip, 7 fail, and 1 build failure. The 8 failure
outcomes repeat exactly in both runs. Diagnostics expose a vet rejection, a
stale converter test API, a printer fixture mismatch, and a changed reflect
type surface. These are observations under the pinned environment, not inferred
causes or permission to weaken the success threshold.

Raw timestamps, durations, output text, and Go 1.27 `OutputType` annotations are
not compared. The witness compares sorted final package/test outcomes. It
accepts the fixed test actions and the two build actions documented by Go 1.27;
every other action is a known guardrail violation.

## Meta-programming connection

The chain is executable rather than descriptive:

`merged reference evidence -> pinned checkout -> event observer -> indicators -> decision`

`FOUNDATION` binds the reference and toolchain, `COHERENCE` binds two real runs
and their normalized replay, and `REGRESSION` binds the fixed negative suite.
Unavailable evidence lowers resolution to `COARSE` and fails closed. An unknown
top-level reference decision is never treated as a fixed point.

## Boundaries

The witness performs exactly 2 external executions, 0 source writes, 0 external
checkout writes, 0 official mutations, and 0 promotions. This PR establishes a
reproducible incompatibility baseline. It cannot raise the language-assurance
denominator from `11/12` to `12/12`; a later capability-level adapter must earn
that transition without rewriting this result.

The event contract follows the [Go 1.27 release notes](https://go.dev/doc/go1.27),
[`test2json`](https://go.dev/src/cmd/test2json/main.go), and
[`buildjson`](https://go.dev/src/cmd/go/internal/help/helpdoc.go).
