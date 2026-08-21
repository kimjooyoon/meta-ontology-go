package analyzer

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"
)

func TestAnalyzeAndAdaptSemanticRejectsDiagnosticsWithoutMutation(t *testing.T) {
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	before := irSnapshot(base)
	_, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: base,
		Sources: []SourceFile{{
			Filename:    "invalid-annotation.go",
			PackagePath: "billing",
			Source: []byte(`package billing

//gooo:semantic entity id="not-an-identity" namespace=billing
type Order struct{}
`),
		}},
		Registry: NewRegistry(), Policy: billingPolicy(t, RelationUses),
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		ToolchainIdentity: "go1.26.5|test/amd64",
	})
	var adapterErr AdapterError
	if !errors.As(err, &adapterErr) || adapterErr.Code != AdapterAnalysisDiagnostics ||
		adapterErr.WriteEffect != ReconcileNoWrite {
		t.Fatalf("diagnostic rejection = %v, want analysis-diagnostics no-write", err)
	}
	if !strings.Contains(adapterErr.Detail, "semantic identity must be a valid URI") {
		t.Fatalf("diagnostic rejection detail = %q", adapterErr.Detail)
	}
	if got := irSnapshot(base); got != before {
		t.Fatalf("diagnostic rejection mutated base IR: before=%q after=%q", before, got)
	}
}
func TestSourceBundleDigestIsOrderIndependentAndByteSensitive(t *testing.T) {
	first := []SourceFile{
		{Filename: "z.go", PackagePath: "billing", Source: []byte("package billing\nvar Z = 1\n")},
		{Filename: "a.go", PackagePath: "billing", Source: []byte("package billing\nvar A = 1\n")},
	}
	second := []SourceFile{first[1], first[0]}
	left, err := SourceBundleDigest(first)
	if err != nil {
		t.Fatal(err)
	}
	right, err := SourceBundleDigest(second)
	if err != nil {
		t.Fatal(err)
	}
	if left != right {
		t.Fatalf("source order changed digest: %s != %s", left, right)
	}
	mutated := append([]SourceFile(nil), first...)
	mutated[0].Source = []byte("package billing\nvar Z = 2\n")
	changed, err := SourceBundleDigest(mutated)
	if err != nil {
		t.Fatal(err)
	}
	if changed == left {
		t.Fatal("source byte mutation retained digest")
	}
}
