package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func literalAuthorityMutationCases() []productionInputMutationCase {
	return []productionInputMutationCase{
		{name: "input-rehash-all-packet-authority", mutate: mutateRehashAllPacketAuthority},
		{name: "input-rehash-all-external-authority", mutate: mutateRehashAllExternalAuthority},
	}
}
func detectorAuthorityForCorpus(index int) production.AuthorityContext {
	snapshot := "a96c0e268b59e4cf96b821bf38290a6adc7a67298c1bb56f9bb3cf72b0c4b665"
	if index == 1 || index == 9 {
		snapshot = "c033a3f54b319d78cdced9eed61ce7e99859b08987ea6d4b62208d34ad73c644"
	}
	if index == 10 {
		snapshot = "c033a3f54b319d78cdced9eed61ce7e99859b08987ea6d4b62208d34ad73c644"
	}
	registry := production.Registry{Schema: production.RegistrySchemaV1, Digest: "8338cd4a0c97ab010ccb82150d50c7418cdb1c96c543ae0fad097493a0a773c9", Surfaces: []production.Surface{
		{SurfaceID: bridgeID("urn:gooo:surface:billing/pay-order"), CodeSymbolID: bridgeID("urn:gooo:code:billing/pay-order"), SemanticOwnerID: bridgeID("urn:gooo:owner:billing/pay-order"), Binding: production.SourceMapBinding{SourceMapID: bridgeSourceMapID("sm.billing.pay-order"), BindingDigest: "8215244803bce19bff427e5843a64e920330ab4b4a377c731338cf81df77eac8"}},
		{SurfaceID: bridgeID("urn:gooo:surface:billing/pay-order-helper"), CodeSymbolID: bridgeID("urn:gooo:code:billing/pay-order-helper"), SemanticOwnerID: bridgeID("urn:gooo:owner:billing/pay-order-helper"), Binding: production.SourceMapBinding{SourceMapID: bridgeSourceMapID("sm.billing.pay-order-helper"), BindingDigest: "d0f939d45a374c9a06164b10b4eaba1d66ee91cbda58d80c24dd82779f244a10"}},
	}}
	baseline := production.BaselineConfig{Schema: production.BaselineSchemaV1, FullSuiteRequired: true, Digest: "06fc34747e4cadf50f2b013d08ba817817cff1a0be52b09a6f6cedb44f012f02"}
	return production.AuthorityContext{Schema: production.AuthorityContextSchemaV1, Registry: registry, ToolchainDigest: "d214b2b2087b7acc5e3f8ebb3eea7419adbdfbeb0e8e14833feaaa487e475da7", ProfileDigest: "ddd660840b0bae00220bf57cb12f542373f6eddcde3bc0851a269322fb39fb3b", SnapshotDigest: snapshot, ExpectedProviderDigest: "0397c5f0be6ec940d353c05ccac9039256bb1589be30de89924df412ce5db183", ExpectedObserverDigest: "d3e90cc0c61f4342a180c18abf2403dca3bdc6c2284458f3c3df68c7ffb40eba", Baseline: baseline, ExternalReceiptRequired: true}
}
func cloneProductionInput(input production.Input) production.Input {
	output := input
	output.Registry.Surfaces = append([]production.Surface(nil), input.Registry.Surfaces...)
	output.Manifest.Entries = append([]production.ManifestEntry(nil), input.Manifest.Entries...)
	output.Receipts = append([]production.CouplingReceipt(nil), input.Receipts...)
	for i := range output.Receipts {
		output.Receipts[i].OriginPathIDs = append([]semantic.ID(nil), input.Receipts[i].OriginPathIDs...)
		output.Receipts[i].EvidenceRefs = append([]semantic.EvidenceReference(nil), input.Receipts[i].EvidenceRefs...)
		if input.Receipts[i].AuthoritativeSource != nil {
			source := *input.Receipts[i].AuthoritativeSource
			output.Receipts[i].AuthoritativeSource = &source
		}
	}
	if input.ExternalReceipt != nil {
		external := *input.ExternalReceipt
		if input.ExternalReceipt.CPUWorkUnits != nil {
			value := *input.ExternalReceipt.CPUWorkUnits
			external.CPUWorkUnits = &value
		}
		if input.ExternalReceipt.PeakMemoryBytes != nil {
			value := *input.ExternalReceipt.PeakMemoryBytes
			external.PeakMemoryBytes = &value
		}
		if input.ExternalReceipt.DeterministicWorkUnits != nil {
			value := *input.ExternalReceipt.DeterministicWorkUnits
			external.DeterministicWorkUnits = &value
		}
		output.ExternalReceipt = &external
	}
	output.InferencePath.Edges = append([]semantic.InferenceEdge(nil), input.InferencePath.Edges...)
	output.InferencePath.Claims = append([]semantic.SemanticChangeClaim(nil), input.InferencePath.Claims...)
	output.InferencePath.Evidence = append([]semantic.InferenceEvidence(nil), input.InferencePath.Evidence...)
	for i := range output.InferencePath.Edges {
		output.InferencePath.Edges[i].SourceRoots = append([]semantic.ID(nil), input.InferencePath.Edges[i].SourceRoots...)
		output.InferencePath.Edges[i].Evidence = append([]semantic.EvidenceReference(nil), input.InferencePath.Edges[i].Evidence...)
	}
	for i := range output.InferencePath.Claims {
		output.InferencePath.Claims[i].Evidence = append([]semantic.EvidenceReference(nil), input.InferencePath.Claims[i].Evidence...)
	}
	for i := range output.InferencePath.Evidence {
		output.InferencePath.Evidence[i].Controls = input.InferencePath.Evidence[i].Controls
	}
	return output
}
