package counterexamplefirst

import (
	"encoding/json"
	"fmt"
	"os"
)

func DecodeContract(raw []byte) (Contract, error) {
	var value Contract
	if err := json.Unmarshal(raw, &value); err != nil {
		return Contract{}, fmt.Errorf("decode counterexample contract: %w", err)
	}
	if !ValidContract(value) {
		return Contract{}, fmt.Errorf("COUNTEREXAMPLE_CONTRACT_DRIFT")
	}
	return value, nil
}

func DecodeCorpus(raw []byte) (ScenarioCorpus, error) {
	var value ScenarioCorpus
	if err := json.Unmarshal(raw, &value); err != nil {
		return ScenarioCorpus{}, fmt.Errorf("decode counterexample scenarios: %w", err)
	}
	if value.Schema != CorpusSchema || value.Version != 1 || len(value.Scenarios) != CaseCount {
		return ScenarioCorpus{}, fmt.Errorf("COUNTEREXAMPLE_CORPUS_DRIFT")
	}
	return value, nil
}

func WriteReceipts(path string, receipts []DecisionReceipt) error {
	raw, err := json.MarshalIndent(receipts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func DecodeReceipts(raw []byte) ([]DecisionReceipt, error) {
	var value []DecisionReceipt
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("decode counterexample receipts: %w", err)
	}
	return value, nil
}

func WriteReport(path string, report Report) error {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
