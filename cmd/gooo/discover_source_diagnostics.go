package main

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Keep the parser/lowerer's source origin while the domain loader returns its
// existing discovery-stage code. No diagnostic text is parsed for coordinates.
type discoveryLocatedError struct {
	cause error
	span  syntax.Span
}

func (err *discoveryLocatedError) Error() string { return err.cause.Error() }
func (err *discoveryLocatedError) Unwrap() error { return err.cause }

func discoveryFailureSpan(err error) syntax.Span {
	var located *discoveryLocatedError
	if errors.As(err, &located) {
		return located.span
	}
	return semanticFailureSpan(nil, err)
}

func firstDiscoveryErrorSpan(diagnostics syntax.Diagnostics) syntax.Span {
	ordered := diagnostics.Errors().SortBySpan()
	if len(ordered) == 0 {
		return syntax.Span{}
	}
	return ordered[0].Span
}
