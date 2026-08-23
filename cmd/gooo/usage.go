package main

import (
	"fmt"
	"io"
)

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: gooo <check|generate|roundtrip|query|inspect|graph|analyze|format|fix|provenance|selective-ci|lsp|version> [args]")
}
