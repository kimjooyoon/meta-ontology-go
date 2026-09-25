package sourceauthorityeval

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthority"

func newReport(bundle Bundle) Report {
	return Report{
		Schema:                 ReportSchema,
		MetricID:               sourceauthority.MetricID,
		MetaOperation:          sourceauthority.MetaOperation,
		ProofChoice:            sourceauthority.ProofChoice,
		SubjectSHA:             bundle.SubjectSHA,
		ContractDigest:         sourceauthority.Digest(),
		EvidenceContractDigest: bundle.ContractDigest,
		Receipts:               []FactReceipt{},
	}
}

func finalize(report Report) Report {
	report.Summary.AcceptedFacts = len(report.Receipts)
	for _, receipt := range report.Receipts {
		switch receipt.Observation {
		case "SATISFIED":
			report.Summary.BackedFacts++
		case "NOT_SATISFIED":
			report.Summary.FailedFacts++
		case "UNKNOWN":
			report.Summary.UnknownFacts++
		default:
			report.Summary.ErrorFacts++
		}
	}
	report.Summary.CoverageBPS =
		report.Summary.BackedFacts * 10000 / report.Summary.AcceptedFacts
	switch {
	case report.Summary.ErrorFacts > 0:
		setOutcome(&report, "ERROR", "EXACT", "BLOCK", "FACT_EVALUATION_ERROR")
	case report.Summary.UnknownFacts > 0:
		setOutcome(&report, "UNKNOWN", "INVARIANT_ONLY", "BLOCK",
			"SOURCE_AUTHORITY_EVIDENCE_UNKNOWN")
	case report.Summary.FailedFacts > 0:
		setOutcome(&report, "NOT_SATISFIED", "EXACT", "BLOCK",
			"SOURCE_AUTHORITY_NOT_SATISFIED")
	default:
		setOutcome(&report, "SATISFIED", "EXACT", "ALLOW",
			"SOURCE_AUTHORITY_EXACT")
	}
	return seal(report)
}

func setOutcome(report *Report, observation, resolution, enforcement, reason string) {
	report.Observation = observation
	report.Resolution = resolution
	report.Enforcement = enforcement
	report.Reason = reason
}
