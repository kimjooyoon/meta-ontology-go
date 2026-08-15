package analyzer

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const generatedBillingSlotID = "billing://activity/pay-order/implementation"

func TestGeneratedBillingProtectedSlotIsStableAcrossReplayAndOrder(t *testing.T) {
	source := generatedBillingSource(t)
	support := SourceFile{Filename: "support.go", PackagePath: "billing", Source: []byte("package billing\n")}
	sources := []SourceFile{source, support}
	policy := generatedBillingPolicy(t)
	first := adaptGeneratedBillingSources(t, sources, generatedBillingRegistry(t), policy)
	second := adaptGeneratedBillingSources(t, sources, generatedBillingRegistry(t), policy)
	permuted := adaptGeneratedBillingSources(t, []SourceFile{support, source},
		reversedGeneratedBillingRegistry(t), policy)

	for _, result := range []SemanticAdapterResult{first, second, permuted} {
		if len(result.SlotObservations) != 1 || len(result.NormalizedDelta.DeferredSlots) != 1 {
			t.Fatalf("protected slots = %d/%d, want one", len(result.SlotObservations),
				len(result.NormalizedDelta.DeferredSlots))
		}
		slot := result.SlotObservations[0]
		if slot.SlotID != generatedBillingSlotID || slot.Status != ProtectedSlotDeferred ||
			slot.SourceFile != source.Filename || slot.SourceDigest != result.SourceDigest ||
			slot.BaseDigest != result.NormalizedDelta.SignatureFacts[0].Binding.BaseDigest ||
			slot.PolicyDigest != result.PolicyDigest || slot.ToolchainDigest != result.ToolchainDigest ||
			slot.RegistryDigest != result.RegistryDigest ||
			slot.BodySpan.End.Offset <= slot.BodySpan.Start.Offset {
			t.Fatalf("protected slot = %#v", slot)
		}
		if !validDigest(slot.BodyDigest) || !validDigest(slot.Fingerprint()) ||
			!validDigest(result.SlotObservationDigest) {
			t.Fatalf("protected slot digest binding = %#v, digest=%q", slot, result.SlotObservationDigest)
		}
	}
	if first.SourceDigest != second.SourceDigest || first.SourceDigest != permuted.SourceDigest ||
		first.SlotObservations[0].Span != second.SlotObservations[0].Span ||
		first.SlotObservations[0].Span != permuted.SlotObservations[0].Span ||
		first.SlotObservations[0].BodySpan != second.SlotObservations[0].BodySpan ||
		first.SlotObservations[0].BodySpan != permuted.SlotObservations[0].BodySpan ||
		first.SlotObservationDigest != second.SlotObservationDigest ||
		first.SlotObservationDigest != permuted.SlotObservationDigest ||
		first.NormalizedDelta.Digest != second.NormalizedDelta.Digest ||
		first.NormalizedDelta.Digest != permuted.NormalizedDelta.Digest {
		t.Fatal("protected slot handoff changed across replay or source/registry order")
	}
}

func TestGeneratedBillingProtectedSlotMutationChangesOnlyDeferredObservation(t *testing.T) {
	source := generatedBillingSource(t)
	originalBytes := append([]byte(nil), source.Source...)
	policy := generatedBillingPolicy(t)
	first := adaptGeneratedBillingSource(t, source, generatedBillingRegistry(t), policy)
	mutated := mutateGeneratedBillingSlotBody(t, source)
	second := adaptGeneratedBillingSource(t, mutated, generatedBillingRegistry(t), policy)
	comparison := semantic.CompareIR(first.IR, second.IR)
	if !comparison.SemanticEqual {
		t.Fatalf("slot body mutation changed authoritative contract facts: %#v", comparison)
	}
	if first.SlotObservations[0].BodyDigest == second.SlotObservations[0].BodyDigest ||
		first.SlotObservations[0].Fingerprint() == second.SlotObservations[0].Fingerprint() ||
		first.SlotObservationDigest == second.SlotObservationDigest ||
		first.ImplementationObservationDigest == second.ImplementationObservationDigest ||
		first.BindingDigest == second.BindingDigest ||
		first.NormalizedDelta.Digest == second.NormalizedDelta.Digest {
		t.Fatal("slot body mutation did not change deferred observation evidence")
	}
	before := irSnapshot(first.IR)
	reconcile := ReconcileSemantic(second, first.IR, first.SourceDigest, first.PolicyDigest,
		first.ToolchainDigest, first.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "source-or-observation-mismatch" || reconcile.SourceMatch ||
		reconcile.ObservationMatch {
		t.Fatalf("slot body mutation reconcile = %#v, want fail-closed no-write", reconcile)
	}
	if got := irSnapshot(first.IR); got != before {
		t.Fatalf("rejected slot body mutation changed IR: before=%q after=%q", before, got)
	}
	if !bytes.Equal(source.Source, originalBytes) {
		t.Fatal("slot analysis mutated the original generated source bytes")
	}
}

func TestGeneratedBillingProtectedSlotIdentityFailuresAreNoWrite(t *testing.T) {
	original := generatedBillingSource(t)
	cases := []struct {
		name   string
		mutate func(string) string
	}{
		{name: "missing identity", mutate: func(source string) string {
			return strings.Replace(source, `id="`+generatedBillingSlotID+`"`, `id=""`, 1)
		}},
		{name: "stale end identity", mutate: func(source string) string {
			return strings.Replace(source,
				`//gooo:slot:end id="`+generatedBillingSlotID+`"`,
				`//gooo:slot:end id="billing://activity/pay-order/stale"`, 1)
		}},
		{name: "duplicate identity", mutate: func(source string) string {
			return source + "\nfunc duplicateSlot() {\n" +
				"\t//gooo:slot:start id=\"" + generatedBillingSlotID + "\"\n" +
				"\treturn\n" +
				"\t//gooo:slot:end id=\"" + generatedBillingSlotID + "\"\n}\n"
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			base := semantic.NewIR("billing", semantic.Namespace("billing"))
			before := irSnapshot(base)
			source := original
			source.Source = []byte(testCase.mutate(string(original.Source)))
			_, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
				Base: base, Sources: []SourceFile{source}, Registry: generatedBillingRegistry(t),
				Policy: generatedBillingPolicy(t), Producer: semantic.GoHostedCompilerID,
				EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: billingToolchain(),
			})
			var adapterErr AdapterError
			if !errors.As(err, &adapterErr) || adapterErr.Code != AdapterSlotConfig ||
				adapterErr.WriteEffect != ReconcileNoWrite {
				t.Fatalf("slot error = %v, want slot-config/no-write", err)
			}
			if got := irSnapshot(base); got != before {
				t.Fatalf("slot configuration failure changed IR: before=%q after=%q", before, got)
			}
			if !bytes.Equal(source.Source, []byte(testCase.mutate(string(original.Source)))) {
				t.Fatal("slot validation mutated rejected source bytes")
			}
		})
	}
}

func adaptGeneratedBillingSources(
	t *testing.T, sources []SourceFile, registry *Registry, policy MappingPolicy,
) SemanticAdapterResult {
	t.Helper()
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: semantic.NewIR("billing", semantic.Namespace("billing")), Sources: sources,
		Registry: registry, Policy: policy, Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: billingToolchain(),
	})
	if err != nil {
		t.Fatalf("adapt generated billing sources: %v", err)
	}
	return result
}

func mutateGeneratedBillingSlotBody(t *testing.T, source SourceFile) SourceFile {
	t.Helper()
	const body = "\treturn Payment{}\n"
	mutated := strings.Replace(string(source.Source), body, "\treturn Payment{} // slot body mutation\n", 1)
	if mutated == string(source.Source) {
		t.Fatal("generated billing slot body was not found")
	}
	result := source
	result.Source = []byte(mutated)
	return result
}
