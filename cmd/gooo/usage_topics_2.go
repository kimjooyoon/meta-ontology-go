package main

var topicHelpGroup2 = map[string]string{
	"discover": `Discover a capability from natural language and Gooo source

Usage:
  gooo discover [--json] --query <question> [--domain-contract <contract.gooo>]
    [--generation <body-codegen.json> [--execute-cases <cases.json>
      [--go-bin <go1.27.1>]]] <file.gooo>

This deterministic JEV integration binds the question to the exact Gooo source
and normalized semantic IR. A separate Gooo domain contract supplies the
denominator for declaration coverage; without it, coverage remains UNKNOWN.
Add --generation to replay a saved source-owned IR search or typed path result.
Expected domain activity signatures supply the generation coverage denominator.
Projection replay checks finite selection observations without a model.
With --execute-cases, the command builds the source-replayed Integer -> Integer
projection and executes the supplied cases twice. It records fresh runtime and
reverse observations, counts repeated independent inputs once, and preserves
partial behavior. Independent here means absent from the effective selection
suite; model-training exposure remains unknown. --go-bin selects the Go tool.

See docs/language/capability-discovery.md.
`,
	"init": `Create a starter project

Usage:
  gooo init [--template app|library|diagnostic] [--module <path>] <new-directory>

Examples:
  gooo init hello-gooo
  gooo init --template library my-library
  gooo init --template library --module github.com/acme/my-library my-library
  gooo init --template diagnostic my-diagnostic

Then check and generate:
  cd hello-gooo
  gooo check main.gooo
  gooo body-codegen --json --activity Clamp main.gooo

The starters work without a model. Library and diagnostic module paths customize
the generated workspace and source identity. The diagnostic starter assembles a
Gooo tool from declared choices and includes saved-program replay instructions.
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
