package reproducibilitysemantics

import (
	"encoding/json"
	"fmt"
	"os"
)

func WriteJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}
	return nil
}

func ReadReceipt(path string) (Receipt, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Receipt{}, fmt.Errorf("read receipt: %w", err)
	}
	var receipt Receipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return Receipt{}, fmt.Errorf("decode receipt: %w", err)
	}
	return receipt, nil
}
