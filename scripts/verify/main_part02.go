package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func run(root, from, to, head, base, branch, expectedHead string, capsOnly, skipCaps bool) error {
	if !skipCaps {
		if err := printSourceMetrics(root); err != nil {
			return err
		}
		policy := verify.DefaultLinePolicy()
		if err := verify.CheckSourcePolicy(root, nil, policy); err != nil {
			return err
		}
	}
	if capsOnly {
		return nil
	}
	if err := checkAgentPushBranch(branch); err != nil {
		return err
	}
	if err := validateScopeRevisions(from, to, expectedHead); err != nil {
		return err
	}
	if err := verifyPRCheckoutIdentity(root, from, to, expectedHead); err != nil {
		return err
	}
	if validRevision(from) && validRevision(to) && from != to {
		if !revisionAvailable(root, from) {
			return fmt.Errorf("scope base revision %q is unavailable", from)
		}
		if !revisionAvailable(root, to) {
			return fmt.Errorf("scope head revision %q is unavailable", to)
		}
		changed, err := changedPaths(root, from, to)
		if err != nil {
			return err
		}
		scopeBranch := branch
		if scopeBranch == "" {
			scopeBranch = head
		}
		promotion := base == "main" && head == "dev"
		if !promotion {
			if err := verify.CheckPathScopeForBranch(changed, scopeBranch); err != nil {
				return err
			}
		}
		if scopeBranch == "agent/go-version" && !promotion {
			diff, err := changedDiff(root, from, to, "go.mod")
			if err != nil {
				return err
			}
			if err := verify.CheckGoModToolchainDiff(diff); err != nil {
				return err
			}
		}
	}
	if base != "" || head != "" {
		if err := verify.CheckPullRequestPolicy(head, base); err != nil {
			return err
		}
	}
	return nil
}

func printSourceMetrics(root string) error {
	report, err := linecaps.AnalyzeLineMetrics(root)
	if err != nil {
		return err
	}
	total := report.Total()
	fmt.Printf("source metrics: total_files=%d total_dirs=%d go_lines=%d gooo_lines=%d\n", total.RecursiveFiles, total.RecursiveFolders, total.GoLines, total.GoooLines)
	fmt.Printf("source language totals: go_files=%d gooo_files=%d\n", total.GoFiles, total.GoooFiles)
	fmt.Printf("source metrics detail:\n%s", report.Text())
	return nil
}
