package analyzer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestUnresolvedIdentifierSelectorAndTypeBecomeDeferredDetails(t *testing.T) {
	source := []byte(`package billing

//gooo:semantic activity id="billing://activity/run" namespace=billing
func Run() {
	missingIdentifier
	missingSelector.Value
	missingType{}
}
`)
	result, err := AnalyzeSource("unknown.go", source, NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Delta.Added) != 0 || len(result.Delta.Candidates) != 0 {
		t.Fatalf("unknown delta crossed semantic authority: %#v", result.Delta)
	}
	if got := len(result.Delta.ImplementationDetails); got != 3 {
		t.Fatalf("deferred details = %d, want 3: %#v", got, result.Delta.ImplementationDetails)
	}
	want := []string{"missingIdentifier", "missingSelector.Value", "missingType"}
	for index, detail := range result.Delta.ImplementationDetails {
		if detail.Reference != want[index] || detail.IdentityState != IdentityUnresolved || detail.Reason == "" {
			t.Fatalf("detail[%d] = %#v, want reference/state/reason", index, detail)
		}
		if detail.Span.Filename != "unknown.go" || detail.Span.End.Offset <= detail.Span.Start.Offset {
			t.Fatalf("detail[%d] lost source span: %#v", index, detail)
		}
	}
	if evidence := result.GoHostedEvidence().ImplementationEvidence(); len(evidence) != 3 {
		t.Fatalf("implementation evidence = %d, want 3", len(evidence))
	}
}

func TestUnresolvedDetailsAreStableAcrossReplayAndFileOrder(t *testing.T) {
	files := []SourceFile{
		{Filename: "activity.go", PackagePath: "billing", Source: []byte(`package billing

//gooo:semantic activity id="billing://activity/run"
func Run() { missingSelector.Value; missingType{} }
`)},
		{Filename: "support.go", PackagePath: "billing", Source: []byte(`package billing

//gooo:semantic activity id="billing://activity/support"
func Support() { missingValue }
`)},
	}
	first, err := AnalyzePackage(files, NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	replay, err := AnalyzePackage(files, NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	reversed := []SourceFile{files[1], files[0]}
	permuted, err := AnalyzePackage(reversed, NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Delta, replay.Delta) || !reflect.DeepEqual(first.Delta, permuted.Delta) {
		t.Fatalf("deferred details changed across replay/order:\nfirst=%#v\nreplay=%#v\npermuted=%#v", first.Delta, replay.Delta, permuted.Delta)
	}
	if len(first.Delta.ImplementationDetails) != 3 {
		t.Fatalf("deferred details = %#v, want activity selector/type and support identifier", first.Delta.ImplementationDetails)
	}
}

func TestUnresolvedDetailsStayOutsideAdapterAuthorityAndRejectTampering(t *testing.T) {
	source := SourceFile{Filename: "unknown.go", PackagePath: "billing", Source: []byte(`package billing

//gooo:semantic activity id="billing://activity/run" namespace=billing
func Run() { missingIdentifier }
`)}
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	baseBefore := irSnapshot(base)
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: base, Sources: []SourceFile{source}, Registry: NewRegistry(),
		Policy: emptyPolicy(t), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: "go1.26.5|test/amd64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := irSnapshot(base); got != baseBefore {
		t.Fatalf("base IR changed during deferred adaptation: before=%q after=%q", baseBefore, got)
	}
	if len(result.IR.Graph.DeterministicFacts()) != 0 || len(result.IR.Graph.Candidates()) != 0 {
		t.Fatalf("unknown detail entered semantic graph: %#v", result.IR.Graph)
	}
	if len(result.NormalizedDelta.DeferredDetails) != 1 ||
		result.NormalizedDelta.DeferredDetails[0].Detail.IdentityState != IdentityUnresolved {
		t.Fatalf("normalized deferred details = %#v", result.NormalizedDelta.DeferredDetails)
	}
	before := irSnapshot(result.IR)
	result.NormalizedDelta.DeferredDetails[0].Detail.IdentityState = IdentityState("tampered")
	reconcile := ReconcileSemantic(result, result.IR, result.SourceDigest, result.PolicyDigest,
		result.ToolchainDigest, result.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.WriteEffect != ReconcileNoWrite || reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("tampered deferred detail reconcile = %#v, want no-write rejection", reconcile)
	}
	if got := irSnapshot(result.IR); got != before {
		t.Fatalf("reconcile mutated IR: before=%q after=%q", before, got)
	}
}

func TestTopLevelUnresolvedDetailTamperFailsReconcileWithoutWrite(t *testing.T) {
	source := SourceFile{Filename: "unknown.go", PackagePath: "billing", Source: []byte(`package billing

//gooo:semantic activity id="billing://activity/run" namespace=billing
func Run() { missingIdentifier }
`)}
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base:    semantic.NewIR("billing", semantic.Namespace("billing")),
		Sources: []SourceFile{source}, Registry: NewRegistry(), Policy: emptyPolicy(t),
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		ToolchainIdentity: "go1.26.5|test/amd64",
	})
	if err != nil {
		t.Fatal(err)
	}
	before := irSnapshot(result.IR)
	result.ImplementationDetails[0].Reference = "tamperedReference"

	reconcile := ReconcileSemantic(result, result.IR, result.SourceDigest, result.PolicyDigest,
		result.ToolchainDigest, result.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.DeltaValid || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("top-level detail tamper reconcile = %#v, want invalid no-write", reconcile)
	}
	if got := irSnapshot(result.IR); got != before {
		t.Fatalf("reconcile mutated IR: before=%q after=%q", before, got)
	}
}

func TestUnresolvedDetailMissingSourceFileFailsReconcileWithoutWrite(t *testing.T) {
	source := SourceFile{Filename: "unknown.go", PackagePath: "billing", Source: []byte(`package billing

//gooo:semantic activity id="billing://activity/run" namespace=billing
func Run() { missingIdentifier }
`)}
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base:    semantic.NewIR("billing", semantic.Namespace("billing")),
		Sources: []SourceFile{source}, Registry: NewRegistry(), Policy: emptyPolicy(t),
		Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence,
		ToolchainIdentity: "go1.26.5|test/amd64",
	})
	if err != nil {
		t.Fatal(err)
	}
	before := irSnapshot(result.IR)
	result.NormalizedDelta.DeferredDetails[0].Detail.Span.Filename = ""
	result.NormalizedDelta.Digest = result.NormalizedDelta.StableHash()
	result.BindingDigest = semanticAdapterBindingDigest(result)

	reconcile := ReconcileSemantic(result, result.IR, result.SourceDigest, result.PolicyDigest,
		result.ToolchainDigest, result.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.DeltaValid || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("missing source file reconcile = %#v, want invalid no-write", reconcile)
	}
	if got := irSnapshot(result.IR); got != before {
		t.Fatalf("reconcile mutated IR: before=%q after=%q", before, got)
	}
}

func TestMalformedSourceRejectsBeforeDeferredCommit(t *testing.T) {
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	before := irSnapshot(base)
	_, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: base, Sources: []SourceFile{{Filename: "broken.go", Source: []byte("package billing\nfunc Run(\n")}},
		Registry: NewRegistry(), Policy: emptyPolicy(t), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: "go1.26.5|test/amd64",
	})
	if err == nil {
		t.Fatal("malformed source was accepted")
	}
	var adapterErr AdapterError
	if errors.As(err, &adapterErr) {
		t.Fatalf("malformed source returned adapter error instead of parse error: %v", err)
	}
	if got := irSnapshot(base); got != before {
		t.Fatalf("malformed source changed base IR: before=%q after=%q", before, got)
	}
}
