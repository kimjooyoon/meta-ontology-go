package bindingcoverage

import (
	_ "embed"
	"encoding/json"
	"sort"
)

//go:embed corpus.json
var corpusJSON []byte

type CorpusFile struct {
	Schema          string       `json:"schema"`
	CanonicalDigest string       `json:"canonical_digest"`
	Cases           []CorpusCase `json:"cases"`
}

type CorpusCase struct {
	Name           string `json:"name"`
	Input          Input  `json:"input"`
	Expected       Vector `json:"expected"`
	ExpectedDigest string `json:"expected_digest"`
}

func LoadCorpus() (CorpusFile, error) {
	var corpus CorpusFile
	if err := decodeStrict(corpusJSON, &corpus); err != nil {
		return CorpusFile{}, err
	}
	return corpus, nil
}

func canonicalCorpusJSON(corpus CorpusFile) ([]byte, error) {
	cases := append([]CorpusCase{}, corpus.Cases...)
	sort.Slice(cases, func(i, j int) bool { return cases[i].Name < cases[j].Name })
	return json.Marshal(struct {
		Schema string       `json:"schema"`
		Cases  []CorpusCase `json:"cases"`
	}{Schema: corpus.Schema, Cases: cases})
}

func CorpusDigest(corpus CorpusFile) (string, error) {
	data, err := canonicalCorpusJSON(corpus)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}
