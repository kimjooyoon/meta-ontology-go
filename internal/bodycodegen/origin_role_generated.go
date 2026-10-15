package bodycodegen

//gooo:generated:start id="urn:gooo:origin-display:role" kind="activity"
func OriginRole(input string) string {

	if input == "input" {
		return "input port"
	}
	if input == "literal" {
		return "literal"
	}
	if input == "expression" {
		return "operation"
	}
	if input == "copy" {
		return "local definition"
	}
	if input == "write" {
		return "local update"
	}
	if input == "read" {
		return "value read"
	}
	if input == "return" {
		return "return value"
	}
	if input == "join" {
		return "possible value merge"
	}
	if input == "choice" {
		return "permitted alternatives"
	}
	if input == "guard" {
		return "branch condition"
	}
	if input == "guard_join" {
		return "condition merge"
	}
	if input == "call" {
		return "helper result"
	}
	if input == "bind" {
		return "helper input"
	}
	return "unclassified source role"

}

//gooo:generated:end id="urn:gooo:origin-display:role" kind="activity"
