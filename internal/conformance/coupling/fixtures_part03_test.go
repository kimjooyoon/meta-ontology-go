package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func makeResourceReceipts(binding ResourceBindingConfig) []ExternalResourceReceipt {
	values := []ExternalResourceReceipt{{ReceiptID: "urn:gooo:resource:cpu", Metric: "cpu-core-ns", Value: 10, Unit: "ns"}, {ReceiptID: "urn:gooo:resource:memory", Metric: "peak-memory-bytes", Value: 20, Unit: "bytes"}, {ReceiptID: "urn:gooo:resource:work", Metric: "work-units", Value: 30, Unit: "units"}}
	for i := range values {
		values[i].ProviderDigest, values[i].ObserverDigest, values[i].SnapshotDigest, values[i].SourceDigest = binding.ProviderDigest, binding.ObserverDigest, binding.SnapshotDigest, binding.SourceDigest
		values[i].Present, values[i].Independent, values[i].State = true, true, "CURRENT"
		values[i].BindingDigest = resourceBindingDigest(values[i])
	}
	return values
}
func cloneInput(input Input) Input {
	output := input
	output.SemanticBefore = cloneSemanticIR(input.SemanticBefore)
	output.SemanticAfter = cloneSemanticIR(input.SemanticAfter)
	output.Registry = append([]CodeBinding(nil), input.Registry...)
	output.Changes = append([]CodeChange(nil), input.Changes...)
	output.Receipts = append([]CouplingReceipt(nil), input.Receipts...)
	for i := range output.Receipts {
		output.Receipts[i].EvidenceRefs = append([]string(nil), input.Receipts[i].EvidenceRefs...)
	}
	output.ResourceReceipts = append([]ExternalResourceReceipt(nil), input.ResourceReceipts...)
	output.Roots = append([]string(nil), input.Roots...)
	output.Path.Edges = append([]semantic.InferenceEdge(nil), input.Path.Edges...)
	output.Path.Claims = append([]semantic.SemanticChangeClaim(nil), input.Path.Claims...)
	output.Path.Evidence = append([]semantic.InferenceEvidence(nil), input.Path.Evidence...)
	for i := range output.Path.Edges {
		output.Path.Edges[i].SourceRoots = append([]semantic.ID(nil), input.Path.Edges[i].SourceRoots...)
		output.Path.Edges[i].Evidence = append([]semantic.EvidenceReference(nil), input.Path.Edges[i].Evidence...)
	}
	for i := range output.Path.Claims {
		output.Path.Claims[i].Evidence = append([]semantic.EvidenceReference(nil), input.Path.Claims[i].Evidence...)
	}
	return output
}
func digestText(value string) string { return digestBytes([]byte(value)) }
func TestFixtureBuilderSanity(t *testing.T) {
	for _, row := range testCorpus() {
		if row.Name == "" || row.Input.FixtureID == "" || row.Expected.Decision == "" || row.Expected.Reason == "" {
			t.Fatal("fixture metadata missing")
		}
	}
}
