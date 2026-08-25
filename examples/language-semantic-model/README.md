# Language semantic model receipt

This use case produces a deterministic, read-only receipt for the staged Gooo semantic model.

The fixed denominator is 24 cases:

- 19 repository `.gooo` files lowered twice to normalized `semantic.IR`
- 3 authority laws derived from an observed IR
- 2 fail-closed syntax rejections inherited from the exact-head syntax receipt

The accepted upstream coordinate is 22 syntax cases over 24 observed `.gooo` files, 304 physical lines, and 2 package units. Package members remain package inputs rather than additional standalone semantic models.

The receipt exposes 19 indicators: 1 outcome, 10 drivers, and 8 guardrails. A static file or catalog row cannot satisfy the result. The producer is `languagesemantic.Evaluate`, the consumer is `self-improvement-cycle`, and every indicator names its meta-operation.

The command writes only outside the source repository:

```sh
go run ./cmd/language-semantic-witness \
  --root "$GITHUB_WORKSPACE" \
  --head "$HEAD_SHA" \
  --registry examples/language-semantic-model/corpus.json \
  --syntax-artifact "$RUNNER_TEMP/language-syntax-roundtrip/first/artifact.json" \
  --output "$RUNNER_TEMP/language-semantic-model/first/artifact.json"
```

Unknown upstream decisions, missing files, registry drift, stage-order violations, or observed effects lower the resolution and fail closed.
