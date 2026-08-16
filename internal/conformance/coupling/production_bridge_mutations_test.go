//go:build detector_bridge

package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func literalProductionMutationExpectations() map[string]productionVector {
	return decodeLiteralProductionVectors(literalProducerMutationGZIP)
}

type productionResultMutation struct {
	name     string
	mutate   func(*production.Result)
	truth    productionVector
	observed productionVector
}

func literalResultMutations(base production.Result, input production.Input, authority production.AuthorityContext) []productionResultMutation {
	observed := literalProducerResultObservedExpectations()
	truth := literalProductionCorpusExpectations()["positive-no-delta"]
	return []productionResultMutation{
		{name: "result-wrong-decision", mutate: func(result *production.Result) {
			result.Status = production.StatusFailClosed
			result.Reasons = []production.Reason{{Code: production.ReasonDigestMismatch, Detail: "producer-only mutation"}}
		}, truth: truth, observed: observed["result-wrong-decision"]},
		{name: "result-wrong-reason", mutate: func(result *production.Result) {
			result.Reasons = []production.Reason{{Code: production.ReasonDigestMismatch, Detail: "producer-only mutation"}}
		}, truth: truth, observed: observed["result-wrong-reason"]},
		{name: "result-missing-accepted-surface", mutate: func(result *production.Result) {
			result.AcceptedSurfaceIDs = nil
		}, truth: truth, observed: observed["result-missing-accepted-surface"]},
		{name: "result-extra-accepted-surface", mutate: func(result *production.Result) {
			result.AcceptedSurfaceIDs = []semantic.ID{bridgeID("urn:gooo:surface:billing/pay-order"), bridgeID("urn:gooo:surface:unexpected")}
		}, truth: truth, observed: observed["result-extra-accepted-surface"]},
		{name: "result-count-drift", mutate: func(result *production.Result) {
			result.Observation.InferenceRecords.Value++
		}, truth: truth, observed: observed["result-count-drift"]},
		{name: "result-resource-drift", mutate: func(result *production.Result) {
			result.Observation.CPU.Value++
		}, truth: truth, observed: observed["result-resource-drift"]},
		{name: "result-input-digest-drift", mutate: func(result *production.Result) {
			result.InputDigest = bridgeHash("producer-only-input-digest")
		}, truth: truth, observed: observed["result-input-digest-drift"]},
		{name: "result-result-digest-drift", mutate: func(result *production.Result) {
			result.Digest = bridgeHash("producer-only-result-digest")
		}, truth: truth, observed: observed["result-result-digest-drift"]},
	}
}
func literalInputMutations(base production.Input, authority production.AuthorityContext) []struct {
	name            string
	input           production.Input
	authorityBefore production.AuthorityContext
	want            productionVector
} {
	wants := literalProductionMutationExpectations()
	mutations := []struct {
		name            string
		input           production.Input
		authorityBefore production.AuthorityContext
		want            productionVector
	}{}
	add := func(name string, mutate func(*production.Input)) {
		input := cloneProductionInput(base)
		mutate(&input)
		mutations = append(mutations, struct {
			name            string
			input           production.Input
			authorityBefore production.AuthorityContext
			want            productionVector
		}{name: name, input: input, authorityBefore: authority, want: wants[name]})
	}
	add("input-stale-receipt", func(input *production.Input) {
		input.Receipts[0].SnapshotDigest = bridgeHash("stale-receipt")
	})
	add("input-missing-receipt", func(input *production.Input) {
		input.Receipts = nil
	})
	add("input-arbitrary-provider", func(input *production.Input) {
		input.ExternalReceipt.ProviderDigest = bridgeHash("arbitrary-provider")
	})
	add("input-arbitrary-observer", func(input *production.Input) {
		input.ExternalReceipt.ObserverDigest = bridgeHash("arbitrary-observer")
	})
	add("input-wrong-path-endpoint", func(input *production.Input) {
		last := len(input.InferencePath.Edges) - 1
		input.InferencePath.Edges[last].ObjectID = bridgeID("urn:gooo:evidence:wrong-endpoint")
		input.InferencePath.Edges[last].InferenceRecord.ObjectID = input.InferencePath.Edges[last].ObjectID
	})
	add("input-omitted-evidence", func(input *production.Input) {
		input.Receipts[0].EvidenceRefs = nil
	})
	add("input-extra-unrelated-evidence", func(input *production.Input) {
		input.Receipts[0].EvidenceRefs = append(input.Receipts[0].EvidenceRefs, semantic.EvidenceReference{ID: bridgeID("urn:gooo:evidence:unrelated"), Digest: bridgeHash("unrelated-evidence")})
	})
	add("input-duplicate-evidence", func(input *production.Input) {
		input.Receipts[0].EvidenceRefs = append(input.Receipts[0].EvidenceRefs, input.Receipts[0].EvidenceRefs[0])
	})
	add("input-reordered-path-id-presentation", func(input *production.Input) {
		reverseProductionIDs(input.Receipts[0].OriginPathIDs)
	})
	add("input-disconnected-selected-edge", func(input *production.Input) {
		edge := input.InferencePath.Edges[1]
		edge.RecordID = bridgeID("urn:gooo:path:disconnected")
		edge.SubjectID, edge.ObjectID = bridgeID("urn:gooo:term:disconnected"), bridgeID("urn:gooo:code:disconnected")
		edge.InferenceRecord.RecordID, edge.InferenceRecord.SubjectID, edge.InferenceRecord.ObjectID = edge.RecordID, edge.SubjectID, edge.ObjectID
		input.InferencePath.Edges = append(input.InferencePath.Edges, edge)
		input.Receipts[0].OriginPathIDs = append(input.Receipts[0].OriginPathIDs, edge.RecordID)
	})
	add("input-forked-selected-path", func(input *production.Input) {
		edge := input.InferencePath.Edges[1]
		edge.RecordID, edge.ObjectID = bridgeID("urn:gooo:path:fork"), bridgeID("urn:gooo:surface:fork")
		edge.InferenceRecord.RecordID, edge.InferenceRecord.ObjectID = edge.RecordID, edge.ObjectID
		input.InferencePath.Edges = append(input.InferencePath.Edges, edge)
		input.Receipts[0].OriginPathIDs = append(input.Receipts[0].OriginPathIDs, edge.RecordID)
	})
	add("input-cyclic-selected-path", func(input *production.Input) {
		edge := input.InferencePath.Edges[1]
		edge.RecordID, edge.SubjectID, edge.ObjectID = bridgeID("urn:gooo:path:cycle"), bridgeID("urn:gooo:evidence:verification"), bridgeID("urn:gooo:code:billing/pay-order")
		edge.InferenceRecord.RecordID, edge.InferenceRecord.SubjectID, edge.InferenceRecord.ObjectID = edge.RecordID, edge.SubjectID, edge.ObjectID
		input.InferencePath.Edges = append(input.InferencePath.Edges, edge)
		input.Receipts[0].OriginPathIDs = append(input.Receipts[0].OriginPathIDs, edge.RecordID)
	})
	add("input-wrong-start-end", func(input *production.Input) {
		input.InferencePath.Edges[0].SubjectID = bridgeID("urn:gooo:source:wrong")
		input.InferencePath.Edges[0].InferenceRecord.SubjectID = input.InferencePath.Edges[0].SubjectID
		last := len(input.InferencePath.Edges) - 1
		input.InferencePath.Edges[last].SubjectID = bridgeID("urn:gooo:surface:wrong")
		input.InferencePath.Edges[last].InferenceRecord.SubjectID = input.InferencePath.Edges[last].SubjectID
	})
	add("input-rehash-all-packet-authority", mutateRehashAllPacketAuthority)
	add("input-rehash-all-external-authority", mutateRehashAllExternalAuthority)
	return mutations
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
