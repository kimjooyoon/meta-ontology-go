package main

var topicHelpGroup4 = map[string]string{
	"body-construct": `Construct a program using caller feedback

` + bodyConstructUsage + `

Record-choice helpers keep their own Gooo cases. Each completed combination is
compiled and run on --construction-cases; --cases runs after selection. The
optional model ranks local choices once during initial construction. Without it,
the order is deterministic. Saved construction rechecks every attempted program
without inference. Consumed caller inputs are reported separately from new ones.
See examples/caller-guided-construction/README.md.
`,
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
  package, body-compose, body-construct, body-execute, body-search-run, body-realize, body-refine
  profile, debug, decide, invoke, lsp, version

Use ` + "`gooo help <command>`" + ` for a guide to a core command. Advanced
revision, provenance, repair, and self-improvement commands are listed in the
repository README and implemented as specialized workflows.
`,
}
