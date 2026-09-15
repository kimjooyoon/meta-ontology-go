package semantic

import (
	"fmt"
	"sort"
	"strings"
)

func (p InferencePathV1) Normalized() (InferencePathV1, error) {
	issues := &InferencePathErrors{}
	version := strings.TrimSpace(p.Version)
	if version != InferencePathSchemaVersion {
		issues.add("unknown-version", "", fmt.Sprintf("got %q", p.Version))
	}
	out := normalizeInferenceRecords(p, version, issues)
	validateInferencePathRecords(out, issues)
	if len(issues.Issues) != 0 {
		sortInferenceIssues(issues)
		return InferencePathV1{}, issues
	}
	return out, nil
}
func normalizeInferenceRecords(p InferencePathV1, version string, issues *InferencePathErrors) InferencePathV1 {
	out := InferencePathV1{
		Version: version, Edges: make([]InferenceEdge, 0, len(p.Edges)),
		Claims:   make([]SemanticChangeClaim, 0, len(p.Claims)),
		Evidence: make([]InferenceEvidence, 0, len(p.Evidence)),
	}
	for _, raw := range p.Evidence {
		normalized, err := raw.normalized()
		if err != nil {
			issues.add("evidence", raw.ID, err.Error())
			continue
		}
		out.Evidence = append(out.Evidence, normalized)
	}
	for _, raw := range p.Edges {
		normalized, err := raw.normalized()
		if err != nil {
			issues.add("edge", raw.RecordID, err.Error())
			continue
		}
		out.Edges = append(out.Edges, normalized)
	}
	for _, raw := range p.Claims {
		normalized, err := raw.normalized()
		if err != nil {
			issues.add("claim", raw.RecordID, err.Error())
			continue
		}
		out.Claims = append(out.Claims, normalized)
	}
	sort.Slice(out.Evidence, func(i, j int) bool { return out.Evidence[i].ID < out.Evidence[j].ID })
	sort.Slice(out.Edges, func(i, j int) bool { return out.Edges[i].RecordID < out.Edges[j].RecordID })
	sort.Slice(out.Claims, func(i, j int) bool { return out.Claims[i].RecordID < out.Claims[j].RecordID })
	return out
}
