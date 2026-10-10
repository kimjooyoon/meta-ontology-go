package main

var topicHelpGroup1 = map[string]string{
	"start": rootHelp,
	"language": `Gooo source basics

A source file declares a package, namespace, semantic entities, and activities.
An activity gives an input type, output type, and a plain-language intent.

Example:
  package sample
  namespace sample
  entity Integer id "sample://integer"
  activity Clamp(Integer) -> Integer computes "if input < 0 { return 0 } else { return input }"

Try it:
  gooo check main.gooo

See docs/language-direction.ko.md and docs/language/language-semantic-model.md
for the current language model and supported syntax.
`,
	"models": `Use a local decision model with Gooo

Gooo declares the available code paths and required output/condition cases.
A compatible local model proposes candidate order; Gooo checks each candidate.
Omit --path-model (body-codegen) or --model (body-construct/body-compose)
for deterministic order. A supplied unreadable or incompatible file is an error.

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

Included record-field example:
  gooo body-context --activity Describe \
    --model examples/scalar-identity/model/model.json examples/scalar-identity/source.gooo.fixture
  See gooo help body-compose for record assembly and saved replay.

Model metadata selects the input format. Inspection makes zero predictions and
candidate tests. READY_FOR_RANKING means that input is supported; source and
caller case results measure the constructed program. Saved replay makes zero
new predictions. Native execution requires Go1.27.2 on the host.

See docs/declared-contract-model.md and docs/record-model-preflight.md.
`,
	"body-codegen": `Generate a Gooo activity body

Usage:
  gooo body-codegen [--json] [body-generation options] --activity <name> <file.gooo>

Example:
  gooo body-codegen --json --activity NonNegative examples/body-codegen/source-ir-fill-probe-choice.gooo.fixture

The output includes generated Go, selected source, finite-case results, candidate
coverage, and unresolved completeness dimensions. Candidate selection can use
deterministic ordering or a configured decision model. Models rank only
typed candidates declared by the plan. The Gooo source defines the allowed
alternatives and case obligations. Fixed pure activities in the same
source can be called from a body; body-compose --entry selects an execution root.
A finite test score covers the recorded inputs and expectations.

See docs/language/body-codegen.md and docs/source-assembly.md.
`,
}
