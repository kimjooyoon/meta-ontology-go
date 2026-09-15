package main

import (
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
	original, err := bidir.Lower(file)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("semantic lowering: %w", err)
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("document adaptation: %w", err)
	}
	model, err := bidir.Get(document)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("Get: %w", err)
	}
	written, err := bidir.Put(document, model)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("Put: %w", err)
	}
	roundTripped, err := bidir.LowerDocument(written)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("lower written document: %w", err)
	}
	getPutErr := bidir.CheckGetPut(document)
	putGetErr := bidir.CheckPutGet(document, model)
	return roundTripResult{
		original: original, roundTripped: roundTripped,
		equivalent: bidir.EquivalentAfterRoundTrip(original, roundTripped),
		getPut:     getPutErr == nil, putGet: putGetErr == nil,
		getPutErr: getPutErr, putGetErr: putGetErr,
	}, nil
}
