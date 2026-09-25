package generation

import (
	"encoding/json"
	"reflect"
)

func artifactCanonical[T any](value T) bool {
	payload, err := json.Marshal(value)
	if err != nil {
		return false
	}
	var replay T
	if err := json.Unmarshal(payload, &replay); err != nil {
		return false
	}
	return reflect.DeepEqual(value, replay)
}

func artifactBindingVerdict(canonical, matches bool) IndicatorVerdict {
	if !canonical {
		return IndicatorVerdictUnknown
	}
	if !matches {
		return IndicatorVerdictFail
	}
	return IndicatorVerdictPass
}

func artifactProvenanceIndicator(
	id string,
	route TrilemmaRoute,
	verdict IndicatorVerdict,
	evidence ...string,
) ArtifactProvenanceIndicator {
	input := append([]string{id, string(route), string(verdict)}, evidence...)
	return ArtifactProvenanceIndicator{
		ID: id, Route: route, Verdict: verdict,
		EvidenceDigest: digestJSON(input),
	}
}

func executionMatchesProvenance(
	plan Plan, execution ExecutionManifest, digest string, count int,
) bool {
	return execution.BaseSHA == plan.BaseSHA && execution.HeadSHA == plan.HeadSHA &&
		execution.PlanDigest == plan.PlanDigest &&
		execution.IndicatorDecisionLedgerDigest == digest &&
		execution.IndicatorDecisionLedgerCount == count
}

func receiptsMatchProvenance(
	plan Plan, receipts ReceiptReport, digest string, count int,
) bool {
	return receipts.BaseSHA == plan.BaseSHA && receipts.HeadSHA == plan.HeadSHA &&
		receipts.PlanDigest == plan.PlanDigest &&
		receipts.IndicatorDecisionLedgerDigest == digest &&
		receipts.IndicatorDecisionLedgerCount == count
}
