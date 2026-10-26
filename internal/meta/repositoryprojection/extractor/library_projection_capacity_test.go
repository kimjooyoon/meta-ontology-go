package extractor

import "testing"

func TestLibraryNamingRenderedCapacity(t *testing.T) {
	// Keep typed renaming and its callback owner intact in materialized CI.
	// Resolve declarations by identity because their physical files may move.
	requireNamedCapacities(t, "../../../..", []string{
		"internal/bodylibrary#nameLibraryProjection",
		"internal/bodylibrary#renderLibraryProjection",
		"internal/bodylibrary#formatLibraryProjection",
	})
}
