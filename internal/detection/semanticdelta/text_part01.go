package semanticdelta

import (
	"bufio"
	"fmt"
	"strings"
)

// DecodeText parses the deterministic line-oriented format:
//
// version semanticdelta/v1
// scope id billing://activity/pay-order
// delta add-node Activity billing://activity/pay-order
// delta add-fact gooo:invokes billing://activity/pay-order billing://activity/audit
func DecodeText(data []byte) (Request, error) {
	request := Request{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNumber := 0
	seenVersion := false
	for scanner.Scan() {
		lineNumber++
		fields := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		if err := parseTextLine(&request, fields, &seenVersion); err != nil {
			return Request{}, fmt.Errorf("text line %d: %w", lineNumber, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return Request{}, fmt.Errorf("decode semanticdelta text: %w", err)
	}
	if !seenVersion {
		return Request{}, fmt.Errorf("text input is missing version line")
	}
	return request.Normalized()
}
func parseTextLine(request *Request, fields []string, seenVersion *bool) error {
	switch fields[0] {
	case "version":
		if len(fields) != 2 || *seenVersion {
			return fmt.Errorf("expected one version line")
		}
		request.Version = fields[1]
		*seenVersion = true
		return nil
	case "scope":
		if len(fields) != 3 {
			return fmt.Errorf("scope line requires kind and value")
		}
		switch fields[1] {
		case "id":
			request.Allowed.IDs = append(request.Allowed.IDs, fields[2])
		case "prefix":
			request.Allowed.Prefixes = append(request.Allowed.Prefixes, fields[2])
		case "predicate":
			request.Allowed.Predicates = append(request.Allowed.Predicates, fields[2])
		default:
			return fmt.Errorf("unknown scope kind %q", fields[1])
		}
		return nil
	case "delta":
		return parseDeltaLine(request, fields)
	default:
		return fmt.Errorf("unknown record %q", fields[0])
	}
}
