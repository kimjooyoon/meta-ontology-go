package semanticdelta

import (
	"fmt"
	"strings"
)

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
