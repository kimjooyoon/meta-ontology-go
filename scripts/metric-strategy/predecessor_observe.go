package main

import (
	"context"
	"fmt"
	"net/http"

	predecessor "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposalpredecessor"
)

func observeProposalPredecessor(ctx context.Context, client *http.Client, token string, value options) (predecessor.Report, []byte, error) {
	attempts := value.predecessorAttempts
	if attempts == 0 {
		attempts = 1
	}
	collect, selectReport := predecessor.Collect, predecessor.Select
	if attempts > 1 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, predecessorObservationBudget)
		defer cancel()
		collect, selectReport = predecessor.CollectPending, predecessor.SelectPending
	}
	observe := func(ctx context.Context) (predecessor.Report, []byte, error) {
		collection, err := collect(ctx, client, value.githubAPI, token, value.repository, value.predecessorSHA, value.requestedRoute)
		if err != nil {
			return predecessor.Report{}, nil, err
		}
		return selectReport(value.repository, value.subjectSHA, value.predecessorSHA, collection)
	}
	record := func(attempt int, report predecessor.Report) error {
		if attempts == 1 || report.Schema == "" {
			return nil
		}
		return writeJSON(fmt.Sprintf("%s.attempt-%02d.json", value.output, attempt), report)
	}
	return awaitProposalPredecessor(ctx, attempts, observe, record, waitForPredecessorObservation)
}
