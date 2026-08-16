package resourcevector

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
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

func R4F01() Input {
	corpus, err := LoadCorpus()
	if err != nil {
		panic(err)
	}
	for _, row := range corpus.Cases {
		if row.Name == "r4-f-01" {
			return row.Input
		}
	}
	panic("r4-f-01 fixture missing")
}

func R4F01Fixture() Input { return R4F01() }

func CorpusDigest(corpus Corpus) string {
	cases := append([]CorpusCase(nil), corpus.Cases...)
	sort.Slice(cases, func(left, right int) bool { return cases[left].Name < cases[right].Name })
	rows := make([]struct {
		Name  string `json:"name"`
		Input Input  `json:"input"`
	}, len(cases))
	for index, row := range cases {
		rows[index] = struct {
			Name  string `json:"name"`
			Input Input  `json:"input"`
		}{Name: row.Name, Input: row.Input}
	}
	data, _ := json.Marshal(struct {
		Schema string `json:"schema"`
		Cases  any    `json:"cases"`
	}{Schema: corpus.Schema, Cases: rows})
	return digestBytes(data)
}

func decodeStrictJSON(data []byte, target any) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}
