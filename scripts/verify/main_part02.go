package main

import "github.com/kimjooyoon/meta-ontology-go/internal/verify"

func run(root, storageRoot, from, to, head, base, branch, expectedHead string, capsOnly, skipCaps, identityOnly bool) error {
	if !skipCaps {
		if err := checkSourcePolicyForRun(root, storageRoot); err != nil {
			return err
		}
	}
	if capsOnly {
		return nil
	}
	// The lightweight CI preflight intentionally has no PR route inputs.
	// Keep its identity checks fail-closed while deferring ownership and route
	// checks to the later full scope gate.
	if head == "" && base == "" {
		identityOnly = true
	}
	if identityOnly {
		if err := validateScopeRevisions(from, to, expectedHead); err != nil {
			return err
		}
		return verifyPRCheckoutIdentity(root, from, to, expectedHead)
	}
	if err := validateScopeRevisions(from, to, expectedHead); err != nil {
		return err
	}
	if err := verifyPRCheckoutIdentity(root, from, to, expectedHead); err != nil {
		return err
	}
	// Changed-path ownership is not a PR authorization input. Every PR runs
	// the same six canonical checks against this exact head, so new feature
	// branches do not need a pre-registered path allowlist.
	if base != "" || head != "" {
		if err := verify.CheckPullRequestPolicy(head, base); err != nil {
			return err
		}
	}
	return nil
}
