package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func TestGuardAdapterRequiresExplicitPipelineOptIn(t *testing.T) {
	sourcePath := filepath.Join("..", "..", "..", "internal", "meta", "policycompilation", "testdata", "go-error-guard", "profile-original.go.golden")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	handler := "{\n\t\t\treturn exitFailure\n\t\t}"
	program := fmt.Sprintf("package goerrorguard\nnamespace goerrorguard\n"+
		"entity Source id \"gooo://error-guard/source\"\n"+
		"entity RawCandidate id \"gooo://error-guard/raw-candidate\"\n"+
		"entity Candidate id \"gooo://error-guard/candidate\"\n"+
		"activity GuardWrite(Source) -> RawCandidate computes \"go-error-guard:v2;function=runMetaPolicyGenerationProfile;writer=stdout;writer-type=interfaceWriter;source=%s;handler=%s\"\n"+
		"activity Canonicalize(RawCandidate) -> Candidate computes \"go-source-format:v1;toolchain=%s\"\n",
		policycompilation.DigestBytes(source), policycompilation.DigestBytes([]byte(handler)), runtime.Version())
	programPath := filepath.Join(t.TempDir(), "pipeline.gooo")
	if err := os.WriteFile(programPath, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-program", programPath, "-source", sourcePath}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != 1 {
		t.Fatalf("default adapter guessed pipeline: exit=%d stdout=%s", code, stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(append([]string{"-pipeline"}, args...), &stdout, &stderr); code != 0 {
		t.Fatalf("pipeline adapter failed: exit=%d stderr=%s", code, stderr.String())
	}
	var report policycompilation.GoGuardPipelineProposal
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.State != "PROPOSED" || report.Rendering == nil || report.Rendering.State != "RENDERED" ||
		report.CandidateSource == "" || report.Admission.State != "UNKNOWN" ||
		report.MutationAuthority != 0 || report.PromotionAuthority != 0 {
		t.Fatalf("adapter invented acceptance or lost canonical code: %+v", report)
	}
}

func TestGuardAdapterDefaultSingleActivityStillProposes(t *testing.T) {
	sourcePath := filepath.Join("..", "..", "..", "internal", "meta", "policycompilation", "testdata", "go-error-guard", "original.go.golden")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	handler := "{\n\t\t\tfmt.Fprintf(stderr, \"gooo: run package: %v\\n\", err)\n\t\t\treturn exitFailure\n\t\t}"
	program := fmt.Sprintf("package goerrorguard\nnamespace goerrorguard\n"+
		"entity Source id \"gooo://error-guard/source\"\n"+
		"entity Candidate id \"gooo://error-guard/candidate\"\n"+
		"activity GuardWrite(Source) -> Candidate computes \"go-error-guard:v1;function=writeSourcePackageResult;writer=stdout;diagnostic=stderr;source=%s;handler=%s\"\n",
		policycompilation.DigestBytes(source), policycompilation.DigestBytes([]byte(handler)))
	programPath := filepath.Join(t.TempDir(), "guard.gooo")
	if err := os.WriteFile(programPath, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-program", programPath, "-source", sourcePath}, &stdout, &stderr); code != 0 {
		t.Fatalf("default guard adapter changed: exit=%d stderr=%s", code, stderr.String())
	}
	var report policycompilation.GoErrorGuardProposal
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.State != "PROPOSED" || report.CandidateSource == "" || report.Admission.State != "UNKNOWN" {
		t.Fatalf("default proposal changed: %+v", report)
	}
}
