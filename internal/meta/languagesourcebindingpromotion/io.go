package languagesourcebindingpromotion

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

func LoadContract(path string) (Contract, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, err
	}
	contract, err := decodeStrict[Contract](raw)
	if err != nil {
		return Contract{}, err
	}
	if !reflect.DeepEqual(contract, CanonicalContract()) {
		return Contract{}, fmt.Errorf("source binding promotion contract mismatch")
	}
	return contract, nil
}

func LoadIndependence(path string) (IndependenceEvidence, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return IndependenceEvidence{}, err
	}
	return decodeStrict[IndependenceEvidence](raw)
}

func LoadReport(path string) (Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	return decodeStrict[Report](raw)
}

func WriteReport(path string, report Report) error {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}
