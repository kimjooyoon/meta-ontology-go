package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
)

func reportCompileFailure(failure error, stdout, stderr io.Writer) int {
	fmt.Fprintf(stderr, "metric-meta-program: compile: %v\n", failure)
	var refusal *unresolvedpolicy.OperationRefusal
	if !errors.As(failure, &refusal) {
		return 1
	}
	payload, err := encodeJSON(refusal)
	if err != nil {
		fmt.Fprintf(stderr, "metric-meta-program: encode unresolved origin: %v\n", err)
		return 1
	}
	written, err := stdout.Write(payload)
	if err == nil && written != len(payload) {
		err = io.ErrShortWrite
	}
	if err != nil {
		fmt.Fprintf(stderr, "metric-meta-program: write unresolved origin: %v\n", err)
	}
	return 1
}
