//go:build detector_bridge

package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func reverseProductionIDs(values []semantic.ID) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func mutateRehashAllPacketAuthority(input *production.Input) {
	surface := &input.Registry.Surfaces[0]
	surface.CodeSymbolID = bridgeID("urn:gooo:code:billing/pay-order-rehashed")
	surface.Binding.BindingDigest = bridgeBindingDigestValues(surface.SurfaceID.String(), surface.CodeSymbolID.String(), surface.SemanticOwnerID.String(), surface.Binding.SourceMapID.String())
	input.Registry.Digest = bridgeRegistryDigest(input.Registry)
	input.Config.RegistryDigest = input.Registry.Digest
	input.Manifest.RegistryDigest = input.Registry.Digest
	for i := range input.Manifest.Entries {
		if input.Manifest.Entries[i].SurfaceID == surface.SurfaceID {
			input.Manifest.Entries[i].CodeSymbolID = surface.CodeSymbolID
			input.Manifest.Entries[i].AfterBindingDigest = surface.Binding.BindingDigest
			input.Manifest.Entries[i].BeforeBindingDigest = surface.Binding.BindingDigest
		}
	}
	input.Manifest.Digest = bridgeManifestDigest(input.Manifest)
	for i := range input.Receipts {
		if input.Receipts[i].SurfaceID == surface.SurfaceID {
			input.Receipts[i].CodeSymbolID = surface.CodeSymbolID
			input.Receipts[i].SourceMapBindingDigest = surface.Binding.BindingDigest
			input.Receipts[i].RegistryDigest = input.Registry.Digest
		}
	}
	for i := range input.InferencePath.Edges {
		edge := &input.InferencePath.Edges[i]
		if edge.Kind == semantic.InferenceDerivedProjection {
			edge.SubjectID = surface.CodeSymbolID
			edge.InferenceRecord.SubjectID = surface.CodeSymbolID
		}
	}
}
func mutateRehashAllExternalAuthority(input *production.Input) {
	input.Config.ExpectedProviderDigest = bridgeHash("rehashed-provider-authority")
	if input.ExternalReceipt != nil {
		input.ExternalReceipt.ProviderDigest = input.Config.ExpectedProviderDigest
		input.ExternalReceipt.Digest = bridgeExternalDigest(*input.ExternalReceipt)
	}
}
