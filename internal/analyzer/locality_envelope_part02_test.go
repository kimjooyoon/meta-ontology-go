package analyzer

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestGeneratedBillingLocalityRejectsMissingTamperedAndRelabeledNoWrite(t *testing.T) {
	base := localityBase(t, true)
	result := adaptGeneratedBillingWithBase(t, base, generatedBillingRegistry(t))
	tests := map[string]func(*LocalityEnvelope){
		"missing": func(envelope *LocalityEnvelope) { *envelope = LocalityEnvelope{} },
		"tampered": func(envelope *LocalityEnvelope) {
			envelope.Affected = append([]semantic.ID(nil), envelope.Affected...)
			envelope.Affected[0] = semantic.MustIdentity("billing://entity/tampered")
			envelope.Digest = envelope.StableHash()
		},
		"relabeled": func(envelope *LocalityEnvelope) {
			touched := append([]semantic.ID(nil), envelope.Touched...)
			envelope.Touched = append([]semantic.ID(nil), envelope.Affected...)
			envelope.Affected = touched
			envelope.Digest = envelope.StableHash()
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			beforeBase, beforeResult := irSnapshot(base), irSnapshot(result.IR)
			mutated := result
			mutated.Locality.Touched = append([]semantic.ID(nil), result.Locality.Touched...)
			mutated.Locality.Affected = append([]semantic.ID(nil), result.Locality.Affected...)
			mutate(&mutated.Locality)
			err := ValidateLocalityEnvelope(base, mutated)
			var adapterErr AdapterError
			if !errors.As(err, &adapterErr) || adapterErr.Code != AdapterLocalityConfig ||
				adapterErr.WriteEffect != ReconcileNoWrite {
				t.Fatalf("error = %v, want locality-config/no-write", err)
			}
			if got := irSnapshot(base); got != beforeBase {
				t.Fatalf("base changed after rejection: before=%q after=%q", beforeBase, got)
			}
			if got := irSnapshot(result.IR); got != beforeResult {
				t.Fatalf("result changed after rejection: before=%q after=%q", beforeResult, got)
			}
		})
	}
}
