package analyzer

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"os"
	"testing"
)

func TestAnalyzeAndAdaptSemanticBindsRawSourceDigest(t *testing.T) {
	source, err := os.ReadFile("testdata/registered.go")
	if err != nil {
		t.Fatal(err)
	}
	registry := NewRegistry()
	registry.MustRegister(Registration{
		Ref:  SymbolRef{PackagePath: "example.com/fraud", PackageName: "fraud", Name: "Check"},
		Kind: KindActivity, Identity: NewIdentity("fraud", "fraud://activity/check"),
	})
	policy := billingPolicy(t, RelationUses)
	sources := []SourceFile{{Filename: "registered.go", PackagePath: "example.com/fraud", Source: source}}
	digest, err := SourceBundleDigest(sources)
	if err != nil {
		t.Fatal(err)
	}
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: semantic.NewIR("billing", semantic.Namespace("billing")), Sources: sources,
		Registry: registry, Policy: policy, Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: "go1.26.5|darwin/arm64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceDigest != digest {
		t.Fatalf("source digest = %s, want %s", result.SourceDigest, digest)
	}
	if err := result.IR.Validate(); err != nil {
		t.Fatalf("adapted IR invalid: %v", err)
	}
	if len(result.IR.Evidence()) == 0 {
		t.Fatal("source-bound adaptation emitted no evidence")
	}
	if !validDigest(result.PolicyDigest) || !validDigest(result.ToolchainDigest) || !validDigest(result.BindingDigest) {
		t.Fatalf("incomplete source binding: policy=%q toolchain=%q binding=%q", result.PolicyDigest, result.ToolchainDigest, result.BindingDigest)
	}
}
func TestAnalyzeAndAdaptSemanticRequiresToolchainIdentity(t *testing.T) {
	_, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base:    semantic.NewIR("billing", semantic.Namespace("billing")),
		Sources: []SourceFile{{Filename: "main.go", Source: []byte("package billing\n")}},
		Policy:  billingPolicy(t, RelationUses), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence,
	})
	if err == nil {
		t.Fatal("source-bound adaptation accepted missing toolchain identity")
	}
	var adapterErr AdapterError
	if !errors.As(err, &adapterErr) || adapterErr.Code != AdapterSourceConfig {
		t.Fatalf("toolchain error = %v, want source-config", err)
	}
}
