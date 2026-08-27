# Observer-effect ledger

This is a small meta-ontology experiment about the observer, not a second copy
of the repository transformation-effect executor. The source declares separate
entities for an observation result, an observer effect, two receipts, an
independent judgment, and a persistent claim.

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
claim transition. A `PASS` judgment means the evidence is internally valid; it
does not authorize a repository mutation or promotion.
