package main

import (
	"encoding/json"
	"io"
	"os"
)

const splitGoEvidenceOperationID = "gooo/meta/generation/SplitGo"

func applySplitWithEvidence(cfg config, plan splitPlan, output io.Writer) error {
	source, err := os.ReadFile(plan.Parts[0].Path)
	if err != nil {
		return err
	}
	observed := make([]splitEvent, 0, len(plan.Parts)*2+1)
	if err := applySplitObserved(plan, func(event splitEvent) {
		observed = append(observed, event)
	}); err != nil {
		return err
	}
	events, err := canonicalSplitEvents(plan, observed)
	if err != nil {
		return err
	}
	remaining, err := remainingSplitTemporaries(observed)
	if err != nil {
		return err
	}
	candidates := make([]splitEvidenceFile, len(plan.Parts))
	targets := make([]string, len(plan.Parts))
	for index, part := range plan.Parts {
		targets[index] = part.Subject
		candidates[index] = splitEvidenceFile{Path: part.Subject, Data: part.Data}
	}
	evidence := splitEvidence{
		OperationID: splitGoEvidenceOperationID, ExpectedHeadSHA: cfg.sha,
		EvidenceComplete: true,
		Source:           splitEvidenceFile{Path: cfg.subject, Data: source}, Candidates: candidates,
		BuildContexts: splitEvidenceBuildContexts(),
		Write: splitWriteEvidence{Complete: true, ExecutionSucceeded: true,
			DeclaredTargets: targets, Events: events, TemporaryFilesRemaining: remaining},
	}
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(evidence)
}

func splitEvidenceBuildContexts() []splitEvidenceContext {
	return []splitEvidenceContext{
		{GOOS: "linux", GOARCH: "amd64", BuildTags: []string{}},
		{GOOS: "windows", GOARCH: "amd64", BuildTags: []string{}},
	}
}
