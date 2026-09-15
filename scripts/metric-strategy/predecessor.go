package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func writeProposalPredecessor(value options) error {
	token := os.Getenv("GITHUB_TOKEN")
	if value.githubAPI == "" || token == "" || value.repository == "" || value.subjectSHA == "" || value.predecessorSHA == "" || value.requestedRoute == "" || value.selectedProposal == "" {
		return fmt.Errorf("github-api, token, repository, subject-sha, predecessor-sha, route, and selected-proposal are required")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	report, payload, selectionErr := observeProposalPredecessor(context.Background(), client, token, value)
	if report.Schema == "" && selectionErr != nil {
		return selectionErr
	}
	if err := writeJSON(value.output, report); err != nil {
		return err
	}
	if selectionErr != nil {
		return selectionErr
	}
	if err := os.MkdirAll(filepath.Dir(value.selectedProposal), 0o755); err != nil {
		return err
	}
	return os.WriteFile(value.selectedProposal, payload, 0o600)
}
