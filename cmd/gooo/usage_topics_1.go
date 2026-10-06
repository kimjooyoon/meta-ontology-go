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
	"models": `Optional language models

Gooo remains usable without a model. When a model is not configured, decisions
use the compiler's deterministic candidate order.

To use a local Laya server, start it on loopback and set:
  export GOOO_LAYA_URL=http://127.0.0.1:8787/v1/systemone
  gooo body-codegen --json --activity Clamp main.gooo

Laya ranks only candidates already declared by Gooo. The compiler still checks
the selected body and reports declared-case results. See docs/language/laya-decision-provider.md
for configuration details and provider behavior.
`,
	"body-codegen": `Generate a Gooo activity body

Usage:
  gooo body-codegen [--json] [body-generation options] --activity <name> <file.gooo>

Example:
  gooo body-codegen --json --activity NonNegative examples/body-codegen/source-ir-fill-probe-choice.gooo.fixture

The output includes generated Go, selected source, finite-case results, candidate
coverage, and unresolved completeness dimensions. Without GOOO_LAYA_URL, candidate
selection is deterministic. With Laya configured, the model chooses only among
typed candidates declared by the plan. A finite test score is not a proof over
every possible input.

See docs/language/body-codegen.md and docs/source-assembly.md.
`,
}
