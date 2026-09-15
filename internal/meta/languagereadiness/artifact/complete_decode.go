package artifact

import "encoding/json"

func decodeCompleteEvidence[T any](raw []byte) (T, error) {
	value := new(T)
	if err := json.Unmarshal(raw, value); err != nil {
		return *value, err
	}
	return *value, nil
}
