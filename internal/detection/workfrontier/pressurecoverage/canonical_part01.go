package pressurecoverage

import (
	"bytes"
	"encoding/json"
)

// UnmarshalJSON is the strict public boundary for Input.
func (input *Input) UnmarshalJSON(data []byte) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	type wire Input
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var decoded wire
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	if err := validateInput(Input(decoded)); err != nil {
		return err
	}
	*input = Input(decoded)
	return nil
}
func DecodeInput(data []byte) (Input, error) {
	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		return Input{}, err
	}
	return input, nil
}
func CanonicalInputBytes(input Input) ([]byte, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}
	return json.Marshal(normalizeInput(input))
}
func CanonicalInputDigest(input Input) string {
	data, err := CanonicalInputBytes(input)
	if err != nil {
		return digestBytes([]byte("canonical-input-error:" + err.Error()))
	}
	return digestBytes(data)
}
func authorityBindingDigest(input Input, role string) string {
	input.AuthoritySnapshotDigest = ""
	input.PolicyDigest = ""
	input.RegistryDigest = ""
	input.ToolchainOptionsDigest = ""
	return digestBytes([]byte(role + "\x00" + CanonicalInputDigest(input)))
}
