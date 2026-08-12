package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestRunInspectG1ContractAndRepeatDeterminism(t *testing.T) {
	firstOutput := inspectFixtureOutput(t, sourceOrderA)
	repeatOutput := inspectFixtureOutput(t, sourceOrderA)
	if !bytes.Equal(firstOutput, repeatOutput) {
		t.Fatalf("repeated inspect output differs:\nfirst=%s\nrepeat=%s", firstOutput, repeatOutput)
	}
	dump := decodeGraphDump(t, firstOutput)
	if dump.SchemaVersion != graphDumpSchemaVersion || dump.GraphHash == "" || dump.SourceDigest == "" {
		t.Fatalf("incomplete graph identity: %#v", dump)
	}
	if dump.IR.Status != "available" || dump.IR.SemanticDigest == "" {
		t.Fatalf("unexpected IR status: %#v", dump.IR)
	}
	if dump.Evidence.Status != "missing" || dump.Evidence.Reason == "" {
		t.Fatalf("unexpected evidence status: %#v", dump.Evidence)
	}
	if dump.Provenance.Status != "missing" || dump.Provenance.Reason == "" {
		t.Fatalf("unexpected provenance status: %#v", dump.Provenance)
	}
	if dump.Projection.Status != "deferred" || dump.Projection.Reason == "" {
		t.Fatalf("unexpected projection status: %#v", dump.Projection)
	}
	wantAuthorities := graphAuthorities{
		GoooSource: "authoritative", SemanticIR: "authoritative", Handwritten: "authoritative",
		Provenance: "authoritative", Graph: "derived",
	}
	if dump.Authorities != wantAuthorities {
		t.Fatalf("authorities = %#v, want %#v", dump.Authorities, wantAuthorities)
	}
}

func TestRunInspectReportsDiagnosticsAndFailureExit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runInspect([]string{"broken.gooo"}, fixtureReader{
		source: "package billing\nentity Broken id \"x\" @",
	}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitFailure || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("inspect diagnostics = code %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunInspectReadErrorAndUsage(t *testing.T) {
	var stderr bytes.Buffer
	code := runInspect([]string{"missing.gooo"}, fixtureReader{err: os.ErrNotExist}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitFailure || stderr.Len() == 0 {
		t.Fatalf("inspect read error = code %d, stderr=%q", code, stderr.String())
	}
	stderr.Reset()
	code = runInspect(nil, fixtureReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitUsage || stderr.String() != "usage: gooo inspect <file.gooo>\n" {
		t.Fatalf("inspect usage = code %d, stderr=%q", code, stderr.String())
	}
}

func TestRunInspectG1SourcePermutationPreservesSemanticProjection(t *testing.T) {
	first := decodeGraphDump(t, inspectFixtureOutput(t, sourceOrderA))
	second := decodeGraphDump(t, inspectFixtureOutput(t, sourceOrderB))
	if first.SourceDigest == second.SourceDigest {
		t.Fatal("source digest ignored a source-order change")
	}
	if first.GraphHash != second.GraphHash || first.IR.SemanticDigest != second.IR.SemanticDigest {
		t.Fatalf("semantic digests changed with declaration order: first=%#v second=%#v", first, second)
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
		t.Fatalf("canonical graph projection differs by source order:\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}

func TestRunInspectIsReadOnly(t *testing.T) {
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
	var stdout, stderr bytes.Buffer
	if code := runInspect([]string{filename}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("inspect code = %d, stderr=%q", code, stderr.String())
	}
	afterBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeBytes, afterBytes) || !os.SameFile(beforeInfo, afterInfo) ||
		beforeInfo.Mode() != afterInfo.Mode() || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("inspect changed its input file")
	}
	if afterEntries := directoryEntries(t, directory); !reflect.DeepEqual(beforeEntries, afterEntries) {
		t.Fatalf("inspect changed directory entries: before=%v after=%v", beforeEntries, afterEntries)
	}
}

func TestGraphDumpCandidateIsExplicitAndNotInGraphHash(t *testing.T) {
	file, diagnostics := syntax.Parse(sourceOrderA)
	if diagnostics.Error() != nil {
		t.Fatal(diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	beforeHash := authoritativeGraphHash(ir.Graph)
	beforeIRHash := authoritativeIRHash(ir)
	if beforeIRHash != ir.StableHash() {
		t.Fatalf("authoritative IR digest disagrees without candidates: %s != %s", beforeIRHash, ir.StableHash())
	}
	order := semantic.MustIdentity("billing://entity/order")
	payment := semantic.MustIdentity("billing://entity/payment")
	if err := ir.AddCandidate(semantic.NewCandidateFact(order, semantic.WasDerivedFrom, payment, "needs review")); err != nil {
		t.Fatal(err)
	}
	dump := newGraphDump([]byte(sourceOrderA), ir)
	if dump.GraphHash != beforeHash || dump.IR.SemanticDigest != beforeIRHash {
		t.Fatalf("candidate changed authoritative digest: graph %s/%s, IR %s/%s", beforeHash, dump.GraphHash, beforeIRHash, dump.IR.SemanticDigest)
	}
	if !hasGraphRelation(dump.Relations, "candidate", string(order), string(semantic.WasDerivedFrom), string(payment)) {
		t.Fatalf("candidate relation was not explicit: %#v", dump.Relations)
	}
}

func TestGraphDumpEvidenceDoesNotBecomeProvenance(t *testing.T) {
	file, diagnostics := syntax.Parse(sourceOrderA)
	if diagnostics.Error() != nil {
		t.Fatal(diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	facts := ir.Graph.DeterministicFacts()
	if len(facts) == 0 {
		t.Fatal("fixture has no deterministic fact for evidence")
	}
	evidence, err := semantic.NewEvidence(
		semantic.MustIdentity("billing://evidence/run"), semantic.GoHostedCompilerID,
		semantic.CompilerRunEvidence, facts[0].Key(), semantic.StableHash([]byte("fixture")),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddEvidence(evidence); err != nil {
		t.Fatal(err)
	}
	dump := newGraphDump([]byte(sourceOrderA), ir)
	if dump.Evidence.Status != "available" || !reflect.DeepEqual(dump.Evidence.Refs, []string{evidence.ID.String()}) {
		t.Fatalf("evidence status = %#v", dump.Evidence)
	}
	if dump.Provenance.Status != "missing" || dump.Provenance.Refs != nil || dump.Provenance.Reason == "" {
		t.Fatalf("evidence was falsely reported as provenance: %#v", dump.Provenance)
	}
}

func inspectFixtureOutput(t *testing.T, source string) []byte {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := runInspect([]string{"fixture.gooo"}, fixtureReader{source: source}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("inspect code = %d, stderr=%q", code, stderr.String())
	}
	return stdout.Bytes()
}

func decodeGraphDump(t *testing.T, output []byte) graphDump {
	t.Helper()
	var dump graphDump
	if err := json.Unmarshal(output, &dump); err != nil {
		t.Fatalf("decode graph dump: %v; output=%s", err, output)
	}
	return dump
}

func directoryEntries(t *testing.T, directory string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.Name())
	}
	return result
}

func hasGraphRelation(relations []graphRelation, status, subject, predicate, object string) bool {
	for _, relation := range relations {
		if relation.Status == status && relation.Subject == subject && relation.Predicate == predicate && relation.Object == object {
			return true
		}
	}
	return false
}

const sourceOrderA = `package billing
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder(Order) -> Payment
`

const sourceOrderB = `package billing
namespace billing
entity Payment id "billing://entity/payment"
entity Order id "billing://entity/order"
activity PayOrder(Order) -> Payment
`
