package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/conformance/adapter"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
)

func projectionResponse(filename string, result generator.Result, observed adapter.Observed, protectedEqual bool) adapter.Response {
	return adapter.Response{
		Schema:            adapter.ProtocolSchema,
		Fixture:           filename,
		Operation:         adapter.OperationGenerate,
		Status:            adapter.StatusPass,
		PromotionEligible: false,
		Observed:          observed,
		Measurements: adapter.Measurements{
			SourceSpanCount:        len(result.SourceMap.Mappings),
			ProtectedBytesEqual:    protectedEqual,
			RepeatCount:            1,
			CanonicalEqualCount:    1,
			SourceEqualCount:       1,
			SemanticEqualCount:     1,
			RegionEqualCount:       1,
			SourceMapResolvedCount: len(result.SourceMap.Mappings),
		},
		Evidence: adapter.EvidenceArtifact{
			Producer: "gooo",
			Bundle: adapter.EvidenceBundle{
				Schema:   adapter.EvidenceSchema,
				Stage:    adapter.StageGoooAuthoritative,
				Fixture:  filename,
				Decision: string(adapter.StatusPass),
				Facts:    []adapter.EvidenceFact{},
			},
		},
	}
}
