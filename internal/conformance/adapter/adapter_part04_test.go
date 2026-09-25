package adapter

func sampleResponse(status Status, reverse bool) Response {
	regions := []Region{{Kind: "generated", SemanticID: "billing.total", Start: 20, End: 30}, {Kind: "generated", SemanticID: "billing.id", Start: 1, End: 10}}
	slots := []Slot{{ID: "slot.total", OwnerID: "billing.total", Start: 22, End: 28}}
	imports := []Import{{Path: "fmt", UsedBy: []string{"billing.total", "billing.id"}}, {Path: "strings", Alias: "str", UsedBy: []string{"billing.id"}}}
	mappings := []Mapping{{SemanticID: "billing.total", Kind: "field", Source: Range{1, 2}, Generated: Range{20, 30}}}
	if reverse {
		regions[0], regions[1] = regions[1], regions[0]
		imports[0].UsedBy = []string{"billing.id", "billing.total"}
		imports[0], imports[1] = imports[1], imports[0]
	}
	response := Response{
		Schema: ProtocolSchema, Fixture: "billing/main", Operation: OperationGenerate, RunID: "run-001", Status: status,
		PromotionEligible: status == StatusPass,
		Observed: Observed{
			SemanticDigest: "sha256:semantic", SourceDigest: "sha256:source",
			Regions: regions, Slots: slots, Imports: imports, SourceMap: mappings,
			Delta: &Delta{Locality: []string{"billing.total", "billing.id"}},
		},
		Measurements: Measurements{RepeatCount: 2, CanonicalEqualCount: 2, SourceEqualCount: 2, SemanticEqualCount: 2, RegionEqualCount: 2},
		Evidence: EvidenceArtifact{Producer: "go", Bundle: EvidenceBundle{
			Schema: EvidenceSchema, Stage: StageGoBaseline, Fixture: "billing/main", Decision: string(status),
			Facts: []EvidenceFact{{ID: "billing/main/status", Kind: "status", Value: string(status)}, {ID: "billing/main/scope", Kind: "scope", Value: "billing"}},
		}},
	}
	if status == StatusFail {
		response.Failure = &Failure{Code: "marker-overlap", Detail: "protected marker changed"}
	}
	if reverse {
		response.Evidence.Bundle.Facts = reverseFacts(response.Evidence.Bundle.Facts)
	}
	return response
}
func reverseFacts(facts []EvidenceFact) []EvidenceFact {
	reversed := append([]EvidenceFact(nil), facts...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	return reversed
}
