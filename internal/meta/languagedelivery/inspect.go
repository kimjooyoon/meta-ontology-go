package languagedelivery

import "encoding/json"

func inspectEvidence(manifest SourceManifest, evidence EvidenceSet, head string) ([]SourceObservation, decodedEvidence) {
	issues := manifestIssues(manifest, evidence, head)
	var decoded decodedEvidence
	observations := make([]SourceObservation, 0, len(sourceOrder))
	for _, source := range sourceOrder {
		entry, _ := manifest.entry(source)
		if issue := issues[source]; issue != "" {
			observations = append(observations, unknownObservation(source, entry, issue))
			continue
		}
		observation := inspectOne(source, evidence.Bytes(source), head, &decoded, entry)
		observations = append(observations, observation)
	}
	return observations, decoded
}

func inspectOne(source SourceName, data []byte, head string, decoded *decodedEvidence, entry ManifestEntry) SourceObservation {
	switch source {
	case SourceUserJourney:
		return inspectJourney(data, head, &decoded.Journey, entry)
	case SourceConformance:
		return inspectConformance(data, head, &decoded.Conformance, entry)
	case SourceLSP:
		return inspectLSP(data, head, &decoded.LSP, entry)
	case SourceRelease:
		return inspectRelease(data, head, &decoded.Release, entry)
	case SourceExecution:
		return inspectExecution(data, head, &decoded.Execution, entry)
	case SourceProfile:
		return inspectProfile(data, head, &decoded.Profile, entry)
	case SourceReadiness:
		return inspectReadiness(data, head, &decoded.Readiness, entry)
	default:
		return unknownObservation(source, entry, "SOURCE_NAME_UNKNOWN")
	}
}

func unmarshalReceipt(data []byte, value any) error {
	return json.Unmarshal(data, value)
}
