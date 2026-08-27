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

The producer derives the four statuses from the source-declared `computes`
programs after `syntax.ParseFile` and `bidir.Lower`. The judge performs the
same source interpretation independently, recomputes the byte and meaning
channels, and rejects a digest-only receipt. Outputs belong in a temporary CI
directory; no repository file is written.
