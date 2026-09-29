# Gooo and Go declaration comparison

The `billing-declaration-signature-go-ast-v1` experiment gives the project its
first exact-head, system-generated comparison with another language. On the
`ubuntu-24.04` / Go 1.27.0 CI runner, it compares the `PayOrder` declaration in
`examples/billing/main.gooo` with the Go AST baseline in
`examples/language-profile/go-baseline.go`.

## Measured indicators

| Indicator | Meaning |
| --- | --- |
| Equivalent output samples | Paired runs whose normalized package, namespace, operation signature, and stable entity IDs have identical SHA-256 digests; expected 5/5. |
| Gooo wall time and allocations | Minimum, median, and maximum for `sourceexecution.Execute`, including source parsing, semantic lowering, entry resolution, and its execution receipt. |
| Go baseline wall time and allocations | Minimum, median, and maximum for Go `parser.ParseFile` with comments, signature extraction, and resolution of the `gooo:id` / `gooo:namespace` annotations. |
| Go/Gooo ratios | Baseline median divided by Gooo median, scaled to parts per million. A value above 1,000,000 means the Go path observed more time or allocations for this fixture on this runner. |

The measurement warms both paths once and alternates which path runs first for
five paired samples. The receipt binds the exact source SHA, both fixture source
digests, comparison executable digest, Go version, operating system, architecture,
logical CPU count, and runner image label. It is uploaded in the existing
exact-head `language-source-execution` artifact and summarized in the CI run.

## Scope and limits

This experiment measures one declaration-signature workload. The Go path uses
Go's standard AST parser and comment annotations to reconstruct the same stable
IDs; the Gooo path uses the public source-execution receipt. Their work boundaries
are not identical, so time and allocation ratios are diagnostic observations,
not a language speed score. The report makes no claim about native compilation,
general language performance, production workloads, business correctness, or
improvement across runners. CI preserves the observation as non-blocking evidence;
it does not use a manually entered target or Guardian decision.

The next useful expansion is another equivalent task family, such as package
resolution or deterministic query execution, with matching output digests and
separate toolchain-bound receipts. Results should remain grouped by task, runner,
and compiler version so a future language comparison cannot hide workload or
environment changes inside one aggregate number.
