package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func run(root, storageRoot, from, to, head, base, branch, expectedHead string, capsOnly, skipCaps bool) error {
	if !skipCaps {
		if storageRoot == "" {
			storageRoot = root
		}
		if err := printSourceMetrics(root, storageRoot); err != nil {
			return err
		}
		policy := verify.DefaultLinePolicy()
		if err := verify.CheckProjectedSourcePolicy(root, storageRoot, nil, policy); err != nil {
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
