package main

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"strings"
)

const (
	maxDiagnosticCount = 256
	maxDiagnosticBytes = 64 << 10
)

var errDiagnosticLimit = errors.New("diagnostic resource limit exceeded")

func reportDiagnostics(diagnostics syntax.Diagnostics, stderr io.Writer) bool {
	output, err := formatDiagnostics(diagnostics)
	if err != nil {
		fmt.Fprintln(stderr, "gooo: diagnostic resource limit exceeded")
		return false
	}
	if len(output) == 0 {
		return true
	}
	if _, err := stderr.Write(output); err != nil {
		return false
	}
	return true
}
func formatDiagnostics(diagnostics syntax.Diagnostics) ([]byte, error) {
	if len(diagnostics) > maxDiagnosticCount {
		return nil, errDiagnosticLimit
	}
	sorted := canonicalDiagnostics(diagnostics)
	var output strings.Builder
	for _, diagnostic := range sorted {
		line := diagnostic.String() + "\n"
		if output.Len()+len(line) > maxDiagnosticBytes {
			return nil, errDiagnosticLimit
		}
		output.WriteString(line)
	}
	return []byte(output.String()), nil
}
