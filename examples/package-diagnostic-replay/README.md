# A reusable Gooo diagnostic tool

This workspace turns counts and a diagnostic detail into a next-step explanation.
`tools/diagnostics` owns the classification and three bounded record-field choices
in Gooo; `app/explain` imports it and formats the result. Its construction recipe
comes from the MIT-licensed [Gooo ecosystem workbench](https://github.com/kimjooyoon/gooo-ecosystem-workbench/blob/d821c346/recipes/diagnostics.gooo).

Build/select once and keep the execution record:

```sh
gooo package execute --json \
  --cases examples/package-diagnostic-replay/cases.json \
  examples/package-diagnostic-replay/gooo.workspace.json > diagnostic-execution.json
```

Add `--assembly-model /path/to/model.json` during this first command to use a
compatible local three-field decision model. The optional [own shared QAT model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary)
has 2,072 parameters. Without a model, the declared candidates use deterministic
ordering. The five construction cases, eight permitted paths and actual attempted
paths remain in the generation record.

Run the saved tool on new input rows, without a model argument:

```sh
gooo package replay --receipt diagnostic-execution.json \
  --inputs examples/package-diagnostic-replay/inputs.json \
  examples/package-diagnostic-replay/gooo.workspace.json
```

This prints one entry value per row, including:

```text
"partial: missing branch result [repair-and-replay]"
```

Use `--json` to retain the fresh execution observations and their source binding.
An input-only invocation reports `OBSERVED`, with zero supplied expectations.
For another finite check, replace `--inputs` with `--cases .../cases.json`.
That suite contains four root inputs and eight activity-output expectations,
including an invalid count above 2^53. A mismatched expectation yields `PROGRESS`.

The receipt and workspace can be copied together to a different directory. Source
identities use manifest-relative paths. Replay rebuilds the package graph and
checks saved selections against current declarations before two fresh native
runs; it performs no model inference or new candidate selection. Earlier model
calls remain historical fields in the construction receipt; current calls are
`result.replay.model_calls` and `result.runtime.model_calls`, both zero.

Supply `--go /path/to/go1.27.1` when the Go binary on PATH differs. This tool suggests
an operation as data; it does not modify another project's source. Its fixed cases
establish the stated finite behavior, with broader task coverage still open.
