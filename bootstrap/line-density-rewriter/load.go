package main

import (
	"encoding/json"
	"fmt"
	"os"
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
		if subject.Reason != "density-rewrite" && subject.Reason != "static-density-rewrite" {
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
