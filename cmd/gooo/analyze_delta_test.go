package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestRunAnalyzeBillingGeneratedDeltaIsCanonicalAndReadOnly(t *testing.T) {
	authority, generated := billingAnalyzeFiles(t, billingAnalyzeAuthority)
	beforeSource, beforeInfo := snapshotFile(t, generated)
	first, firstCode, firstErr := runAnalyzePaths(authority, generated)
	second, secondCode, secondErr := runAnalyzePaths(authority, generated)
	if firstCode != exitOK || secondCode != exitOK || firstErr != "" || secondErr != "" {
		t.Fatalf("billing analyze = %d/%d, stderr=%q/%q", firstCode, secondCode, firstErr, secondErr)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("billing analyze replay changed output:\nfirst=%s\nsecond=%s", first, second)
	}
	var output analyzeDeltaOutput
	if err := json.Unmarshal(first, &output); err != nil {
		t.Fatalf("decode semantic delta: %v; output=%s", err, first)
	}
	if output.SchemaVersion != "analyzer-semantic-delta/v1" || output.Digest == "" || output.AuthoritySemanticDigest == "" || output.AuthoritySemanticDigest != output.ObservedSemanticDigest || !output.SemanticEqual || output.WriteEffect != analyzer.ReconcileNoWrite {
		t.Fatalf("incomplete billing analyze output: %#v", output)
	}
	if len(output.SignatureFacts) != 3 || len(output.CandidateFacts) != 0 || len(output.DeferredImplementation) != 1 || len(output.DeferredSlots) != 1 {
		t.Fatalf("billing delta classes = %d signature, %d candidate, %d implementation, %d slots", len(output.SignatureFacts), len(output.CandidateFacts), len(output.DeferredImplementation), len(output.DeferredSlots))
	}
	for _, fact := range output.SignatureFacts {
		if fact.Fact.Span.File == "" || fact.Fact.Span.End.Offset <= fact.Fact.Span.Start.Offset || fact.Evidence.Span != fact.Fact.Span {
			t.Fatalf("signature fact lost exact source span: %#v", fact)
		}
	}
	if output.DeferredImplementation[0].Origin != analyzer.OriginImplementation || output.DeferredImplementation[0].Object.ID != "billing://entity/payment" {
		t.Fatalf("implementation observation = %#v", output.DeferredImplementation)
	}
	if output.DeferredSlots[0].SlotID != "billing://activity/pay-order/implementation" || output.DeferredSlots[0].Span.End.Offset <= output.DeferredSlots[0].Span.Start.Offset || output.DeferredSlots[0].BodySpan.End.Offset <= output.DeferredSlots[0].BodySpan.Start.Offset {
		t.Fatalf("protected slot = %#v", output.DeferredSlots[0])
	}
	afterSource, afterInfo := snapshotFile(t, generated)
	if !bytes.Equal(beforeSource, afterSource) || !os.SameFile(beforeInfo, afterInfo) || beforeInfo.ModTime() != afterInfo.ModTime() || beforeInfo.Mode() != afterInfo.Mode() {
		t.Fatal("analyze mutated generated Go input")
	}
}

func TestRunAnalyzePreservesStableIDsAcrossDisplayRename(t *testing.T) {
	_, generated := billingAnalyzeFiles(t, billingAnalyzeAuthority)
	authority, _ := writeAnalyzeFile(t, "renamed.gooo", billingAnalyzeRenamedAuthority)
	output, code, stderr := runAnalyzePaths(authority, generated)
	if code != exitOK || stderr != "" {
		t.Fatalf("renamed billing analyze = %d, stderr=%q, output=%s", code, stderr, output)
	}
	var delta analyzeDeltaOutput
	if err := json.Unmarshal(output, &delta); err != nil {
		t.Fatal(err)
	}
	if !delta.SemanticEqual || len(delta.SignatureFacts) != 3 {
		t.Fatalf("renamed billing delta = %#v", delta)
	}
	for _, fact := range delta.SignatureFacts {
		if fact.Fact.Subject.String() != "billing://activity/pay-order" && fact.Fact.Object.String() != "billing://activity/pay-order" {
			t.Fatalf("display rename changed stable semantic IDs: %#v", fact.Fact)
		}
	}
}

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

func TestRunAnalyzeRejectsStaleGeneratedMarkersWithoutWrite(t *testing.T) {
	_, generated := billingAnalyzeFiles(t, billingAnalyzeAuthority)
	source, err := os.ReadFile(generated)
	if err != nil {
		t.Fatal(err)
	}
	mutated := bytes.Replace(source, []byte(`//gooo:generated:end id="billing://entity/order" kind="entity"`), []byte(`//gooo:generated:end id="billing://entity/stale" kind="entity"`), 1)
	if bytes.Equal(source, mutated) {
		t.Fatal("stale marker mutation did not apply")
	}
	if err := os.WriteFile(generated, mutated, 0o640); err != nil {
		t.Fatal(err)
	}
	before, info := snapshotFile(t, generated)
	authority := filepath.Join(filepath.Dir(generated), "authority.gooo")
	if err := os.WriteFile(authority, []byte(billingAnalyzeAuthority), 0o640); err != nil {
		t.Fatal(err)
	}
	output, code, stderr := runAnalyzePaths(authority, generated)
	if code != exitFailure || len(output) != 0 || !strings.Contains(stderr, "generated region") {
		t.Fatalf("stale marker result = code %d, stdout=%q, stderr=%q", code, output, stderr)
	}
	after, afterInfo := snapshotFile(t, generated)
	if !bytes.Equal(before, after) || !os.SameFile(info, afterInfo) || info.ModTime() != afterInfo.ModTime() {
		t.Fatal("stale marker rejection mutated Go input")
	}
}

func billingAnalyzeFiles(t *testing.T, authoritySource string) (string, string) {
	t.Helper()
	authority, _ := writeAnalyzeFile(t, "main.gooo", authoritySource)
	outputDir := filepath.Join(filepath.Dir(authority), "generated")
	var stdout, stderr bytes.Buffer
	if code := runGenerate([]string{authority, "--out", outputDir}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK || stderr.Len() != 0 {
		t.Fatalf("generate billing analyze fixture = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	return authority, filepath.Join(outputDir, generatedFileName)
}

func writeAnalyzeFile(t *testing.T, name, source string) (string, []byte) {
	t.Helper()
	path := name
	if !filepath.IsAbs(path) {
		path = filepath.Join(t.TempDir(), name)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	data := []byte(source)
	if err := os.WriteFile(path, data, 0o640); err != nil {
		t.Fatal(err)
	}
	return path, data
}

func runAnalyzePaths(authority, generated string) ([]byte, int, string) {
	var stdout, stderr bytes.Buffer
	code := runAnalyze([]string{authority, "--go", generated}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr)
	return stdout.Bytes(), code, stderr.String()
}

func snapshotFile(t *testing.T, filename string) ([]byte, os.FileInfo) {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	return data, info
}

const billingAnalyzeAuthority = `package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
`

const billingAnalyzeRenamedAuthority = `package billing
namespace billing

entity PurchaseOrder id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(PurchaseOrder, PaymentMethod) -> Payment
`

const billingAnalyzeAmbiguousAuthority = `package billing
namespace billing

entity Order id "billing://entity/order"
entity AlternateOrder id "billing://entity/order-alt"
entity Payment id "billing://entity/payment"

activity PayOrder(Order) -> Payment
`

const billingAnalyzeAmbiguousGo = `package billing

//gooo:semantic entity id="billing://entity/order" namespace=billing
type Order struct{}

//gooo:semantic entity id="billing://entity/order-alt" namespace=billing
type Order struct{}
type Payment struct{}

func PayOrder(order Order) Payment {
	return Payment{}
}
`
