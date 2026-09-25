package pressureindependence

import (
	_ "embed"
	"encoding/json"
	"sort"
)

//go:embed corpus.json
var corpusJSON []byte

type Corpus struct {
	Schema          string       `json:"schema"`
	CanonicalDigest string       `json:"canonical_digest"`
	Cases           []CorpusCase `json:"cases"`
}

type CorpusCase struct {
	Name     string `json:"name"`
	Input    Input  `json:"input"`
	Expected Output `json:"expected"`
}

func LoadCorpus() (Corpus, error) {
	var corpus Corpus
	if err := decodeStrictJSON(corpusJSON, &corpus); err != nil {
		return Corpus{}, err
	}
	return corpus, nil
}

func CorpusDigest(corpus Corpus) string {
	cases := append([]CorpusCase(nil), corpus.Cases...)
	sort.Slice(cases, func(i, j int) bool { return cases[i].Name < cases[j].Name })
	inputs := make([]struct {
		Name  string `json:"name"`
		Input Input  `json:"input"`
	}, len(cases))
	for index, row := range cases {
		inputs[index] = struct {
			Name  string `json:"name"`
			Input Input  `json:"input"`
		}{Name: row.Name, Input: normalizeInput(row.Input)}
	}
	data, _ := json.Marshal(struct {
		Schema string `json:"schema"`
		Cases  any    `json:"cases"`
	}{Schema: corpus.Schema, Cases: inputs})
	return digestBytes(data)
}
