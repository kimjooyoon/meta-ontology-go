package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compilercompatibility"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type compatibilityRepositoryReader struct{}

func (compatibilityRepositoryReader) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join("../..", name))
}

func TestCompatibilityCertificateSurvivesJSONTransport(t *testing.T) {
	reader := rootedContinuityReader{root: "../..", files: map[string][]byte{}}
	source := []byte(`package contexttransport
namespace contexttransport
entity Request id "contexttransport://entity/request"
entity Rule id "contexttransport://entity/rule"
entity Result id "contexttransport://entity/result"
activity Process(Request, Rule) -> Result
`)
	reader.files["transport.gooo"] = source
	policyData, err := reader.ReadFile(compilercompatibility.CanonicalPolicyPath)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := compatibilitypolicy.Load(compilercompatibility.CanonicalPolicyPath, policyData)
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.ParseFile("transport.gooo", string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	state := compatibilityGenerationState{
		policy: policy, certificate: compilercompatibility.Certificate{CandidateStableID: "synthetic-transport"},
	}
	options := generateOptions{filename: "transport.gooo", outputDir: t.TempDir()}
	replay, err := buildCompatibilityReplay(options, generateInput{source: source, file: file},
		reader, time.Now().Add(commandDeadline), state)
	if err != nil {
		t.Fatal(err)
	}
	// Synthetic receipt fixtures isolate transport from the separate CI test runs.
	successor := replay.current
	authorization, err := compilercompatibility.BuildAuthorization(successor.CandidateStableID,
		successor.SubjectDigest, successor.CompilerImplementationDigest)
	if err != nil {
		t.Fatal(err)
	}
	successor.Role, successor.TestContractResult = "successor", "PASS"
	successor.AuthorizationDigest = authorization.AuthorizationID
	predecessor := successor
	predecessor.Role = "predecessor"
	predecessor.CompilerImplementationDigest = cache.HashBytes([]byte("synthetic predecessor")).String()
	certificate, err := compilercompatibility.BuildCertificate(predecessor, successor, authorization, policy)
	if err != nil {
		t.Fatal(err)
	}
	reader.files["transport-certificate.json"], err = json.MarshalIndent(certificate, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	var out, errors bytes.Buffer
	args := []string{"transport.gooo", "--out", options.outputDir,
		"--compatibility-certificate", "transport-certificate.json", "--json"}
	if code := runGenerate(args, reader, EntityFieldsCLIParser{}, &out, &errors); code != exitOK {
		t.Fatalf("transported certificate rejected: %d %s %s", code, &out, &errors)
	}
}

func TestCompatibilityReplayRetainsOrdinaryGenerationIdentity(t *testing.T) {
	source := []byte(`package contextreuse
namespace contextreuse
entity Request id "contextreuse://entity/request"
entity Rule id "contextreuse://entity/rule"
entity Result id "contextreuse://entity/result"
activity Process(Request, Rule) -> Result
`)
	file, diagnostics := syntax.ParseFile("identity.gooo", string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	input := generateInput{source: source, file: file}
	options := generateOptions{filename: "identity.gooo", outputDir: t.TempDir()}
	var out, errors bytes.Buffer
	deadline := time.Now().Add(commandDeadline)
	ordinary, code := buildGenerateArtifacts(options, input, true, &out, &errors, deadline)
	if code != exitOK {
		t.Fatal(code, &out, &errors)
	}
	ordinary.manifest.GeneratedFile = generatedFileName
	want, err := jsonManifestBytes(ordinary.manifest)
	if err != nil {
		t.Fatal(err)
	}
	state := compatibilityGenerationState{certificate: compilercompatibility.Certificate{CandidateStableID: "synthetic"}}
	replay, err := buildCompatibilityReplay(options, input, compatibilityRepositoryReader{}, deadline, state)
	if err != nil {
		t.Fatal(err)
	}
	if replay.current.SemanticIRDigest != ordinary.ir.StableHash() ||
		!bytes.Equal(replay.current.GeneratedSource, ordinary.result.Source) ||
		!bytes.Equal(replay.current.GeneratedManifest, want) {
		actual, _ := json.Marshal(replay.current)
		t.Fatalf("ordinary generation and consumer identities differ: manifest=%s replay=%s", want, actual)
	}
}

func TestCompatibilityRejectsUnboundPreviousGo(t *testing.T) {
	reader := rootedContinuityReader{root: "../..", files: map[string][]byte{}}
	source := []byte(`package contexttransport
namespace contexttransport
entity Request id "contexttransport://entity/request"
entity Rule id "contexttransport://entity/rule"
entity Result id "contexttransport://entity/result"
activity Process(Request, Rule) -> Result
`)
	reader.files["transport.gooo"] = source
	policyData, err := reader.ReadFile(compilercompatibility.CanonicalPolicyPath)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := compatibilitypolicy.Load(compilercompatibility.CanonicalPolicyPath, policyData)
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.ParseFile("transport.gooo", string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	state := compatibilityGenerationState{
		policy: policy, certificate: compilercompatibility.Certificate{CandidateStableID: "synthetic-transport"},
	}
	options := generateOptions{filename: "transport.gooo", outputDir: t.TempDir()}
	replay, err := buildCompatibilityReplay(options, generateInput{source: source, file: file},
		reader, time.Now().Add(commandDeadline), state)
	if err != nil {
		t.Fatal(err)
	}
	// Synthetic receipt fixtures isolate transport from the separate CI test runs.
	successor := replay.current
	authorization, err := compilercompatibility.BuildAuthorization(successor.CandidateStableID,
		successor.SubjectDigest, successor.CompilerImplementationDigest)
	if err != nil {
		t.Fatal(err)
	}
	successor.Role, successor.TestContractResult = "successor", "PASS"
	successor.AuthorizationDigest = authorization.AuthorizationID
	predecessor := successor
	predecessor.Role = "predecessor"
	predecessor.CompilerImplementationDigest = cache.HashBytes([]byte("synthetic predecessor")).String()
	certificate, err := compilercompatibility.BuildCertificate(predecessor, successor, authorization, policy)
	if err != nil {
		t.Fatal(err)
	}
	reader.files["transport-certificate.json"], err = json.MarshalIndent(certificate, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	reader.files["previous.go"] = []byte("package handwritten\n// caller-owned content\n")
	var out, errors bytes.Buffer
	args := []string{"transport.gooo", "--out", options.outputDir,
		"--compatibility-certificate", "transport-certificate.json", "--previous-go", "previous.go", "--json"}
	code := runGenerate(args, reader, EntityFieldsCLIParser{}, &out, &errors)
	if code != exitUsage || !bytes.Contains(out.Bytes(), []byte("--previous-go")) {
		t.Fatalf("expected actionable usage rejection: %d %s %s", code, &out, &errors)
	}
	entries, err := os.ReadDir(options.outputDir)
	if err != nil || len(entries) != 0 {
		t.Fatal("unsupported combination emitted output", entries, err)
	}
}
