package analyzer

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"runtime"
	"testing"
)

func TestGeneratedBillingDuplicateFilenameRejectsBeforeWrite(t *testing.T) {
	source := generatedBillingSource(t)
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	before := irSnapshot(base)
	_, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: base, Sources: []SourceFile{source, source}, Registry: generatedBillingRegistry(t),
		Policy: generatedBillingPolicy(t), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: "go1.26.5|" + runtime.GOOS + "/" + runtime.GOARCH,
	})
	var adapterErr AdapterError
	if !errors.As(err, &adapterErr) || adapterErr.Code != AdapterSourceConfig {
		t.Fatalf("duplicate filename error = %v, want source-config", err)
	}
	if got := irSnapshot(base); got != before {
		t.Fatalf("base IR changed after duplicate filename: before=%q after=%q", before, got)
	}
}
func TestGeneratedBillingParseFailureDoesNotWrite(t *testing.T) {
	source := generatedBillingSource(t)
	source.Source = []byte("package billing\nfunc broken(\n")
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	before := irSnapshot(base)
	_, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: base, Sources: []SourceFile{source}, Registry: generatedBillingRegistry(t),
		Policy: generatedBillingPolicy(t), Producer: semantic.GoHostedCompilerID,
		EvidenceKind:      semantic.CompilerRunEvidence,
		ToolchainIdentity: "go1.26.5|" + runtime.GOOS + "/" + runtime.GOARCH,
	})
	if err == nil {
		t.Fatal("invalid generated source was accepted")
	}
	if got := irSnapshot(base); got != before {
		t.Fatalf("base IR changed after parse failure: before=%q after=%q", before, got)
	}
}
func TestGeneratedBillingConfigFailureDoesNotWrite(t *testing.T) {
	source := generatedBillingSource(t)
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	before := irSnapshot(base)
	_, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: base, Sources: []SourceFile{source}, Registry: generatedBillingRegistry(t),
		Policy: generatedBillingPolicy(t), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence,
	})
	var adapterErr AdapterError
	if !errors.As(err, &adapterErr) || adapterErr.Code != AdapterSourceConfig {
		t.Fatalf("missing toolchain error = %v, want source-config", err)
	}
	if got := irSnapshot(base); got != before {
		t.Fatalf("base IR changed after config failure: before=%q after=%q", before, got)
	}
}
