package main

import (
	"fmt"
	"os"
)

func rewriteSubjects(root string, inputs []splitSubject) ([]rewriteSubject, error) {
	results := make([]rewriteSubject, 0, len(inputs))
	for _, input := range inputs {
		name, err := densitySourcePath(root, input.Logical)
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		if lines := densityLines(data); lines != input.Lines {
			return nil, fmt.Errorf("line evidence drift for %s: %d != %d",
				input.Logical, lines, input.Lines)
		}
		rewritten, operations, err := compactDensity(name, data)
		if err != nil {
			return nil, err
		}
		after := densityLines(rewritten)
		status := "blocked"
		if after <= 75 {
			info, statErr := os.Stat(name)
			if statErr != nil {
				return nil, statErr
			}
			if err := os.WriteFile(name, rewritten, info.Mode().Perm()); err != nil {
				return nil, err
			}
			status = "applied"
		}
		results = append(results, rewriteSubject{
			Logical: input.Logical, Before: input.Lines, After: after,
			Operations: operations, Status: status, Consumer: "line-density-rewriter",
			Operation: "compact-obvious-lines", Proof: "axiomatic-foundation",
		})
	}
	return results, nil
}

func appliedCount(subjects []rewriteSubject) int {
	count := 0
	for _, subject := range subjects {
		if subject.Status == "applied" {
			count++
		}
	}
	return count
}
