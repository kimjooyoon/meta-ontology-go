package resourcevector

import (
	"encoding/json"
)

func EncodeOutputJSON(output Output) ([]byte, error) {
	data, err := json.MarshalIndent(canonicalOutput(output), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
func CanonicalOutputDigest(output Output) string {
	data, err := json.Marshal(canonicalOutput(output))
	if err != nil {
		return digestBytes([]byte("canonical-output-error:" + err.Error()))
	}
	return digestBytes(data)
}
func ReplayDigest(inputDigest, outputDigest string) string {
	return digestBytes([]byte(inputDigest + "\x00" + outputDigest))
}

type canonicalOutputView struct {
	Schema            string   `json:"schema"`
	FixtureID         string   `json:"fixture_id"`
	InputDigest       string   `json:"input_digest"`
	Selected          *Vector  `json:"selected,omitempty"`
	Full              *Vector  `json:"full,omitempty"`
	Decision          Decision `json:"decision"`
	Reason            Reason   `json:"reason"`
	LimitFailures     []string `json:"limit_failures"`
	FullSuiteRequired bool     `json:"full_suite_required"`
	ProofValid        bool     `json:"proof_valid"`
}

func canonicalOutput(output Output) canonicalOutputView {
	return canonicalOutputView{
		Schema: output.Schema, FixtureID: output.FixtureID, InputDigest: output.InputDigest,
		Selected: output.Selected, Full: output.Full, Decision: output.Decision, Reason: output.Reason,
		LimitFailures: sortedStrings(output.LimitFailures), FullSuiteRequired: output.FullSuiteRequired,
		ProofValid: output.ProofValid,
	}
}
