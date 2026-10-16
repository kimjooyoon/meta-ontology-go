package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
)

// ProjectionPart retains whole generated regions, including both ownership markers.
type ProjectionPart struct {
	Source    []byte
	RegionIDs []string
}

// PartitionProjection packs an import-free projection within a physical line budget.
// A region is indivisible; an oversized region is an error, never a partial result.
func PartitionProjection(source []byte, maximumLines int) ([]ProjectionPart, error) {
	markers, err := parseMarkers(source)
	if err != nil || len(markers.Regions) == 0 || maximumLines < 1 {
		return nil, fmt.Errorf("projection requires complete owned regions and a positive budget: %v", err)
	}
	prefix := source[:markers.Regions[0].Start]
	file, err := parser.ParseFile(token.NewFileSet(), "prefix.go", prefix, 0)
	if err != nil || len(file.Decls) != 0 {
		return nil, fmt.Errorf("partition projection requires an import-free package preamble")
	}
	var parts []ProjectionPart
	var current ProjectionPart
	var body []byte
	previous := markers.Regions[0].Start
	for _, region := range markers.Regions {
		if len(bytes.TrimSpace(source[previous:region.Start])) != 0 {
			return nil, fmt.Errorf("projection contains code outside owned regions")
		}
		chunk := bytes.TrimSpace(source[region.Start:region.End])
		candidate := append(append(append([]byte(nil), body...), chunk...), '\n', '\n')
		raw, err := format.Source(append(append([]byte(nil), prefix...), candidate...))
		if err != nil {
			return nil, err
		}
		if bytes.Count(raw, []byte{'\n'}) > maximumLines {
			if len(current.RegionIDs) == 0 {
				return nil, fmt.Errorf("generated region %s exceeds the physical line budget", region.ID)
			}
			parts = append(parts, current)
			current, body = ProjectionPart{}, nil
			raw, err = format.Source(append(append(append([]byte(nil), prefix...), chunk...), '\n'))
			if err != nil || bytes.Count(raw, []byte{'\n'}) > maximumLines {
				return nil, fmt.Errorf("generated region %s exceeds the physical line budget", region.ID)
			}
			candidate = append(append([]byte(nil), chunk...), '\n', '\n')
		}
		current.Source = raw
		current.RegionIDs = append(current.RegionIDs, region.ID)
		body, previous = candidate, region.End
	}
	if len(bytes.TrimSpace(source[previous:])) != 0 {
		return nil, fmt.Errorf("projection contains trailing unowned code")
	}
	return append(parts, current), nil
}
