package query

import (
	"sort"
)

func newResolutionRow(business, activity, generated ID, candidate bool) ResolutionRow {
	source := FactDeterministic.String()
	if candidate {
		source = FactCandidate.String()
	}
	return ResolutionRow{
		Business: business, Activity: activity, GeneratedEntity: generated,
		Depth: ResolutionMaxDepth, Status: DerivedFactStatus, SourceLayer: source,
	}
}
func resolutionKeys(rows []ResolutionRow) map[resolutionKey]struct{} {
	keys := make(map[resolutionKey]struct{}, len(rows))
	for _, row := range rows {
		keys[resolutionKey{row.Business, row.Activity, row.GeneratedEntity}] = struct{}{}
	}
	return keys
}
func sortResolutionRows(rows []ResolutionRow) {
	sort.Slice(rows, func(i, j int) bool {
		left, right := rows[i], rows[j]
		if left.Business != right.Business {
			return left.Business < right.Business
		}
		if left.Activity != right.Activity {
			return left.Activity < right.Activity
		}
		if left.GeneratedEntity != right.GeneratedEntity {
			return left.GeneratedEntity < right.GeneratedEntity
		}
		return left.SourceLayer < right.SourceLayer
	})
}
func resolutionMetadata(metadata ProjectionMetadata, status string) EnvelopeMetadata {
	result := envelopeMetadata(metadata)
	result.DerivedStatus = DerivedStatusNonAuthoritative
	result.AuthorityLabels = append(result.AuthorityLabels, AuthorityLabel{
		View: "resolution_view", Authority: "derived", Status: status,
	})
	return result
}
