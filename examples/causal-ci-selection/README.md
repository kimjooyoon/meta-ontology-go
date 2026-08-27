# Causal CI selection

This is a read-only meta-program experiment. It makes a CI plan by carrying a
changed file through an explicit claim and impact path to a fixed six-check
catalog. A file suffix or directory pattern is never a selection rule.

The real authority input is [`main.gooo`](main.gooo). [`cases.json`](cases.json)
contains three deliberately fixed scenarios:

| Case | Result | Meaning |
| --- | --- | --- |
| `selection` | `SELECTED` | A known claim path selects `go-test`. |
| `full-fallback` | `FULL_FALLBACK` | An unknown owner path descends to all 6 checks and records `OWNER_RESOLUTION_UNAVAILABLE`. |
| `rejection` | `REJECTED` | An unregistered check reference produces no plan. |

The producer emits `gooo/causal-ci-selection-receipt/v1`. The receipt carries
the source digest, the producer/consumer/meta-operation boundary, the proof
choice for every check, path explanations, fallback coordinates, and an
append-only claim transition chain. The `gooo://verifier/causal-ci-selection`
consumer independently recomputes the case decisions and validates the digest;
it is not a second name for the producer.

The check denominator is fixed at 6 (`gofmt`, `go-vet`, `go-test`,
`go-test-race`, `semantic-conformance`, `ci-policy`). The six binary indicators
are also fixed. This experiment measures the quality of causal explanation and
fail-closed descent; it does not execute checks, change refs, authorize merges,
or claim that the causal graph is complete outside this declared corpus.
