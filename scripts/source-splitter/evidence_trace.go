package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func canonicalSplitEvents(plan splitPlan, observed []splitEvent) ([]splitEvidenceEvent, error) {
	targets := make(map[string]string, len(plan.Parts))
	for _, part := range plan.Parts {
		targets[filepath.Clean(part.Path)] = filepath.ToSlash(part.Subject)
	}
	temporaries := make(map[string]string, len(plan.Parts))
	result := make([]splitEvidenceEvent, 0, len(observed))
	stages := 0
	for _, event := range observed {
		target, known := targets[filepath.Clean(event.Target)]
		if !known && event.Kind == "DIRECTORY_SYNC" {
			target = filepath.ToSlash(filepath.Dir(plan.Parts[0].Subject))
			known = true
		}
		if !known {
			return nil, fmt.Errorf("write event target is undeclared: %s", event.Target)
		}
		temporary := ""
		if event.Temporary != "" {
			key := filepath.Clean(event.Temporary)
			temporary, known = temporaries[key]
			if !known && event.Kind == "STAGE" {
				stages++
				temporary = filepath.ToSlash(filepath.Join(filepath.Dir(target),
					fmt.Sprintf(".source-split-stage-%02d", stages)))
				temporaries[key] = temporary
				known = true
			}
			if !known {
				return nil, fmt.Errorf("write event temporary is undeclared")
			}
		}
		result = append(result, splitEvidenceEvent{Kind: event.Kind, Target: target,
			Temporary: temporary, Success: event.Success})
	}
	return result, nil
}

func remainingSplitTemporaries(observed []splitEvent) (int, error) {
	seen := map[string]bool{}
	for _, event := range observed {
		if event.Kind == "STAGE" && event.Temporary != "" {
			seen[event.Temporary] = true
		}
	}
	remaining := 0
	for temporary := range seen {
		_, err := os.Lstat(temporary)
		if err == nil {
			remaining++
		} else if !os.IsNotExist(err) {
			return 0, err
		}
	}
	return remaining, nil
}
