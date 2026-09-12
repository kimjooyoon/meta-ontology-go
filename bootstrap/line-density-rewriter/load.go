package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadDensitySubjects(name, expectedSHA string) ([]splitSubject, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var plan splitPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	if plan.Schema != "gooo.logical-split-plan.v1" {
		return nil, fmt.Errorf("unsupported split plan %q", plan.Schema)
	}
	if plan.SourceSHA != expectedSHA {
		return nil, fmt.Errorf("plan SHA %s does not match %s", plan.SourceSHA, expectedSHA)
	}
	selected := make([]splitSubject, 0)
	seen := make(map[string]bool)
	for _, subject := range plan.Subjects {
		if subject.Reason != "density-rewrite" && subject.Reason != "static-density-rewrite" &&
			subject.Reason != "large-density-rewrite" {
			continue
		}
		if !subject.Executable {
			continue
		}
		if seen[subject.Logical] || subject.RequiredSave < 1 ||
			(subject.Reason == "density-rewrite" && subject.RequiredSave > 10) {
			return nil, fmt.Errorf("invalid density subject %s", subject.Logical)
		}
		seen[subject.Logical] = true
		selected = append(selected, subject)
	}
	return selected, nil
}

func densitySourcePath(root, logical string) (string, error) {
	if logical == "" || filepath.IsAbs(logical) || strings.ContainsRune(logical, 0) {
		return "", fmt.Errorf("unsafe logical path %q", logical)
	}
	clean := filepath.Clean(filepath.FromSlash(logical))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe logical path %q", logical)
	}
	return filepath.Join(root, clean), nil
}

func densityLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}
