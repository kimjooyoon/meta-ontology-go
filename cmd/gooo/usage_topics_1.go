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
	"models": `Local models for source-declared choices

The Gooo source declares alternatives, types and finite cases. A compatible local
model proposes candidate order. Omitting --model selects deterministic order.

From the compiler source checkout, inspect the included own model:
  gooo body-context --activity Describe \
    --model examples/scalar-identity/model/model.json examples/scalar-identity/source.gooo.fixture

The model metadata selects its input feature. Inspection reports compatibility
and reasons for representation decline with zero predictions and candidate tests.
Use gooo help body-compose for assembly and saved replay. Field and case counts
describe their recorded finite examples. Model-call counts identify actual inference.

See docs/record-model-preflight.md and examples/scalar-identity/model/README.md.
Explicit operation-provider configuration remains documented in
docs/language/laya-decision-provider.md.
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
