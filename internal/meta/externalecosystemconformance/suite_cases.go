package externalecosystemconformance

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed evidence/gomacro.json
var referenceJSON []byte

func Reference() (Capsule, error) {
	var capsule Capsule
	if err := json.Unmarshal(referenceJSON, &capsule); err != nil {
		return Capsule{}, fmt.Errorf("decode reference capsule: %w", err)
	}
	return capsule, nil
}

func cloneCapsule(source Capsule) Capsule {
	clone := source
	clone.Documents = append([]Document(nil), source.Documents...)
	clone.Capabilities = append([]Capability(nil), source.Capabilities...)
	return clone
}

func cloneEvidence(source Evidence) Evidence {
	clone := source
	clone.Readme = append([]byte(nil), source.Readme...)
	clone.GoMod = append([]byte(nil), source.GoMod...)
	return clone
}

func RunCase(subject, caseID string, source Capsule, sourceEvidence Evidence) Report {
	capsule := cloneCapsule(source)
	evidence := cloneEvidence(sourceEvidence)
	switch caseID {
	case "exact":
	case "readme-unavailable":
		evidence.Readme = nil
	case "gomod-unavailable":
		evidence.GoMod = nil
	case "unknown-relation":
		capsule.Capabilities[0].Relation = "UNRECOGNIZED"
	case "readme-digest-mismatch":
		evidence.Readme = append(evidence.Readme, 0)
	case "gomod-digest-mismatch":
		evidence.GoMod = append(evidence.GoMod, 0)
	case "commit-mismatch":
		capsule.CommitSHA = "0000000000000000000000000000000000000000"
	case "license-mismatch":
		capsule.LicenseSPDX = "UNKNOWN"
	case "external-execution":
		evidence.ExternalExecutions = 1
	case "observed-write":
		evidence.RepositoryWrites = 1
	default:
		return fail(baseReport(subject, capsule.ReferenceID, evidence), ResolutionUnknown, ReasonCaseUnknown)
	}
	return Evaluate(subject, capsule, evidence)
}

func expectedCase(caseID string) (string, string) {
	switch caseID {
	case "exact":
		return DecisionReferenceBound, ResolutionExact
	case "readme-unavailable", "gomod-unavailable", "unknown-relation":
		return DecisionFailClosed, ResolutionUnknown
	default:
		return DecisionFailClosed, ResolutionInvariant
	}
}
