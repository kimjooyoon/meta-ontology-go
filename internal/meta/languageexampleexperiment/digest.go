package languageexampleexperiment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func finishReport(report Report) Report {
	facts := struct {
		Interpretation string      `json:"interpretation"`
		Summary        Summary     `json:"summary"`
		Indicators     []Indicator `json:"indicators"`
		Views          []View      `json:"views"`
		NotClaimed     []string    `json:"not_claimed"`
	}{report.Interpretation, report.Summary, report.Indicators, report.Views, report.NotClaimed}
	report.FactsDigest = hashJSON(facts)
	for index := range report.Proofs {
		report.Proofs[index].Evidence = report.FactsDigest
	}
	report.Digest = ""
	report.Digest = hashJSON(report)
	return report
}

func hashJSON(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
