package main

const packageTopicHelp = `Work with a multi-file Gooo workspace

Read recorded candidate paths:
  gooo package origins --receipt execution.json [--candidate <id>] [--json]
See gooo help package origins for source scope and limits.

Interpret recorded native origins using a Gooo workspace:
  gooo package execute --origin-receipt execution.json --origin-kind elements --origin-limit 32 gooo.workspace.json
Kinds: reads, constructions (including empty lists), elements, links.
Use --origin-offset for the next page. The saved producer is not executed;
only the new workspace consumes recorded claims. Original partial observations
and the source receipt digest stay visible in the new result.
See examples/native-origin-reader/README.md.

Start from a library template:
  gooo init --template library --module example.org/team/my-library my-library

For workspace manifests and execution, see docs/language/workspace-manifest.md
and docs/language/language-package-execution.md.

Propose a rename by package and stable activity ID across all declared sources:
  gooo package rename-activity --package example.org/team/rules --id urn:rules:fee --to 요금 --json gooo.workspace.json

The JSON proposal includes source edits and an updated workspace manifest. It
also inspects callers outside the entry closure and leaves source files intact.
See docs/language/workspace-activity-refactoring.ko.md.

Export declaration shapes for documentation and API comparison tools:
  gooo package interface --json gooo.workspace.json

The interface includes imported packages, stable IDs, ordered activity inputs,
and record field types, presence, and cardinality. See docs/language/package-interface.md.

Inspect the entry's prepared source calls before execution:
  gooo package inspect --json gooo.workspace.json

PREPARED shows package exports, source digests, bindings and the entry call closure.
Authored activity IDs remain attached to original and lowered call participants.
Inspection makes zero program executions and model predictions. See
docs/workspace-source-inspection.md for coverage and source-relative call spans.

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
`
