package analyzer

import (
	"os"
	"reflect"
	"testing"
)

func TestHostingPairAnalysisIsRepeatable(t *testing.T) {
	source, err := os.ReadFile("testdata/hosting_pair.go")
	if err != nil {
		t.Fatal(err)
	}
	registry := hostingPairRegistry()
	sources := []SourceFile{{Filename: "testdata/hosting_pair.go", PackagePath: "example.com/billing", Source: source}}
	first, err := AnalyzePackage(sources, registry)
	if err != nil {
		t.Fatal(err)
	}
	second, err := AnalyzePackage(sources, registry)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Delta, second.Delta) || first.GoHostedEvidence().ComparisonDigest() != second.GoHostedEvidence().ComparisonDigest() {
		t.Fatalf("repeated analysis changed evidence:\nfirst=%#v\nsecond=%#v", first.Delta, second.Delta)
	}
}
func hostingPairReport(t *testing.T) EvidenceReport {
	t.Helper()
	source, err := os.ReadFile("testdata/hosting_pair.go")
	if err != nil {
		t.Fatal(err)
	}
	result, err := AnalyzePackage([]SourceFile{{
		Filename: "testdata/hosting_pair.go", PackagePath: "example.com/billing", Source: source,
	}}, hostingPairRegistry())
	if err != nil {
		t.Fatal(err)
	}
	return result.GoHostedEvidence()
}
func hostingPairRegistry() *Registry {
	registry := NewRegistry()
	for namespace, id := range map[string]string{
		"fraud": "fraud://activity/check", "security": "security://activity/check",
	} {
		registry.MustRegister(Registration{
			Ref:  SymbolRef{PackagePath: "example.com/fraud", PackageName: "fraud", Name: "Check"},
			Kind: KindActivity, Identity: NewIdentity(namespace, id),
		})
	}
	return registry
}
func hostingSpan(filename string, startLine, startColumn, endLine, endColumn, startOffset, endOffset int) Span {
	return Span{
		Filename: filename,
		Start:    Position{Offset: startOffset, Line: startLine, Column: startColumn},
		End:      Position{Offset: endOffset, Line: endLine, Column: endColumn},
	}
}
