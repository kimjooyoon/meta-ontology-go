package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageutility"

func generationCellCounts(cell *languageutility.CellResult, evidenceState string) (fulfilled, unknown, refuted int) {
	if cell == nil {
		return 0, 1, 0
	}
	switch cell.State {
	case languageutility.StateOpen:
		// A classified open cell records a known gap in the declared population.
		return 0, 0, 0
	case languageutility.StateRefuted:
		return 0, 0, 1
	case languageutility.StateClosed:
		if cell.EvidencePath != "" && cell.EvidenceDigest != "" && evidenceState == "PASS" {
			return 1, 0, 0
		}
	}
	return 0, 1, 0
}
