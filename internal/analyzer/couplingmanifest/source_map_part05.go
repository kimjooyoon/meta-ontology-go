package couplingmanifest

import (
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func requireCoverage(surfaces map[semantic.ID]detector.Surface, before, head map[semantic.ID]resolved) error {
	for _, surface := range surfaces {
		if _, beforeOK := before[surface.SemanticOwnerID]; !beforeOK {
			if _, headOK := head[surface.SemanticOwnerID]; !headOK {
				return unknownError(CodeUnknownChangedSurface, "registered surface has no before or head observation")
			}
		}
	}
	return nil
}
func sortedSurfaces(values []detector.Surface) []detector.Surface {
	result := append([]detector.Surface(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].SurfaceID < result[j].SurfaceID })
	return result
}
