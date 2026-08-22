package languageconcept

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	Satisfied      bool   `json:"satisfied"`
	EvidenceDigest string `json:"evidence_digest"`
}

func finish(report Report, value observation, ready bool) Report {
	report.Summary = Summary{value.concepts, value.code, value.useCases, value.metrics,
		value.operating, value.conformed, value.unbound, value.novelty, value.writes}
	report.Indicators = []Indicator{
		metric("catalog-readiness-bps", "outcome", "coherence", "evaluate", "bind-language-concept-catalog", bps(ready), 10000, ready),
		metric("concept-code-binding-bps", "driver", "foundation", "codeBindings", "bind-concept-meta-code", coverage(value.code, value.concepts), 10000, value.code == value.concepts),
		metric("concept-use-case-binding-bps", "driver", "coherence", "validUseCases", "bind-concept-use-case", coverage(value.useCases, value.concepts), 10000, value.useCases == value.concepts),
		metric("concept-metric-binding-bps", "driver", "regression", "evaluate", "bind-concept-indicator", coverage(value.metrics, value.concepts), 10000, value.metrics == value.concepts),
		metric("concept-unbound-guardrail", "guardrail", "coherence", "evaluate", "reject-unbound-concept", value.unbound, 0, value.unbound == 0),
		metric("concept-novelty-overclaim-guardrail", "guardrail", "foundation", "evaluate", "reject-unverified-novelty", value.novelty, 0, value.novelty == 0),
		metric("concept-observer-writes-guardrail", "guardrail", "foundation", "Evaluate", "preserve-read-only-concept-catalog", value.writes, 0, value.writes == 0),
	}
	report.Proofs = []Proof{proof("foundation", value.code == value.concepts && value.novelty == 0),
		proof("coherence", value.useCases == value.concepts && value.unbound == 0), proof("regression", value.metrics == value.concepts)}
	report.ReportDigest = digest(report)
	return report
}

func metric(id, class, choice, producer, operation string, value, target int, satisfied bool) Indicator {
	return Indicator{"gooo.metric.meta." + id + ".v1", class, choice, "languageconcept." + producer,
		"self-improvement-cycle", operation, value, target, satisfied}
}

func proof(choice string, satisfied bool) Proof {
	operation := "justify-language-concept-by-" + choice
	return Proof{choice, operation, satisfied, digest(choice + operation)}
}

func digest(value any) string {
	encoded, _ := json.Marshal(value)
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func coverage(part, whole int) int { if whole == 0 { return 0 }; return part * 10000 / whole }
func bps(value bool) int { if value { return 10000 }; return 0 }
