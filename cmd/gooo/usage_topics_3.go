package main

var topicHelpGroup3 = map[string]string{
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
  gooo init --template library --module example.org/team/my-library my-library

For workspace manifests and execution, see docs/language/workspace-manifest.md
and docs/language/language-package-execution.md.

Export declaration shapes for documentation and API comparison tools:
  gooo package interface --json gooo.workspace.json

The interface includes imported packages, stable IDs, ordered activity inputs,
and record field types, presence, and cardinality. See docs/language/package-interface.md.

Reuse a saved package execution without a model:
  gooo package replay --receipt execution.json --inputs inputs.json gooo.workspace.json

Continue saved record choices under an explicit Gooo policy:
  gooo package resume --receipt execution.json --assembly-policy-workspace policy.workspace.json --cases cases.json gooo.workspace.json

Reconsider imported helpers using caller examples:
  gooo package construct --json --construction-cases feedback.json --cases evaluation.json --attempts 8 gooo.workspace.json
  gooo package construct --json --receipt construction.json --cases evaluation.json gooo.workspace.json
  gooo package construct --receipt construction.json --inputs inputs.json gooo.workspace.json

Save the first command's JSON output as construction.json. Models are optional;
--model orders record choices and --fill-model selects initial source fills.
Saved construction rechecks all attempts with zero new inference. See
examples/package-caller-construction/README.md for a runnable package example.
Use exactly one of --cases (finite checks) or --inputs (actual values).
Input-only JSON is OBSERVED; the original construction decision stays separate.
`,
}
