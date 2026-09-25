package closure

import (
	"encoding/json"
	"fmt"
)

func decodeDocuments(in Input) (programDocument, verificationDocument, error) {
	var program programDocument
	if err := json.Unmarshal(in.ProgramJSON, &program); err != nil {
		return program, verificationDocument{}, fmt.Errorf("decode program: %w", err)
	}
	var verification verificationDocument
	if err := json.Unmarshal(in.VerificationJSON, &verification); err != nil {
		return program, verification, fmt.Errorf("decode verification: %w", err)
	}
	return program, verification, nil
}
