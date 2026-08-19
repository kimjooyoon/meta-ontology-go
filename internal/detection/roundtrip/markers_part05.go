package roundtrip

import (
	"fmt"
)

func canonicalRegionBody(source []byte) ([]byte, error) {
	var result []byte
	insideSlot := false
	for _, line := range splitSourceLines(source) {
		marker, _, ok, err := parseMarker(line.Text)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line.Number, err)
		}
		if !ok {
			if !insideSlot {
				result = append(result, source[line.Start:line.End]...)
			}
			continue
		}
		switch marker {
		case "slot-start":
			if insideSlot {
				return nil, fmt.Errorf("line %d: nested slot", line.Number)
			}
			insideSlot = true
			result = append(result, source[line.Start:line.End]...)
		case "slot-end":
			if !insideSlot {
				return nil, fmt.Errorf("line %d: slot ends without start", line.Number)
			}
			insideSlot = false
			result = append(result, source[line.Start:line.End]...)
		default:
			return nil, fmt.Errorf("line %d: generated marker is not nested legally", line.Number)
		}
	}
	if insideSlot {
		return nil, fmt.Errorf("unterminated slot")
	}
	return result, nil
}
