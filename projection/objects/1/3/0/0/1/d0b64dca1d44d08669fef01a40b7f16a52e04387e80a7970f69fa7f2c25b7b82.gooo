package selectiveci

import (
	"encoding/json"
	"fmt"
	"io"
)

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("strict JSON: trailing value")
		}
		return fmt.Errorf("strict JSON: trailing data: %w", err)
	}
	return nil
}
