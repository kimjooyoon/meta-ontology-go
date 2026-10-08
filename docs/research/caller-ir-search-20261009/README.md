# Caller-guided IR expressions — 2026-10-09

Plan/example commit: `3d3658a5`. Implementation and mixed example frozen before
model use: `de5aba5b5103dd6e2f1c73afc9c3e09c7b46ba08`. The clean local compiler's
full identity is in `build.json`. This is development source after 0.6.12.

Previously the same scalar example stopped in PREFLIGHT with
`joint construction requires a record-choice contract at Choose`.

| Program / ordering | Retained eligible combinations | Native program attempts | Selected local expectations | Evaluation |
| --- | ---: | ---: | ---: | ---: |
| Integer search / deterministic | 5 | 2 | 1/1 | 3/3 |
| Integer search + record / deterministic | 40 | 16 | 2/2 | 4/4 |
| Integer search + record / existing graph model | 40 | 9 | 2/2 | 4/4 |

The mixed program fills an integer expression and three record fields. The
unchanged graph model suggested the record ordering once; the integer search
used the deterministic source grammar. Its metadata SHA is
`3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202`, weights SHA
`76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.

Both mixed paths produced byte-identical Gooo source. Saved replay re-executed
their construction history and again matched 4/4 evaluation expectations with
zero new model calls. Four evaluation root inputs differ from both consumed
caller inputs; this does not establish independence from model training. The
exact integer 9007199254740993 is included. These are two programs and three
construction runs, followed by two saved replays, not independent accuracy trials.

`observations.tar.gz` contains original command output, selected Gooo and Go,
construction receipts, evaluation and build identity. The workbench's v2 reader
recounts scalar search observations separately from record cases. Its regression
fixtures retain these original outputs in compressed form.

Focused tests exercise changed selectors, expression and observed values, grammar
coverage, conflicting local/caller expectations, source and program limits,
explicit binds, mixed choices, cancellation, and old v1 record receipts.

Local `go vet ./...` and the billing semantic check passed. The broad macOS
`go test ./...` invocation had failures: existing platform-specific tool cases
and native two-second execution timeouts during the parallel package run.
`baseline-tests.log` reruns two tool failures on unchanged compiler source
`4a0cfdd8`; both reproduce, while the sampled old caller/runtime tests pass.
The focused race result is retained separately. Linux canonical CI remains the
whole-repository check; local broad-test success is not claimed.

Candidate execution errors retain a failure and stop the current request. Joint
construction of multi-hole `source_fill` remains a further language integration.
