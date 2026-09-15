package lanefrontier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '{' {
		return fmt.Errorf("input must be an object")
	}
	seen := map[string]struct{}{}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return err
		}
		name, ok := key.(string)
		if !ok {
			return fmt.Errorf("input key is not a string")
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate input field %q", name)
		}
		seen[name] = struct{}{}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return err
		}
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	return requireEOF(decoder)
}
func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}
func CanonicalJSON(c Case) []byte {
	value := struct {
		Name     string `json:"name"`
		Input    Input  `json:"input"`
		Expected struct {
			Decision Decision `json:"decision"`
			Reason   Reason   `json:"reason"`
		} `json:"expected"`
	}{Name: c.Name, Input: normalizedInput(c.Input), Expected: struct {
		Decision Decision `json:"decision"`
		Reason   Reason   `json:"reason"`
	}{c.Expected.Decision, c.Expected.Reason}}
	body, _ := json.Marshal(value)
	return body
}
func CanonicalDigest(c Case) string {
	return digestResult(c.Input, c.Expected.Decision, c.Expected.Reason)
}
