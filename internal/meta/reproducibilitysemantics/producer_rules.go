package reproducibilitysemantics

func digestText(value string) string {
	if value == "" {
		return ""
	}
	return digestBytes([]byte(value))
}

func compare(reference, candidate string) (string, string) {
	if reference == candidate {
		return StatusDischarged, "EVIDENCE_MATCH"
	}
	return StatusRefuted, "EVIDENCE_MISMATCH"
}

func compose(byteStatus, meaningStatus string) (string, string) {
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
