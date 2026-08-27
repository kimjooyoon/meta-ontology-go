package main

func hasGoooRelations(source string) bool {
	return contains(source, "activity ObservePartialObservation") && contains(source, "activity DescendToInvariantOnly") && contains(source, "activity AdjudicateReceipt")
}

func contains(source, part string) bool {
	for index := 0; index+len(part) <= len(source); index++ {
		if source[index:index+len(part)] == part {
			return true
		}
	}
	return false
}
