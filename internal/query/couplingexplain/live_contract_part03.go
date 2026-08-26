package couplingexplain

import (
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
)

func acceptedIDs(result detector.Result) []string {
	ids := make([]string, 0, len(result.AcceptedSurfaceIDs))
	for _, id := range result.AcceptedSurfaceIDs {
		ids = append(ids, id.String())
	}
	return sortedStrings(ids)
}
