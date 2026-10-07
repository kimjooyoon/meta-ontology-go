package domaincompleteness

//gooo:generated:start id="domain-completeness://activity/can-compare-domain-completeness-axes" kind="activity"
func CanCompareDomainCompletenessAxes(input0 int64, input1 int64, input2 int64, input3 int64, input4 string, input5 string) bool {
	return input0 == 0 && input1 == 0 && input2 == 0 && input3 == 0 && (input4 == "PASS" || input4 == "PROGRESS") && (input5 == "PASS" || input5 == "PROGRESS")
}

//gooo:generated:end id="domain-completeness://activity/can-compare-domain-completeness-axes" kind="activity"
