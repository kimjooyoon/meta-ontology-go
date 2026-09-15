package rollbackintegrityshadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func buildIndicators(summary Summary, baseline bool) []Indicator {
	return []Indicator{
		indicator("shadow-case-coverage", "OUTCOME", "COHERENCE", summary.CasesPassed,
			caseTotal, summary.CasesPassed == caseTotal),
		indicator("meta-report-validity", "DRIVER", "FOUNDATION", summary.MetaReportsValid,
			caseTotal, summary.MetaReportsValid == caseTotal),
		indicator("assurance-baseline", "DRIVER", "FOUNDATION", boolInt(baseline), 1, baseline),
		indicator("evaluator-repository-writes", "GUARDRAIL", "REGRESSION", 0, 0, true),
		indicator("shadow-promotion-applied", "GUARDRAIL", "FOUNDATION", 0, 0, true),
		indicator("unknown-resolution-preserved", "GUARDRAIL", "COHERENCE",
			boolInt(summary.UnknownDecisionCases == 1), 1, summary.UnknownDecisionCases == 1),
	}
}

func indicator(id, class, proof string, value, target int, satisfied bool) Indicator {
	return Indicator{MetricID: "gooo.metric.operation.rollback-integrity." + id + ".v1",
		Class: class, ProofChoice: proof, Producer: "rollbackintegrityshadow.Evaluate",
		Consumer: "language-assurance-promotion-gate", MetaOperation: MetaOperation,
		Value: value, Target: target, Satisfied: satisfied}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func digestBytes(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	payload, _ := json.Marshal(value)
	return digestBytes(payload)
}

func validDigest(value string) bool {
	raw := strings.TrimPrefix(value, "sha256:")
	_, err := hex.DecodeString(raw)
	return len(raw) == 64 && raw == strings.ToLower(raw) && err == nil
}

func seal(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}

func Encode(report Report) []byte {
	payload, _ := json.MarshalIndent(report, "", "  ")
	return append(payload, '\n')
}
