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

Declared-case model for typed integer bodies:
  gooo body-context --activity Choose --include-plan \
    --model out/choice-model/model-choice-v1.json \
    examples/body-codegen/source-condition-cases.gooo.fixture

This export includes every authored output case, exact integer features, source
choices and the verified model format. See gooo help models for download and
construction commands. Gooo checks intermediate condition cases during assembly.

Typed branches and variables:
  gooo body-context --activity Choose --include-plan \
    --model examples/scalar-identity/model/model.json examples/caller-typed-paths/unary.gooo.fixture

The current record model reports THREE_DECISION_COUNT_UNSUPPORTED for this one
branch choice. Deterministic assembly remains available. --include-plan shows
source choices, legal options, fallbacks and the separate finite cases. Joint
models also export their complete framed input when compatible.
--value-flow applies to record assembly and Integer assembly with fixed criteria.
It exports predicate input/result roles, source graphs and dependency edges from
the prepared pure-call closure. These source views do not execute the predicates.
Integer criteria can use the scalar-criteria-features/v1 model contract.
Without a model, ranking stays deterministic. Inspect the predicate source:
  gooo body-context --value-flow --activity Fee \
    examples/criteria-dependencies/main.gooo.fixture

An explicit --feature-version must match the
loaded model. Whole-candidate judges and source IR search have separate contracts.
See docs/scalar-source-model.ko.md, docs/source-criteria-dependencies.md, docs/typed-model-preflight.md,
docs/record-model-preflight.md and gooo help body-compose.
`
