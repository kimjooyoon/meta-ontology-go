package roundtrip

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	generatedStartPrefix = "//gooo:generated:start"
	generatedEndPrefix   = "//gooo:generated:end"
	slotStartPrefix      = "//gooo:slot:start"
	slotEndPrefix        = "//gooo:slot:end"
)

type generatedRegion struct {
	ID   semantic.ID
	Kind string
	Body []byte
}
type generatedFile struct {
	Regions map[semantic.ID]generatedRegion
}
type markerLine struct {
	Number     int
	Text       string
	Start, End int
}
type openRegion struct {
	ID           semantic.ID
	Kind         string
	ContentStart int
}

func parseGeneratedFile(source []byte) (generatedFile, error) {
	state := markerState{
		result:        generatedFile{Regions: make(map[semantic.ID]generatedRegion)},
		source:        source,
		closedSlotIDs: make(map[semantic.ID]struct{}),
	}
	for _, line := range splitSourceLines(source) {
		if err := state.apply(line); err != nil {
			return generatedFile{}, err
		}
	}
	return state.finish()
}

type markerState struct {
	result        generatedFile
	source        []byte
	open          *openRegion
	slotID        semantic.ID
	closedSlotIDs map[semantic.ID]struct{}
}
