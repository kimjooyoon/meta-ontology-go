package metacircularboundary

import (
	"fmt"
	"strconv"
	"strings"
)

// parseAttempts reads the executable value programs emitted by the Gooo source.
// The Go contract supplies only the fixed case IDs; all attempt facts come from
// the parsed/lowered source computation values.
func parseAttempts(source SourceObservation) (map[string]Attempt, error) {
	attempts := make(map[string]Attempt, len(contractCases()))
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
	for _, definition := range contractCases() {
		if _, ok := attempts[definition.ID]; !ok {
			return nil, fmt.Errorf("missing case computation %q", definition.ID)
		}
	}
	return attempts, nil
}

func parseComputation(computation Computation, semanticDigest string) (string, Attempt, error) {
	parts := strings.Split(computation.Program, "|")
	if len(parts) == 0 || parts[0] != "meta-circular-boundary.case" {
		return "", Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("unknown computation prefix in %q", computation.Activity)
	}
	fields := map[string]string{}
	contradictory := false
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || value == "" {
			return "", Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("malformed computation field in %q", computation.Activity)
		}
		if previous, exists := fields[key]; exists {
			if previous != value {
				contradictory = true
			} else {
				return "", Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("duplicate computation field %q", key)
			}
		}
		fields[key] = value
	}
	id, ok := fields["id"]
	if !ok || id == "" {
		return "", Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("case computation has no id")
	}
	if expectedActivity := activityForCase(id); expectedActivity == "" {
		return id, Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("unknown case id %q", id)
	} else if computation.Activity != expectedActivity {
		return id, Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("case computation activity %q does not bind %q", computation.Activity, id)
	}
	for key := range fields {
		switch key {
		case "id", "description", "request", "request_execution", "description_authority":
		default:
			return id, Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("unknown computation field %q", key)
		}
	}
	description, ok := fields["description"]
	if !ok {
		return id, Attempt{FactActivity: computation.Activity, Unknown: true}, nil
	}
	descriptionDigest := semanticDigest
	if description != "source" {
		descriptionDigest = digestBytes([]byte(description))
	}
	requestExecution, err := strconv.ParseBool(fields["request_execution"])
	if err != nil {
		return id, Attempt{FactActivity: computation.Activity, Unknown: true}, nil
	}
	requestKind, requestOK := fields["request"]
	if !requestOK || (requestKind != RequestNone && requestKind != RequestReadOnly) {
		attempt := Attempt{FactActivity: computation.Activity, DescriptionDigest: descriptionDigest, RequestExecution: requestExecution, Contradictory: contradictory, Unknown: true}
		return id, attempt, nil
	}
	attempt := Attempt{FactActivity: computation.Activity, DescriptionDigest: descriptionDigest, RequestKind: requestKind, RequestExecution: requestExecution, Contradictory: contradictory}
	if authority, ok := fields["description_authority"]; ok {
		attempt.DescriptionAuthorityClaim = authority == "GRANTED"
		if authority != "GRANTED" && authority != "NONE" {
			attempt.Unknown = true
		}
	}
	attempt.Predicate = predicateForRequest(requestKind, attempt.DescriptionAuthorityClaim)
	return id, attempt, nil
}

func predicateForRequest(requestKind string, descriptionAuthorityClaim bool) string {
	if descriptionAuthorityClaim {
		return PredicateDescriptionOnly
	}
	if requestKind == RequestNone {
		return PredicateDescriptionOnly
	}
	return PredicateExplicitGrant
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

func contractCases() []CaseDefinition {
	return []CaseDefinition{
		{ID: "description-only", Predicate: PredicateDescriptionOnly, ExpectedDecision: DecisionFailClosed, ExpectedAuthorization: AuthorizationDenied, ExpectedExecution: ExecutionBlocked, ExpectedReason: ReasonDescriptionOnly, ProofChoice: ProofRegression, MetaOperation: "deny-description-authority-escalation"},
		{ID: "explicit-read-only-capability", Predicate: PredicateExplicitGrant, ExpectedDecision: DecisionPass, ExpectedAuthorization: AuthorizationGranted, ExpectedExecution: ExecutionAllowed, ExpectedReason: ReasonExplicitCapability, ProofChoice: ProofCoherence, MetaOperation: "accept-explicit-read-only-capability"},
		{ID: "forged-capability", Predicate: PredicateForgedGrant, ExpectedDecision: DecisionFailClosed, ExpectedAuthorization: AuthorizationDenied, ExpectedExecution: ExecutionBlocked, ExpectedReason: ReasonForgedCapability, ProofChoice: ProofRegression, MetaOperation: "reject-forged-capability"},
		{ID: "write-capability-out-of-scope", Predicate: PredicateOutOfScopeGrant, ExpectedDecision: DecisionFailClosed, ExpectedAuthorization: AuthorizationDenied, ExpectedExecution: ExecutionBlocked, ExpectedReason: ReasonOutOfScopeCapability, ProofChoice: ProofRegression, MetaOperation: "reject-out-of-scope-capability"},
	}
}
