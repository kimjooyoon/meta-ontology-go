package claimledger

func validStage(stage string) bool {
	switch stage {
	case "SOURCE", "PARSE", "BIND", "COMPILE", "EMIT", "TRANSPORT", "OBSERVE", "CONFORM", "PROMOTE":
		return true
	default:
		return false
	}
}

func validProofRoute(route string) bool {
	return route == "FOUNDATION" || route == "COHERENCE" || route == "REGRESSION"
}

func validOperator(operator string) bool {
	return operator == "EQUALS" || operator == "NON_NULL" || operator == "POSITIVE_INTEGER"
}
