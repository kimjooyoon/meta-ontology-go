package executioncontinuation

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/candidatedelivery"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"
	candidate "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
	observation "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementobservation"
	transport "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
	"os"
	"strings"
	"testing"
)

func observation1617() []byte {
	d := Digest([]byte("fresh consumer 1617"))
	o := observation.Observation{Schema: transport.ObservationSchema, Metaprogram: "internal/meta/selfimprovementobservation",
		SubjectSHA: strings.Repeat("1", 40), SourceWorkflowRunID: 9161301, ContractID: "billing-operation-manifest-v2",
		Decision: "OBSERVED", Resolution: "EXACT", Reason: "READ_ONLY_MINIMAL_VALUE_BOUND",
		Interpretation: "READ_ONLY_IMPROVEMENT_INPUT", InputDigest: d,
		Summary: observation.ObservationSummary{Coordinates: observation.CountSummary{Satisfied: 16, Total: 16, BasisPoints: 10000},
			SourceCoordinates:   observation.CountSummary{Satisfied: 15, Total: 15, BasisPoints: 10000},
			Counterexamples:     observation.CountSummary{Satisfied: 6, Total: 6, BasisPoints: 10000},
			GoooDefinitionFiles: 2, ResourceSamples: 5, MaxWallMS: 14, MaxRSSKiB: 12000, BinaryBytes: 12000000},
		NotClaimed: []string{"business correctness", "value-level computation", "production readiness",
			"performance beyond this runner and fixed sample set", "general-purpose code generation",
			"improvement candidate quality", "automatic execution or adoption"}}
	for range 3 {
		o.Artifacts = append(o.Artifacts, observation.ArtifactRef{Kind: "artifact", Schema: "schema", FileDigest: d, SemanticDigest: d, Decision: "PASS"})
	}
	for range 16 {
		o.Indicators = append(o.Indicators, observation.Indicator{ID: "fresh1613", Class: "DRIVER", ProofChoice: "FOUNDATION", MetaOperation: "observe", Value: 1, Target: 1, Satisfied: true})
	}
	for _, view := range []struct {
		audience string
		count    int
	}{{"USER", 5}, {"TOOL_AUTHOR", 12}, {"GOVERNOR", 16}} {
		audience, count := view.audience, view.count
		ids := make([]string, count)
		for i := range ids {
			ids[i] = "fresh1613"
		}
		o.Views = append(o.Views, observation.View{Audience: audience, Resolution: "view", Satisfied: count, Total: count, BasisPoints: 10000, IndicatorIDs: ids})
	}
	for _, choice := range []string{"FOUNDATION", "COHERENCE", "REGRESSION"} {
		o.Proofs = append(o.Proofs, observation.Proof{Choice: choice, Claim: "claim", MetaOperation: "observe", EvidenceDigest: d, Passed: true})
	}
	o.Digest = digestValue(o)
	raw, _ := json.Marshal(o)
	return raw
}

type fixture1617 struct {
	prepared                           Prepared
	contractArchive, grantArchive      []byte
	grantRun, contractRun, consumerRun Run
	grantArtifact, contractArtifact    candidatedelivery.Artifact
}

func fixtureConsumer1617(t *testing.T) fixture1617 {
	t.Helper()
	root := os.DirFS("../../..")
	subject := strings.Repeat("1", 40)
	cr := candidate.Evaluate(root, "examples/self-improvement/candidate.gooo", subject, 9161301, observation1617())
	if cr.Decision != "PROPOSED" {
		t.Fatal(cr.Reason)
	}
	raw, _ := json.Marshal(cr)
	req, err := candidate.BuildAuthorizationRequest(root, "examples/self-improvement/authorization.gooo", raw, candidate.ArtifactMetadata{Repository: "kimjooyoon/meta-ontology-go", SubjectSHA: subject, RunID: 9161701, RunAttempt: 1, ArtifactID: 9161702, ArtifactName: "candidate", ArchiveDigest: Digest([]byte("candidate zip")), SizeBytes: 100})
	if err != nil {
		t.Fatal(err)
	}
	auth := candidate.ResolveAuthorization(req)
	cp, err := contract.CompilePolicy(root, contract.PolicyPath)
	if err != nil {
		t.Fatal(err)
	}
	ci := contract.ProjectAuthorizationRequest(req, contract.KnownRegistry())
	resolution := contract.Evaluate(cp, ci)
	live := contract.LiveReport{ContractResolution: resolution, Verification: contract.Verify(cp, ci, resolution)}
	if resolution.Decision != contract.DecisionClosed || !live.Verification.Verified {
		t.Fatal("fresh contract", resolution.Reason)
	}
	cb := zipJSON1617(t, map[string]any{"live-contract.json": live, "source-authorization-request.json": req, "source-authorization-resolution.json": auth})
	gp, err := grant.CompilePolicy(root, grant.PolicyPath)
	if err != nil {
		t.Fatal(err)
	}
	source := grant.SourceArtifact{Repository: "kimjooyoon/meta-ontology-go", WorkflowRunID: 9161720, WorkflowRunAttempt: 1, ArtifactID: 9161721, ArtifactDigest: Digest(cb), ObservedArtifactDigest: Digest(cb), ArtifactExpiryKnown: true, ArtifactRetrieved: true}
	gr := grant.BuildRequest(gp, grant.ProjectV24(req, auth), grant.ProjectV25(resolution), source)
	gi := grant.GrantInput{Request: gr, Live: true}
	resolved := grant.Evaluate(gp, gi)
	g := grant.LiveReport{Schema: grant.Schema, Policy: gp.Evidence, Request: gr, GrantDecision: grant.DecisionAllow, DecisionSource: grant.DecisionSourceSystem, Resolution: resolved, Verification: grant.Verify(gp, gi, resolved), Metrics: resolved.Metrics}
	g.Digest = digestValue(g)
	gb := zipJSON1617(t, map[string]any{"live-grant.json": g})
	e := Envelope{Repository: source.Repository, WorkflowSHA: strings.Repeat("b", 40), SubjectSHA: subject, RunID: 9161730, Attempt: 1, ArtifactID: 9161731, ArchiveDigest: Digest(gb), RequestDigest: gr.Digest, Wait: "CREATED", PolicyDigest: policy.SourceDigest()}
	m := Metadata{Source: e, ConsumerRunID: 9161740, ConsumerAttempt: 1, ConsumerWorkflowSHA: strings.Repeat("c", 40), ContractRunID: source.WorkflowRunID, ContractAttempt: 1, ContractArtifactID: source.ArtifactID, ContractDigest: source.ArtifactDigest, ContractWorkflowSHA: strings.Repeat("a", 40), SubjectBlobSHA: strings.Repeat("d", 40), SubjectFileDigest: resolution.ExecutionInput.Source.Digest, PolicyDigest: policy.SourceDigest()}
	m.Digest = digestValue(m)
	p := Prepared{Pair: Pair{Grant: g, Contract: live, AuthorizationRequest: req, AuthorizationResolution: auth}, Metadata: m, SourceBytes: resolution.ExecutionInput.Source.Bytes}
	if _, err := validatePrepared(root, p); err != nil {
		t.Fatal("fresh source pair", err)
	}
	f := fixture1617{prepared: p, contractArchive: cb, grantArchive: gb, grantRun: run1617("grant", e.RunID, e.WorkflowSHA), contractRun: run1617("contract", m.ContractRunID, m.ContractWorkflowSHA), consumerRun: run1617("consumer", m.ConsumerRunID, m.ConsumerWorkflowSHA)}
	f.grantArtifact = artifact1617(f.grantRun, e.ArtifactID, policy.ArtifactLabel("grant", subject), gb)
	f.contractArtifact = artifact1617(f.contractRun, m.ContractArtifactID, policy.ArtifactLabel("contract", subject), cb)
	return f
}

func run1617(kind string, id int64, head string) Run {
	r := Run{Run: candidatedelivery.Run{ID: id, Attempt: 1, Head: head, Event: "workflow_dispatch", Path: policy.WorkflowPath(kind), Status: "completed", Conclusion: "success"}, Branch: "dev"}
	r.Repository.Name = "kimjooyoon/meta-ontology-go"
	return r
}
func artifact1617(r Run, id int64, name string, raw []byte) candidatedelivery.Artifact {
	a := candidatedelivery.Artifact{ID: id, Name: name, Digest: Digest(raw), Size: int64(len(raw))}
	a.Run.ID = r.ID
	a.Run.Head = r.Head
	return a
}
func zipJSON1617(t *testing.T, files map[string]any) []byte {
	t.Helper()
	var raw bytes.Buffer
	z := zip.NewWriter(&raw)
	for _, name := range []string{"live-contract.json", "source-authorization-request.json", "source-authorization-resolution.json", "live-grant.json", "reservation.json"} {
		v, ok := files[name]
		if !ok {
			continue
		}
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		w, err := z.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(b); err != nil {
			t.Fatal(err)
		}
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return raw.Bytes()
}
