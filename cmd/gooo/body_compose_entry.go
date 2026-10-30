package main

import (
	"errors"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

var errSavedCompositionEntry = errors.New("saved composition entry mismatch")

// A repeated selector asserts the saved root; it never selects another root.
// Decode and execution still independently validate the saved source and plan.
func readSavedCompositionForEntry(path, entry string) (bodyexecution.Composition, error) {
	raw, err := readBodyExecutionFile(path, 32<<20)
	if err != nil {
		return bodyexecution.Composition{}, err
	}
	prior, err := bodyexecution.DecodeComposition(raw)
	if err != nil {
		return bodyexecution.Composition{}, err
	}
	if entry != "" && entry != prior.Plan.EntryActivity {
		return bodyexecution.Composition{}, fmt.Errorf("%w: requested %q, saved %q; use the saved entry",
			errSavedCompositionEntry, entry, prior.Plan.EntryActivity)
	}
	return prior, nil
}
