package main

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"os"
	"path/filepath"
	"testing"
)

func TestRunAnalyzeDefersImplementationDetailsAndHelpers(t *testing.T) {
	_, generated := billingAnalyzeFiles(t, billingAnalyzeAuthority)
	source, err := os.ReadFile(generated)
	if err != nil {
		t.Fatal(err)
	}
	source = bytes.Replace(source, []byte("package billing\n"), []byte("package billing\n\nimport \"strings\"\n\nfunc helper(Order) {}\n"), 1)
	source = bytes.Replace(source, []byte("\treturn Payment{}"), []byte("\tnormalized := strings.TrimSpace(\"order\")\n\t_ = normalized\n\thelper()\n\treturn Payment{}"), 1)
	source = bytes.Replace(source, []byte("func helper(Order) {}"), []byte("func helper() {}"), 1)
	if err := os.WriteFile(generated, source, 0o640); err != nil {
		t.Fatal(err)
	}
	authority := filepath.Join(filepath.Dir(generated), "authority.gooo")
	if err := os.WriteFile(authority, []byte(billingAnalyzeAuthority), 0o640); err != nil {
		t.Fatal(err)
	}
	output, code, stderr := runAnalyzePaths(authority, generated)
	if code != exitOK || stderr != "" {
		t.Fatalf("implementation-detail analyze = %d, stderr=%q, output=%s", code, stderr, output)
	}
	var delta analyzeDeltaOutput
	if err := json.Unmarshal(output, &delta); err != nil {
		t.Fatal(err)
	}
	if len(delta.DeferredDetails) < 3 {
		t.Fatalf("deferred details = %#v, want strings.TrimSpace and helper", delta.DeferredDetails)
	}
	wanted := map[string]bool{"strings.TrimSpace": false, "helper": false}
	for _, detail := range delta.DeferredDetails {
		if _, ok := wanted[detail.Detail.Reference]; ok {
			wanted[detail.Detail.Reference] = true
		} else if detail.Detail.Reference != "normalized" {
			t.Fatalf("unexpected implementation detail: %#v", detail)
		}
	}
	for reference, found := range wanted {
		if !found {
			t.Fatalf("deferred details omitted %q: %#v", reference, delta.DeferredDetails)
		}
	}
	if len(delta.SignatureFacts) != 3 || len(delta.DeferredImplementation) != 1 {
		t.Fatalf("implementation detail changed authoritative classes: %#v", delta)
	}
}
func TestRunAnalyzeRetainsAmbiguousRegistryCandidates(t *testing.T) {
	root := t.TempDir()
	authority, _ := writeAnalyzeFile(t, filepath.Join(root, "authority.gooo"), billingAnalyzeAmbiguousAuthority)
	goPath, _ := writeAnalyzeFile(t, filepath.Join(root, "annotated.go"), billingAnalyzeAmbiguousGo)
	output, code, stderr := runAnalyzePaths(authority, goPath)
	if code != exitOK || stderr != "" {
		t.Fatalf("ambiguous analyze = %d, stderr=%q, output=%s", code, stderr, output)
	}
	var delta analyzeDeltaOutput
	if err := json.Unmarshal(output, &delta); err != nil {
		t.Fatal(err)
	}
	if len(delta.CandidateFacts) != 1 || len(delta.CandidateFacts[0].Options) != 2 || len(delta.CandidateFacts[0].Facts) == 0 || delta.CandidateFacts[0].Facts[0].Status != semantic.FactCandidate {
		t.Fatalf("ambiguous candidate was not retained: %#v", delta.CandidateFacts)
	}
	if len(delta.SignatureFacts) != 1 {
		t.Fatalf("ambiguous candidate altered deterministic signature set: %#v", delta.SignatureFacts)
	}
}
