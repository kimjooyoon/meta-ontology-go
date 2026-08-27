package reproducibilitysemantics

func summarizeJudgment(cases []JudgmentCase, sourceBound, semanticCausal bool) Summary {
	byteDischarged, meaningDischarged, jointDischarged, counterexamples, openCases := 0, 0, 0, 0, 0
	for _, item := range cases {
		if item.ByteStatus == StatusDischarged {
			byteDischarged++
		}
		if item.MeaningStatus == StatusDischarged {
			meaningDischarged++
		}
		if item.Status == StatusDischarged {
			jointDischarged++
		}
		if item.Status == StatusRefuted {
			counterexamples++
		}
		if item.Status == StatusOpen {
			openCases++
		}
	}
	total := len(cases)
	return Summary{CaseMatrix: coordinate(total, total), ByteClaim: coordinate(byteDischarged, total),
		MeaningClaim: coordinate(meaningDischarged, total), JointClaim: coordinate(jointDischarged, total),
		Counterexamples: coordinate(counterexamples, total), OpenCases: coordinate(openCases, total),
		SourceDigestBinding: coordinate(statusInt(sourceBound)*total, total),
		SemanticCausality:   coordinate(statusInt(semanticCausal)*total, total)}
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
