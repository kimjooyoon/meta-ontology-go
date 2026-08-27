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
same source interpretation independently in the separate consumer package,
recomputes the byte and meaning channels and all per-case transitions, and
rejects a digest-only receipt. A successful conformance judgment also reports
subject resolution separately; an OPEN case remains `OPEN/LOWER_RESOLUTION`.

The CI-only intervention command persists two separate cases in an artifact:

```sh
go run ./scripts/reproducibility-semantics \
  -mode intervention \
  -source examples/reproducibility-semantics/main.gooo \
  -semantic-source "$RUNNER_TEMP/semantic-intervention.gooo" \
  -presentation-source "$RUNNER_TEMP/presentation-intervention.gooo" \
  -head-sha "$HEAD_SHA" \
  -output "$RUNNER_TEMP/reproducibility-intervention.json"
```

Its fixed denominator is `2`; it deliberately has no aggregate score. Outputs
belong in a temporary CI directory; no repository file is written.
