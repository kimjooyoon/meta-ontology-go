package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
	"strings"
)

func validateScopeRevisions(from, to, expectedHead string) error {
	if from == "" && to == "" {
		if expectedHead != "" {
			return fmt.Errorf("pull-request scope revisions must not be empty")
		}
		return nil
	}
	if !validRevision(from) || !validRevision(to) {
		return fmt.Errorf("scope revisions must both be valid or both be empty")
	}
	if from == to {
		return fmt.Errorf("scope base and head revisions must differ")
	}
	if expectedHead != "" && to != expectedHead {
		return fmt.Errorf("scope head %q does not match expected PR head %q", to, expectedHead)
	}
	return nil
}
func checkAgentPushBranch(branch string) error {
	if !strings.HasPrefix(branch, "agent/") {
		return nil
	}
	return verify.CheckPathScopeForBranch(nil, branch)
}
func verifyPRCheckoutIdentity(root, base, to, expectedHead string) error {
	if expectedHead == "" {
		return nil
	}
	actualHead, err := runGit(root, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	actualHead = strings.TrimSpace(actualHead)
	if actualHead != expectedHead {
		return fmt.Errorf("checked out revision %q does not match expected PR head %q", actualHead, expectedHead)
	}
	if to != expectedHead {
		return fmt.Errorf("scope head %q does not match expected PR head %q", to, expectedHead)
	}
	if !revisionAvailable(root, base) {
		return fmt.Errorf("scope base revision %q is unavailable", base)
	}
	return nil
}
func validRevision(value string) bool {
	if len(value) != 40 || value == strings.Repeat("0", 40) {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
func revisionAvailable(root, revision string) bool {
	if !validRevision(revision) {
		return false
	}
	_, err := runGit(root, "rev-parse", "--verify", revision+"^{commit}")
	return err == nil
}
