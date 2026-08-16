package shadow

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// corpusFile is embedded so the Phase A fixtures are immutable at runtime.
//
//go:embed corpus.json
var corpusFile embed.FS

// LoadCorpus strictly decodes the checked-in corpus. It does not evaluate or
// rewrite any fixture values.
func LoadCorpus() (Corpus, error) {
	data, err := corpusFile.ReadFile("corpus.json")
	if err != nil {
		return Corpus{}, err
	}
	if err := decodeStrictCorpus(data); err != nil {
		return Corpus{}, err
	}
	var corpus Corpus
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&corpus); err != nil {
		return Corpus{}, err
	}
	if corpus.Schema != CorpusSchema || corpus.Cases == nil {
		return Corpus{}, fmt.Errorf("invalid shadow corpus schema")
	}
	return corpus, nil
}

// CorpusDigest returns the SHA-256 of the canonical decoded corpus.
func CorpusDigest() string {
	data, _ := corpusFile.ReadFile("corpus.json")
	var corpus Corpus
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&corpus); err != nil {
		return ""
	}
	canonical, err := json.Marshal(corpus)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:])
}

// ExpectedVectorDigest binds the expected decision vector independently of
// the raw file bytes. The authored case order is part of the immutable vector.
func ExpectedVectorDigest(corpus Corpus) string {
	cases := append([]Case(nil), corpus.Cases...)
	for i := range cases {
		cases[i].Files = Files{}
		cases[i].Expected.CanonicalDigest = ""
	}
	// The corpus is authored in stable name order; retaining that order makes
	// accidental fixture reordering visible in the vector digest.
	value := struct {
		Schema string `json:"schema"`
		Cases  []Case `json:"cases"`
	}{corpus.Schema, cases}
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func decodeStrictCorpus(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("corpus must be a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	if err := scanJSON(decoder); err != nil {
		return fmt.Errorf("corpus strict JSON: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return err
	}
	decoder = json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var corpus Corpus
	if err := decoder.Decode(&corpus); err != nil {
		return fmt.Errorf("decode corpus: %w", err)
	}
	return requireEOF(decoder)
}
