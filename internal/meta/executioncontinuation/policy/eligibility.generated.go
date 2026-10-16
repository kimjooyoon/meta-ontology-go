package policy

//gooo:generated:start id="gooo://self-improvement/consumer/permit-continuation" kind="activity"
func PermitContinuation(decision string, resolution string, remaining int64, consumed int64, executions int64, verified bool) bool {

	return decision == "CLOSED" && resolution == "GRANTED_UNCONSUMED" && remaining == 1 && consumed == 0 && executions == 0 && verified

}

//gooo:generated:end id="gooo://self-improvement/consumer/permit-continuation" kind="activity"
