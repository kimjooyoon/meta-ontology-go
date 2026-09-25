package main

import (
	"fmt"
	"io"
)

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: gooo <run|compare|propose-repair|consume-repair|revise-source|revise-from-handoff|evaluate-revision|verify-revision-contract|run-accepted-revision|compare-accepted-revision|stage-accepted-revision|profile|debug|test|emit|check|generate|roundtrip|query|inspect|graph|claim|analyze|format|fix|provenance|selective-ci|invoke|lsp|version> [args]")
}
