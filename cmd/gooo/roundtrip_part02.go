package main

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"time"
)

func evaluateRoundTripWithDeadline(file *syntax.File, timeout time.Duration) (roundTripResult, error) {
	if timeout <= 0 {
		return roundTripResult{}, errCommandDeadline
	}
	type evaluation struct {
		result roundTripResult
		err    error
	}
	result := make(chan evaluation, 1)
	go func() {
		value, err := evaluateRoundTrip(file)
		result <- evaluation{result: value, err: err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case value := <-result:
		return value.result, value.err
	case <-timer.C:
		return roundTripResult{}, errCommandDeadline
	}
}
func evaluateRoundTrip(file *syntax.File) (roundTripResult, error) {
	support := syntax.EntityFieldsSupportForFile(file, syntax.EntityFieldsV4Support())
	original, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, support)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("semantic lowering: %w", err)
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("document adaptation: %w", err)
	}
	model, err := bidir.GetWithEntityFieldsSupport(document, support)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("Get: %w", err)
	}
	written, err := bidir.PutWithEntityFieldsSupport(document, model, support)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("Put: %w", err)
	}
	roundTripped, err := bidir.LowerDocumentWithEntityFieldsSupport(written, support)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("lower written document: %w", err)
	}
	getPutErr := bidir.CheckGetPutWithEntityFieldsSupport(document, support)
	putGetErr := bidir.CheckPutGetWithEntityFieldsSupport(document, model, support)
	return roundTripResult{
		original: original, roundTripped: roundTripped,
		equivalent: bidir.EquivalentAfterRoundTrip(original, roundTripped),
		getPut:     getPutErr == nil, putGet: putGetErr == nil,
		getPutErr: getPutErr, putGetErr: putGetErr,
	}, nil
}
