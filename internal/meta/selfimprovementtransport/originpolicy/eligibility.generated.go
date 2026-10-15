package originpolicy

//gooo:generated:start id="gooo://self-improvement/transport/continuation-eligible" kind="activity"
func ContinuationEligible(decision string, outcome string, verified bool) bool {

	return decision == "CLOSED" && outcome == "AUTHORIZED" && verified

}

//gooo:generated:end id="gooo://self-improvement/transport/continuation-eligible" kind="activity"
