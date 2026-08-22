package main

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorresolution"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func resolveAncestry(ctx context.Context, client *githubClient, cfg config,
	immediate string) (predecessorselection.Result, predecessorresolution.Report, error) {
	attempts := make([]predecessorresolution.Attempt, 0,
		predecessorresolution.SearchLimit)
	ancestor := immediate
	for depth := range predecessorresolution.SearchLimit {
		result, err := selectAncestor(ctx, client, cfg, ancestor)
		if err != nil {
			return predecessorselection.Result{}, predecessorresolution.Report{}, err
		}
		attempt := predecessorresolution.Attempt{Depth: depth,
			AncestorSHA: ancestor, Selection: result}
		terminal := result.Report.Decision == predecessorselection.DecisionSelected ||
			result.Report.Reason != predecessorselection.ReasonNotFound ||
			depth == predecessorresolution.SearchLimit-1
		if !terminal {
			attempt.ParentSHA, err = resolveParent(ctx, client, cfg.repository, ancestor)
			if err != nil {
				return predecessorselection.Result{}, predecessorresolution.Report{}, err
			}
		}
		attempts = append(attempts, attempt)
		if terminal {
			return finishAncestry(cfg, immediate, result, attempts)
		}
		ancestor = attempt.ParentSHA
	}
	return predecessorselection.Result{}, predecessorresolution.Report{},
		fmt.Errorf("unreachable ancestor resolution state")
}

func selectAncestor(ctx context.Context, client *githubClient, cfg config,
	ancestor string) (predecessorselection.Result, error) {
	input, err := collect(ctx, client, cfg, ancestor)
	if err != nil {
		return predecessorselection.Result{}, err
	}
	result, err := predecessorselection.Select(input)
	if err != nil {
		return predecessorselection.Result{}, err
	}
	replay, replayErr := predecessorselection.Select(input)
	if replayErr != nil || replay.Report.ReportDigest != result.Report.ReportDigest {
		return predecessorselection.Result{}, fmt.Errorf("ancestor selection replay mismatch")
	}
	return result, nil
}
