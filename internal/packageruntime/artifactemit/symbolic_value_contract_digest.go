package artifactemit

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func canonicalSymbolicValueContract(contract SymbolicValueContract) ([]byte, error) {
	contract.Digest = ""
	raw, err := json.Marshal(contract)
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

func validSymbolicValueSHA256(value string) bool {
	if !strings.HasPrefix(value, "sha256:") {
		return false
	}
	return validSymbolicValueHexDigest(strings.TrimPrefix(value, "sha256:"), 32)
}

func validSymbolicValueHexDigest(value string, bytesLength int) bool {
	if value != strings.ToLower(value) || len(value) != bytesLength*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == bytesLength
}
