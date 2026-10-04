# Record expression context for a small Gooo model

The compiler can export actual record-field expression alternatives and their
Korean/English intent for `triple_record_field_context_v1_joint_v1`:

```sh
gooo body-context --activity Select \
  --feature-version triple_record_field_context_v1_joint_v1 \
  examples/body-codegen/record-field-assembly.gooo.fixture
```

This performs source/type/case-shape preflight with zero predictions and zero
candidate outcomes. `--include-plan` includes the separate finite case plan;
input values and expected outputs are excluded from the model context.

Each of three choices carries the complete field name, first expression, second
expression and intention. The Go runtime projects 32 structural slots per
alternative and 192 intent slots into the fixed 768-element array. Copying the
same field, selecting another field, constants and ordered concatenation have
different structural observations. Literal and intent fragments use bounded
hash channels. Parentheses and consistent field/receiver renaming preserve the
observed structural meaning. Enclosing branch logic and local binding resolution
remain in Gooo's typed program and finite evaluator, outside this feature map.

An explicitly supplied model with this feature version uses this context during
`body-codegen` and construction inside `body-compose`:

```sh
gooo body-codegen --json --activity Select \
  --model /path/to/record-field-model/model.json \
  examples/body-codegen/record-field-assembly.gooo.fixture
```

The runtime requires record-specific metadata and trained weights. Existing
integer-trained models keep their explicit ordinal transfer context. An omitted
model keeps deterministic mask order. Wrong arity or oversized complete context
retains its full attempted parts and continues deterministically with zero
predictions. The field model currently ranks three binary choices with one
prediction; observed finite case outcomes select the best partial or complete
body within the source's attempt budget. Replay reconstructs those observations
without calling the model again.

The receipt names the feature version, complete input hash, actual prediction,
candidate order and case/field counts. These are counts for supplied examples.
Prediction probability is an ordering hint, separate from verified completeness.

The dense 768/24/8 layout has 18,656 parameters and a 3,200-byte caller workspace.
FP32 tensors occupy 74,624 bytes. Ternary matrices store five trits per byte;
the 3,854-byte weight file includes FP32 biases and decodes to 18,752 tensor
bytes plus eight scale bytes. JSON and expression parsing allocate temporary
storage; these tensor sizes do not describe whole-process RAM.

The [paired-intent model](https://huggingface.co/asketeddy/gooo-record-field-tiny-v2)
learns all eight combinations of the three authored field requirements. In the
published comparison, FP32 picked the complete mask first on 424/1,536 source
views and 81/512 new-body/new-wording views. Finite continuation completed every
full-budget native graph. These separate counts show the current wording and
representation limits; probability does not replace execution evidence.

## Shared field judgment

SDK v0.2.23 adds `triple_record_field_context_v1_shared_v1`. It keeps the complete
source context and feature array above, then uses the same 256/8/2 judge on each
field. Eight summed path scores come from one prediction. This requires weights
trained for that explicit contract. FP32 tensors occupy 8,288 bytes; ternary files
occupy 446 bytes and decode into 2,096 tensor bytes plus eight scale bytes. The
caller workspace remains 3,200 bytes.

```sh
gooo body-context --activity Select \
  --feature-version triple_record_field_context_v1_shared_v1 \
  examples/body-codegen/record-field-assembly.gooo.fixture
gooo body-codegen --json --activity Select --model /path/to/shared/model.json \
  examples/body-codegen/record-field-assembly.gooo.fixture
```

The receipt records `local_shared_field_prediction_then_finite_tdd`. A retained
model can serve concurrent requests with separate selection state. Unsupported
record arity/context size continues deterministically with zero predictions.
A field-specific model applied to a scalar body also continues deterministically
and reports `FIELD_MODEL_REQUIRES_RECORD_BODY`. Reconstruction makes zero new
predictions. Each field is judged independently, so cross-field dependencies
must be checked by the typed body and finite cases.

The same model can rank [local field update alternatives](record-field-updates.md).
Its expression context retains the exact expressions; local reaching definitions
remain outside the trained feature map. Sequential dependence is evaluated
by the source body for every attempted mask.

## Conditional bodies without an else

Source-owned bodies also accept `if condition { ... }` followed by ordinary
statements, early guard returns and nested conditional updates. The absent
branch falls through. The compiler retains the source form and requires the
whole function to return on every path; branch locals stay inside their scope.
