package evidencequorumconsumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumwire"
)

func decodeStrict[T any](raw []byte) (T, error) {
	var value T
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return value, fmt.Errorf("trailing JSON input")
	}
	return value, nil
}

func DecodeReceipt(raw []byte) (evidencequorumwire.Receipt, error) {
	return decodeStrict[evidencequorumwire.Receipt](raw)
}

func WriteReport(path string, report Report) error {
	if path == "" {
		return fmt.Errorf("report output is required")
	}
	raw, err := json.Marshal(report)
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func LoadReport(path string) (Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	return decodeStrict[Report](raw)
}
