# Input-only composition plan

This plan is written before enabling the CLI route. Existing internal input-only
decoding and typed graph validation remain the execution authority.

Expose `body-compose --inputs` as an alternative to `--cases` or `--case-series`.
Keep the original input document as `inputs.json` and record its schema in the
CLI envelope. Expected-case documents keep their existing files and behavior.

Predeclared checks:

1. Generate and repeat a mixed scalar graph with independent roots, joins,
   fan-out and partly bound arguments. Retain actual inputs and intermediate
   values; report zero expectations, zero passes and UNKNOWN correctness.
2. Replay the saved graph using only new inputs, with zero new predictions.
   Supply expected cases separately afterwards and score the same saved graph.
3. Use the own record-field model on the Korean scalar identity fixture. Preserve
   9007199254740993, Unicode, false, zero and empty text exactly. Original model
   calls belong to construction; native runs and saved replay make no calls.
4. Resume a partial nested-helper composition with input-only observations.
   Keep source-owned assembly cases and saved candidate order authoritative.
5. Reject mixed modes, wrong schemas, expectations in input-only documents,
   unknown/missing/ill-typed roots, and caller values for bound ports before
   loading a model or starting native compilation. Cases still need an oracle.

No accuracy percentage is inferred from unscored observations. This adds a
usable native entry point; saved multi-activity workbench feedback remains a
separate extension.
