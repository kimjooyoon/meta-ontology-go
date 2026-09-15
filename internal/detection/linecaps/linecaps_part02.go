package linecaps

import (
	"fmt"
	"os"
	"path/filepath"
)

// Analyze checks the supplied repository-relative Go paths. An empty paths
// slice discovers all Go files below root. I/O and parse failures are findings,
// allowing one invocation to report every independently unverifiable file.
func Analyze(root string, paths []string, limits Limits) (Report, error) {
	if err := limits.validate(); err != nil {
		return Report{}, err
	}
	if root == "" {
		return Report{}, fmt.Errorf("linecaps root must not be empty")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return Report{}, err
	}
	if len(paths) == 0 {
		paths, err = Discover(absoluteRoot)
		if err != nil {
			return Report{}, err
		}
	} else {
		paths, err = normalizePaths(absoluteRoot, paths)
		if err != nil {
			return Report{}, err
		}
	}
	findings := make([]Finding, 0)
	for _, path := range paths {
		displayPath, fullPath, err := resolvePath(absoluteRoot, path)
		if err != nil {
			return Report{}, err
		}
		source, err := os.ReadFile(fullPath)
		if err != nil {
			findings = append(findings, Finding{Path: displayPath, Rule: RuleReadFile, Detail: err.Error()})
			continue
		}
		fileFindings, err := AnalyzeSource(displayPath, source, limits)
		if err != nil {
			return Report{}, err
		}
		findings = append(findings, fileFindings...)
	}
	sortFindings(findings)
	return Report{Findings: findings}, nil
}
