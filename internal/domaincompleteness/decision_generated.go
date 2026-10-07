package domaincompleteness

//gooo:generated:start id="domain-completeness://activity/select-domain-completeness-outcome" kind="activity"
func SelectDomainCompletenessOutcome(input0 bool, input1 bool, input2 bool, input3 bool) string {
	if input0 || input1 {
		return "FAIL_CLOSED"
	} else if input2 {
		return "UNKNOWN"
	} else if input3 {
		return "PROGRESS"
	} else {
		return "PASS"
	}
}

//gooo:generated:end id="domain-completeness://activity/select-domain-completeness-outcome" kind="activity"
