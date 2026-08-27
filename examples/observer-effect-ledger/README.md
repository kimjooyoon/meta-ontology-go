# Observer-effect ledger

This is a small meta-ontology experiment about the observer, not a second copy
of the repository transformation-effect executor. The source declares separate
entities for an observation result, an observer effect, two receipts, an
independent judgment, a persistent claim, static trigger topology, causal
topology edges, and runner-scoped evidence.

The normal run reads the real `main.gooo` source and a repository root. It does
not write the repository. It does write three external artifacts: the ledger,
the observation receipt, and the observer-effect receipt. That output is itself
recorded as an `OUTPUT` effect, so “repository writes = 0” is not incorrectly
reported as “the observer had no effects.”

The `violate` mode requires an explicit `-allow-intentional-write` flag and
writes a marker only to the supplied disposable root. The resulting ledger is
expected to be `FAIL_CLOSED`, with a fixed denominator of 12 and one failed
repository-write guard. The `unknown` mode preserves the same denominator but
records an `UNKNOWN` stage, step, and reason instead of silently removing an
indicator.

The observer emits the ledger and two role-specific receipts. The separate
`observer-effect-judge` executable re-derives the subject decision, effect
domains, receipt chain, fixed denominator, authority boundary, and persistent
claim transition. It also re-reads the five workflow files so a trigger,
branch-filter, or concurrency-key mutation becomes `FAIL_CLOSED`. A `PASS`
judgment means the evidence is internally valid; it does not authorize a
repository mutation or promotion.

The static topology evidence keeps exact counts separate from the 12-indicator
success denominator: subscribers audited `5/5`, branch-filtered `5/5`,
duplicate PR observation paths `2 -> 1`, and expected skipped CI-child run
objects per PR completion `4 -> 0`. The Actions API snapshot of skipped `59`
and queued `41` is labeled `RUNNER_SCOPED`, time-dependent, and excluded from
the fixed denominator. Because this PR changes protected workflow policy, the
ledger also reports `CI-ROOT-OF-TRUST-001` as the expected
`FAIL_CLOSED`/`BOOTSTRAP_EXPECTED_NEGATIVE` guardian result; it is not hidden or
turned into a required context by this PR.
