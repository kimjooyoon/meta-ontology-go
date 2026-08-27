package experimentpromotion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func LoadContract(filename string) (Contract, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return Contract{}, err
	}
	var contract Contract
	if err := decodeStrict(raw, &contract); err != nil {
		return Contract{}, fmt.Errorf("contract: %w", err)
	}
	if err := ValidateContract(contract); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

func LoadObservation(filename string) (ObservationBundle, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return ObservationBundle{}, err
	}
	return DecodeObservation(raw)
}

func DecodeObservation(raw []byte) (ObservationBundle, error) {
	var bundle ObservationBundle
	if err := decodeStrict(raw, &bundle); err != nil {
		return ObservationBundle{}, fmt.Errorf("observation bundle: %w", err)
	}
	return bundle, nil
}

func WriteReport(filename string, report Report) error {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(filename, raw, 0o644)
}

func WriteVerification(filename string, verification Verification) error {
	raw, err := json.MarshalIndent(verification, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(filename, raw, 0o644)
}

func LoadReport(filename string) (Report, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return Report{}, err
	}
	var report Report
	if err := decodeStrict(raw, &report); err != nil {
		return Report{}, fmt.Errorf("report: %w", err)
	}
	return report, nil
}

func decodeStrict(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON")
		}
		return err
	}
	return nil
}
