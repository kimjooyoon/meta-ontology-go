package lanefrontier

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

//go:embed corpus.json
var corpusFile embed.FS

func LoadCorpus() (Corpus, error) {
	data, err := corpusFile.ReadFile("corpus.json")
	if err != nil {
		return Corpus{}, err
	}
	var corpus Corpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		return Corpus{}, err
	}
	return corpus, nil
}

func DecodeInput(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, err
	}
	var input Input
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return Input{}, err
	}
	if err := requireEOF(decoder); err != nil {
		return Input{}, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return Input{}, fmt.Errorf("input must be an object")
	}
	return input, nil
}

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

func CorpusDigest() string {
	data, _ := corpusFile.ReadFile("corpus.json")
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
