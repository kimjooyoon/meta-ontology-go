# Repository topology use case

This use case answers the first observable questions a repository reader asks:
how many physical Go and Gooo lines each file contains, and how many direct and
recursive files and folders each directory contains. It consumes the existing
`line-metrics` report instead of introducing a second scanner.

The receipt has a versioned denominator of 10 coordinates: 1 outcome, 5
drivers, and 4 guardrails. A passing receipt means only that these 10 exact-head
coordinates agree. It does not imply that the language is complete, novel,
fast, memory-efficient, or suitable for a particular user workload.

## Reader-dependent resolution

All views project the same receipt and digest. `READER` exposes 4/4 row and
language-total coordinates. `IMPLEMENTER` exposes 7/7 coordinates by adding
ontology, meta-binding, and decision-vocabulary evidence. `GOVERNOR` exposes
all 10/10 coordinates, including exact subject identity, root exceptions, and
the no-mutation guardrail. A lower view never promotes a higher view.

Known upstream `FAIL_CLOSED` decisions remain visible as an exact count and do
not become unknown evidence or PASS. Only values outside the versioned decision
vocabulary lower the receipt resolution.

The project root is observed for raw direct and recursive counts. Its topology
limits are exactly two `NOT_APPLICABLE` indicators, and its missing README rule
is exactly one `NOT_APPLICABLE` indicator. Any other applicability value is
`FAIL_CLOSED / REPOSITORY_TOPOLOGY_DECISION_UNKNOWN` at `LOWER_RESOLUTION`.

## Meta-code binding

The receipt hashes and validates both
`examples/root-readme-indicator/main.gooo` and
`examples/meta-binding-coverage/main.gooo`. Every source indicator must name a
producer, consumer, meta operation, and Munchausen proof choice. The explicit
unbound-indicator witness must be present with value zero.

CI generates the bound source metrics and receipt twice, compares bytes, runs a
fresh check, exercises unknown-decision and count-mismatch cases, publishes the
full artifact, and writes the three audience projections to the job summary.
