package domaincompleteness

//gooo:generated:start id="domain-completeness://activity/classify-domain-completeness" kind="activity"
func ClassifyDomainCompleteness(input0 int64, input1 int64, input2 int64, input3 bool) string {
	if input3 || input0 < 0 || input1 < 0 || input2 < 0 || input0 > input1 {
		return "FAIL_CLOSED"
	} else if input1 == 0 || input2 > 0 {
		return "UNKNOWN"
	} else if input0 == input1 {
		return "PASS"
	} else {
		return "PROGRESS"
	}
}

//gooo:generated:end id="domain-completeness://activity/classify-domain-completeness" kind="activity"
