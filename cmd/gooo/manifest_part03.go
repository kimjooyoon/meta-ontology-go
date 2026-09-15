package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/conformance/adapter"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func observedProjection(result generator.Result) (adapter.Observed, error) {
	regions := make([]adapter.Region, 0)
	slots := make([]adapter.Slot, 0)
	mappings := make([]adapter.Mapping, 0, len(result.SourceMap.Mappings))
	for _, mapping := range result.SourceMap.Mappings {
		start, end := mapping.Generated.Start.Offset, mapping.Generated.End.Offset
		if start < 0 || end < start || end > len(result.Source) {
			return adapter.Observed{}, fmt.Errorf("source map %q has invalid generated range %d:%d", mapping.SemanticID, start, end)
		}
		mapped := adapter.Mapping{
			SemanticID: mapping.SemanticID,
			Kind:       mapping.Kind,
			Source:     adapter.Range{Start: mapping.Source.Start.Offset, End: mapping.Source.End.Offset},
			Generated:  adapter.Range{Start: start, End: end},
		}
		mappings = append(mappings, mapped)
		if mapping.Kind == "slot" {
			slots = append(slots, adapter.Slot{ID: mapping.SemanticID, Start: start, End: end})
			continue
		}
		regions = append(regions, adapter.Region{
			Kind:       mapping.Kind,
			SemanticID: mapping.SemanticID,
			Start:      start,
			End:        end,
			BodyDigest: semantic.StableHash(result.Source[start:end]),
		})
	}
	for index := range slots {
		for _, region := range regions {
			if region.Start <= slots[index].Start && slots[index].End <= region.End {
				slots[index].OwnerID = region.SemanticID
				break
			}
		}
		if slots[index].OwnerID == "" {
			return adapter.Observed{}, fmt.Errorf("slot %q is not inside a generated region", slots[index].ID)
		}
		if slots[index].End > len(result.Source) || slots[index].Start < 0 {
			return adapter.Observed{}, fmt.Errorf("slot %q has invalid range", slots[index].ID)
		}
		slots[index].BodyDigest = semantic.StableHash(result.Source[slots[index].Start:slots[index].End])
	}
	return adapter.Observed{SourceDigest: semantic.StableHash(result.Source), Regions: regions, Slots: slots, SourceMap: mappings}, nil
}
func sourceMapDigest(sourceMap generator.SourceMap) string {
	var canonical strings.Builder
	for _, mapping := range sourceMap.Mappings {
		fmt.Fprintf(&canonical, "%s\x00%s\x00%d\x00%d\x00%d\x00%d\n", mapping.SemanticID, mapping.Kind,
			mapping.Source.Start.Offset, mapping.Source.End.Offset, mapping.Generated.Start.Offset, mapping.Generated.End.Offset)
	}
	return semantic.StableHash([]byte(canonical.String()))
}
func previousDigest(previous []byte) string {
	if len(previous) == 0 {
		return ""
	}
	return semantic.StableHash(previous)
}

type ownedSlot struct {
	id        string
	bodyStart int
}
