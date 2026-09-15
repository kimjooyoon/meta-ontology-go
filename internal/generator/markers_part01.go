package generator

import (
	"fmt"
	"strconv"
)

const (
	generatedStartPrefix = "//gooo:generated:start"
	generatedEndPrefix   = "//gooo:generated:end"
	slotStartPrefix      = "//gooo:slot:start"
	slotEndPrefix        = "//gooo:slot:end"
)

type sourceLine struct {
	start int
	end   int
	text  string
}
type generatedRegion struct {
	ID        string
	Kind      string
	Start     int
	End       int
	StartLine int
	EndLine   int
	Slots     []parsedSlot
}
type parsedSlot struct {
	ID         string
	RegionID   string
	RegionKind string
	Start      int
	End        int
	StartLine  int
	EndLine    int
	Body       []byte
}
type parsedMarkers struct {
	Regions []generatedRegion
	Slots   map[string]parsedSlot
}

func generatedMarker(prefix, id, kind string) string {
	if kind == "" {
		return prefix + " id=" + strconv.Quote(id)
	}
	return prefix + " id=" + strconv.Quote(id) + " kind=" + strconv.Quote(kind)
}
func slotMarker(prefix, id string) string {
	return prefix + " id=" + strconv.Quote(id)
}
func parseMarkers(source []byte) (parsedMarkers, error) {
	state := markerState{result: parsedMarkers{Slots: make(map[string]parsedSlot)}}
	for index, line := range splitSourceLines(source) {
		marker, attrs, ok, err := parseMarker(line.text)
		if err != nil {
			return parsedMarkers{}, fmt.Errorf("generator: malformed marker on line %d: %w", index+1, err)
		}
		if !ok {
			continue
		}
		if err := state.apply(marker, attrs, line, index, source); err != nil {
			return parsedMarkers{}, err
		}
	}
	return state.finish()
}
