package main

const modelsHelp = `Use a local decision model with Gooo

Gooo declares the available code paths and required output/condition cases.
A compatible local model proposes candidate order; Gooo checks each candidate.
Omit --path-model (body-codegen) or --model (body-construct/body-compose)
for deterministic order. A supplied unreadable or incompatible file is an error.

Check the compiler's included SDK with gooo version --build --json.
The public0.6.27 binary includes SDK0.2.37. This development source includes
SDK0.2.39 and supports the interaction model described below.

Typed integer branches, comparisons and variables:
  1. From the matching source checkout, get model-choice-v1.json and its checksum
     from https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.37-experimental
     See docs/releases/0.6.26-dev.md for download and checksum commands.

  2. Inspect the exact source/case input without making a prediction:
     gooo body-context --activity Choose --model out/choice-model/model-choice-v1.json \
       examples/body-codegen/source-condition-cases.gooo.fixture

  3. Assemble and check a body:
     gooo body-codegen --json --activity Choose \
       --path-model out/choice-model/model-choice-v1.json --path-step-attempts 1 \
       examples/body-codegen/source-condition-cases.gooo.fixture

  Run the same source deterministically:
     gooo body-codegen --json --activity Choose \
       examples/body-codegen/source-condition-cases.gooo.fixture

  Use body-construct --model for native construction, then replay its saved
  construction.json without a model file. See gooo help body-construct and
  docs/releases/0.6.26-dev.md for the full source, cases and output arguments.
  The declared-case model makes one initial prediction; later candidates reuse
  that ordering. Its fixed ranking mode does not combine with sampling,
  repeated model feedback or an external execution oracle.

Ordered source, output and intermediate-condition interaction model:
  Get model-interaction-v1.json and its checksum from
  https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.39-experimental
  See docs/interaction-contract-model.md for verified download commands.

  gooo body-context --activity Choose --model out/interaction-model/model-interaction-v1.json \
    examples/body-codegen/source-interaction-condition-cases.gooo.fixture
  gooo body-codegen --json --activity Choose \
    --path-model out/interaction-model/model-interaction-v1.json \
    examples/body-codegen/source-interaction-condition-cases.gooo.fixture

  Scores order candidates; authored output and condition checks record what
  each candidate satisfied. The guide also covers assignments and saved replay.

Included record-field example:
  gooo body-context --activity Describe \
    --model examples/scalar-identity/model/model.json examples/scalar-identity/source.gooo.fixture
  See gooo help body-compose for record assembly and saved replay.

Model metadata selects the input format. Inspection makes zero predictions and
candidate tests. READY_FOR_RANKING means that input is supported; source and
caller case results measure the constructed program. Saved replay makes zero
new predictions. Native execution requires Go1.27.2 on the host.

See docs/declared-contract-model.md and docs/record-model-preflight.md.
`
