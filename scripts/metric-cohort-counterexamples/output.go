package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func inputPayloads(inputs co.SourceInputs) map[string][]byte {
	return map[string][]byte{
		"strategy-plan.json": inputs.Strategy, "strategy-verification.json": inputs.StrategyVerification,
		"source-metrics.json": inputs.SourceMetrics, "intervention-ledger.json": inputs.Intervention,
		"intervention-verification.json": inputs.InterventionVerification, "program.json": inputs.Program,
		"program.gooo": inputs.ProgramSource, "verification.json": inputs.ProgramVerification,
	}
}

func inputDigests(v variant) map[string]string {
	payloads := inputPayloads(v.Inputs)
	payloads["concept-operation-binding.json"] = v.Receipt
	digests := make(map[string]string, len(payloads))
	for name, raw := range payloads {
		digests[name] = artifact.ContentDigest(raw)
	}
	return digests
}

func writeReport(directory string, value report) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "report.json"), append(raw, '\n'), 0640)
}
