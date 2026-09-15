package main

import (
	"errors"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var errCLIEntityFieldsDeferred = errors.New("parse.entity-fields-deferred")

func rejectCLIEntityFieldsIR(ir semantic.IR) error {
	support := syntax.CurrentEntityFieldsSupport()
	if err := validateCLIEntityFieldsSupport(support); err != nil {
		return err
	}
	if semanticIRHasFields(ir) && support.State == syntax.EntityFieldsDeferred {
		return errCLIEntityFieldsDeferred
	}
	return nil
}

func isCLIEntityFieldsDeferredError(err error) bool {
	return err != nil && strings.Contains(err.Error(), errCLIEntityFieldsDeferred.Error())
}

func hasCLIEntityFieldsDeferredDiagnostic(diagnostics syntax.Diagnostics) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == syntax.DiagEntityFieldsDeferred {
			return true
		}
	}
	return false
}

func semanticDiagnosticCode(err error) string {
	if isCLIEntityFieldsDeferredError(err) {
		return "parse.entity-fields-deferred"
	}
	if errors.Is(err, errCommandDeadline) {
		return "semantic.deadline"
	}
	if strings.Contains(err.Error(), "unknown declaration") {
		return "semantic.invalid-endpoint"
	}
	if errors.Is(err, semantic.ErrUnknownRelation) {
		return "semantic.invalid-relation"
	}
	if strings.Contains(err.Error(), "cannot connect") || errors.Is(err, semantic.ErrInvalidFact) {
		return "semantic.invalid-kind"
	}
	return "semantic.invalid"
}
