package generation

import (
	"encoding/json"
	"sort"
)

type executionInput struct {
	PlanDigest                    string          `json:"plan_digest"`
	IndicatorDecisionLedgerDigest string          `json:"indicator_decision_ledger_digest,omitempty"`
	IndicatorDecisionLedgerCount  int             `json:"indicator_decision_ledger_count"`
	Steps                         []ExecutionStep `json:"steps"`
}

func finishExecutionManifest(manifest ExecutionManifest) ExecutionManifest {
	if manifest.Steps == nil {
		manifest.Steps = []ExecutionStep{}
	}
	sort.Slice(manifest.Steps, func(i, j int) bool {
		return manifest.Steps[i].ActionIndicatorID <
			manifest.Steps[j].ActionIndicatorID
	})
	manifest.InputDigest = digestJSON(executionInput{
		PlanDigest: manifest.PlanDigest, IndicatorDecisionLedgerDigest: manifest.IndicatorDecisionLedgerDigest,
		IndicatorDecisionLedgerCount: manifest.IndicatorDecisionLedgerCount, Steps: manifest.Steps,
	})
	manifest.ManifestDigest, manifest.ReplayDigest = "", ""
	manifest.ManifestDigest = digestJSON(manifest)
	manifest.ReplayDigest = digestPair(
		manifest.InputDigest,
		manifest.ManifestDigest,
	)
	return manifest
}

func EncodeExecutionManifest(manifest ExecutionManifest) ([]byte, error) {
	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}
