package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
)

func reportGenerateIOError(jsonMode bool, stdout, stderr io.Writer, filename, code, prefix string, err error) int {
	if jsonMode {
		return reportFailure(true, stdout, stderr, "generate", filename, code, err.Error(), syntax.Span{})
	}
	fmt.Fprintf(stderr, "gooo: %s: %s: %v\n", filename, prefix, err)
	return exitFailure
}
