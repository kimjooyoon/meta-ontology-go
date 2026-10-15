package candidatedelivery

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	observation "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementobservation"
	transport "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

const repo1613 = "kimjooyoon/meta-ontology-go"

func observation1613() []byte {
	d := Digest([]byte("fresh delivery 1613"))
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
	o.Digest = digestJSON(o)
	raw, _ := json.Marshal(o)
	return raw
}

func bundle1613(t *testing.T) (Bundle, Manifest) {
	t.Helper()
	root := os.DirFS("../../..")
	obs := observation1613()
	subject, workflow := strings.Repeat("1", 40), strings.Repeat("b", 40)
	candidate := selfimprovementcandidate.Evaluate(root, "examples/self-improvement/candidate.gooo", subject, 9161301, obs)
	if err := selfimprovementcandidate.Validate(candidate, subject, 9161301); err != nil || candidate.Decision != "PROPOSED" {
		t.Fatal("fresh candidate", candidate.Reason, err)
	}
	producer, err := transport.Produce(root, "examples/self-improvement/transport.gooo", transport.ProducerInput{
		Repository: repo1613, SubjectSHA: subject, CheckoutSHA: subject, WorkflowSHA: workflow,
		WorkflowRef: repo1613 + "/.github/workflows/self-improvement-language-observation.yml@refs/heads/dev",
		RunID:       9161302, RunAttempt: 1, Job: "observation", ArtifactName: transport.ArtifactName}, obs)
	if err != nil {
		t.Fatal(err)
	}
	producerRaw, _ := json.Marshal(producer)
	archive := archive1613(t, map[string][]byte{"first.json": obs, "producer.json": producerRaw})
	meta, err := transport.BindTransportMetadata(transport.LifecycleReceipt{Repository: repo1613, ExpectedRunID: 9161302,
		ExpectedRunAttempt: 1, ProducerEvent: "workflow_run", OrchestrationHeadSHA: workflow,
		WorkflowPath: ".github/workflows/self-improvement-language-observation.yml", ArtifactID: 9161303,
		ArtifactName: transport.ArtifactName, ArtifactDigest: Digest(archive), ArtifactSizeBytes: int64(len(archive)), ArtifactInstanceCount: 1, ArtifactTypeCount: 1}, archive)
	if err != nil {
		t.Fatal(err)
	}
	metadataRaw, _ := json.Marshal(meta)
	receipt := transport.Evaluate(root, "examples/self-improvement/transport.gooo", repo1613, 9161302, obs, producerRaw, metadataRaw, Digest(archive))
	if err := transport.CheckReadOnly(receipt); err != nil {
		t.Fatal(err)
	}
	cr, _ := json.Marshal(candidate)
	tr, _ := json.Marshal(receipt)
	bundle := Bundle{Candidate: cr, Transport: tr}
	manifest, err := Produce(Manifest{RunID: 9161310, Attempt: 2, Event: "workflow_run", SubjectSHA: subject,
		WorkflowSHA: workflow, RunHeadSHA: workflow}, "self-improvement-nonexecuting-candidate-"+subject, bundle)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest, _ = json.Marshal(manifest)
	return bundle, manifest
}

func archive1613(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var raw bytes.Buffer
	writer := zip.NewWriter(&raw)
	for _, name := range []string{"first.json", "producer.json", "output/first.json", "transport/receipt.json", "output/delivery.json"} {
		data, found := files[name]
		if !found {
			continue
		}
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = entry.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return raw.Bytes()
}

func selected1613(t *testing.T, bundle Bundle) (Receipt, []byte) {
	t.Helper()
	archive := archive1613(t, map[string][]byte{"output/first.json": bundle.Candidate, "transport/receipt.json": bundle.Transport, "output/delivery.json": bundle.Manifest})
	run := Run{ID: 9161310, Attempt: 2, Head: strings.Repeat("b", 40), Event: "workflow_run",
		Path: ".github/workflows/self-improvement-language-candidate.yml", Status: "completed", Conclusion: "success"}
	run.Repository.Name = repo1613
	a := Artifact{ID: 9161320, Name: "self-improvement-nonexecuting-candidate-" + strings.Repeat("1", 40), Digest: Digest(archive), Size: int64(len(archive))}
	a.Run.ID = run.ID
	a.Run.Head = run.Head
	raw, _ := json.Marshal(run)
	list, _ := json.Marshal(map[string]any{"total_count": 1, "artifacts": []Artifact{a}})
	return Select(raw, list, Selection{Repository: repo1613, RunID: run.ID, Attempt: run.Attempt}), archive
}
