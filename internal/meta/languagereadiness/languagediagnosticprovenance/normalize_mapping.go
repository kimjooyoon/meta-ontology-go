package languagediagnosticprovenance

import "github.com/kimjooyoon/meta-ontology-go/internal/generator"

func resolveSourceMapping(observation Observation) (generator.SourceMapping, *ProvenanceError) {
	matches := []generator.SourceMapping{}
	for _, mapping := range observation.SourceMap.Mappings {
		if observation.GeneratedOffset >= mapping.Generated.Start.Offset &&
			observation.GeneratedOffset < mapping.Generated.End.Offset {
			matches = append(matches, mapping)
		}
	}
	if len(matches) == 0 {
		return generator.SourceMapping{}, provenanceError("SOURCE_MAP_MISSING")
	}
	if len(matches) != 1 {
		return generator.SourceMapping{}, provenanceError("SOURCE_MAP_AMBIGUOUS")
	}
	mapping := matches[0]
	if mapping.SemanticID == "" || mapping.Kind == "" || mapping.Source.URI == "" {
		return generator.SourceMapping{}, provenanceError("SOURCE_MAP_SEMANTIC_UNKNOWN")
	}
	semantic := sourceSpan(mapping)
	if !validSpan(semantic) {
		return generator.SourceMapping{}, provenanceError("SOURCE_MAP_RANGE_INVALID")
	}
	return mapping, nil
}

func sourceSpan(mapping generator.SourceMapping) Span {
	return Span{
		Start: Position{
			Filename: mapping.Source.URI, Offset: mapping.Source.Start.Offset,
			Line: mapping.Source.Start.Line, Column: mapping.Source.Start.Column,
		},
		End: Position{
			Filename: mapping.Source.URI, Offset: mapping.Source.End.Offset,
			Line: mapping.Source.End.Line, Column: mapping.Source.End.Column,
		},
	}
}
