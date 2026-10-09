# Run values now, check expectations later

Use this when you have a typed Gooo program and some inputs, and want to see its
actual values before writing caller tests. From the compiler checkout:

```sh
gooo-dev body-compose --source examples/body-codegen/native-input-joins.gooo.fixture \
  --inputs examples/composition-inputs/joins.json --repeat 2 --out out/observed-joins
```

`inputs.json` is preserved byte for byte in the output. `runtime.json` records
each root input, producer result and consumer input. The envelope's `input_schema`
identifies this mode. The second execution reuses the retained executable.
The underlying native executor runs each request twice to check replay.

An input-only result has zero expectations and zero passes. Correctness is
`UNKNOWN` with `NO_RUNTIME_EXPECTATIONS`; it has no accuracy percentage.
Assembly still checks the finite cases declared in the source.

Replay without generating or predicting again:

```sh
gooo-dev body-compose --source out/observed-joins/original.gooo \
  --inputs examples/composition-inputs/joins.json \
  --composition out/observed-joins/composition.json
```

Then supply the existing named expectations for the same saved graph:

```sh
gooo-dev body-compose --source out/observed-joins/original.gooo \
  --cases examples/body-codegen/native-input-joins-cases.json \
  --composition out/observed-joins/composition.json
```

Those seven input rows provide 49 named expectations. They score caller behavior
after generation and do not change the selected graph. `record.json` exercises
the own record-field model through `--entry Describe --model
examples/scalar-identity/model/model.json` and the scalar identity source.
`helpers.json` exercises nested helpers through `--entry Main` and the dependent
continuation source. Input-only mode also supports `--resume-composition` with an
explicit assembly policy, using the retained candidate ranking.

Root names, types and required ports remain checked. Caller values cannot replace
bound producer outputs. Use exactly one of `--inputs`, `--cases`, `--case-series`;
input-only documents cannot include expectations. Limits remain 1..128 rows,
32 KiB per input document and at most 16 executions including repeats.

The 0.6.23 development source includes this route. Public v0.6.22-dev predates it.
See the [0.6.23 guide](../../docs/releases/0.6.23-dev.md) for the native release
profile, then use `gooo` in the commands above with the matching installed version.
