package candidatedelivery

import (
	"encoding/json"
	"os"
	"testing"

	candidate "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
)

func TestEligibilityPreservesUnknownFresh1613(t *testing.T) {
	resolution := candidate.BuildUnavailableAuthorizationResolution(candidate.ArtifactMetadata{Repository: repo1613, RunID: 9161300})
	raw, _ := json.Marshal(resolution)
	got := ObserveEligibility(nil, raw, nil)
	var unknown map[string]any
	_ = json.Unmarshal(got.Unknown, &unknown)
	if got.State != "UNKNOWN" || got.Continue || len(unknown) != 6 || got.Reason != resolution.Reason {
		t.Fatalf("unavailable authorization was not preserved: %+v", got)
	}
}

func TestEligibilityVerificationFresh1613(t *testing.T) {
	b, _ := bundle1613(t)
	located, archive := selected1613(t, b)
	delivery, _ := Consume(located, archive)
	request, err := candidate.BuildAuthorizationRequest(os.DirFS("../../.."), "examples/self-improvement/authorization.gooo", b.Candidate, Metadata(delivery))
	if err != nil {
		t.Fatal(err)
	}
	resolution := candidate.ResolveAuthorization(request)
	verification, err := candidate.BuildAuthorizationVerification(request, resolution)
	if err != nil {
		t.Fatal(err)
	}
	requestRaw, _ := json.Marshal(request)
	resolutionRaw, _ := json.Marshal(resolution)
	for _, test := range []struct {
		name   string
		change func(*candidate.AuthorizationVerification)
		state  string
	}{
		{"verified", func(*candidate.AuthorizationVerification) {}, "CLOSED"},
		{"changed-value-retained-digest", func(v *candidate.AuthorizationVerification) { v.DecisionVerified = false }, "REFUTED"},
		{"changed-digest", func(v *candidate.AuthorizationVerification) { v.Digest = Digest([]byte("different")) }, "REFUTED"},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := verification
			test.change(&value)
			raw, _ := json.Marshal(value)
			got := ObserveEligibility(requestRaw, resolutionRaw, raw)
			if got.State != test.state || got.Continue != (test.state == "CLOSED") {
				t.Fatalf("eligibility: %+v", got)
			}
		})
	}
}
