package feedbackstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	Target        int    `json:"target"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	Satisfied      bool   `json:"satisfied"`
	EvidenceDigest string `json:"evidence_digest"`
}

type observation struct {
	identity, payload, replay, semantic bool
	descents, falseFixed, writes        int
}

func finish(report Report, observed observation) Report {
	ready := report.Decision == "READY"
	report.Summary = Summary{boolBPS(ready), observed.descents, observed.falseFixed, observed.writes}
	report.Indicators = indicators(observed, ready)
	report.Proofs = []Proof{
		proof("foundation", "bind-predecessor-semantic-evidence", observed.identity && observed.payload && observed.writes == 0),
		proof("coherence", "validate-predecessor-semantic-transition", observed.semantic && observed.falseFixed == 0),
		proof("regression", "replay-predecessor-semantic-receipt", observed.replay),
	}
	report.ReportDigest = digest(report)
	return report
}

func proof(choice, operation string, satisfied bool) Proof {
	state := "0"
	if satisfied {
		state = "10000"
	}
	value := choice + "|" + operation + "|" + state
	return Proof{choice, operation, satisfied, digest(value)}
}

func digest(value any) string {
	encoded, _ := json.Marshal(value)
	return bytesDigest(encoded)
}

func bytesDigest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func boolBPS(value bool) int {
	if value {
		return 10000
	}
	return 0
}
