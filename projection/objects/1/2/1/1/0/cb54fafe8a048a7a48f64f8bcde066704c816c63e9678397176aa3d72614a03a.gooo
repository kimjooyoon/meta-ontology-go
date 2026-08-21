package shadow

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func decodeStrictCorpus(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("corpus must be a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	if err := scanJSON(decoder); err != nil {
		return fmt.Errorf("corpus strict JSON: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return err
	}
	decoder = json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var corpus Corpus
	if err := decoder.Decode(&corpus); err != nil {
		return fmt.Errorf("decode corpus: %w", err)
	}
	return requireEOF(decoder)
}
