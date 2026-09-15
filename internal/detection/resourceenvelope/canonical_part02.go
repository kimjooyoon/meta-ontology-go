package resourceenvelope

import (
	"encoding/json"
	"fmt"
)

// EncodeResultJSON returns one stable JSON result followed by LF.
func EncodeResultJSON(result Result) ([]byte, error) {
	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("encode resource result: %w", err)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("encode resource result: %w", err)
	}
	return append(payload, '\n'), nil
}
