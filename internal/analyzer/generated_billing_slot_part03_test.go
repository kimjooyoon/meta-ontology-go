package analyzer

import (
	"bytes"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"
)

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
