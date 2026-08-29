package verify

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckProjectedSourcePolicyRevision reports only changed-surface regressions.
// Existing baseline findings remain in the source-metrics inventory.
func CheckProjectedSourcePolicyRevision(root, storageRoot string, files []string, policy LinePolicy, base, head string) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if storageRoot == "" || base == "" || head == "" {
		return fmt.Errorf("source policy revision context is incomplete")
	}
	if err := verifyRevision(root, base); err != nil {
		return err
	}
	if err := verifyCheckedOutRevision(root, head); err != nil {
		return err
	}
	if len(files) == 0 {
		var err error
		files, err = changedSourceFiles(root, base, head)
		if err != nil {
			return err
		}
	}
	current := make([]Violation, 0)
	for _, path := range sortedUnique(files) {
		current = append(current, checkSourceFile(root, path, policy)...)
	}
	current = append(current, checkDirectoryLayout(storageRoot, policy)...)
	regressions, err := policyRegressions(root, base, current, policy)
	if err != nil {
		return err
	}
	if len(regressions) == 0 {
		return nil
	}
	return fmt.Errorf("source policy regressions:\n%s", formatViolations(regressions))
}

func verifyRevision(root, revision string) error {
	command := exec.Command("git", "-C", root, "rev-parse", "--verify", revision+"^{commit}")
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("source policy base revision unavailable: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func verifyCheckedOutRevision(root, expected string) error {
	command := exec.Command("git", "-C", root, "rev-parse", "HEAD")
	output, err := command.Output()
	if err != nil || strings.TrimSpace(string(output)) != expected {
		return fmt.Errorf("source policy checkout does not match expected head")
	}
	return nil
}
