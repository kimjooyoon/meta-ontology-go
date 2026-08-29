package verify

import (
	"fmt"
	"os/exec"
	"strings"
)

func changedSourceFiles(root, base, head string) ([]string, error) {
	command := exec.Command("git", "-C", root, "diff", "--name-only", "-z", "--diff-filter=ACMRTUXB", base, head, "--", "*.go", "*.gooo")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("read source policy changed paths: %w", err)
	}
	paths := make([]string, 0)
	for _, path := range strings.Split(string(output), "\x00") {
		if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".gooo") {
			paths = append(paths, path)
		}
	}
	return sortedUnique(paths), nil
}

func revisionViolations(root, base, path string, policy LinePolicy) ([]Violation, error) {
	listed, err := revisionPath(root, base, path)
	if err != nil || !listed {
		return nil, err
	}
	command := exec.Command("git", "-C", root, "show", base+":"+path)
	source, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("read source policy baseline %s: %w", path, err)
	}
	return checkSourceBytes(path, source, policy), nil
}

func revisionPath(root, base, path string) (bool, error) {
	command := exec.Command("git", "-C", root, "ls-tree", "-r", "--name-only", base, "--", path)
	output, err := command.Output()
	if err != nil {
		return false, fmt.Errorf("read source policy baseline path %s: %w", path, err)
	}
	return strings.TrimSpace(string(output)) != "", nil
}

func policyViolationRegressed(current Violation, previous []Violation) bool {
	for _, candidate := range previous {
		if candidate.Path == current.Path && candidate.Rule == current.Rule && candidate.Detail == current.Detail {
			return current.Actual > candidate.Actual
		}
	}
	return true
}

// ValidatePolicyClosure keeps a CLOSED claim separate from its receipt fact.
func ValidatePolicyClosure(closedClaimed, receiptObserved bool) error {
	if closedClaimed && !receiptObserved {
		return fmt.Errorf("closed policy claim has no receipt")
	}
	return nil
}
