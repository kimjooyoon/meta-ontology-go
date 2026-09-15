package toolchainconformance

func inspectCases(definition SurfaceDefinition, envelope artifactEnvelope,
	summary *Summary) int {
	total, totalOK := summaryInteger(envelope.Summary, "total")
	satisfied, satisfiedOK := summaryInteger(envelope.Summary, "satisfied")
	executed, executedOK := summaryInteger(envelope.Summary, "executed")
	unresolved, unresolvedOK := summaryInteger(envelope.Summary, "unresolved")
	if !totalOK || !satisfiedOK || !executedOK || !unresolvedOK ||
		total != definition.Cases || satisfied != total || executed != total {
		summary.CaseMismatches++
	}
	if unresolvedOK {
		summary.Unresolved += unresolved
	} else {
		summary.Unresolved++
	}
	if satisfiedOK {
		summary.CasesSatisfied += satisfied
	}
	if executedOK {
		summary.ExecutedCases += executed
	}
	return total
}

func inspectIndicators(definition SurfaceDefinition, envelope artifactEnvelope,
	summary *Summary) int {
	satisfied := 0
	for _, indicator := range envelope.Indicators {
		if indicator.Satisfied {
			satisfied++
		}
	}
	summary.IndicatorsSatisfied += satisfied
	if len(envelope.Indicators) != definition.Indicators ||
		satisfied != definition.Indicators {
		summary.IndicatorFailures++
	}
	return len(envelope.Indicators)
}

func inspectProofs(definition SurfaceDefinition, envelope artifactEnvelope,
	summary *Summary) int {
	passed := 0
	for _, proof := range envelope.Proofs {
		if proof.Passed {
			passed++
		}
	}
	summary.ProofsPassed += passed
	if len(envelope.Proofs) != definition.Proofs || passed != definition.Proofs {
		summary.ProofFailures++
	}
	return len(envelope.Proofs)
}
