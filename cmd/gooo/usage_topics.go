package main

var topicHelp = map[string]string{
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
	"discover": `Discover a capability from natural language and Gooo source

Usage:
  gooo discover [--json] --query <question> <file.gooo>

This deterministic JEV integration binds the question to the exact Gooo source
and normalized semantic IR. It does not call a model, generate code, execute
the source, or authorize work. The shared completeness receipt records the
discovery observation while leaving generation, independent use cases, runtime
execution, and reverse observation unresolved.

See docs/declared-completeness-receipt.md.
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
	"run": `Resolve or execute a Gooo activity

Usage:
  gooo run [--json] --entry <activity> [--input <input.json> | --record-input <record.json>] [--runtime-plan <runtime-plan.json>] [--iterations N] <file.gooo|package-directory>

Example:
  gooo run --entry Clamp --input input.json main.gooo

Without --input or --record-input, this resolves the activity declaration and
reports its typed inputs and output only when the source has no runtime binds;
it does not evaluate computes or generated Go. A bound source without input
fails closed with SOURCE_RUNTIME_BINDINGS_UNSUPPORTED. With an explicit input,
the value-plan path executes registered value operations and declared typed
binds. For bound sources it uses runtime-plan.json from the current directory
when present; ` + "`--runtime-plan`" + ` selects a different validated artifact.
Use --json for a machine-readable receipt.

See docs/language/language-source-execution.md and
docs/language/language-package-execution.md for the two execution scopes.
`,
	"generate": `Generate a project from Gooo declarations

Usage:
  gooo generate <file.gooo> --out <directory>

Example:
  gooo generate main.gooo --out generated

When the source declares typed runtime binds, generation also writes a
validated runtime-plan.json into the output directory. Pass --runtime-plan to
choose another relative filename.
`,
	"format": `Format Gooo source

Usage:
  gooo format [--check] [--json] <file.gooo>

Use --check to report whether formatting is needed without changing the file.
`,
	"package": `Work with a multi-file Gooo workspace

Start from a library template:
  gooo init --template library my-library

For workspace manifests and execution, see docs/language/workspace-manifest.md
and docs/language/language-package-execution.md.
`,
	"inspect": `Inspect source declarations

Usage:
  gooo inspect <file.gooo>

The result summarizes declarations and stable semantic identifiers.
`,
	"query": `Query Gooo declarations

Usage:
  gooo query [--json] <file.gooo> [--id <stable-id>] [--kind <kind>] [--predicate <relation>]

Example:
  gooo query --json main.gooo --kind activity
`,
	"version": `Show Gooo version

Usage:
  gooo version [--build] [--json]

Use --build to include compiler build information.
`,
	"commands": `Core commands

  init, check, test, run, body-codegen, generate, emit
  format, fix, inspect, query, graph, receipt-schema
  package, body-compose, body-execute, body-search-run, body-realize
  profile, debug, decide, invoke, lsp, version

Use ` + "`gooo help <command>`" + ` for a guide to a core command. Advanced
revision, provenance, repair, and self-improvement commands are listed in the
repository README and implemented as specialized workflows.
`,
}
