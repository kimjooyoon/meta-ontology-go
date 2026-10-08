package main

import (
	"fmt"
	"io"
	"strings"
)

const rootHelp = `Gooo — Go Of Ontology

Build a small Gooo program:
  gooo init my-app
  cd my-app
  gooo check main.gooo

Generate Go from an activity:
  gooo body-codegen --json --activity Clamp main.gooo

Explore a source-bound capability without executing it:
  gooo discover --query "What can Gooo generate here?" --json main.gooo

The compiler works without a language model. A local Laya service can optionally
rank choices already declared in Gooo; it cannot add code outside those choices.

Commands:
  init             Create an app or library starter
  check            Parse and validate a Gooo source file
  test             Check declared activity-output test markers
  run              Resolve an activity or execute a supported value plan
  body-codegen     Fill declared IR holes and generate Go
  body-construct   Reconsider body choices using whole-program examples
  generate         Generate a project from Gooo declarations
  discover         Map a natural-language question to a source-bound capability
  format, fix      Format or repair Gooo source
  inspect, query   Inspect declarations and semantic relationships
  package          Check and execute a multi-file Gooo workspace
  version          Show the compiler version and build information

Use ` + "`gooo help <topic>`" + ` for guides and examples.
Topics: start, language, models, body-codegen, body-construct, discover, init, check, test, run,
        generate, format, package, inspect, query, version, commands
`

func printUsage(writer io.Writer) {
	fmt.Fprint(writer, rootHelp)
}

func runHelp(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, rootHelp)
		return exitOK
	}
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: gooo help [topic]")
		return exitUsage
	}
	topic := strings.TrimSpace(args[0])
	content, ok := topicHelp[topic]
	if !ok {
		fmt.Fprintf(stderr, "gooo help: no guide for %q; run `gooo help` for available topics\n", topic)
		return exitUsage
	}
	fmt.Fprint(stdout, content)
	return exitOK
}

func helpRequest(args []string) (topic string, requested bool) {
	if len(args) == 0 {
		return "", false
	}
	if args[0] == "help" {
		return strings.Join(args[1:], " "), true
	}
	if args[0] == "--help" || args[0] == "-h" {
		return strings.Join(args[1:], " "), true
	}
	if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
		return args[0], true
	}
	return "", false
}
