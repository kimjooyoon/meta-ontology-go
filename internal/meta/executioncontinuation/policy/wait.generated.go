package policy

//gooo:generated:start id="gooo://self-improvement/consumer/permit-wait" kind="activity"
func PermitWait(wait string, childNeedsCompletion bool) bool {

	if childNeedsCompletion && wait == "COMPLETED" {
		return false
	}
	return wait == "CREATED" || wait == "NONE"

}

//gooo:generated:end id="gooo://self-improvement/consumer/permit-wait" kind="activity"
