package selfimprovementtransport

import (
	"encoding/json"
	"strings"
	"testing"
)

type originCase1610 struct {
	name, event, run, state, reason  string
	payloadMismatch, archiveMismatch bool
}

func TestTransportOriginLifecycleFresh1610(t *testing.T) {
	a, b, c := strings.Repeat("a", 40), strings.Repeat("b", 40), strings.Repeat("c", 40)
	cases := []originCase1610{
		{name: "completion observes other source", event: "workflow_run", run: b, state: ResolutionClosed},
		{name: "PR workflow merge commit differs", event: "pull_request", run: a, state: ResolutionClosed},
		{name: "push binds source", event: "push", run: a, state: ResolutionClosed},
		{name: "completion contradicts workflow", event: "workflow_run", run: c, state: ResolutionRefuted,
			reason: "PRODUCER_WORKFLOW_RUN_MISMATCH"},
		{name: "PR contradicts subject", event: "pull_request", run: b, state: ResolutionRefuted,
			reason: "PRODUCER_SUBJECT_RUN_MISMATCH"},
		{name: "missing event", run: a, state: ResolutionUnknown, reason: "PRODUCER_EVENT_UNKNOWN"},
		{name: "unmodeled event", event: "schedule", run: a, state: ResolutionUnknown,
			reason: "PRODUCER_EVENT_UNSUPPORTED"},
		{name: "completion still binds payload", event: "workflow_run", run: b, state: ResolutionRefuted,
			reason: "PRODUCER_PAYLOAD_DIGEST_MISMATCH", payloadMismatch: true},
		{name: "archive contradiction dominates missing event", run: a, state: ResolutionRefuted,
			reason: "PRODUCER_ARTIFACT_DIGEST_MISMATCH", archiveMismatch: true},
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) { checkOriginCase1610(t, row) })
	}
}

func checkOriginCase1610(t *testing.T, row originCase1610) {
	t.Helper()
	repository, observation, producer, _, _ := fixture(t)
	if row.payloadMismatch {
		observation = []byte(strings.ReplaceAll(string(observation), strings.Repeat("a", 40), strings.Repeat("d", 40)))
	}
	archive := zipArchive(t, map[string][][]byte{"producer.json": {producer}, "first.json": {observation}})
	metadata := originMetadata1610(t, row, archive)
	actual := digestBytes(archive)
	if row.archiveMismatch {
		actual = digestBytes([]byte("contradicting archive"))
	}
	metadataRaw, _ := json.Marshal(metadata)
	report := Evaluate(repository, "transport.gooo", metadata.Repository, 101, observation, producer, metadataRaw, actual)
	if err := ValidateReport(report); err != nil {
		t.Fatal(err)
	}
	if report.ProvenanceState != row.state || (row.reason != "" && report.Provenance.Reason != row.reason) {
		t.Fatalf("unexpected resolution: %+v", report.Provenance)
	}
	if row.state == ResolutionClosed {
		if err := CheckReadOnly(report); err != nil {
			t.Fatal(err)
		}
	} else if CheckReadOnly(report) == nil {
		t.Fatal("unresolved or contradictory origin accepted")
	}
	if !validDigest(report.Provenance.RunOrigin.PolicyDigest) || report.Provenance.RunOrigin.Event != row.event {
		t.Fatal("source policy or event provenance lost")
	}
}

func originMetadata1610(t *testing.T, row originCase1610, archive []byte) TransportMetadata {
	t.Helper()
	repository, _, _, _, _ := fixture(t)
	run := workflowRunAPI{ID: 101, RunAttempt: 2, HeadSHA: row.run, Event: row.event,
		Path: ".github/workflows/observation.yml"}
	runRaw, _ := json.Marshal(run)
	listRaw, _ := json.Marshal(artifactListAPI{Artifacts: []artifactAPI{{ID: 202, Name: ArtifactName,
		Digest: digestBytes(archive), SizeInBytes: int64(len(archive)),
		WorkflowRun: artifactWorkflowRunAPI{ID: 101, HeadSHA: row.run}}}})
	selection := ArtifactSelectionInput{Repository: "kimjooyoon/meta-ontology-go", ExpectedRunID: 101,
		ExpectedRunAttempt: 2, ArtifactName: ArtifactName}
	selected, err := SelectTransportMetadata(runRaw, listRaw, selection)
	if err != nil || selected.ProducerEvent != row.event {
		t.Fatalf("selector lost API event: %v", err)
	}
	_, receipt := ObserveArtifactLifecycle(repository, "transport.gooo", runRaw, listRaw,
		ArtifactLifecycleInput{Selection: selection})
	receipt = CompleteArtifactLifecycle(receipt, archive, 0)
	if receipt.Decision != DecisionPass || receipt.ProducerEvent != row.event {
		t.Fatalf("lifecycle lost API event: %+v", receipt)
	}
	metadata, err := BindTransportMetadata(receipt, archive)
	if err != nil || metadata.ProducerEvent != row.event {
		t.Fatalf("archive binding lost API event: %v", err)
	}
	return metadata
}
