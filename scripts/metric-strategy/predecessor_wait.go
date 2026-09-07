package main

import (
	"context"
	"fmt"
	"time"

	predecessor "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposalpredecessor"
)

const maximumPredecessorObservations = 13
const predecessorObservationInterval = 15 * time.Second
const predecessorObservationBudget = 4 * time.Minute

type predecessorObservation func(context.Context) (predecessor.Report, []byte, error)
type predecessorObservationRecord func(int, predecessor.Report) error
type predecessorObservationWait func(context.Context) error

func awaitProposalPredecessor(ctx context.Context, attempts int, observe predecessorObservation, record predecessorObservationRecord, wait predecessorObservationWait) (predecessor.Report, []byte, error) {
	var last predecessor.Report
	if attempts < 1 || attempts > maximumPredecessorObservations {
		return last, nil, fmt.Errorf("predecessor observations must be bounded from 1 to %d", maximumPredecessorObservations)
	}
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return last, nil, err
		}
		report, payload, err := observe(ctx)
		if report.Schema != "" {
			last = report
		}
		if recordErr := record(attempt, report); recordErr != nil {
			return last, nil, recordErr
		}
		if err == nil {
			if !report.Ready() {
				return last, nil, fmt.Errorf("predecessor observation returned success without selection")
			}
			return report, payload, nil
		}
		if attempt == attempts || !predecessor.AwaitablePending(report) {
			return last, nil, err
		}
		if err := wait(ctx); err != nil {
			return last, nil, err
		}
	}
	return last, nil, fmt.Errorf("predecessor observation budget exhausted")
}

func waitForPredecessorObservation(ctx context.Context) error {
	timer := time.NewTimer(predecessorObservationInterval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
