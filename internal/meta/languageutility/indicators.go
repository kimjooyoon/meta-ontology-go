package languageutility

func buildIndicators(summary Summary) []Indicator {
	return []Indicator{
		indicator("gooo.metric.language-utility.cells-closed.v1", "OUTCOME", "foundation",
			summary.ClosedCells, summary.CellsTotal, "cells", equality(summary.ClosedCells, summary.CellsTotal)),
		indicator("gooo.metric.language-utility.use-cases-complete.v1", "OUTCOME", "coherence",
			summary.CompleteUseCases, summary.UseCasesTotal, "use_cases", equality(summary.CompleteUseCases, summary.UseCasesTotal)),
		indicator("gooo.metric.language-utility.remaining-cells.v1", "DRIVER", "foundation",
			summary.RemainingCells, 0, "cells", zero(summary.RemainingCells)),
		indicator("gooo.metric.language-utility.accepted-floor.v1", "GUARDRAIL", "regression",
			summary.ClosedCells, summary.ClosedFloor, "cells", floor(summary.ClosedCells, summary.ClosedFloor)),
		indicator("gooo.metric.language-utility.unknown-cells.v1", "GUARDRAIL", "foundation",
			summary.UnknownCells, 0, "cells", zero(summary.UnknownCells)),
		indicator("gooo.metric.language-utility.refuted-cells.v1", "GUARDRAIL", "coherence",
			summary.RefutedCells, 0, "cells", zero(summary.RefutedCells)),
		indicator("gooo.metric.language-utility.evidence-artifacts.v1", "DRIVER", "coherence",
			summary.EvidenceArtifacts, summary.UseCasesTotal, "artifacts", equality(summary.EvidenceArtifacts, summary.UseCasesTotal)),
		indicator("gooo.metric.language-utility.repository-writes.v1", "GUARDRAIL", "regression",
			summary.RepositoryWrites, 0, "writes", zero(summary.RepositoryWrites)),
	}
}

func indicator(id, class, proof string, observed, target int, unit, status string) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, Observed: observed,
		Target: target, Unit: unit, Status: status}
}

func equality(observed, target int) string {
	if observed == target {
		return "SATISFIED"
	}
	return "GAP"
}

func zero(observed int) string { return equality(observed, 0) }

func floor(observed, target int) string {
	if observed >= target {
		return "SATISFIED"
	}
	return "REFUTED"
}
