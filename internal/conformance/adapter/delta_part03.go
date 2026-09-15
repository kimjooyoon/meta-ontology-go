package adapter

func hasDuplicateStrings(values []string) bool {
	for i := 1; i < len(values); i++ {
		if values[i] == values[i-1] {
			return true
		}
	}
	return false
}
func emptyStringsIfNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
func emptyRegionsIfNil(values []Region) []Region {
	if values == nil {
		return []Region{}
	}
	return values
}
func emptySlotsIfNil(values []Slot) []Slot {
	if values == nil {
		return []Slot{}
	}
	return values
}
func emptyImportsIfNil(values []Import) []Import {
	if values == nil {
		return []Import{}
	}
	return values
}
func emptyMappingsIfNil(values []Mapping) []Mapping {
	if values == nil {
		return []Mapping{}
	}
	return values
}
func emptyFactsIfNil(values []Fact) []Fact {
	if values == nil {
		return []Fact{}
	}
	return values
}
