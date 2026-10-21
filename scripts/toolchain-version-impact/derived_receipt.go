package main

import (
	"fmt"
	"slices"
	"strings"
)

func appendDerivedCandidates(root string, plan projectionPlan, projected map[string][]byte,
	planID, decisionID string, receipt *projectionReceipt, graph *[]any) ([]candidateImage, error) {
	tracked, err := trackedPaths(root)
	if err != nil {
		return nil, err
	}
	images := make([]candidateImage, 0, len(plan.DerivedOutputs))
	for _, rule := range plan.DerivedOutputs {
		if !slices.Contains(tracked, rule.TargetPath) {
			return nil, fmt.Errorf("derived target is not tracked: %s", rule.TargetPath)
		}
		if _, alreadyProjected := projected[rule.TargetPath]; alreadyProjected {
			return nil, fmt.Errorf("derived target overlaps another output: %s", rule.TargetPath)
		}
		derived, err := deriveCandidate(root, rule, projected)
		if err != nil {
			return nil, err
		}
		images = append(images, candidateImage{rule.TargetPath, derived.After, derived.Mode})
		file := projectedFile{Path: rule.TargetPath, InputSHA256: digest(derived.Before),
			OutputSHA256: digest(derived.After), DerivationRule: rule.RuleID,
			SourcePath: rule.SourcePath, SourceOutputSHA256: digest(projected[rule.SourcePath])}
		receipt.Files = append(receipt.Files, file)
		*graph = append(*graph, derivedGraphNodes(plan.PlanCoreSHA256, planID, decisionID, file)...)
		projected[rule.TargetPath] = derived.After
	}
	return images, nil
}

func derivedGraphNodes(planSHA, planID, decisionID string, file projectedFile) []any {
	pathID := strings.TrimPrefix(digest([]byte(file.Path)), "sha256:")
	inputID := "urn:gooo:toolchain-projection:input:" + pathID + ":" + strings.TrimPrefix(file.InputSHA256, "sha256:")
	outputID := projectionOutputID(file.Path, file.OutputSHA256)
	sourceID := projectionOutputID(file.SourcePath, file.SourceOutputSHA256)
	activityID := "urn:gooo:toolchain-projection:derive:" + pathID + ":" + strings.TrimPrefix(planSHA, "sha256:")
	derivedFrom := []any{map[string]string{"@id": sourceID}}
	if file.Path == readmeTargetPath {
		derivedFrom = append(derivedFrom, map[string]string{"@id": inputID})
	}
	return []any{
		map[string]any{"@id": inputID, "@type": "prov:Entity", "gooo:path": file.Path, "gooo:content_sha256": file.InputSHA256},
		map[string]any{"@id": activityID, "@type": "prov:Activity", "gooo:rule_id": file.DerivationRule,
			"prov:used": []any{map[string]string{"@id": sourceID}, map[string]string{"@id": inputID},
				map[string]string{"@id": planID}, map[string]string{"@id": decisionID}}},
		map[string]any{"@id": outputID, "@type": "prov:Entity", "gooo:path": file.Path,
			"gooo:content_sha256": file.OutputSHA256, "prov:wasGeneratedBy": map[string]string{"@id": activityID},
			"prov:wasDerivedFrom": derivedFrom},
	}
}
