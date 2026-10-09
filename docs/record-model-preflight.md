# Check a source and model before assembly

`body-context --model` reads an explicit local record-field model and chooses its
input representation. The output shows the model identity, source binding and
whether the compiler can represent a Korean or English Gooo body for that model.

From the compiler checkout:

```sh
go run ./cmd/gooo body-context --activity Describe \
  --model examples/scalar-identity/model/model.json \
  examples/scalar-identity/source.gooo.fixture
```

The fixture contains existing metadata and 446 bytes of packed weights. The
command validates both files using the same loader as assembly. It prepares and
typechecks source alternatives, then exports their input. Model predictions and
candidate tests remain zero. This is a source-development feature after the
0.6.21 release candidate; check your executable's source or release notes.

## Read the result

| Result | Meaning | Next action |
| --- | --- | --- |
| `READY_FOR_RANKING` / `SOURCE_INPUT_ENCODED` | This verified model can consume the exported source representation. | Run `body-compose --model` to rank and check candidates. |
| `DECLINED_TO_DETERMINISTIC` with a reason | The source exceeds or differs from the model's representation. | Inspect the reason and use ordinary deterministic assembly. |
| `FAIL_CLOSED`, nonzero exit | Reading, model validation or source validation failed. | Correct that input before assembly. |

Inspection returns success for a recorded representation decline: the reason is
the result. Scripts can inspect `model_compatibility.status` before requesting
ranking. This own model requires three field choices; one choice reports
`THREE_FIELD_CHOICES_REQUIRED` and exports no model input text.

`model_compatibility.model` retains metadata and weight digests, feature and
arithmetic versions, setup time and resident tensor bytes. These identify the
artifact actually read. Process memory includes the compiler, validation and Go
runtime in addition to model arrays. Setup timing varies between invocations.

`READY_FOR_RANKING` describes representation compatibility. Correctness comes
from separately declared cases and native execution. Preflight provides no
candidate score, chosen program or accuracy estimate. Default model input excludes
case inputs and expected outputs. `--include-plan` explicitly adds the separate
source plan, including finite cases; `--value-flow` adds the source provenance graph.

## Supported scope

This option currently covers source-owned record field assembly and its four
field-expression, shared, origin and graph model contracts. An explicit
`--feature-version` must match the loaded model. Other `body-context` exports use
their existing explicit features and currently reject `--model`.

The usability idea came from [tinyjs's capabilities and requirements APIs](https://github.com/tarwin/tinyjsapp):
show available behavior and missing conditions before work starts. Gooo applies
that principle to a source/model pair, retaining finite verification at assembly.

[Graph input](record-graph-model-input.md) · [Field assembly](record-field-assembly.md)
· [Scalar identity example](language/scalar-identity.md)
