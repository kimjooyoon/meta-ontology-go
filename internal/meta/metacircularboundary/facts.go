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
	attempts := make(map[string]Attempt, CaseTotal)
	for _, computation := range source.Computations {
		id, attempt, err := parseComputation(computation, source.SourceDigest)
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

func parseComputation(computation Computation, sourceDigest string) (string, Attempt, error) {
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
		case "id", "description", "capability", "issuer", "subject", "operation", "scope", "handle", "request_execution":
		default:
			return id, Attempt{FactActivity: computation.Activity, Unknown: true}, fmt.Errorf("unknown computation field %q", key)
		}
	}
	description, ok := fields["description"]
	if !ok {
		return id, Attempt{FactActivity: computation.Activity, Unknown: true}, nil
	}
	descriptionDigest := sourceDigest
	if description != "source" {
		descriptionDigest = digestBytes([]byte(description))
	}
	requestExecution, err := strconv.ParseBool(fields["request_execution"])
	if err != nil {
		return id, Attempt{FactActivity: computation.Activity, Unknown: true}, nil
	}
	attempt := Attempt{FactActivity: computation.Activity, DescriptionDigest: descriptionDigest, RequestExecution: requestExecution, Contradictory: contradictory}
	capabilityKind, ok := fields["capability"]
	if !ok {
		attempt.Unknown = true
		return id, attempt, nil
	}
	if capabilityKind == "none" {
		for _, key := range []string{"issuer", "subject", "operation", "scope", "handle"} {
			if _, exists := fields[key]; exists {
				attempt.Unknown = true
				return id, attempt, nil
			}
		}
		return id, attempt, nil
	}
	if capabilityKind != "present" {
		attempt.Unknown = true
		return id, attempt, nil
	}
	issuer, issuerOK := fields["issuer"]
	subject, subjectOK := fields["subject"]
	operation, operationOK := fields["operation"]
	scope, scopeOK := fields["scope"]
	handle, handleOK := fields["handle"]
	if !issuerOK || !subjectOK || !operationOK || !scopeOK || !handleOK {
		attempt.Unknown = true
		return id, attempt, nil
	}
	subjectDigest := subject
	if subject == "source" {
		subjectDigest = sourceDigest
	}
	handleValue := handle
	switch handle {
	case "fixture":
		handleValue = capabilityHandle(sourceDigest)
	case "forged":
		handleValue = digestBytes([]byte("forged|" + sourceDigest))
	}
	attempt.Capability = &Capability{Issuer: issuer, SubjectDigest: subjectDigest, Operation: operation, Scope: scope, Handle: handleValue}
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

func contractCases() []CaseDefinition {
	return []CaseDefinition{
		{ID: "description-only", ProofChoice: ProofRegression, MetaOperation: "deny-description-authority-escalation"},
		{ID: "explicit-read-only-capability", ProofChoice: ProofCoherence, MetaOperation: "accept-explicit-read-only-capability"},
		{ID: "forged-capability", ProofChoice: ProofRegression, MetaOperation: "reject-forged-capability"},
		{ID: "write-capability-out-of-scope", ProofChoice: ProofRegression, MetaOperation: "reject-out-of-scope-capability"},
	}
}
