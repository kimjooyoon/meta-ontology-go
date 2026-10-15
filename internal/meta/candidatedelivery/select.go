package candidatedelivery

import (
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport/originpolicy"
)

func Select(runRaw, listRaw []byte, input Selection) Receipt {
	r := Unavailable(input, "DELIVERY_API_UNKNOWN")
	var list struct {
		Total     int        `json:"total_count"`
		Artifacts []Artifact `json:"artifacts"`
	}
	if json.Unmarshal(runRaw, &r.Run) != nil || json.Unmarshal(listRaw, &list) != nil || list.Artifacts == nil {
		return closeReceipt(r, "UNKNOWN", "DELIVERY_API_UNKNOWN")
	}
	if input.Repository == "" || input.RunID <= 0 || input.Attempt <= 0 {
		return closeReceipt(r, "UNKNOWN", "DELIVERY_REQUEST_UNKNOWN")
	}
	if r.Run.ID == 0 || r.Run.Attempt == 0 || r.Run.Head == "" || r.Run.Repository.Name == "" ||
		r.Run.Event == "" || r.Run.Path == "" || r.Run.Status == "" || r.Run.Conclusion == "" {
		return closeReceipt(r, "UNKNOWN", "DELIVERY_RUN_ORIGIN_UNKNOWN")
	}
	if r.Run.ID != input.RunID || r.Run.Attempt != input.Attempt || r.Run.Repository.Name != input.Repository ||
		r.Run.Path != ".github/workflows/self-improvement-language-candidate.yml" ||
		r.Run.Status != "completed" || r.Run.Conclusion != "success" || !validSHA(r.Run.Head) {
		return closeReceipt(r, "REFUTED", "DELIVERY_RUN_MISMATCH")
	}
	if list.Total != len(list.Artifacts) {
		return closeReceipt(r, "UNKNOWN", "DELIVERY_POPULATION_INCOMPLETE")
	}
	var candidates []Artifact
	for _, a := range list.Artifacts {
		if originpolicy.CandidateArtifactRole(a.Name) != "" {
			candidates = append(candidates, a)
		}
	}
	if len(candidates) == 0 {
		return closeReceipt(r, "UNKNOWN", "DELIVERY_ARTIFACT_MISSING")
	}
	if len(candidates) != 1 {
		return closeReceipt(r, "REFUTED", "DELIVERY_ROLE_DUPLICATED")
	}
	r.Artifact = candidates[0]
	if r.Artifact.Run.ID == 0 || r.Artifact.Run.Head == "" || r.Artifact.ID == 0 || r.Artifact.Size == 0 || r.Artifact.Digest == "" {
		return closeReceipt(r, "UNKNOWN", "DELIVERY_ARTIFACT_METADATA_UNKNOWN")
	}
	if r.Artifact.Run.ID != r.Run.ID || r.Artifact.Run.Head != r.Run.Head ||
		r.Artifact.ID <= 0 || r.Artifact.Size <= 0 || !validDigest(r.Artifact.Digest) {
		return closeReceipt(r, "REFUTED", "DELIVERY_ARTIFACT_RUN_MISMATCH")
	}
	if r.Artifact.Expired {
		return closeReceipt(r, "UNKNOWN", "DELIVERY_ARTIFACT_EXPIRED")
	}
	return closeReceipt(r, "LOCATED", "DELIVERY_ARTIFACT_LOCATED")
}

func Unavailable(input Selection, reason string) Receipt {
	return closeReceipt(Receipt{Schema: Schema, Repository: input.Repository, Expected: input,
		PolicyDigest: originpolicy.SourceDigest()}, "UNKNOWN", reason)
}

func closeReceipt(r Receipt, state, reason string) Receipt {
	r.State, r.Reason, r.Unknown, r.Digest = state, reason, nil, ""
	if state == "UNKNOWN" {
		r.Unknown = &selfimprovementtransport.CausalUnknown{Stage: "DELIVERY", Step: "bind-candidate-origin",
			Reason: reason, UnknownClass: "DIRECT_MISSING", NextOperation: "observe_exact_candidate_delivery",
			BlockedBy: []string{"candidate-delivery"}}
		if strings.Contains(reason, "EXPIRED") {
			r.Unknown.UnknownClass = "EXPIRED"
		}
	}
	r.Digest = digestJSON(r)
	return r
}

func validSHA(value string) bool { return validHex(value, 40) }
func validDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:") && validHex(strings.TrimPrefix(value, "sha256:"), 64)
}
func validHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, ch := range value {
		if !(ch >= 'a' && ch <= 'f' || ch >= '0' && ch <= '9') {
			return false
		}
	}
	return true
}

func stringsUnknown(reason string) bool {
	return strings.HasSuffix(reason, "_UNKNOWN") || reason == "PRODUCER_EVENT_UNSUPPORTED"
}
