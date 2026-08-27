# Observer-effect ledger

This is a small meta-ontology experiment about the observer, not a second copy
of the repository transformation-effect executor. The source declares separate
entities for an observation result, an observer effect, two receipts, an
independent judgment, a persistent claim, static trigger topology, causal
topology edges, and runner-scoped evidence.

The normal run reads the real `main.gooo` source and a repository root. It does
not write the repository. The three external artifacts are emitted after the
report is built, so the `OUTPUT` coordinate is deliberately
`planned=true`/`OPEN` with no claimed observed write count until those writes
are instrumented and the report is resealed. “Repository writes = 0” is not
reported as “the observer had no effects.”

The `violate` mode requires an explicit `-allow-intentional-write` flag and
writes a marker only to the supplied disposable root. The resulting ledger is
expected to be `FAIL_CLOSED`, with a fixed denominator of 12 and one failed
repository-write guard. The normal result is `UNKNOWN/LOWER_RESOLUTION` at
`11/12` (`9166` basis points) because OUTPUT is open; the `unknown` mode is
`9/12` (`7500` basis points) and records an exact stage, step, and reason
instead of silently removing an indicator. The violation mode is `10/12`
(`8333` basis points).

The observer emits the ledger and two role-specific receipts. The separate
`observer-effect-judge` executable uses a consumer-owned wire model and
re-derives the source policy, semantic intervention outcomes, subject decision,
effect domains, receipt chain, fixed denominator, authority boundary, and
persistent claim transition. Its producer dependency count is `0/0`. One
semantic policy intervention must change the OUTPUT coordinate, decision, and
claim transition; one comment/quoted-text intervention must preserve them. It
also re-reads the five workflow files so a trigger, branch-filter, or
concurrency-key mutation becomes `FAIL_CLOSED`. A `PASS` judgment means the
evidence is internally valid; it does not authorize a repository mutation or
promotion.

The static topology evidence keeps exact counts separate from the 12-indicator
success denominator: subscribers audited `5/5`, branch-filtered `5/5`,
duplicate PR observation paths `2 -> 1`, and expected skipped CI-child run
objects per PR completion `4 -> 0`. The Actions API snapshot of skipped `59`
and queued `41` is labeled `RUNNER_SCOPED`, `HISTORICAL_FIXTURE`, and `OPEN`,
with unknown observation time/query and `current_evidence=false`; it is
time-dependent and excluded from the fixed denominator. The source is bound to
canonical parse/lowering and a computed `OUTPUT_OPEN` policy. Its semantic
policy intervention is causal, while the comment/quoted-text intervention
preserves the canonical IR; raw declaration-looking text cannot establish Gooo
semantics. Because this PR changes protected workflow policy, the
ledger also reports `CI-ROOT-OF-TRUST-001` as the expected
`FAIL_CLOSED`/`BOOTSTRAP_EXPECTED_NEGATIVE` guardian result; it is not hidden or
turned into a required context by this PR.
