package shadow

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
)

const CorrectionSchema = "gooo/selective-ci-shadow-correction/v1"

//go:embed correction_predecessor.json
var correctionFile embed.FS

//go:embed correction_predecessor_2.json
var secondCorrectionFile embed.FS

type CorrectionRecord struct {
	Schema     string            `json:"schema"`
	ReasonCode string            `json:"reason_code"`
	Superseded CorrectionDigests `json:"superseded"`
	Corrected  CorrectionDigests `json:"corrected"`
}

type CorrectionDigests struct {
	CorpusDigest         string `json:"corpus_digest"`
	ExpectedVectorDigest string `json:"expected_vector_digest"`
}

func LoadCorrection() (CorrectionRecord, error) {
	return loadCorrection(correctionFile, "correction_predecessor.json")
}

func LoadSecondCorrection() (CorrectionRecord, error) {
	return loadCorrection(secondCorrectionFile, "correction_predecessor_2.json")
}

func loadCorrection(files embed.FS, name string) (CorrectionRecord, error) {
	data, err := files.ReadFile(name)
	if err != nil {
		return CorrectionRecord{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var record CorrectionRecord
	if err := decoder.Decode(&record); err != nil {
		return CorrectionRecord{}, fmt.Errorf("decode correction record: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return CorrectionRecord{}, err
	}
	if record.Schema != CorrectionSchema || record.ReasonCode == "" || record.Superseded.CorpusDigest == "" || record.Superseded.ExpectedVectorDigest == "" || record.Corrected.CorpusDigest == "" || record.Corrected.ExpectedVectorDigest == "" {
		return CorrectionRecord{}, fmt.Errorf("invalid correction record")
	}
	return record, nil
}
