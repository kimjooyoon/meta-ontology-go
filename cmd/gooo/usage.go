package main

import (
	"fmt"
	"io"
)

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: gooo <run|profile|debug|test|emit|check|generate|roundtrip|query|inspect|graph|analyze|format|fix|provenance|selective-ci|invoke|lsp|version> [args]")
}
