# Metric transition

`metric-transition` turns source line metrics into a canonical repository state
and binds that state to a verified transformation-effect artifact set.

The state retains every `.go` and `.gooo` file line count and every logical and
physical directory's direct and recursive folder/file counts. Host paths are
excluded so exact-head replay is deterministic.

At an exact fixed point, the after state may alias the before state only after
the transformation ledger, receipts, provenance, and patch all pass their own
verifier. The transition then records an explicit zero delta.

Project-root counts remain observed evidence. Root topology and a root
`README.md` requirement are `NOT_APPLICABLE` and cannot block conformance.
