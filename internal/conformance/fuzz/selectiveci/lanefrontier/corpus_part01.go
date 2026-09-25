package lanefrontier

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
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
