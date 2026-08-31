package artifactemit

import (
	"bytes"
	"encoding/json"
)

func canonicalSymbolicValueReachability(reachability SymbolicValueReachability) ([]byte, error) {
	reachability.Digest = ""
	raw, err := json.Marshal(reachability)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var normalized any
	if err := decoder.Decode(&normalized); err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}
