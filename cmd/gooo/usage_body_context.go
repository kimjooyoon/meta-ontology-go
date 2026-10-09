package main

const bodyContextHelp = `Inspect source input and model compatibility

` + bodyContextUsage + `

From the compiler source checkout:
  gooo body-context --activity Describe \
    --model examples/scalar-identity/model/model.json examples/scalar-identity/source.gooo.fixture

The verified local model's metadata selects its input feature.
READY_FOR_RANKING means the source input can be encoded for that model.
DECLINED_TO_DETERMINISTIC retains the reason, such as THREE_FIELD_CHOICES_REQUIRED
for a record with a different number of selected fields.
Inspection makes zero predictions and candidate tests. It prepares and typechecks
source alternatives; assembly checks their finite cases. A recorded representation
decline is a successful inspection result. Invalid files or source return failure.

Typed branches and variables:
  gooo body-context --activity Choose --include-plan \
    --model examples/scalar-identity/model/model.json examples/caller-typed-paths/unary.gooo.fixture

The current record model reports THREE_DECISION_COUNT_UNSUPPORTED for this one
branch choice. Deterministic assembly remains available. --include-plan shows
source choices, legal options, fallbacks and the separate finite cases. Joint
models also export their complete framed input when compatible.
--value-flow applies to records. An explicit --feature-version must match the
loaded model. Whole-candidate judges and source IR search have separate contracts.
See docs/typed-model-preflight.md, docs/record-model-preflight.md and gooo help body-compose.
`
