package generation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

func (envelope *ArtifactProvenance) UnmarshalJSON(data []byte) error {
	type wire ArtifactProvenance
	var decoded wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return fmt.Errorf("decode artifact provenance: %w", err)
	}
	if err := ensureIndicatorLedgerEOF(decoder); err != nil {
		return fmt.Errorf("decode artifact provenance: %w", err)
	}
	candidate := ArtifactProvenance(decoded)
	if candidate.SchemaVersion != ArtifactProvenanceSchemaVersion ||
		!validArtifactProvenanceIndicators(candidate.Indicators) {
		return fmt.Errorf("invalid artifact provenance schema or indicators")
	}
	if !validIndicatorDecisionLedgerDigest(candidate.IndicatorDecisionLedgerDigest) ||
		candidate.IndicatorDecisionLedgerCount < 0 ||
		!validDigest(candidate.PlanDigest) ||
		!validDigest(candidate.ExecutionManifestDigest) ||
		!validDigest(candidate.ReceiptReportDigest) {
		return fmt.Errorf("invalid artifact provenance digests")
	}
	canonical := candidate
	canonical.InputDigest, canonical.EnvelopeDigest, canonical.ReplayDigest = "", "", ""
	canonical = finishArtifactProvenance(canonical)
	if !reflect.DeepEqual(candidate, canonical) {
		return fmt.Errorf("artifact provenance canonical replay mismatch")
	}
	*envelope = candidate
	return nil
}

func validArtifactProvenanceIndicators(
	indicators []ArtifactProvenanceIndicator,
) bool {
	expected := map[string]TrilemmaRoute{
		"foundation.plan-ledger":      TrilemmaRouteFoundation,
		"coherence.execution-ledger":  TrilemmaRouteCoherence,
		"coherence.receipt-ledger":    TrilemmaRouteCoherence,
		"regression.canonical-replay": TrilemmaRouteRegression,
	}
	if len(indicators) != len(expected) {
		return false
	}
	for index, indicator := range indicators {
		route, exists := expected[indicator.ID]
		if !exists || route != indicator.Route || !validDigest(indicator.EvidenceDigest) ||
			(index != 0 && indicators[index-1].ID >= indicator.ID) {
			return false
		}
		switch indicator.Verdict {
		case IndicatorVerdictPass, IndicatorVerdictFail, IndicatorVerdictUnknown:
		default:
			return false
		}
	}
	return true
}
