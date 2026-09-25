package languagetestexperiment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func ReadInput(path string) (Input, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return Input{}, err
	}
	var input Input
	if err := json.Unmarshal(payload, &input); err != nil {
		return Input{}, err
	}
	return input, nil
}

func Marshal(report Report) ([]byte, error) {
	if report.Digest != reportDigest(report) {
		return nil, fmt.Errorf("LANGUAGE_TEST_REPORT_DIGEST_MISMATCH")
	}
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func WriteReport(path string, report Report) error {
	payload, err := Marshal(report)
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func CheckReport(path string, report Report) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	payload, err := Marshal(report)
	if err != nil {
		return err
	}
	if !bytes.Equal(existing, payload) {
		return fmt.Errorf("LANGUAGE_TEST_REPORT_REPLAY_MISMATCH")
	}
	return nil
}
