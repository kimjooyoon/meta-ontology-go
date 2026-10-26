package main

const packageTopicHelp = `Work with a multi-file Gooo workspace

Start an origin reader and inspect its editable input:
  gooo init --template origins my-origins
  cd my-origins
  gooo package execute --inputs inputs.json gooo.workspace.json

Successful input-only execution prints one JSON value per input row.
Add --json for the complete generation, execution and origin record.
Add --receipt-compression gzip with --json to package execute, replay or resume
for lossless compressed output. Saved receipt readers detect gzip automatically.
Inspect gooo version --build --json for package_receipt_transport byte limits.
The original JSON digest and partial observation state remain unchanged.
See docs/language/project-starters.md for the practice checks and source editing.

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

Select original field locations by exact stable IDs before paging:
  gooo package execute --origin-receipt execution.json --origin-field-id urn:your:field --origin-limit 9 gooo.workspace.json

Optional --origin-activity-id and --origin-root-activity-id narrow the same
selection. Subject IDs default to field-input-locations; field-writes and
field-uses are also supported. source_rows counts retained rows before selection;
total_rows counts matches. Reuse the same IDs with the next page offset.
Lineage column inputs keep the full original case journal.
See docs/language/origin-subject-queries.ko.md.

Start from a library template:
  gooo init --template library --module example.org/team/my-library my-library

For workspace manifests and execution, see docs/language/workspace-manifest.md
and docs/language/language-package-execution.md.

Propose a rename by package and stable activity ID across all declared sources:
  gooo package rename-activity --package example.org/team/rules --id urn:rules:fee --to 요금 --json gooo.workspace.json

The JSON proposal includes source edits and an updated workspace manifest. It
also inspects callers outside the entry closure and leaves source files intact.
See docs/language/workspace-activity-refactoring.ko.md.

Propose a source split at a stable declaration ID:
  gooo package split-source --boundaries --package named/inputs --source main.gooo --json workspace.json
  gooo package split-source --package named/inputs --source main.gooo --before urn:gooo:example:named-inputs:net --into net.gooo --json workspace.json

The proposal includes both sources, their original byte ranges and the updated
workspace source list. It preserves declaration identity and leaves files intact.
Apply a saved proposal after checking the current package inputs:
  gooo package split-source --apply split.json --json workspace.json
Application rechecks inputs after staging and verifies the installed package.
See examples/named-inputs/README.md for execution checks and consumer scope.

Export declaration shapes for documentation and API comparison tools:
  gooo package interface --json gooo.workspace.json

The interface includes imported packages, stable IDs, ordered activity inputs,
and record field types, presence, and cardinality. See docs/language/package-interface.md.

Generate an importable Go package from authored computes bodies:
  gooo package generate --go-package rules gooo.workspace.json > rules_generated.go

Evaluate calls the entry and its bound producers. R_<record> and F_<field> expose
source names; stable field IDs remain in gooo struct tags. Generation needs no
cases, model, native build or program execution. Assembling contracts are unsupported.
See docs/language/package-library-generation.md for input and validation limits.

Inspect the entry's prepared source calls before execution:
  gooo package inspect --json gooo.workspace.json
  gooo package inspect --inputs-template gooo.workspace.json > inputs.json

PREPARED shows package exports, source digests, bindings and the entry call closure.
Authored activity IDs remain attached to original and lowered call participants.
The input template contains editable zero values for unbound caller ports.
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
