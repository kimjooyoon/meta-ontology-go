# Reproducibility semantics witness

Run from the repository root in CI:

```sh
go run ./scripts/reproducibility-semantics \
  -mode produce \
  -source examples/reproducibility-semantics/main.gooo \
  -head-sha "$HEAD_SHA" \
  -output "$RUNNER_TEMP/reproducibility-receipt.json"

go run ./scripts/reproducibility-semantics \
  -mode judge \
  -source examples/reproducibility-semantics/main.gooo \
  -head-sha "$HEAD_SHA" \
  -receipt "$RUNNER_TEMP/reproducibility-receipt.json" \
  -output "$RUNNER_TEMP/reproducibility-judgment.json" \
  -check
```

The producer is deterministic, while the judge recomputes the four statuses
without calling the producer. Outputs belong in a temporary CI directory; no
repository file is written.
