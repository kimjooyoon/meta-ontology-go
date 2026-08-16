package roundtrip

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// CheckLocality reports generated regions that changed outside the accepted
// semantic locality. Text outside generated markers is deliberately ignored:
// it belongs to handwritten implementation slots or other source owners.
func CheckLocality(input LocalityInput) Report {
	before, beforeErr := parseGeneratedFile(input.Before)
	after, afterErr := parseGeneratedFile(input.After)
	var report Report
	if beforeErr != nil {
		report.add(Violation{Rule: RuleMarker, Path: "before-go", Detail: beforeErr.Error()})
	}
	if afterErr != nil {
		report.add(Violation{Rule: RuleMarker, Path: "after-go", Detail: afterErr.Error()})
	}
	if !report.OK() {
		report.normalize()
		return report
	}
	allowed := make(map[semantic.ID]struct{}, len(input.AllowedIDs))
	for _, rawID := range input.AllowedIDs {
		id, err := semantic.ParseIdentity(rawID.String())
		if err != nil {
			report.add(Violation{Rule: RuleSnapshot, Path: "allowed-id", Identity: rawID.String(), Detail: err.Error()})
			continue
		}
		allowed[id] = struct{}{}
	}
	if !report.OK() {
		report.normalize()
		return report
	}
	ids := regionIDs(before.Regions, after.Regions)
	for _, id := range ids {
		old, oldExists := before.Regions[id]
		current, currentExists := after.Regions[id]
		if oldExists && currentExists && sameRegion(old, current) {
			continue
		}
		if _, permitted := allowed[id]; permitted {
			continue
		}
		detail := "generated region changed outside semantic locality"
		if !oldExists {
			detail = "generated region was added outside semantic locality"
		}
		if !currentExists {
			detail = "generated region was removed outside semantic locality"
		}
		report.add(Violation{Rule: RuleLocality, Path: "generated-go", Identity: id.String(), Detail: detail})
	}
	report.normalize()
	return report
}

func sameRegion(left, right generatedRegion) bool {
	return left.Kind == right.Kind && string(left.Body) == string(right.Body)
}

func regionIDs(left, right map[semantic.ID]generatedRegion) []semantic.ID {
	values := make(map[semantic.ID]struct{}, len(left)+len(right))
	for id := range left {
		values[id] = struct{}{}
	}
	for id := range right {
		values[id] = struct{}{}
	}
	result := make([]semantic.ID, 0, len(values))
	for id := range values {
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func parseAttributes(input string) (map[string]string, error) {
	attributes := make(map[string]string)
	for index := 0; index < len(input); {
		index = skipSpaces(input, index)
		if index == len(input) {
			break
		}
		keyStart := index
		for index < len(input) && attributeKeyChar(input[index]) {
			index++
		}
		if keyStart == index || index == len(input) || input[index] != '=' {
			return nil, fmt.Errorf("expected quoted attribute near byte %d", keyStart)
		}
		key := input[keyStart:index]
		index++
		if index == len(input) || input[index] != '"' {
			return nil, fmt.Errorf("attribute %q must use a quoted value", key)
		}
		valueStart := index
		index++
		escaped, closed := false, false
		for index < len(input) {
			if escaped {
				escaped = false
				index++
				continue
			}
			switch input[index] {
			case '\\':
				escaped = true
			case '"':
				index++
				closed = true
				goto valueDone
			}
			index++
		}
	valueDone:
		if !closed {
			return nil, fmt.Errorf("unterminated attribute %q", key)
		}
		value, err := strconv.Unquote(input[valueStart:index])
		if err != nil {
			return nil, fmt.Errorf("invalid attribute %q: %w", key, err)
		}
		if _, exists := attributes[key]; exists {
			return nil, fmt.Errorf("duplicate attribute %q", key)
		}
		attributes[key] = value
	}
	return attributes, nil
}

func splitSourceLines(source []byte) []markerLine {
	if len(source) == 0 {
		return nil
	}
	lines := make([]markerLine, 0, bytes.Count(source, []byte{'\n'})+1)
	for start := 0; start < len(source); {
		newline := bytes.IndexByte(source[start:], '\n')
		end := len(source)
		if newline >= 0 {
			end = start + newline + 1
		}
		textEnd := end
		if textEnd > start && source[textEnd-1] == '\n' {
			textEnd--
		}
		if textEnd > start && source[textEnd-1] == '\r' {
			textEnd--
		}
		lines = append(lines, markerLine{Number: len(lines) + 1, Text: string(source[start:textEnd]), Start: start, End: end})
		if newline < 0 {
			break
		}
		start = end
	}
	return lines
}

func skipSpaces(value string, index int) int {
	for index < len(value) && (value[index] == ' ' || value[index] == '\t') {
		index++
	}
	return index
}

func attributeKeyChar(value byte) bool {
	return value == '_' || value == '-' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}
