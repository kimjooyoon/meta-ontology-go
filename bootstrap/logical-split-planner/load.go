package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadSubjects(name, expectedSHA string) ([]inputSubject, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var evidence projectionEvidence
	if err := json.Unmarshal(data, &evidence); err != nil {
		return nil, err
	}
	if evidence.Schema != "gooo.repository-projection-evidence.v1" {
		return nil, fmt.Errorf("unsupported projection evidence %q", evidence.Schema)
	}
	if evidence.SourceSHA != expectedSHA {
		return nil, fmt.Errorf("evidence SHA %s does not match %s",
			evidence.SourceSHA, expectedSHA)
	}
	selected := make([]inputSubject, 0, len(evidence.Subjects))
	seen := make(map[string]bool)
	for _, subject := range evidence.Subjects {
		if subject.Indicator != "source.line-cap-debt" {
			continue
		}
		if seen[subject.Logical] {
			return nil, fmt.Errorf("duplicate subject %s", subject.Logical)
		}
		if subject.Limit != 75 || subject.Value <= subject.Limit {
			return nil, fmt.Errorf("invalid line-cap evidence for %s", subject.Logical)
		}
		seen[subject.Logical] = true
		selected = append(selected, subject)
	}
	return selected, nil
}
