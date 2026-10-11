package main

var topicHelpGroup4 = map[string]string{
	"body-outcomes-delta": `Compare saved workflow outcomes

` + bodyOutcomesDeltaUsage + `

Read body-construct output, its saved evaluation.json, body-compose output,
or a composition runtime v1/v2/v3. Match complete caller input tuples and stable
activity IDs, even when case order changes. Requirements and actual outcomes
are compared separately. Regressions require the same observed expectation;
conflicting duplicates and missing observations remain explicit.

This command does not execute programs, call models, or write files. Exit0
means the comparison was produced; it does not mean every requirement passed.
Use --json for exact integer values, original observations and grouped counts.
Use --markdown for a shareable report with explicitly named denominators.
Choose one output format. A group with no eligible population is n/a, not 0%.
See docs/workflow-outcome-delta.md.
`,
	"body-construct": bodyConstructHelp,
	"body-refine": `Run a Gooo policy between bounded construction rounds

` + bodyRefineUsage + `

The policy owns continuation and retained-round decisions. Feedback cases guide
the loop; optional evaluation cases run after selection. Each round and the
selected program are saved for replay. See examples/assembly-feedback/README.md.
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
  package, body-context, body-compose, body-construct, body-execute, body-search-run, body-realize, body-refine
  body-outcomes-delta, profile, debug, decide, invoke, lsp, version

Use ` + "`gooo help <command>`" + ` for a guide to a core command. Advanced
revision, provenance, repair, and self-improvement commands are listed in the
repository README and implemented as specialized workflows.
`,
}
