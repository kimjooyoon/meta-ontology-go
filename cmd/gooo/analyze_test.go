package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestRunAnalyzeValidFixPlanContract(t *testing.T) {
	output, code, stderr := analyzeFixtureOutput(t, sourceOrderA)
	if code != exitOK || len(stderr) != 0 {
		t.Fatalf("analyze result = code %d, stderr=%q", code, stderr)
	}
	plan := decodeFixPlan(t, output)
	if plan.SchemaVersion != fixPlanSchemaVersion || plan.Status != fixPlanReady || plan.SourceDigest == "" || plan.GraphHash == "" {
		t.Fatalf("incomplete fix plan identity: %#v", plan)
	}
	if !bytes.Contains(output, []byte(`"diagnostics":[]`)) || bytes.Contains(output, []byte(`"diagnostics":null`)) {
		t.Fatalf("valid fix plan diagnostics is not an empty JSON array: %s", output)
	}
	if plan.IR.Status != "available" || plan.IR.SemanticDigest == "" || len(plan.Diagnostics) != 0 {
		t.Fatalf("unexpected valid plan state: %#v", plan)
	}
	if plan.Evidence.Status != "missing" || plan.Provenance.Status != "missing" {
		t.Fatalf("evidence status = %#v, provenance = %#v", plan.Evidence, plan.Provenance)
	}
	if plan.Repairs.Status != "deferred" || plan.GraphPatch.Status != "deferred" || plan.Workspace.Status != "deferred" || plan.SemanticLoop.Status != "deferred" {
		t.Fatalf("write or loop status was not deferred: %#v", plan)
	}
	wantAuthorities := graphAuthorities{
		GoooSource: "authoritative", SemanticIR: "authoritative", Handwritten: "authoritative",
		Provenance: "authoritative", Graph: "derived",
	}
	if plan.Authorities != wantAuthorities {
		t.Fatalf("authorities = %#v, want %#v", plan.Authorities, wantAuthorities)
	}
}

func TestRunAnalyzeSyntaxDiagnosticsAreStableAndOrdered(t *testing.T) {
	const malformed = "package billing\nentity Broken id \"x\" @"
	first, firstCode, firstErr := analyzeFixtureOutput(t, malformed)
	second, secondCode, secondErr := analyzeFixtureOutput(t, malformed)
	if firstCode != exitFailure || secondCode != exitFailure || len(firstErr) != 0 || len(secondErr) != 0 {
		t.Fatalf("syntax plan result = %d/%d, stderr=%q/%q", firstCode, secondCode, firstErr, secondErr)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("repeated syntax plan differs:\nfirst=%s\nsecond=%s", first, second)
	}
	plan := decodeFixPlan(t, first)
	if plan.Status != fixPlanSyntaxInvalid || plan.IR.Status != "unavailable" || len(plan.Diagnostics) == 0 {
		t.Fatalf("unexpected syntax plan: %#v", plan)
	}
	previousOffset := -1
	for _, diagnostic := range plan.Diagnostics {
		if diagnostic.Phase != "syntax" || diagnostic.RepairID == "" || diagnostic.Status != "deferred" || diagnostic.Applicability != "potential" {
			t.Fatalf("unexpected syntax diagnostic: %#v", diagnostic)
		}
		if diagnostic.Span.File != "fixture.gooo" || diagnostic.Span.Start.Offset < previousOffset {
			t.Fatalf("diagnostics are not source ordered: %#v", plan.Diagnostics)
		}
		previousOffset = diagnostic.Span.Start.Offset
	}
}

func TestRunAnalyzeSemanticInvalidInputIsDeferred(t *testing.T) {
	const semanticInvalid = `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Missing) -> Order
`
	output, code, stderr := analyzeFixtureOutput(t, semanticInvalid)
	if code != exitFailure || len(stderr) != 0 {
		t.Fatalf("semantic plan result = code %d, stderr=%q", code, stderr)
	}
	plan := decodeFixPlan(t, output)
	if plan.Status != fixPlanSemanticInvalid || plan.IR.Status != "unavailable" || len(plan.Diagnostics) != 1 {
		t.Fatalf("unexpected semantic plan: %#v", plan)
	}
	diagnostic := plan.Diagnostics[0]
	if diagnostic.Phase != "semantic" || diagnostic.Code != "semantic.lowering" || diagnostic.Span.File != "fixture.gooo" || diagnostic.Status != "deferred" || diagnostic.Applicability != "not-evaluated" {
		t.Fatalf("unexpected semantic diagnostic: %#v", diagnostic)
	}
	if plan.GraphHash != "" || plan.GraphPatch.Status != "deferred" || plan.Workspace.Status != "deferred" {
		t.Fatalf("semantic-invalid plan exposed unavailable output: %#v", plan)
	}
}

func TestRunAnalyzePermutationPreservesCanonicalPlan(t *testing.T) {
	first := decodeFixPlan(t, analyzeFixtureBytes(t, sourceOrderA))
	second := decodeFixPlan(t, analyzeFixtureBytes(t, sourceOrderB))
	if first.SourceDigest == second.SourceDigest || first.GraphHash != second.GraphHash || first.IR.SemanticDigest != second.IR.SemanticDigest {
		t.Fatalf("source permutation changed semantic plan identity: first=%#v second=%#v", first, second)
	}
	first.SourceDigest, second.SourceDigest = "", ""
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("canonical plan differs by source order:\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}

func TestRunAnalyzeIsReadOnly(t *testing.T) {
	directory := t.TempDir()
	filename := filepath.Join(directory, "input.gooo")
	if err := os.WriteFile(filename, []byte(sourceOrderA), 0o640); err != nil {
		t.Fatal(err)
	}
	beforeBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeEntries := directoryEntries(t, directory)
	output, code, stderr := runAnalyzeFile(filename)
	if code != exitOK || len(stderr) != 0 || len(output) == 0 {
		t.Fatalf("read-only analyze = code %d, stderr=%q, output=%d bytes", code, stderr, len(output))
	}
	afterBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeBytes, afterBytes) || !os.SameFile(beforeInfo, afterInfo) || beforeInfo.Mode() != afterInfo.Mode() || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("analyze changed its input file")
	}
	if afterEntries := directoryEntries(t, directory); !reflect.DeepEqual(beforeEntries, afterEntries) {
		t.Fatalf("analyze changed directory entries: before=%v after=%v", beforeEntries, afterEntries)
	}
}

func TestApplyFixPlanIREvidenceRefsAreInsertionIndependent(t *testing.T) {
	firstIR := lowerInspectFixtureIR(t)
	secondIR := lowerInspectFixtureIR(t)
	fact := firstIR.Graph.DeterministicFacts()[0]
	evidenceA := mustAnalyzeEvidence(t, "billing://evidence/a", fact.Key())
	evidenceB := mustAnalyzeEvidence(t, "billing://evidence/b", fact.Key())
	if err := firstIR.AddEvidence(evidenceB); err != nil {
		t.Fatal(err)
	}
	if err := firstIR.AddEvidence(evidenceA); err != nil {
		t.Fatal(err)
	}
	if err := secondIR.AddEvidence(evidenceA); err != nil {
		t.Fatal(err)
	}
	if err := secondIR.AddEvidence(evidenceB); err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.ParseFile("fixture.gooo", sourceOrderA)
	if diagnostics.Error() != nil {
		t.Fatal(diagnostics.Error())
	}
	firstPlan := newFixPlan([]byte(sourceOrderA), nil, file)
	secondPlan := newFixPlan([]byte(sourceOrderA), nil, file)
	applyFixPlanIR(&firstPlan, firstIR)
	applyFixPlanIR(&secondPlan, secondIR)
	if !reflect.DeepEqual(firstPlan.Evidence, secondPlan.Evidence) || !sort.StringsAreSorted(firstPlan.Evidence.Refs) {
		t.Fatalf("evidence refs depend on insertion order: first=%#v second=%#v", firstPlan.Evidence, secondPlan.Evidence)
	}
}

func analyzeFixtureOutput(t *testing.T, source string) ([]byte, int, []byte) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := runAnalyze([]string{"fixture.gooo"}, fixtureReader{source: source}, SyntaxSourceParser{}, &stdout, &stderr)
	return stdout.Bytes(), code, stderr.Bytes()
}

func analyzeFixtureBytes(t *testing.T, source string) []byte {
	t.Helper()
	output, code, stderr := analyzeFixtureOutput(t, source)
	if code != exitOK || len(stderr) != 0 {
		t.Fatalf("analyze fixture = code %d, stderr=%q", code, stderr)
	}
	return output
}

func decodeFixPlan(t *testing.T, output []byte) fixPlan {
	t.Helper()
	var plan fixPlan
	if err := json.Unmarshal(output, &plan); err != nil {
		t.Fatalf("decode fix plan: %v; output=%s", err, output)
	}
	return plan
}

func runAnalyzeFile(filename string) ([]byte, int, []byte) {
	var stdout, stderr bytes.Buffer
	code := runAnalyze([]string{filename}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr)
	return stdout.Bytes(), code, stderr.Bytes()
}

func mustAnalyzeEvidence(t *testing.T, id string, fact semantic.FactKey) semantic.Evidence {
	t.Helper()
	evidence, err := semantic.NewEvidence(
		semantic.MustIdentity(id), semantic.GoHostedCompilerID, semantic.CompilerRunEvidence,
		fact, semantic.StableHash([]byte(id)),
	)
	if err != nil {
		t.Fatal(err)
	}
	return evidence
}
