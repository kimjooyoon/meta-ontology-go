package couplingexplain

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateConnectedPath(path PathSummary, ownerID, termID string) *envelopeIssue {
	visited := map[string]struct{}{path.StartID: {}}
	for index, step := range path.Steps {
		if step.FromID == "" || step.ToID == "" || step.FromID == step.ToID || !step.Kind.Valid() ||
			!step.Phase.Phase.Valid() || !validDigest(step.InputDigest) || !validDigest(step.OutputDigest) {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonAmbiguous, code: "malformed-origin-path", ids: []string{path.PathID}}
		}
		if index == 0 && step.FromID != path.StartID {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonAmbiguous, code: "disconnected-origin-path", ids: []string{path.PathID}}
		}
		if index > 0 && step.FromID != path.Steps[index-1].ToID {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonAmbiguous, code: "disconnected-origin-path", ids: []string{path.PathID}}
		}
		if _, exists := visited[step.ToID]; exists {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonAmbiguous, code: "cyclic-origin-path", ids: []string{path.PathID}}
		}
		visited[step.ToID] = struct{}{}
		if step.Kind == semantic.InferenceObservationCandidate {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonAmbiguous, code: "candidate-in-verified-path", ids: []string{path.PathID}}
		}
	}
	last := path.Steps[len(path.Steps)-1]
	if path.Steps[0].ToID != ownerID {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonUnregistered, code: "origin-owner-mismatch", ids: []string{path.PathID}}
	}
	termReached := false
	for _, step := range path.Steps {
		if step.FromID == ownerID && step.ToID == termID {
			termReached = true
		}
	}
	if !termReached {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonMissing, code: "origin-term-missing", ids: []string{path.PathID}}
	}
	if last.ToID != path.EndID || last.Kind != semantic.InferenceIndependentVerification || last.EvidenceRef == "" {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonNotVerified, code: "origin-path-verifier-missing", ids: []string{path.PathID}}
	}
	return nil
}
func requestNoLink(binding SnapshotBinding, status Status, reason LinkReason, code string, ids ...string) Explanation {
	diagnostics := []Diagnostic{{Code: code, IDs: sortedStrings(ids)}}
	value := struct {
		Status      Status          `json:"status"`
		Binding     SnapshotBinding `json:"binding"`
		NoLink      NoLink          `json:"no_link"`
		Diagnostics []Diagnostic    `json:"diagnostics"`
	}{Status: status, Binding: binding, NoLink: NoLink{Reason: reason}, Diagnostics: diagnostics}
	data, _ := json.Marshal(value)
	return Explanation{Status: status, EvidenceDigest: DigestBytes(data), Binding: binding,
		NoLink: &NoLink{Reason: reason}, Diagnostics: diagnostics}
}
func validChangeClaim(value ChangeClaim) bool { return value == ClaimDelta || value == ClaimNoDelta }
func validReason(value LinkReason) bool {
	switch value {
	case ReasonAmbiguous, ReasonStale, ReasonUnregistered, ReasonMissing, ReasonNotVerified:
		return true
	default:
		return false
	}
}
