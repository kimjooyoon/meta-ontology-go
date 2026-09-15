package reproducibilitysemanticsconsumer

func summarize(cases []JudgmentCase, sourceBound, semanticCausal bool) Summary {
	byteDischarged, meaningDischarged, jointDischarged, counterexamples, openCases := 0, 0, 0, 0, 0
	for _, item := range cases {
		byteDischarged += statusInt(item.ByteStatus == StatusDischarged)
		meaningDischarged += statusInt(item.MeaningStatus == StatusDischarged)
		jointDischarged += statusInt(item.Status == StatusDischarged)
		counterexamples += statusInt(item.Status == StatusRefuted)
		openCases += statusInt(item.Status == StatusOpen)
	}
	total := len(cases)
	return Summary{CaseMatrix: coordinate(total, total), ByteClaim: coordinate(byteDischarged, total),
		MeaningClaim: coordinate(meaningDischarged, total), JointClaim: coordinate(jointDischarged, total),
		Counterexamples: coordinate(counterexamples, total), OpenCases: coordinate(openCases, total),
		SourceDigestBinding: coordinate(statusInt(sourceBound)*total, total),
		SemanticCausality:   coordinate(statusInt(semanticCausal)*total, total)}
}

func statusInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func judgeCompose(byteStatus, meaningStatus string) (string, string) {
	if byteStatus == StatusDischarged && meaningStatus == StatusDischarged {
		return StatusDischarged, "BOTH_CLAIMS_DISCHARGED"
	}
	if byteStatus == StatusDischarged && meaningStatus == StatusRefuted {
		return StatusRefuted, "REPRODUCIBLE_BUT_WRONG"
	}
	if byteStatus == StatusRefuted && meaningStatus == StatusDischarged {
		return StatusRefuted, "MEANINGFUL_BUT_UNREPRODUCED"
	}
	return StatusOpen, "CLAIMS_OPEN"
}

func subjectOutcome(cases []JudgmentCase) (string, string, string) {
	for _, item := range cases {
		if item.Status == StatusOpen {
			return StatusOpen, "LOWER_RESOLUTION", "OPEN_EVIDENCE_REMAINS"
		}
	}
	for _, item := range cases {
		if item.Status == StatusRefuted {
			return StatusRefuted, "EXACT", "COUNTEREXAMPLE_DISCHARGED"
		}
	}
	return StatusDischarged, "EXACT", "ALL_CASES_DISCHARGED"
}
