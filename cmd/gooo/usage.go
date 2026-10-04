package main

import (
	"fmt"
	"io"
)

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: gooo <run|compare|propose-repair|consume-repair|revise-source|revise-from-handoff|evaluate-revision|verify-revision-contract|run-accepted-revision|compare-accepted-revision|stage-accepted-revision|profile|debug|test|emit|receipt-schema|check|decide|generate|body-codegen|body-path-stream|body-path-run|body-execute|body-realize|completeness-delta|body-context|roundtrip|query|inspect|graph|claim|analyze|format|fix|provenance|selective-ci|invoke|lsp|version> [args]")
}
