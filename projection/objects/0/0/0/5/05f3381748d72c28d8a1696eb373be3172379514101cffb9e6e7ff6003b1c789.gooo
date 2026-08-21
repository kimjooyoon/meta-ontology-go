package generator

func makeSourceMap(source []byte, ir SemanticIR) (SourceMap, error) {
	markers, err := parseMarkers(source)
	if err != nil {
		return SourceMap{}, err
	}
	if _, err := canonicalMarkerManifestV1(source, markers, ir); err != nil {
		return SourceMap{}, err
	}
	entities, activities := sourceMapNodes(ir)
	result := SourceMap{Mappings: make([]SourceMapping, 0, len(markers.Regions))}
	for _, region := range markers.Regions {
		if err := appendSourceMapRegion(&result, source, region, entities, activities); err != nil {
			return SourceMap{}, err
		}
	}
	return result, nil
}
func sourceMapNodes(ir SemanticIR) (map[string]Entity, map[string]Activity) {
	entities := make(map[string]Entity, len(ir.Entities))
	activities := make(map[string]Activity, len(ir.Activities))
	for _, entity := range ir.Entities {
		entities[entity.ID] = entity
	}
	for _, activity := range ir.Activities {
		activities[activity.ID] = activity
	}
	return entities, activities
}
func appendSourceMapRegion(result *SourceMap, source []byte, region generatedRegion, entities map[string]Entity, activities map[string]Activity) error {
	result.Mappings = append(result.Mappings, SourceMapping{
		SemanticID: region.ID,
		Kind:       region.Kind,
		Ordinal:    len(result.Mappings),
		Source:     regionSourceSpan(region, entities, activities),
		Generated:  rangeForOffsets(source, region.Start, region.End),
	})
	if entity, ok := entities[region.ID]; ok && len(entity.Fields) > 0 {
		if err := appendEntityFieldMappings(result, source, region, entity); err != nil {
			return err
		}
	}
	return appendSlotMappings(result, source, region, activities)
}
func regionSourceSpan(region generatedRegion, entities map[string]Entity, activities map[string]Activity) SourceSpan {
	switch region.Kind {
	case "entity":
		if entity, ok := entities[region.ID]; ok {
			return entity.Source
		}
	case "activity":
		if activity, ok := activities[region.ID]; ok {
			return activity.Source
		}
	}
	return SourceSpan{}
}
