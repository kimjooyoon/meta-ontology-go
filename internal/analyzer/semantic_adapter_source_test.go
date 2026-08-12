package analyzer

import (
	"errors"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestSourceBundleDigestRejectsDuplicateFilename(t *testing.T) {
	_, err := SourceBundleDigest([]SourceFile{
		{Filename: "same.go", Source: []byte("package p")},
		{Filename: "same.go", Source: []byte("package p")},
	})
	if !errors.Is(err, ErrSemanticAdapter) {
		t.Fatalf("duplicate source error = %v, want ErrSemanticAdapter", err)
	}
	var adapterErr AdapterError
	if !errors.As(err, &adapterErr) || adapterErr.Code != AdapterSourceConfig {
		t.Fatalf("duplicate source error = %v, want source-config", err)
	}
}
