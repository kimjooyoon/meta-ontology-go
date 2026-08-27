package metacircularboundaryconsumer

import (
	"fmt"
	"strconv"
	"strings"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

// parseAttempts independently reads case facts from the source computations.
// Only the wire model and the general parser/lowerer are shared with producer.
func parseAttempts(source contract.SourceObservation) (map[string]contract.Attempt, error) {
	attempts := make(map[string]contract.Attempt, len(expectedCases()))
	for _, computation := range source.Computations {
		id, attempt, err := parseComputation(computation, source.SemanticDigest)
		if err != nil {
			return nil, err
		}
		if _, exists := attempts[id]; exists {
			return nil, fmt.Errorf("duplicate case computation %q", id)
		}
		attempts[id] = attempt
	}
	for _, definition := range expectedCases() {
		if _, ok := attempts[definition.ID]; !ok {
			return nil, fmt.Errorf("missing case computation %q", definition.ID)
		}
	}
	return attempts, nil
}

func parseComputation(computation contract.Computation, semanticDigest string) (string, contract.Attempt, error) {
	parts := strings.Split(computation.Program, "|")
	if len(parts) == 0 || parts[0] != "meta-circular-boundary.case" {
		return "", contract.Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("unknown computation prefix in %q", computation.Activity)
	}
	fields := map[string]string{}
	contradictory := false
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || value == "" {
			return "", contract.Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("malformed computation field in %q", computation.Activity)
		}
		if previous, exists := fields[key]; exists {
			if previous != value {
				contradictory = true
			} else {
				return "", contract.Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("duplicate computation field %q", key)
			}
		}
		fields[key] = value
	}
	id, ok := fields["id"]
	if !ok || id == "" {
		return "", contract.Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("case computation has no id")
	}
	if expectedActivity := activityForCase(id); expectedActivity == "" {
		return id, contract.Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("unknown case id %q", id)
	} else if computation.Activity != expectedActivity {
		return id, contract.Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("case computation activity %q does not bind %q", computation.Activity, id)
	}
	for key := range fields {
		switch key {
		case "id", "description", "request", "request_execution", "description_authority":
		default:
			return id, contract.Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("unknown computation field %q", key)
		}
	}
	description, ok := fields["description"]
	if !ok {
		return id, contract.Attempt{FactActivity: computation.Activity, Unknown: true}, nil
	}
	descriptionDigest := semanticDigest
	if description != "source" {
		descriptionDigest = digestBytes([]byte(description))
	}
	requestExecution, err := strconv.ParseBool(fields["request_execution"])
	if err != nil {
		return id, contract.Attempt{FactActivity: computation.Activity, Unknown: true}, nil
	}
	requestKind, requestOK := fields["request"]
	if !requestOK || (requestKind != requestNone && requestKind != requestReadOnly) {
		return id, contract.Attempt{FactActivity: computation.Activity, DescriptionDigest: descriptionDigest, RequestExecution: requestExecution, Contradictory: contradictory, Unknown: true}, nil
	}
	attempt := contract.Attempt{FactActivity: computation.Activity, DescriptionDigest: descriptionDigest, RequestKind: requestKind, RequestExecution: requestExecution, Contradictory: contradictory}
	if authority, ok := fields["description_authority"]; ok {
		attempt.DescriptionAuthorityClaim = authority == "GRANTED"
		if authority != "GRANTED" && authority != "NONE" {
			attempt.Unknown = true
		}
	}
	if attempt.DescriptionAuthorityClaim || requestKind == requestNone {
		attempt.Predicate = predicateDescriptionOnly
	} else {
		attempt.Predicate = predicateExplicitGrant
	}
	return id, attempt, nil
}

func activityForCase(id string) string {
	switch id {
	case "description-only":
		return "DescriptionOnlyAttempt"
	case "explicit-read-only-capability":
		return "ExplicitReadOnlyCapabilityAttempt"
	case "forged-capability":
		return "ForgedCapabilityAttempt"
	case "write-capability-out-of-scope":
		return "WriteCapabilityAttempt"
	default:
		return ""
	}
}
