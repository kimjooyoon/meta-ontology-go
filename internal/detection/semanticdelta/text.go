package semanticdelta

import (
	"bufio"
	"fmt"
	"strconv"
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

func parseDeltaLine(request *Request, fields []string) error {
	if len(fields) < 3 {
		return fmt.Errorf("delta line requires operation and kind")
	}
	operation := fields[1]
	switch operation {
	case "add-node", "remove-node":
		if len(fields) != 4 {
			return fmt.Errorf("%s requires kind and ID", operation)
		}
		node := Node{Kind: fields[2], ID: fields[3]}
		if operation == "add-node" {
			request.Delta.AddedNodes = append(request.Delta.AddedNodes, node)
		} else {
			request.Delta.RemovedNodes = append(request.Delta.RemovedNodes, node)
		}
	case "add-fact", "remove-fact":
		if len(fields) != 5 {
			return fmt.Errorf("%s requires predicate, subject, and object", operation)
		}
		fact := Fact{Predicate: fields[2], Subject: fields[3], Object: fields[4]}
		if operation == "add-fact" {
			request.Delta.AddedFacts = append(request.Delta.AddedFacts, fact)
		} else {
			request.Delta.RemovedFacts = append(request.Delta.RemovedFacts, fact)
		}
	default:
		return fmt.Errorf("unknown delta operation %q", operation)
	}
	return nil
}

// EncodeText returns the canonical line-oriented representation.
func EncodeText(request Request) ([]byte, error) {
	normalized, err := request.Normalized()
	if err != nil {
		return nil, err
	}
	var builder strings.Builder
	builder.WriteString("version\t")
	builder.WriteString(normalized.Version)
	builder.WriteByte('\n')
	for _, id := range normalized.Allowed.IDs {
		writeTextRecord(&builder, "scope", "id", id)
	}
	for _, prefix := range normalized.Allowed.Prefixes {
		writeTextRecord(&builder, "scope", "prefix", prefix)
	}
	for _, predicate := range normalized.Allowed.Predicates {
		writeTextRecord(&builder, "scope", "predicate", predicate)
	}
	for _, node := range normalized.Delta.AddedNodes {
		writeTextRecord(&builder, "delta", "add-node", node.Kind, node.ID)
	}
	for _, node := range normalized.Delta.RemovedNodes {
		writeTextRecord(&builder, "delta", "remove-node", node.Kind, node.ID)
	}
	for _, fact := range normalized.Delta.AddedFacts {
		writeTextRecord(&builder, "delta", "add-fact", fact.Predicate, fact.Subject, fact.Object)
	}
	for _, fact := range normalized.Delta.RemovedFacts {
		writeTextRecord(&builder, "delta", "remove-fact", fact.Predicate, fact.Subject, fact.Object)
	}
	return []byte(builder.String()), nil
}

func writeTextRecord(builder *strings.Builder, fields ...string) {
	builder.WriteString(strings.Join(fields, "\t"))
	builder.WriteByte('\n')
}

// EncodeReportText returns a stable tab-separated report intended for logs.
func EncodeReportText(report Report) []byte {
	report.Normalize()
	var builder strings.Builder
	writeTextRecord(&builder, "allowed", strconv.FormatBool(report.Passes()))
	writeTextRecord(&builder, "violations", strconv.Itoa(len(report.Violations)))
	for _, violation := range report.Violations {
		writeTextRecord(&builder, "violation", string(violation.Operation), string(violation.Change),
			violation.ID, violation.Kind, violation.Subject, violation.Predicate,
			violation.Object, violation.Endpoint, strconv.Quote(violation.Reason))
	}
	return []byte(builder.String())
}
