package main

var topicHelpGroup2 = map[string]string{
	"discover": `Discover a capability from natural language and Gooo source

Usage:
  gooo discover [--json] --query <question> <file.gooo>

This deterministic JEV integration binds the question to the exact Gooo source
and normalized semantic IR. It does not call a model, generate code, execute
the source, or authorize work. The shared completeness receipt records the
discovery observation while leaving generation, independent use cases, runtime
execution, and reverse observation unresolved.

See docs/language/capability-discovery.md.
`,
	"init": `Create a starter project

Usage:
  gooo init [--template app|library] <new-directory>

Examples:
  gooo init hello-gooo
  gooo init --template library my-library

Then check and generate:
  cd hello-gooo
  gooo check main.gooo
  gooo body-codegen --json --activity Clamp main.gooo

The starter is deterministic and does not require a model.
`,
	"check": `Check Gooo source

Usage:
  gooo check [--semantic] [--provenance-store <ledger.jsonl>] [--json] <file.gooo>

Example:
  gooo check main.gooo

Use --semantic to include semantic validation and --json for a machine-readable
report. Checking does not execute the generated Go program.
`,
	"test": `Check Gooo language-test markers

Usage:
  gooo test [--json] <file.gooo>

Add a marker that names an activity and its expected output entity:
  package sample
  namespace sample
  entity User id "sample://user"
  entity BuildProducesUser id "gooo://test/activity/Build/output/User"
  activity Build() -> User

gooo test checks that Build declares User as its output entity. It does not
test runtime values, side effects, or generated Go. See examples/language-test/README.md.
`,
}
