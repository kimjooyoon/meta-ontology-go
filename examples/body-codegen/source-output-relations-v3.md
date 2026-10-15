# V3 output relationships and finite construction

This source-owned example returns input/4 below67 and input+13 otherwise. Two
structural choices select comparison order and branch layout. Six declared output
cases and three intermediate Boolean cases guide finite body assembly.

```sh
gooo body-context --activity Choose --model model-v3.json \
  examples/body-codegen/source-output-relations-v3-cases.gooo.fixture
gooo body-construct \
  --source examples/body-codegen/source-output-relations-v3-cases.gooo.fixture \
  --entry Main --model model-v3.json --attempts 4 \
  --construction-cases examples/body-codegen/source-output-relations-v3-construction.json \
  --cases examples/body-codegen/source-output-relations-v3-evaluation.json \
  --out output-v3-construction
gooo body-construct --source output-v3-construction/original.gooo \
  --construction output-v3-construction/construction.json \
  --cases examples/body-codegen/source-output-relations-v3-evaluation.json
```

Use a compiler version with the V3 adapter and a matching pinned artifact from
[the guide](../../docs/source-relative-v3-contract-model.md). Inspection reports
all original requirements and output relationships with zero model calls.
Construction makes one initial model call; saved replay works after removing the
model file and makes zero new calls. Omit --model for deterministic construction.

Caller83->96 is separate from the source's six examples and the eight evaluation
inputs. Native cases include exact Integer9007199254740993. The release profile
uses fresh zero weights to establish interface behavior; fitted V3 first-choice
quality remains the separately published study.
