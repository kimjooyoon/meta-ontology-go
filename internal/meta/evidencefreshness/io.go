package evidencefreshness

import (
	"fmt"
	"os"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
)

func LoadContract(path string) (model.Contract, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return model.Contract{}, err
	}
	contract, err := model.DecodeStrict[model.Contract](raw)
	if err != nil {
		return model.Contract{}, err
	}
	if !reflect.DeepEqual(contract, CanonicalContract()) {
		return model.Contract{}, fmt.Errorf("evidence freshness contract mismatch")
	}
	return contract, nil
}

func LoadIndependence(path string) (model.IndependenceEvidence, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return model.IndependenceEvidence{}, err
	}
	return model.DecodeStrict[model.IndependenceEvidence](raw)
}

func LoadReport(path string) (model.Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return model.Report{}, err
	}
	return model.DecodeStrict[model.Report](raw)
}

func WriteJSON(path string, value any) error {
	raw, err := model.Marshal(value)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
