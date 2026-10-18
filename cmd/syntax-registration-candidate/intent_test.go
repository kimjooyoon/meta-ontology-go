package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func TestDirectIntentExportsPinnedCandidate(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/languageassurance/verticalsliceclosureshadow",
	} {
		if err := os.CopyFS(filepath.Join(root, directory), os.DirFS(filepath.Join("../..", directory))); err != nil {
			t.Fatal(err)
		}
	}
	corpus := "examples/language-syntax-roundtrip/corpus.json"
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, corpus)), 0700); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join("../..", corpus))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, corpus), raw, 0600); err != nil {
		t.Fatal(err)
	}
	versions, err := filepath.Glob(filepath.Join(root,
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v*.json"))
	if err != nil {
		t.Fatal(err)
	}
	baseVersion := 0
	for _, path := range versions {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var evidence struct {
			Version int `json:"version"`
		}
		if err := json.Unmarshal(raw, &evidence); err != nil {
			t.Fatal(err)
		}
		if evidence.Version > baseVersion {
			baseVersion = evidence.Version
		}
	}
	if baseVersion == 0 {
		t.Fatal("missing baseline denominator")
	}
	source := "examples/one-step-intent/main.gooo"
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, source)), 0700); err != nil {
		t.Fatal(err)
	}
	const program = "package onestep\nnamespace onestep\n" +
		"entity Observation id \"gooo://one-step/observation\" fields {\n" +
		"    field State id \"gooo://one-step/state\" type string required one\n}\n" +
		"activity Capture(Observation) -> Observation computes \"record.forward:v1\"\n"
	if err := os.WriteFile(filepath.Join(root, source), []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	intent, err := json.Marshal(syntaxregistration.CaseIntent{
		ID: "one-step-intent", Path: source, ProofChoice: "COHERENCE", Profile: "text-fields",
		BaseVersion: baseVersion,
	})
	if err != nil {
		t.Fatal(err)
	}
	intentPath := filepath.Join(t.TempDir(), "intent.json")
	if err := os.WriteFile(intentPath, intent, 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "candidate")
	if err := run(root, "", intentPath, output, false, time.Now()); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(filepath.Join(output, "candidate.json"))
	if err != nil {
		t.Fatal(err)
	}
	var candidate syntaxregistration.Candidate
	if err := json.Unmarshal(raw, &candidate); err != nil {
		t.Fatal(err)
	}
	if candidate.State != "PROPOSAL_ONLY" || candidate.Admission != "UNASSESSED" ||
		candidate.ApplyAuthorized || candidate.PromotionAllowed || candidate.RepositoryWrites != 0 ||
		candidate.Emitted != candidate.Required || candidate.RequiredArtifacts != 9 ||
		candidate.ExecutionBinding.Identity.ExecutableDigest == "" {
		t.Fatalf("direct intent lost proposal or source binding: %+v", candidate)
	}
	if current, err := os.ReadFile(filepath.Join(root, source)); err != nil || string(current) != program {
		t.Fatal("direct intent changed its source")
	}
}
